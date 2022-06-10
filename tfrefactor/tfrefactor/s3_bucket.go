package tfrefactor

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type ProviderAwsS3BucketMigrator struct {
	ignoreArguments     []string
	ignoreResourceNames []string
	newResourceNames    []string
}

func NewProviderAwsS3BucketMigrator(ignoreArguments, ignoreResourceNames []string) (Migrator, error) {
	return &ProviderAwsS3BucketMigrator{
		ignoreArguments:     ignoreArguments,
		ignoreResourceNames: ignoreResourceNames,
	}, nil
}

func (m *ProviderAwsS3BucketMigrator) SkipResourceName(resourceName string) bool {
	if m == nil {
		return false
	}

	if len(m.ignoreResourceNames) == 0 {
		return false
	}

	for _, rn := range m.ignoreResourceNames {
		if rn == resourceName {
			return true
		}
	}

	return false
}

func (m *ProviderAwsS3BucketMigrator) SkipArgument(arg string) bool {
	if m == nil {
		return false
	}

	if len(m.ignoreArguments) == 0 {
		return false
	}

	for _, argument := range m.ignoreArguments {
		if argument == arg {
			return true
		}
	}

	return false
}

func (m *ProviderAwsS3BucketMigrator) Migrate(f *hclwrite.File) error {
	if err := m.migrateS3BucketResources(f); err != nil {
		return err
	}

	return nil
}

func (m *ProviderAwsS3BucketMigrator) Migrations() []string {
	if m == nil {
		return nil
	}
	return m.newResourceNames
}

func (m *ProviderAwsS3BucketMigrator) migrateS3BucketResources(f *hclwrite.File) error {
	if f == nil || f.Body() == nil {
		return fmt.Errorf("error migrating (%s) resources: empty file", ResourceTypeAwsS3Bucket)
	}

	for _, block := range f.Body().Blocks() {
		if block == nil {
			continue
		}

		labels := block.Labels()
		if len(labels) != 2 || labels[0] != ResourceTypeAwsS3Bucket {
			continue
		}

		if m.SkipResourceName(labels[1]) {
			continue
		}

		bucketPath := strings.Join(labels, ".")
		log.Printf("[INFO] Found %s\n", bucketPath)

		// Special Attribute Handling i.e. for_each and count
		countAttr := block.Body().GetAttribute("count")
		forEachAttr := block.Body().GetAttribute("for_each")

		/////////////////////////////////////////// Attribute Handling /////////////////////////////////////////////////
		// 1. acceleration_status
		// 2. acl
		// 3. policy
		// 4. request_payer
		////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		var aclResourceBlock *hclwrite.Block

		for k, v := range block.Body().Attributes() {
			if m.SkipArgument(k) {
				continue
			}
			switch k {
			case AccelerationStatus:
				block.Body().RemoveAttribute(k)
				f.Body().AppendNewline()

				newlabels := []string{ResourceMap[k], fmt.Sprintf("%s_%s", labels[1], AccelerateConfiguration)}
				newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

				newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
					hcl.TraverseRoot{
						Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
					},
				})

				newBlock.Body().SetAttributeRaw("status", v.Expr().BuildTokens(nil))

				log.Printf("	  ✓ Created %s.%s", ResourceMap[k], newlabels[1])
				m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceMap[k], newlabels[1], bucketPath))
			case Acl, Policy:
				block.Body().RemoveAttribute(k)
				f.Body().AppendNewline()

				newlabels := []string{ResourceMap[k], fmt.Sprintf("%s_%s", labels[1], k)}
				aclResourceBlock = f.Body().AppendNewBlock(block.Type(), newlabels)

				aclResourceBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
					hcl.TraverseRoot{
						Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
					},
				})

				aclResourceBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))

				log.Printf("	  ✓ Created %s.%s", ResourceMap[k], newlabels[1])
				m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceMap[k], newlabels[1], bucketPath))
			case RequestPayer:
				block.Body().RemoveAttribute(k)
				f.Body().AppendNewline()

				newlabels := []string{ResourceMap[k], fmt.Sprintf("%s_%s", labels[1], RequestPaymentConfiguration)}
				newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

				newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
					hcl.TraverseRoot{
						Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
					},
				})

				newBlock.Body().SetAttributeRaw("payer", v.Expr().BuildTokens(nil))

				log.Printf("	  ✓ Created %s.%s", ResourceMap[k], newlabels[1])
				m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceMap[k], newlabels[1], bucketPath))
			}
		}

		///////////////////////////////////////////// Block Handling ///////////////////////////////////////////////////
		// 1. Cors Rules
		// 2. Grants
		// 3. Lifecycle Rules
		// 4. Logging
		// 5. Object Lock Configuration
		// 6. Replication Configuration
		// 7. Server Side Encryption Configuration
		// 8. Website
		// 9. Versioning
		////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		var corsRules []*hclwrite.Block
		var grants []*hclwrite.Block
		var lifecycleRules []*hclwrite.Block
		var logging *hclwrite.Block
		var objectLockConfig *hclwrite.Block
		var replicationConfig *hclwrite.Block
		var serverSideEncryptionConfig *hclwrite.Block
		var website *hclwrite.Block
		var versioning *hclwrite.Block

		for _, subBlock := range block.Body().Blocks() {
			if m.SkipArgument(subBlock.Type()) {
				continue
			}

			block.Body().RemoveBlock(subBlock)

			switch t := subBlock.Type(); t {
			case CorsRule:
				corsRules = append(corsRules, subBlock)
			case Grant:
				grants = append(grants, subBlock)
			case LifecycleRule:
				lifecycleRules = append(lifecycleRules, subBlock)
			case Logging:
				logging = subBlock
			case ObjectLockConfiguration:
				objectLockConfig = subBlock
			case ReplicationConfiguration:
				replicationConfig = subBlock
			case ServerSideEncryptionConfiguration:
				serverSideEncryptionConfig = subBlock
			case Versioning:
				versioning = subBlock
			case Website:
				website = subBlock
			case "dynamic":
				// TODO: Account for "dynamic" blocks ... yikes ...
				// Maybe we can recreate them ??
				argument := subBlock.Labels()[0] // e.g. "website"

				forEachAttr := subBlock.Body().GetAttribute("for_each")

				for _, b := range subBlock.Body().Blocks() {
					// Expected: content
					if b.Type() != "content" {
						continue
					}

					switch argument {
					case CorsRule:
						// There can be many defined so we can maintain the dynamic block?
						corsRules = append(corsRules, subBlock)
					case Logging:
						// Set block with additional for_each data
						if forEachAttr != nil {
							b.Body().SetAttributeRaw("for_each", forEachAttr.Expr().BuildTokens(nil))
						}
						logging = b
					case Website:
						// Set block with additional for_each data
						if forEachAttr != nil {
							b.Body().SetAttributeRaw("for_each", forEachAttr.Expr().BuildTokens(nil))
						}
						website = b
					}
				}
			}
		}

		if len(corsRules) > 0 {
			// Create new Cors resource
			f.Body().AppendNewline()

			newlabels := []string{ResourceTypeAwsS3BucketCorsConfiguration.String(), fmt.Sprintf("%s_%s", labels[1], CorsConfiguration)}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			for _, crBlock := range corsRules {
				if crBlock.Type() == "dynamic" {
					if forEach := crBlock.Body().GetAttribute("for_each"); forEach != nil {
						newBlock.Body().AppendUnstructuredTokens(hclwrite.Tokens{
							{
								Type:  hclsyntax.TokenComment,
								Bytes: []byte("# TODO: Replace with your intended 'for_each' value\n"),
							},
						})
						newBlock.Body().SetAttributeRaw("# for_each ", forEach.Expr().BuildTokens(nil))
					}
				}
			}

			if countAttr != nil || forEachAttr != nil {
				newBlock.Body().AppendUnstructuredTokens(hclwrite.Tokens{
					{
						Type: hclsyntax.TokenComment,
						Bytes: []byte(fmt.Sprintf(`
# TODO: Replace 'bucket' argument value with correct instance index
# e.g. aws_s3_bucket.%s[count.index].id
`, labels[1])),
					},
				})
			}

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for _, b := range corsRules {
				if b.Type() == "dynamic" {
					// Update content to use "each.value"
					for _, bb := range b.Body().Blocks() {
						if bb.Type() == "content" {
							for k, v := range bb.Body().Attributes() {
								bb.Body().SetAttributeTraversal(k, hcl.Traversal{
									hcl.TraverseRoot{
										Name: strings.Replace(string(v.Expr().BuildTokens(nil).Bytes()), "cors_rule.value", "each.value", 1),
									},
								})
							}
						}
					}
				}
				newBlock.Body().AppendBlock(b)
			}

			log.Printf("	  ✓ Created %s.%s", ResourceTypeAwsS3BucketCorsConfiguration, newlabels[1])
			m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceTypeAwsS3BucketCorsConfiguration, newlabels[1], bucketPath))
		}

		if len(grants) > 0 {
			if aclResourceBlock == nil {
				// Create new aws_s3_bucket_acl resource
				f.Body().AppendNewline()

				newlabels := []string{ResourceTypeAwsS3BucketAcl.String(), fmt.Sprintf("%s_%s", labels[1], Acl)}
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
						// Expected: id, type, uri, permissions
						if k == "permissions" {
							for _, t := range v.BuildTokens(nil) {
								if p := string(t.Bytes); len(p) > 1 && p != k {
									permissions = append(permissions, p)
								}
							}
						} else {
							grantee.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
						}
					}

					if len(permissions) == 0 {
						continue
					}

					grantBlock.Body().SetAttributeValue("permission", cty.StringVal(permissions[0]))

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

							grantBlock.Body().SetAttributeValue("permission", cty.StringVal(permission))
						}
					}
				}

				log.Printf("	  ✓ Created %s.%s", ResourceTypeAwsS3BucketAcl, newlabels[1])
				m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceTypeAwsS3BucketAcl, newlabels[1], bucketPath))
			} // TODO: Account for case where "acl" and "grant" are configured
		}

		if len(lifecycleRules) > 0 {
			f.Body().AppendNewline()

			newlabels := []string{ResourceTypeAwsS3BucketLifecycleConfiguration.String(), fmt.Sprintf("%s_%s", labels[1], LifecycleConfiguration)}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for _, lifecycleRuleBlock := range lifecycleRules {
				ruleBlock := newBlock.Body().AppendNewBlock("rule", nil)

				m := make(map[string]*hclwrite.Attribute)

				for k, v := range lifecycleRuleBlock.Body().Attributes() {
					// Expected: id, prefix, tags, enabled, abort_incomplete_multipart_upload_days
					switch k {
					case "abort_incomplete_multipart_upload_days":
						// This is represented as a abort_incomplete_multipart_upload block in the new resource
						abortBlock := ruleBlock.Body().AppendNewBlock("abort_incomplete_multipart_upload", nil)
						abortBlock.Body().SetAttributeRaw("days_after_initiation", v.Expr().BuildTokens(nil))
					case "enabled":
						// This is represented as "status" in the new resource
						value := strings.TrimSpace(string(v.Expr().BuildTokens(nil).Bytes()))
						if value == "true" {
							ruleBlock.Body().SetAttributeValue("status", cty.StringVal("Enabled"))
						} else if value == "false" {
							ruleBlock.Body().SetAttributeValue("status", cty.StringVal("Disabled"))
						}
					case "id":
						ruleBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
					case "prefix", "tags":
						m[k] = v
					}
				}

				if vTags, ok := m["tags"]; ok {
					filterBlock := ruleBlock.Body().AppendNewBlock("filter", nil)
					andBlock := filterBlock.Body().AppendNewBlock("and", nil)
					andBlock.Body().SetAttributeRaw("tags", vTags.Expr().BuildTokens(nil))
					if vPrefix, vOk := m["prefix"]; vOk {
						andBlock.Body().SetAttributeRaw("prefix", vPrefix.Expr().BuildTokens(nil))
					} else {
						andBlock.Body().SetAttributeValue("prefix", cty.StringVal(""))
					}
				} else if vPrefix, vOk := m["prefix"]; vOk {
					filterBlock := ruleBlock.Body().AppendNewBlock("filter", nil)
					filterBlock.Body().SetAttributeRaw("prefix", vPrefix.Expr().BuildTokens(nil))
				}

				for _, b := range lifecycleRuleBlock.Body().Blocks() {
					// Expected: expiration, noncurrent_version_expiration, transition, noncurrent_version_transition
					switch b.Type() {
					case "expiration", "transition":
						ruleBlock.Body().AppendBlock(b)
					case "noncurrent_version_expiration":
						nve := ruleBlock.Body().AppendNewBlock("noncurrent_version_expiration", nil)
						for k, v := range b.Body().Attributes() {
							// Expected: days
							if k != "days" {
								continue
							}
							// "days" is represented as "noncurrent_days" in the new resource
							nve.Body().SetAttributeRaw("noncurrent_days", v.Expr().BuildTokens(nil))
						}
					case "noncurrent_version_transition":
						nvt := ruleBlock.Body().AppendNewBlock("noncurrent_version_transition", nil)
						for k, v := range b.Body().Attributes() {
							// Expected: days, storage_class
							switch k {
							case "days":
								// "days" is represented as "noncurrent_days" in the new resource
								nvt.Body().SetAttributeRaw("noncurrent_days", v.Expr().BuildTokens(nil))
							case "storage_class":
								nvt.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
							}
						}
					}
				}
			}

			log.Printf("	  ✓ Created %s.%s", ResourceTypeAwsS3BucketLifecycleConfiguration, newlabels[1])
			m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceTypeAwsS3BucketLifecycleConfiguration, newlabels[1], bucketPath))
		}

		if logging != nil {
			f.Body().AppendNewline()

			newlabels := []string{ResourceTypeAwsS3BucketLogging.String(), fmt.Sprintf("%s_%s", labels[1], Logging)}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			// Account for dynamic blocks of this argument
			var hasForEach bool
			loggingForEachAttribute := logging.Body().GetAttribute("for_each")
			if loggingForEachAttribute != nil {
				hasForEach = true
				newBlock.Body().SetAttributeRaw("for_each", loggingForEachAttribute.Expr().BuildTokens(nil))
				newBlock.Body().AppendNewline()
			}

			bucketAttribute := fmt.Sprintf("%s.%s.id", labels[0], labels[1])

			if (countAttr != nil || forEachAttr != nil) && hasForEach {
				bucketAttribute = fmt.Sprintf("%s.%s[each.key].id", labels[0], labels[1])
			}

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: bucketAttribute,
				},
			})

			for k, v := range logging.Body().Attributes() {
				// Expected: target_bucket, target_prefix
				if hasForEach {
					val := strings.Replace(string(v.Expr().BuildTokens(nil).Bytes()), "logging.value", "each.value", 1)
					newBlock.Body().SetAttributeTraversal(k, hcl.Traversal{
						hcl.TraverseRoot{
							Name: val,
						},
					})
				} else {
					newBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
				}
			}

			log.Printf("	  ✓ Created %s.%s", ResourceTypeAwsS3BucketLogging, newlabels[1])
			m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceTypeAwsS3BucketLogging, newlabels[1], bucketPath))
		}

		if versioning != nil {
			f.Body().AppendNewline()

			newlabels := []string{ResourceTypeAwsS3BucketVersioning.String(), fmt.Sprintf("%s_%s", labels[1], Versioning)}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			versioningConfigBlock := newBlock.Body().AppendNewBlock("versioning_configuration", nil)

			for k, v := range versioning.Body().Attributes() {
				// Expected: enabled, mfa_delete
				switch k {
				case "enabled":
					value := strings.TrimSpace(string(v.Expr().BuildTokens(nil).Bytes()))
					if value == "true" {
						expr := hclwrite.NewExpressionLiteral(cty.StringVal("Enabled"))
						versioningConfigBlock.Body().SetAttributeRaw("status", expr.BuildTokens(nil))
					} else if value == "false" {
						// This might not be accurate as "false" can indicate never enable versioning
						expr := hclwrite.NewExpressionLiteral(cty.StringVal("Suspended"))
						versioningConfigBlock.Body().SetAttributeRaw("status", expr.BuildTokens(nil))
					}
				case "mfa_delete":
					value := strings.TrimSpace(string(v.Expr().BuildTokens(nil).Bytes()))
					if value == "true" {
						expr := hclwrite.NewExpressionLiteral(cty.StringVal("Enabled"))
						versioningConfigBlock.Body().SetAttributeRaw("mfa_delete", expr.BuildTokens(nil))
					} else if value == "false" {
						expr := hclwrite.NewExpressionLiteral(cty.StringVal("Disabled"))
						versioningConfigBlock.Body().SetAttributeRaw("mfa_delete", expr.BuildTokens(nil))
					}
				}
			}

			log.Printf("	  ✓ Created %s.%s", newlabels[1], ResourceTypeAwsS3BucketVersioning)
			m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceTypeAwsS3BucketVersioning, newlabels[1], bucketPath))
		}

		if objectLockConfig != nil {
			f.Body().AppendNewline()

			newlabels := []string{ResourceTypeAwsS3BucketObjectLockConfiguration.String(), fmt.Sprintf("%s_%s", labels[1], ObjectLockConfiguration)}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for k, v := range objectLockConfig.Body().Attributes() {
				// Expected: object_lock_enabled
				if k != "object_lock_enabled" {
					continue
				}
				newBlock.Body().SetAttributeRaw("object_lock_enabled", v.Expr().BuildTokens(nil))
			}

			for _, ob := range objectLockConfig.Body().Blocks() {
				// we only expect 1 rule as defined in the aws_s3_bucket schema
				if ob.Type() != "rule" {
					continue
				}
				newBlock.Body().AppendBlock(ob)
			}

			log.Printf("	  ✓ Created %s.%s", ResourceTypeAwsS3BucketObjectLockConfiguration, newlabels[1])
			m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceTypeAwsS3BucketObjectLockConfiguration, newlabels[1], bucketPath))
		}

		if replicationConfig != nil {
			f.Body().AppendNewline()

			newlabels := []string{ResourceTypeAwsS3BucketReplicationConfiguration.String(), fmt.Sprintf("%s_%s", labels[1], ReplicationConfiguration)}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: fmt.Sprintf("%s.%s.id", labels[0], labels[1]),
				},
			})

			for k, v := range replicationConfig.Body().Attributes() {
				// Expected: role
				if k != "role" {
					continue
				}
				newBlock.Body().SetAttributeRaw("role", v.Expr().BuildTokens(nil))
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
						// This is represented as a block in the new resource
						deleteMarkerBlock := ruleBlock.Body().AppendNewBlock("delete_marker_replication", nil)
						deleteMarkerBlock.Body().SetAttributeRaw("status", v.Expr().BuildTokens(nil))
					}
				}

				for _, innerRuleBlock := range b.Body().Blocks() {
					// Expected: filter, source_selection_criteria, destination
					switch innerRuleBlock.Type() {
					case "destination":
						destBlock := ruleBlock.Body().AppendNewBlock("destination", nil)

						for k, v := range innerRuleBlock.Body().Attributes() {
							// Expected: account_id, bucket, storage_class, replica_kms_key_id
							switch k {
							case "account_id":
								// This is represented as "account" in the new resource
								destBlock.Body().SetAttributeRaw("account", v.Expr().BuildTokens(nil))
							case "bucket", "storage_class":
								destBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
							case "replica_kms_key_id":
								// This is represented as an encryption_configuration block in the new resource
								encryptionBlock := destBlock.Body().AppendNewBlock("encryption_configuration", nil)
								encryptionBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
							}
						}

						for _, irb := range innerRuleBlock.Body().Blocks() {
							// Expected: access_control_translation, replication_time, metrics
							switch irb.Type() {
							case "access_control_translation":
								destBlock.Body().AppendBlock(irb)
							case "metrics":
								// This is represented as metrics.event_threshold.minutes and metrics.status in the new resource
								metricsBlock := destBlock.Body().AppendNewBlock("metrics", nil)
								for k, v := range irb.Body().Attributes() {
									// Expect: minutes, status
									switch k {
									case "minutes":
										// Need to wrap in a "event_threshold" block
										etBlock := metricsBlock.Body().AppendNewBlock("event_threshold", nil)
										etBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
									case "status":
										metricsBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
									}
								}
							case "replication_time":
								// This is represented as replication_time.time.minutes and replication_time.status in the new resource
								repTimeBlock := destBlock.Body().AppendNewBlock("replication_time", nil)
								for k, v := range irb.Body().Attributes() {
									// Expect: minutes, status
									switch k {
									case "minutes":
										// Need to wrap in a "time" block
										timeBlock := repTimeBlock.Body().AppendNewBlock("time", nil)
										timeBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
									case "status":
										repTimeBlock.Body().SetAttributeRaw(k, v.Expr().BuildTokens(nil))
									}
								}
							}
						}

					case "filter":
						filterBlock := ruleBlock.Body().AppendNewBlock("filter", nil)

						m := make(map[string]*hclwrite.Attribute)

						for k, v := range innerRuleBlock.Body().Attributes() {
							// Expected: prefix and/or tags
							switch k {
							case "prefix", "tags":
								m[k] = v
							}
						}

						if vTags, ok := m["tags"]; ok {
							andBlock := filterBlock.Body().AppendNewBlock("and", nil)
							andBlock.Body().SetAttributeRaw("tags", vTags.Expr().BuildTokens(nil))
							if vPrefix, vOk := m["prefix"]; vOk {
								andBlock.Body().SetAttributeRaw("prefix", vPrefix.Expr().BuildTokens(nil))
							} else {
								andBlock.Body().SetAttributeValue("prefix", cty.StringVal(""))
							}
						} else if vPrefix, ok := m["prefix"]; ok {
							filterBlock.Body().SetAttributeRaw("prefix", vPrefix.Expr().BuildTokens(nil))
						}
					case "source_selection_criteria":
						sscBlock := ruleBlock.Body().AppendNewBlock("source_selection_criteria", nil)

						for _, innerSscBlock := range innerRuleBlock.Body().Blocks() {
							switch innerSscBlock.Type() {
							case "sse_kms_encrypted_objects":
								sseBlock := sscBlock.Body().AppendNewBlock("sse_kms_encrypted_objects", nil)
								for k, v := range innerSscBlock.Body().Attributes() {
									if k != "enabled" {
										continue
									}

									value := strings.TrimSpace(string(v.Expr().BuildTokens(nil).Bytes()))

									if value == "true" {
										sseBlock.Body().SetAttributeValue("status", cty.StringVal("Enabled"))
									} else if value == "false" {
										sseBlock.Body().SetAttributeValue("status", cty.StringVal("Disabled"))
									}
								}
							}
						}
					}
				}
			}

			log.Printf("	  ✓ Created %s.%s", ResourceTypeAwsS3BucketReplicationConfiguration, newlabels[1])
			m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceTypeAwsS3BucketReplicationConfiguration, newlabels[1], bucketPath))
		}

		if serverSideEncryptionConfig != nil {
			f.Body().AppendNewline()

			newlabels := []string{ResourceTypeAwsS3BucketServerSideEncryptionConfiguration.String(), fmt.Sprintf("%s_%s", labels[1], ServerSideEncryptionConfiguration)}
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

			log.Printf("	  ✓ Created %s.%s", ResourceTypeAwsS3BucketServerSideEncryptionConfiguration, newlabels[1])
			m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceTypeAwsS3BucketServerSideEncryptionConfiguration, newlabels[1], bucketPath))
		}

		if website != nil {
			f.Body().AppendNewline()

			newlabels := []string{ResourceTypeAwsS3BucketWebsiteConfiguration.String(), fmt.Sprintf("%s_%s", labels[1], WebsiteConfiguration)}
			newBlock := f.Body().AppendNewBlock(block.Type(), newlabels)

			// Account for dynamic blocks of this argument
			var hasForEach bool
			websiteForEachAttribute := website.Body().GetAttribute("for_each")
			if websiteForEachAttribute != nil {
				hasForEach = true
				newBlock.Body().SetAttributeRaw("for_each", websiteForEachAttribute.Expr().BuildTokens(nil))
				newBlock.Body().AppendNewline()
			}

			bucketAttribute := fmt.Sprintf("%s.%s.id", labels[0], labels[1])

			if (countAttr != nil || forEachAttr != nil) && hasForEach {
				bucketAttribute = fmt.Sprintf("%s.%s[each.key].id", labels[0], labels[1])
			}

			newBlock.Body().SetAttributeTraversal("bucket", hcl.Traversal{
				hcl.TraverseRoot{
					Name: bucketAttribute,
				},
			})

			for k, v := range website.Body().Attributes() {
				switch k {
				case "index_document":
					indexDocBlock := newBlock.Body().AppendNewBlock("index_document", nil)

					if hasForEach {
						// Is it safe to assume this value will always be <argument_name>.value ?
						val := strings.Replace(string(v.Expr().BuildTokens(nil).Bytes()), "website.value", "each.value", 1)
						indexDocBlock.Body().SetAttributeTraversal("suffix", hcl.Traversal{
							hcl.TraverseRoot{
								Name: val,
							},
						})
					} else {
						indexDocBlock.Body().SetAttributeRaw("suffix", v.Expr().BuildTokens(nil))
					}

				case "error_document":
					errDocBlock := newBlock.Body().AppendNewBlock("error_document", nil)

					if hasForEach {
						val := strings.Replace(string(v.Expr().BuildTokens(nil).Bytes()), "website.value", "each.value", 1)
						errDocBlock.Body().SetAttributeTraversal("key", hcl.Traversal{
							hcl.TraverseRoot{
								Name: val,
							},
						})
					} else {
						errDocBlock.Body().SetAttributeRaw("key", v.Expr().BuildTokens(nil))
					}
				case "redirect_all_requests_to":
					redirectBlock := newBlock.Body().AppendNewBlock("redirect_all_requests_to", nil)

					if hasForEach {
						val := strings.Replace(string(v.Expr().BuildTokens(nil).Bytes()), "website.value", "each.value", 1)
						redirectBlock.Body().SetAttributeTraversal("host_name", hcl.Traversal{
							hcl.TraverseRoot{
								Name: val,
							},
						})
					} else {
						redirectBlock.Body().SetAttributeRaw("host_name", v.Expr().BuildTokens(nil))
					}
				case "routing_rules":
					var unmarshalledRules []*s3.RoutingRule    // if we can parse string as JSON
					var customUnmarshalledRules []*RoutingRule // if we can't parse string as JSON, try as YAML (e.g. when jsonencode func is used in terraform)

					routingRulesStr := strings.TrimSpace(string(v.Expr().BuildTokens(nil).Bytes()))
					indexOfOpenBracket := strings.Index(routingRulesStr, "[")
					indexOfCloseBracket := strings.LastIndex(routingRulesStr, "]")

					if indexOfOpenBracket == -1 || indexOfCloseBracket == -1 {
						log.Printf("[WARN] Unable to set 'routing_rule' in %s.%s.%s as configuration blocks from value", ResourceTypeAwsS3BucketWebsiteConfiguration, labels[1], WebsiteConfiguration)
						newBlock.Body().AppendUnstructuredTokens(hclwrite.Tokens{
							{
								Type:  hclsyntax.TokenComment,
								Bytes: []byte("# TODO: Replace with your 'routing_rule' configuration\n"),
							},
						})
						continue
					}

					routingRulesStr = routingRulesStr[indexOfOpenBracket : indexOfCloseBracket+1]

					if err := json.Unmarshal([]byte(routingRulesStr), &unmarshalledRules); err != nil {
						log.Printf("[DEBUG] Unable to json unmarshal 'routing_rule' in %s.%s_%s: %s. Trying yaml unmarshal...", ResourceTypeAwsS3BucketWebsiteConfiguration, labels[1], WebsiteConfiguration, err)
						if yamlErr := yaml.Unmarshal([]byte(routingRulesStr), &customUnmarshalledRules); yamlErr != nil {
							log.Printf("[DEBUG] Unable to yaml unmarshal 'routing_rule' in %s.%s_%s: %s", ResourceTypeAwsS3BucketWebsiteConfiguration, labels[1], WebsiteConfiguration, yamlErr)
						}
					}

					if len(unmarshalledRules) == 0 && len(customUnmarshalledRules) == 0 {
						log.Printf("[WARN] Unable to set 'routing_rule' in %s.%s_%s: no routing rules parsed", ResourceTypeAwsS3BucketWebsiteConfiguration, labels[1], WebsiteConfiguration)
						newBlock.Body().AppendUnstructuredTokens(hclwrite.Tokens{
							{
								Type:  hclsyntax.TokenComment,
								Bytes: []byte("# TODO: Replace with your 'routing_rule' configuration\n"),
							},
						})
						continue
					}

					for _, rule := range customUnmarshalledRules {
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

					for _, rule := range unmarshalledRules {
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

			log.Printf("	  ✓ Created %s.%s", ResourceTypeAwsS3BucketWebsiteConfiguration, newlabels[1])
			m.newResourceNames = append(m.newResourceNames, fmt.Sprintf("%s.%s,%s", ResourceTypeAwsS3BucketWebsiteConfiguration, newlabels[1], bucketPath))
		}
	}

	return nil
}
