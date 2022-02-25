[![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)](https://github.com/emersion/stability-badges#experimental)

# tfrefactor

## Features

- Migrate `aws_s3_bucket` resource arguments to independent resources available since `v4.0.0` of the Terraform AWS Provider.
- Update version constraints of the Terraform AWS Provider defined in configurations.
- Get a table (in `.csv` format) of each new resource with its parent `aws_s3_bucket` to enable resource import.

## Limitations

- Migrating `dynamic` arguments. This is done as a _best-effort_ attempt. Current _best_effort_ support available is for the `cors_rule`, `logging`, and `website` arguments.
- Migrating `aws_s3_bucket` `routing_rules` (String) to `aws_s3_bucket_website_configuration` `routing_rule` configuration blocks
if the given literal value is not a JSON or YAML representation of RoutingRules. 

For example, given the following configuration:
```shell
$ cat main.tf

resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  dynamic "website" {
    for_each = length(keys(var.website)) == 0 ? [] : [var.website]

    content {
      index_document           = lookup(website.value, "index_document", null)
      error_document           = lookup(website.value, "error_document", null)
      redirect_all_requests_to = lookup(website.value, "redirect_all_requests_to", null)
      routing_rules            = lookup(website.value, "routing_rules", null)
    }
  }
}
```

The `aws_s3_bucket_website_configuration` resource in `main_migrated.tf` will look like:
```terraform
resource "aws_s3_bucket_website_configuration" "this_website_configuration" {
  for_each = length(keys(var.website)) == 0 ? [] : [var.website]

  bucket = aws_s3_bucket.this[each.key].id
  # TODO: Replace with your 'routing_rule' configuration
  index_document {
    suffix = lookup(each.value, "index_document", null)
  }
  error_document {
    key = lookup(each.value, "error_document", null)
  }
  redirect_all_requests_to {
    host_name = lookup(each.value, "redirect_all_requests_to", null)
  }
}
```

### Download

Download the latest compiled binaries and put it anywhere in your executable path.

https://github.com/anGie44/ohmyhcl/releases

### Source

If you have Go 1.17+ development environment:

```
$ git clone https://github.com/anGie44/ohmyhcl
$ cd ohmyhcl/
$ make install
```

## Usage
```shell
tfrefactor --help
Usage: tfrefactor [--version] [--help] <command> [<args>]

Available commands are:
    resource    Migrate resource arguments to individual resources
```

### resource

```shell
$ tfrefactor resource --help
Usage: tfrefactor resource <RESOURCE_TYPE> <PATH> [options]
Arguments
  RESOURCE_TYPE      The provider resource type (e.g. aws_s3_bucket)
  PATH               A path of file or directory to update
Options:
  --ignore-arguments       The arguments in the <RESOURCE_TYPE> to ignore
                           Set the flag with values separated by commas (e.g. --ignore-arguments="acl,grant") or set the flag multiple times.
  --ignore-names           The resource names of <RESOURCE_TYPE> to ignore
                           Set the flag with values separated by commas (e.g. --ignore-names="example,log_bucket") or set the flag multiple times.
  -i  --ignore-paths       Regular expressions for path to ignore
                           Set the flag with values separated by commas or set the flag multiple times.
  -c  --csv    			   Generate a CSV file of new resources and their parent resource (default: false)
  -p  --provider-version   The provider version constraint (default: v4.0.0)
  -r  --recursive          Check a directory recursively (default: false)
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

$ tfrefactor resource aws_s3_bucket main.tf

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

Set the environment variable `TFREFACTOR_LOG` to the log-level of choice. Valid values include: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`.

## Credit

Credit is due to [@minamijoyo](https://github.com/minamijoyo) (and community contributors) and their [`tfupdate`](https://github.com/minamijoyo/tfupdate) and [`hcledit`](https://github.com/minamijoyo/hcledit) projects,
as `tfrefactor` is in essence an extension to that functionality and much of the foundation and CLI implementation of this tool takes from their existing code and patterns. 
The goal of this project was to develop a tool, quickly and with a familiar UX, and the existing projects made it come to fruition sooner than expected.
