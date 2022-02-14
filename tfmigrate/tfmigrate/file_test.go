package tfmigrate

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestMigrateFileExist(t *testing.T) {
	cases := []struct {
		name                      string
		filename                  string
		src                       string
		o                         Option
		want                      string
		expectedErr               error
		expectedMigrationFilename string
	}{
		{
			filename: "valid.tf",
			src: `
terraform {
  required_providers {
    aws = {
	  version = "3.74.0"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-acc-test-1234"
  acl    = "private"
}
`,
			o: Option{
				MigratorType: "resource",
				ResourceType: ResourceTypeAwsS3Bucket,
			},
			expectedMigrationFilename: "valid_migrated.tf",
			want: `
terraform {
  required_providers {
    aws = {
      version = "3.74.0"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-acc-test-1234"
}

resource "aws_s3_bucket_acl" "test_acl" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`,
		},
		{
			filename: "valid.tf",
			src: `
terraform {
  required_providers {
    aws = {
	  version = "3.74.0"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-acc-test-1234"
  acl    = "private"
}
`,
			o: Option{
				MigratorType:    "resource",
				ProviderVersion: "latest",
				ResourceType:    ResourceTypeAwsS3Bucket,
			},
			expectedMigrationFilename: "valid_migrated.tf",
			want: `
terraform {
  required_providers {
    aws = {
      version = "4.0.0"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-acc-test-1234"
}

resource "aws_s3_bucket_acl" "test_acl" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`,
		},
		{
			filename: "valid.tf",
			src: `
terraform {
  required_providers {
    aws = {
	  version = "3.74.0"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-acc-test-1234"
  acl    = "private"
}
`,
			o: Option{
				MigratorType:    "resource",
				ProviderVersion: "4.1.0",
				ResourceType:    ResourceTypeAwsS3Bucket,
			},
			expectedMigrationFilename: "valid_migrated.tf",
			want: `
terraform {
  required_providers {
    aws = {
      version = "4.1.0"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-acc-test-1234"
}

resource "aws_s3_bucket_acl" "test_acl" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`,
		},
		{
			filename: "unformatted_mo_match.tf",
			src: `
			terraform {
			required_providers {
			  aws = "3.74.0"
			}
			}

		    resource "aws_s3_bucket" "test" {
			bucket = "tf-acc-test-1234"
			}
			`,
			o: Option{
				MigratorType:    "resource",
				ResourceType:    ResourceTypeAwsS3Bucket,
				ProviderVersion: "latest",
			},
		},
	}
	for _, tc := range cases {
		fs := afero.NewMemMapFs()
		err := afero.WriteFile(fs, tc.filename, []byte(tc.src), 0644)
		if err != nil {
			t.Fatalf("failed to write file: %s", err)
		}

		err = MigrateFile(fs, tc.filename, tc.o)
		if tc.expectedErr == nil && err != nil {
			t.Errorf("MigrateFile() with filename = %s, o = %#v returns unexpected err: %+v", tc.filename, tc.o, err)
		}

		if tc.expectedErr != nil && err == nil {
			t.Errorf("MigrateFile() with filename = %s, o = %#v expects to return an error, but no error", tc.filename, tc.o)
		}

		if tc.expectedMigrationFilename == "" {
			migrationFile := fmt.Sprintf("%s_migrated.tf", tc.filename)
			if _, err := os.Stat(migrationFile); err == nil {
				t.Errorf("MigrateFile() with no migrations expects to return no migration file, but found migration file: %s", migrationFile)
			}
		}

		if tc.expectedMigrationFilename != "" {
			got, err := afero.ReadFile(fs, tc.expectedMigrationFilename)
			if err != nil {
				t.Fatalf("failed to read migration file: %s", err)
			}

			if string(got) != tc.want {
				t.Errorf("MigrateFile() with filename = %s, o = %#v returns %s, but want = %s", tc.filename, tc.o, string(got), tc.want)
			}
		}
	}
}
