package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/cli"
	"github.com/zclconf/go-cty/cty"
)

const S3Mode = "s3"

var (
	mode       = flag.String("mode", "", "service to migrate e.g. s3; required")
	inputFile  = flag.String("input", "", "input file; required")
	outputFile = flag.String("output", "", "output file; required")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] -mode <mode> -input <input-file> -output <output-file>\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *mode == "" || *inputFile == "" || *outputFile == "" {
		flag.Usage()
		os.Exit(2)
	}

	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	if *mode != S3Mode {
		ui.Error(fmt.Sprintf("Mode (%s) not implemented", *mode))
		os.Exit(0)
	}

	fBytes, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		ui.Error(fmt.Sprintf("error loading configuration: %s", err))
		os.Exit(1)
	}

	f, diags := hclwrite.ParseConfig(fBytes, *inputFile, hcl.Pos{Line: 1, Column: 1})

	if diags != nil {
		for _, diag := range diags {
			if diag.Error() != "" {
				ui.Error(fmt.Sprintf("error loading configuration: %s", diag.Error()))
			}
		}
		os.Exit(1)
	}

	var newResources []string

	for _, block := range f.Body().Blocks() {
		if block == nil {
			continue
		}

		labels := block.Labels()
		if len(labels) != 2 || labels[0] != "aws_s3_bucket" {
			continue
		}

		bucketPath := strings.Join(labels, ".")
		log.Printf("[INFO] Migrating %s \n", bucketPath)

		if v, ok := block.Body().Attributes()["acceleration_status"]; ok {
			// Create Acceleration Status resource
			block.Body().RemoveAttribute("acceleration_status")

			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_accelerate_configuration", fmt.Sprintf("%s_acceleration_configuration", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			newBlock.Body().SetAttributeRaw("status", v.Expr().BuildTokens(nil))

			log.Printf("	  ✓ Created aws_s3_bucket_accelerate_configuration.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_accelerate_configuration.%s,%s", newlabels[1], bucketPath))
		}

		var aclResourceBlock *hclwrite.Block
		if v, ok := block.Body().Attributes()["acl"]; ok {
			// Create ACL resource
			block.Body().RemoveAttribute("acl")

			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_acl", fmt.Sprintf("%s_acl", labels[1])}
			aclResourceBlock = f.Body().AppendNewBlock(block.Type(), newlabels)

			aclResourceBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			aclResourceBlock.Body().SetAttributeRaw("acl", v.Expr().BuildTokens(nil))

			log.Printf("	  ✓ Created aws_s3_bucket_acl.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_acl.%s,%s", newlabels[1], bucketPath))
		}

		if v, ok := block.Body().Attributes()["policy"]; ok {
			// Create ACL resource
			block.Body().RemoveAttribute("policy")

			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_policy", fmt.Sprintf("%s_policy", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			newBlock.Body().SetAttributeRaw("policy", v.Expr().BuildTokens(nil))

			log.Printf("	  ✓ Created aws_s3_bucket_policy.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_policy.%s,%s", newlabels[1], bucketPath))
		}

		if v, ok := block.Body().Attributes()["request_payer"]; ok {
			// Create ACL resource
			block.Body().RemoveAttribute("request_payer")

			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_request_payment_configuration", fmt.Sprintf("%s_request_payment_configuration", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			newBlock.Body().SetAttributeRaw("payer", v.Expr().BuildTokens(nil))

			log.Printf("	  ✓ Created aws_s3_bucket_request_payment_configuration.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_request_payment_configuration.%s,%s", newlabels[1], bucketPath))
		}

		// Block Handling
		// 1. Cors Rules
		// 2. Grants
		var corsRules []*hclwrite.Block
		var grants []*hclwrite.Block
		//var lifecycleRules []*hclwrite.Block
		var logging *hclwrite.Block
		var objectLockConfig *hclwrite.Block
		var replicationConfig *hclwrite.Block
		var serverSideEncryptionConfig *hclwrite.Block
		var website *hclwrite.Block
		var versioning *hclwrite.Block

		for _, subBlock := range block.Body().Blocks() {
			block.Body().RemoveBlock(subBlock)

			switch t := subBlock.Type(); t {
			case "cors_rule":
				corsRules = append(corsRules, subBlock)
			case "grant":
				grants = append(grants, subBlock)
			//case "lifecycle_rule":
			//	lifecycleRules = append(lifecycleRules, subBlock)
			case "logging":
				logging = subBlock
			case "object_lock_configuration":
				objectLockConfig = subBlock
			case "replication_configuration":
				replicationConfig = subBlock
			case "server_side_encryption_configuration":
				serverSideEncryptionConfig = subBlock
			case "versioning":
				versioning = subBlock
			case "website":
				website = subBlock
			}
		}

		if len(corsRules) > 0 {
			// Create new Cors resource
			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_cors_configuration", fmt.Sprintf("%s_cors_configuration", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for _, b := range corsRules {
				newBlock.Body().AppendBlock(b)
			}

			log.Printf("	  ✓ Created aws_s3_bucket_cors_configuration.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_cors_configuration.%s,%s", newlabels[1], bucketPath))
		}

		if len(grants) > 0 {
			if aclResourceBlock == nil {
				// Create new aws_s3_bucket_acl resource
				f.Body().AppendNewline()

				newlabels := []string{"aws_s3_bucket_acl", fmt.Sprintf("%s_acl", labels[1])}
				newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

				newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
					hcl.TraverseRoot{
						Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
					},
				})

				acpBlock := newBlock.Body().AppendNewBlock("access_control_policy", nil)

				for _, grant := range grants {
					grantBlock := acpBlock.Body().AppendNewBlock("grant", nil)
					grantee := grantBlock.Body().AppendNewBlock("grantee", nil)

					var permissions []string

					for k, v := range grant.Body().Attributes() {
						if k == "permissions" {
							for _, t := range v.BuildTokens(nil) {
								if p := string(t.Bytes); len(p) > 1 && p != k {
									permissions = append(permissions, fmt.Sprintf("%q", p))
								}
							}
						} else {
							grantee.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
						}
					}

					if len(permissions) == 0 {
						continue
					}

					// Need to do a traversal since we don't want to escape the "" around the string
					grantBlock.Body().SetAttributeTraversal("permission", hcl.Traversal{
						hcl.TraverseRoot{
							Name: permissions[0],
						},
					})

					if len(permissions) > 1 {
						// Create a new grant block for this permission
						for _, permission := range permissions[1:] {
							grantBlock := acpBlock.Body().AppendNewBlock("grant", nil)
							grantee := grantBlock.Body().AppendNewBlock("grantee", nil)

							for k, v := range grant.Body().Attributes() {
								if k == "permissions" {
									continue
								}
								grantee.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
							}

							// Need to do a traversal since we don't want to escape the "" around the string
							grantBlock.Body().SetAttributeTraversal("permission", hcl.Traversal{
								hcl.TraverseRoot{
									Name: permission,
								},
							})
						}
					}
				}

				log.Printf("	  ✓ Created aws_s3_bucket_acl.%s", newlabels[1])
				newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_acl.%s,%s", newlabels[1], bucketPath))
			}
		}

		if logging != nil {
			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_logging", fmt.Sprintf("%s_logging", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for k, v := range logging.Body().Attributes() {
				newBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
			}

			log.Printf("	  ✓ Created aws_s3_bucket_logging.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_logging.%s,%s", newlabels[1], bucketPath))
		}

		if versioning != nil {
			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_versioning", fmt.Sprintf("%s_versioning", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			versioningConfigBlock := newBlock.Body().AppendNewBlock("versioning_configuration", nil)

			for k, v := range versioning.Body().Attributes() {
				if k == "enabled" {
					value := strings.TrimSpace(string(v.Expr().BuildTokens(nil).Bytes()))
					if value == "true" {
						expr := hclwrite.NewExpressionLiteral(cty.StringVal("Enabled"))
						versioningConfigBlock.Body().SetAttributeRaw("status", expr.BuildTokens(nil))
					} else if value == "false" {
						// This might not be accurate as "false" can indicate never enable versioning
						expr := hclwrite.NewExpressionLiteral(cty.StringVal("Suspended"))
						versioningConfigBlock.Body().SetAttributeRaw("status", expr.BuildTokens(nil))
					}
				}
			}

			log.Printf("	  ✓ Created aws_s3_bucket_versioning.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_versioning.%s,%s", newlabels[1], bucketPath))
		}

		if objectLockConfig != nil {
			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_object_lock_configuration", fmt.Sprintf("%s_object_lock_configuration", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for k, v := range objectLockConfig.Body().Attributes() {
				switch k {
				case "object_lock_enabled":
					newBlock.Body().SetAttributeRaw("object_lock_enabled", v.Expr().BuildTokens(nil))
				}
			}

			for _, ob := range objectLockConfig.Body().Blocks() {
				// we only expect 1 rule as defined in the aws_s3_bucket schema
				if ob.Type() != "rule" {
					continue
				}
				newBlock.Body().AppendBlock(ob)
			}

			log.Printf("	  ✓ Created aws_s3_bucket_object_lock_configuration.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_object_lock_configuration.%s,%s", newlabels[1], bucketPath))
		}

		if replicationConfig != nil {
			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_replication_configuration", fmt.Sprintf("%s_replication_configuration", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for k, v := range replicationConfig.Body().Attributes() {
				if k == "role" {
					newBlock.Body().SetAttributeRaw("role", v.Expr().BuildTokens(nil))
				}
			}

			for _, b := range replicationConfig.Body().Blocks() {
				ruleBlock := newBlock.Body().AppendNewBlock("rule", nil)

				if b.Type() != "rules" {

					// not expected to hit this as the replication_configuration block only has the rules block
					continue
				}

				for k, v := range b.Body().Attributes() {
					// Expected: id, prefix, status, priority, delete_marker_replication_status
					switch k {
					case "id", "prefix", "status", "priority":
						ruleBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
					case "delete_marker_replication_status":
					}
				}
			}

			log.Printf("	  ✓ Created aws_s3_bucket_replication_configuration.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_replication_configuration.%s,%s", newlabels[1], bucketPath))
		}

		if serverSideEncryptionConfig != nil {
			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_server_side_encryption_configuration", fmt.Sprintf("%s_server_side_encryption_configuration", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for _, b := range serverSideEncryptionConfig.Body().Blocks() {
				// we only expect 1 rule as defined in the aws_s3_bucket schema
				if b.Type() != "rule" {
					continue
				}
				newBlock.Body().AppendBlock(b)
			}

			log.Printf("	  ✓ Created aws_s3_bucket_server_side_encryption_configuration.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_server_side_encryption_configuration.%s,%s", newlabels[1], bucketPath))
		}

		if website != nil {
			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_website_configuration", fmt.Sprintf("%s_website_configuration", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for k, v := range website.Body().Attributes() {
				switch k {
				case "index_document":
					indexDocBlock := newBlock.Body().AppendNewBlock("index_document", nil)
					indexDocBlock.Body().SetAttributeRaw("suffix", v.Expr().BuildTokens(nil))
				case "error_document":
					errDocBlock := newBlock.Body().AppendNewBlock("error_document", nil)
					errDocBlock.Body().SetAttributeRaw("key", v.Expr().BuildTokens(nil))
				case "redirect_all_requests_to":
					redirectBlock := newBlock.Body().AppendNewBlock("redirect_all_requests_to", nil)
					redirectBlock.Body().SetAttributeRaw("host_name", v.Expr().BuildTokens(nil))
				case "routing_rules":
					var unmarshaledRules []*s3.RoutingRule

					routingRulesStr := strings.TrimPrefix(strings.TrimSpace(string(v.Expr().BuildTokens(nil).Bytes())), "<<EOF")
					routingRulesStr = strings.TrimSuffix(routingRulesStr, "EOF")

					if err := json.Unmarshal([]byte(routingRulesStr), &unmarshaledRules); err != nil {
						log.Printf("[WARN] Unable to set 'routing_rule' in aws_s3_bucket_website_configuration.%s: %s", labels[1], err)
					}

					for _, rule := range unmarshaledRules {
						routingRuleBlock := newBlock.Body().AppendNewBlock("routing_rule", nil)
						if c := rule.Condition; c != nil {
							conditionBlock := routingRuleBlock.Body().AppendNewBlock("condition", nil)
							if c.HttpErrorCodeReturnedEquals != nil {
								expr := hclwrite.NewExpressionLiteral(cty.StringVal(aws.StringValue(c.HttpErrorCodeReturnedEquals)))
								conditionBlock.Body().SetAttributeRaw("http_error_code_returned_equals", expr.BuildTokens(nil))
							}
							if c.KeyPrefixEquals != nil {
								expr := hclwrite.NewExpressionLiteral(cty.StringVal(aws.StringValue(c.KeyPrefixEquals)))
								conditionBlock.Body().SetAttributeRaw("key_prefix_equals", expr.BuildTokens(nil))
							}
						}

						if r := rule.Redirect; r != nil {
							redirectBlock := routingRuleBlock.Body().AppendNewBlock("redirect", nil)
							if r.HostName != nil {
								expr := hclwrite.NewExpressionLiteral(cty.StringVal(aws.StringValue(r.HostName)))
								redirectBlock.Body().SetAttributeRaw("host_name", expr.BuildTokens(nil))
							}
							if r.HttpRedirectCode != nil {
								expr := hclwrite.NewExpressionLiteral(cty.StringVal(aws.StringValue(r.HttpRedirectCode)))
								redirectBlock.Body().SetAttributeRaw("http_redirect_code", expr.BuildTokens(nil))
							}
							if r.Protocol != nil {
								expr := hclwrite.NewExpressionLiteral(cty.StringVal(aws.StringValue(r.Protocol)))
								redirectBlock.Body().SetAttributeRaw("protocol", expr.BuildTokens(nil))
							}
							if r.ReplaceKeyPrefixWith != nil {
								expr := hclwrite.NewExpressionLiteral(cty.StringVal(aws.StringValue(r.ReplaceKeyPrefixWith)))
								redirectBlock.Body().SetAttributeRaw("replace_key_prefix_with", expr.BuildTokens(nil))
							}
							if r.ReplaceKeyWith != nil {
								expr := hclwrite.NewExpressionLiteral(cty.StringVal(aws.StringValue(r.ReplaceKeyWith)))
								redirectBlock.Body().SetAttributeRaw("replace_key_with", expr.BuildTokens(nil))
							}
						}
					}
				}
			}

			log.Printf("	  ✓ Created aws_s3_bucket_website_configuration.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_website_configuration.%s,%s", newlabels[1], bucketPath))
		}
	}

	tmp := hclwrite.Format(f.Bytes())

	os.MkdirAll("output", 0755)

	path, err := os.Getwd()

	newFileName := filepath.Join(path, fmt.Sprintf("output/%s", *outputFile))
	nf, err := os.Create(newFileName)

	defer nf.Close()

	nf.Write(tmp)

	newFile, err := os.Create(filepath.Join(path, fmt.Sprintf("output/%s", "resources.csv")))
	for _, r := range newResources {
		newFile.WriteString(fmt.Sprintf("%s\n", r))
	}
	newFile.Close()
}
