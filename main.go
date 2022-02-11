package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/cli"
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
				ui.Error(fmt.Sprintf("error loading configuration: %s", err))
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
			expr, err := buildExpression("bucket", fmt.Sprintf("%s.%s.id", labels[0], labels[1]))
			if err != nil {
				continue
			}
			newBlock.Body().SetAttributeRaw("bucket", expr.BuildTokens(nil))
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
			expr, err := buildExpression("bucket", fmt.Sprintf("%s.%s.id", labels[0], labels[1]))
			if err != nil {
				continue
			}
			aclResourceBlock.Body().SetAttributeRaw("bucket", expr.BuildTokens(nil))
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
			expr, err := buildExpression("bucket", fmt.Sprintf("%s.%s.id", labels[0], labels[1]))
			if err != nil {
				continue
			}
			newBlock.Body().SetAttributeRaw("bucket", expr.BuildTokens(nil))
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
			expr, err := buildExpression("bucket", fmt.Sprintf("%s.%s.id", labels[0], labels[1]))
			if err != nil {
				continue
			}
			newBlock.Body().SetAttributeRaw("bucket", expr.BuildTokens(nil))
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
		//var objectLockConfig *hclwrite.Block
		//var replicationConfig *hclwrite.Block
		//var serverSideEncryptionConfig *hclwrite.Block
		//var website *hclwrite.Block
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
			//case "object_lock_configuration":
			//	objectLockConfig = subBlock
			//case "replication_configuration":
			//	replicationConfig = subBlock
			//case "server_side_encryption_configuration":
			//	serverSideEncryptionConfig = subBlock
			case "versioning":
				versioning = subBlock
				//case "website":
				//	website = subBlock
			}
		}

		if len(corsRules) > 0 {
			// Create new Cors resource
			f.Body().AppendNewline()

			newlabels := []string{"aws_s3_bucket_cors_configuration", fmt.Sprintf("%s_cors_configuration", labels[1])}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)
			expr, err := buildExpression("bucket", fmt.Sprintf("%s.%s.id", labels[0], labels[1]))
			if err != nil {
				continue
			}

			newBlock.Body().SetAttributeRaw("bucket", expr.BuildTokens(nil))

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
				expr, err := buildExpression("bucket", fmt.Sprintf("%s.%s.id", labels[0], labels[1]))
				if err != nil {
					continue
				}

				newBlock.Body().SetAttributeRaw("bucket", expr.BuildTokens(nil))
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

					expr, err := buildExpression("permission", permissions[0])
					if err != nil {
						continue
					}
					grantBlock.Body().SetAttributeRaw("permission", expr.BuildTokens(nil))

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

							expr, err := buildExpression("permission", permission)
							if err != nil {
								continue
							}

							grantBlock.Body().SetAttributeRaw("permission", expr.BuildTokens(nil))
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
			expr, err := buildExpression("bucket", fmt.Sprintf("%s.%s.id", labels[0], labels[1]))
			if err != nil {
				continue
			}

			newBlock.Body().SetAttributeRaw("bucket", expr.BuildTokens(nil))
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
			expr, err := buildExpression("bucket", fmt.Sprintf("%s.%s.id", labels[0], labels[1]))
			if err != nil {
				continue
			}

			newBlock.Body().SetAttributeRaw("bucket", expr.BuildTokens(nil))
			versioningConfigBlock := newBlock.Body().AppendNewBlock("versioning_configuration", nil)

			for k, v := range versioning.Body().Attributes() {
				if k == "enabled" {
					value := strings.TrimSpace(string(v.Expr().BuildTokens(nil).Bytes()))
					if value == "true" {
						expr, err := buildExpression("status", fmt.Sprintf("%q", "Enabled"))
						if err != nil {
							continue
						}
						versioningConfigBlock.Body().SetAttributeRaw("status", expr.BuildTokens(nil))
					} else if value == "false" {
						// This might not be accurate as "false" can indicate never enable versioning
						expr, err := buildExpression("status", fmt.Sprintf("%q", "Suspended"))
						if err != nil {
							continue
						}
						versioningConfigBlock.Body().SetAttributeRaw("status", expr.BuildTokens(nil))
					}
				}
			}

			log.Printf("	  ✓ Created aws_s3_bucket_versioning.%s", newlabels[1])
			newResources = append(newResources, fmt.Sprintf("aws_s3_bucket_versioning.%s,%s", newlabels[1], bucketPath))
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

// Helper Functions
// Author: https://github.com/minamijoyo/hcledit
func safeParseConfig(src []byte, filename string, start hcl.Pos) (f *hclwrite.File, e error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[DEBUG] failed to parse input: %s\nstacktrace: %s", filename, string(debug.Stack()))
			// Set a return value from panic recover
			e = fmt.Errorf(`failed to parse input: %s
panic: %s
This may be caused by a bug in the hclwrite parser`, filename, err)
		}
	}()

	f, diags := hclwrite.ParseConfig(src, filename, start)

	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse input: %s", diags)
	}

	return f, nil
}

func buildExpression(name string, value string) (*hclwrite.Expression, error) {
	src := name + " = " + value
	f, err := safeParseConfig([]byte(src), "generated_by_buildExpression", hcl.Pos{Line: 1, Column: 1})
	if err != nil {
		return nil, fmt.Errorf("failed to build expression at the parse phase: %s", err)
	}

	attr := f.Body().GetAttribute(name)
	if attr == nil {
		return nil, fmt.Errorf("failed to build expression at the get phase. name = %s, value = %s", name, value)
	}

	return attr.Expr(), nil
}
