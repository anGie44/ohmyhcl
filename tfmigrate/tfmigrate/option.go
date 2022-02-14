package tfmigrate

import (
	"fmt"
	"regexp"
)

// Option is a set of parameters to migrate.
type Option struct {
	MigratorType string

	// ResourceType to migrate e.g. aws_s3_bucket
	ResourceType string

	// a new provider version constraint
	ProviderVersion string

	// If a recursive flag is true, it checks and updates directories recursively.
	Recursive bool

	// An array of arguments to ignore
	IgnoreArguments []string

	// An array of resource names to ignore
	IgnoreResourceNames []string

	// An array of regular expression for paths to ignore.
	IgnorePaths []*regexp.Regexp
}

// NewOption returns an option.
func NewOption(migratorType, resourceType, providerVersion string, recursive bool, ignoreArguments, ignoreResourceNames, ignorePaths []string) (Option, error) {
	regexps := make([]*regexp.Regexp, 0, len(ignorePaths))
	for _, ignorePath := range ignorePaths {
		if len(ignorePath) == 0 {
			continue
		}

		r, err := regexp.Compile(ignorePath)
		if err != nil {
			return Option{}, fmt.Errorf("faild to compile regexp for ignorePath: %s", err)
		}
		regexps = append(regexps, r)
	}

	return Option{
		MigratorType:        migratorType,
		ResourceType:        resourceType,
		ProviderVersion:     providerVersion,
		Recursive:           recursive,
		IgnoreArguments:     ignoreArguments,
		IgnoreResourceNames: ignoreResourceNames,
		IgnorePaths:         regexps,
	}, nil
}

// MatchIgnorePaths returns whether any of the ignore conditions are met.
func (o *Option) MatchIgnorePaths(path string) bool {
	for _, r := range o.IgnorePaths {
		if r.MatchString(path) {
			return true
		}
	}

	return false
}
