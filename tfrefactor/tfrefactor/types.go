package tfrefactor

// Custom Routing Rule structs for parsing yaml-like strings

type Condition struct {
	HttpErrorCodeReturnedEquals *string `yaml:"HttpErrorCodeReturnedEquals"`
	KeyPrefixEquals             *string `yaml:"KeyPrefixEquals"`
}
type Redirect struct {
	HostName             *string `yaml:"HostName"`
	HttpRedirectCode     *string `yaml:"HttpRedirectCode"`
	Protocol             *string `yaml:"Protocol"`
	ReplaceKeyPrefixWith *string `yaml:"ReplaceKeyPrefixWith"`
	ReplaceKeyWith       *string `yaml:"ReplaceKeyWith"`
}

type RoutingRule struct {
	Condition *Condition `yaml:"Condition"`
	Redirect  *Redirect  `yaml:"Redirect"`
}
