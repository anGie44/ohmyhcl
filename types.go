package main

type Config struct {
	Provider  Provider   `hcl:"provider,block"`
	Resources []Resource `hcl:"resource,block"`
}

type Provider struct {
	ProviderName string `hcl:"provider_name,label"`
}

type Resource struct {
	ResourceTypeName                  string                            `hcl:"resource_type_name,label"`
	ResourceName                      string                            `hcl:"resource_name,label"`
	Bucket                            string                            `hcl:"bucket"`
	BucketPrefix                      string                            `hcl:"bucket_prefix,optional"`
	AccelerationStatus                string                            `hcl:"acceleration_status,optional"`
	Acl                               string                            `hcl:"acl,optional"`
	CorsRules                         []CorsRule                        `hcl:"cors_rule,block"`
	Grants                            []Grant                           `hcl:"grant,block"`
	LifecycleRules                    []LifecycleRule                   `hcl:"lifecycle_rule,optional"`
	Logging                           Logging                           `hcl:"logging,block"`
	ObjectLockConfiguration           ObjectLockConfiguration           `hcl:"object_lock_configuration,optional"`
	Policy                            string                            `hcl:"policy,optional"`
	ReplicationConfiguration          ReplicationConfiguration          `hcl:"replication_configuration,optional"`
	ServerSideEncryptionConfiguration ServerSideEncryptionConfiguration `hcl:"server_side_encryption_configuration,optional"`
	RequestPayer                      string                            `hcl:"request_payer,optional"`
	Versioning                        Versioning                        `hcl:"versioning,optional"`
	Website                           Website                           `hcl:"website,optional"`
}

type CorsRule struct {
	AllowedHeaders []string `hcl:"allowed_headers"`
	AllowedMethods []string `hcl:"allowed_methods"`
	AllowedOrigins []string `hcl:"allowed_origins"`
	ExposeHeaders  []string `hcl:"expose_headers"`
	MaxAgeSeconds  int      `hcl:"max_age_seconds,optional"`
}

type Grant struct {
	Id          string   `hcl:"id,optional"`
	Type        string   `hcl:"type,optional"`
	Uri         string   `hcl:"uri,optional"`
	Permissions []string `hcl:"permissions"`
}

type LifecycleRule struct{}

type Logging struct {
	TargetBucket string `hcl:"target_bucket,optional"`
	TargetPrefix string `hcl:"target_prefix,optional"`
}

type ObjectLockConfiguration struct {
	ObjectLockEnabled string `hcl:"object_lock_enabled"`
}

type ReplicationConfiguration struct{}

type ServerSideEncryptionConfiguration struct{}

type Versioning struct {
	Enabled   bool `hcl:"enabled,optional"`
	MfaDelete bool `hcl:"mfa_delete,optional"`
}

type Website struct {
	IndexDocument         string `hcl:"index_document,optional"`
	ErrorDocument         string `hcl:"error_document,optional"`
	RedirectAllRequestsTo string `hcl:"redirect_all_requests_to"`
	RoutingRules          string `hcl:"routing_rules,optional"`
}
