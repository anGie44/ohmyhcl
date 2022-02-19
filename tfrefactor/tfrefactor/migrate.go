package tfrefactor

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/minamijoyo/tfupdate/tfupdate"
	"github.com/pkg/errors"
)

const v4 = "4.0.0"

type Migrator interface {
	Migrate(file *hclwrite.File) error
	Migrations() []string
}

func NewMigrator(o Option) (Migrator, error) {
	switch o.MigratorType {
	case "resource":
		switch o.ResourceType {
		case "aws_s3_bucket":
			return NewProviderAwsS3BucketMigrator(o.IgnoreArguments, o.IgnoreResourceNames)
		default:
			return nil, errors.Errorf("failed to create new migrator. unknown resource type: %s", o.ResourceType)
		}
	default:
		return nil, errors.Errorf("failed to create new migrator. unknown type: %s", o.MigratorType)
	}
}

func MigrateHCL(r io.Reader, w io.Writer, filename string, o Option) ([]string, error) {
	input, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %s", err)
	}

	f, diags := hclwrite.ParseConfig(input, filename, hcl.Pos{Line: 1, Column: 1})
	if diags != nil {
		var errs *multierror.Error
		for _, diag := range diags {
			if diag.Error() != "" {
				errs = multierror.Append(errs, fmt.Errorf(diag.Error()))
			}
		}
		return nil, errs.ErrorOrNil()
	}

	// Migrate Provider Version(s)
	if o.ProviderVersion != "" {
		if o.ProviderVersion == "latest" {
			o.ProviderVersion = v4
		}

		p, err := tfupdate.NewProviderUpdater("aws", o.ProviderVersion)
		if err != nil {
			return nil, fmt.Errorf("error creating tfupdate.ProviderUpdater: %w", err)
		}

		if err := p.Update(f); err != nil {
			return nil, fmt.Errorf("error updating provider configurations to %s: %s", o.ProviderVersion, err)
		}
	}

	m, err := NewMigrator(o)
	if err != nil {
		return nil, err
	}

	if err = m.Migrate(f); err != nil {
		return m.Migrations(), err
	}

	output := f.BuildTokens(nil).Bytes()

	if _, err := w.Write(output); err != nil {
		return m.Migrations(), fmt.Errorf("failed to write output: %s", err)
	}

	return m.Migrations(), nil
}
