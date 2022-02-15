package tfmigrate

import "fmt"

type Resource int64

var ResourceMap map[string]string

const (
	AccelerationStatus                = "acceleration_status"
	Acl                               = "acl"
	AccelerateConfiguration           = "accelerate_configuration"
	CorsConfiguration                 = "cors_configuration"
	CorsRule                          = "cors_rule"
	Grant                             = "grant"
	LifecycleConfiguration            = "lifecycle_configuration"
	LifecycleRule                     = "lifecycle_rule"
	Logging                           = "logging"
	ObjectLockConfiguration           = "object_lock_configuration"
	Policy                            = "policy"
	ReplicationConfiguration          = "replication_configuration"
	RequestPayer                      = "request_payer"
	RequestPaymentConfiguration       = "request_payment_configuration"
	ServerSideEncryptionConfiguration = "server_side_encryption_configuration"
	Versioning                        = "versioning"
	Website                           = "website"
	WebsiteConfiguration              = "website_configuration"

	ResourceTypeAwsS3Bucket                                 = "aws_s3_bucket"
	ResourceTypeAwsS3BucketAccelerateConfiguration Resource = iota
	ResourceTypeAwsS3BucketAcl
	ResourceTypeAwsS3BucketCorsConfiguration
	ResourceTypeAwsS3BucketLifecycleConfiguration
	ResourceTypeAwsS3BucketLogging
	ResourceTypeAwsS3BucketObjectLockConfiguration
	ResourceTypeAwsS3BucketPolicy
	ResourceTypeAwsS3BucketReplicationConfiguration
	ResourceTypeAwsS3BucketRequestPaymentConfiguration
	ResourceTypeAwsS3BucketServerSideEncryptionConfiguration
	ResourceTypeAwsS3BucketVersioning
	ResourceTypeAwsS3BucketWebsiteConfiguration
)

func (r Resource) String() string {
	switch r {
	case ResourceTypeAwsS3BucketAccelerateConfiguration:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, AccelerateConfiguration)
	case ResourceTypeAwsS3BucketAcl:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, Acl)
	case ResourceTypeAwsS3BucketCorsConfiguration:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, CorsConfiguration)
	case ResourceTypeAwsS3BucketLifecycleConfiguration:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, LifecycleConfiguration)
	case ResourceTypeAwsS3BucketLogging:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, Logging)
	case ResourceTypeAwsS3BucketObjectLockConfiguration:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, ObjectLockConfiguration)
	case ResourceTypeAwsS3BucketPolicy:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, Policy)
	case ResourceTypeAwsS3BucketReplicationConfiguration:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, ReplicationConfiguration)
	case ResourceTypeAwsS3BucketRequestPaymentConfiguration:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, RequestPaymentConfiguration)
	case ResourceTypeAwsS3BucketServerSideEncryptionConfiguration:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, ServerSideEncryptionConfiguration)
	case ResourceTypeAwsS3BucketVersioning:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, Versioning)
	case ResourceTypeAwsS3BucketWebsiteConfiguration:
		return fmt.Sprintf("%s_%s", ResourceTypeAwsS3Bucket, WebsiteConfiguration)
	}
	return "unknown"
}

func init() {
	ResourceMap = make(map[string]string)
	ResourceMap["acceleration_status"] = ResourceTypeAwsS3BucketAccelerateConfiguration.String()
	ResourceMap["acl"] = ResourceTypeAwsS3BucketAcl.String()
	ResourceMap["grant"] = ResourceTypeAwsS3BucketAcl.String()
	ResourceMap["policy"] = ResourceTypeAwsS3BucketPolicy.String()
	ResourceMap["request_payer"] = ResourceTypeAwsS3BucketRequestPaymentConfiguration.String()
}
