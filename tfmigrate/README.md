# tfmigrate

## Features

- Migrate `aws_s3_bucket` resource arguments to independent resources available since `v4.0.0` of the Terraform AWS Provider.
- Update version constraints of the Terraform AWS Provider defined in configurations.
- Get a table (in `.csv` format) of each new resource with its parent `aws_s3_bucket` to enable resource import.

### Source

If you have Go 1.17+ development environment:

```
$ git clone https://github.com/anGie44/ohmyhcl
$ cd ohmyhcl/
$ make install
```

## Usage
```shell
tfmigrate --help
Usage: tfmigrate [--version] [--help] <command> [<args>]

Available commands are:
    resource    Migrate resource arguments to individual resources
```

### resource

```shell
$ tfmigrate resource --help
Usage: tfmigrate resource <RESOURCE_TYPE> <PATH> [options]
Arguments
  RESOURCE_TYPE      The provider resource type (e.g. aws_s3_bucket)
  PATH               A path of file or directory to update
Options:
  --ignore-arguments     The arguments to migrate (default: all)
                         Set the flag with values separated by commas (e.g. --ignore-arguments="acl,grant") or set the flag multiple times.
  --ignore-names         The resource names to migrate (default: all)
                         Set the flag with values separated by commas (e.g. --ignore-names="example,log_bucket") or set the flag multiple times.
  -i  --ignore-paths     Regular expressions for path to ignore
                         Set the flag with values separated by commas or set the flag multiple times.
  -p  --provider-version The provider version constraint (default: v4.0.0)
  -r  --recursive        Check a directory recursively (default: false)
```
```shell
$ cat main.tf
provider "aws" {
  version = "2.39.0"
}

resource "aws_s3_bucket" "example" {
  bucket = var.bucket
  acl    = "private"
  
  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.arbitrary.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }
}

$ tfmigrate resource aws_s3_bucket main.tf

$ cat main_migrated.tf
provider "aws" {
  version = "4.0.0"
}

resource "aws_s3_bucket" "example" {
  bucket = var.bucket
}

resource "aws_s3_bucket_acl" "example_acl" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "example_server_side_encryption_configuration" {
  bucket = aws_s3_bucket.example.id
  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.arbitrary.arn
      sse_algorithm     = "aws:kms"
    }
  }
}
```

## Output Logging

Set the environment variable `TFMIGRATE_LOG` to the log-level of choice. Valid values include: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`.
