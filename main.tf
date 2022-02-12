terraform {
  required_providers {
    aws = {
      version = "3.74.0"
    }
  }
}

provider "aws" {}

provider "random" {}

variable "bucket_name" {
  default = "tf-acc-test-abc12345"
  type    = string
}

data "aws_canonical_user_id" "current" {}

data "aws_partition" "current" {}

resource "random_pet" "example" {}

resource "aws_s3_bucket" "b" {
  bucket = random_pet.example.id
  acl    = "public-read"
}

resource "aws_s3_bucket" "log_bucket" {
  bucket = "my-example-log-bucket-44444"
  acl    = "private"
}

resource "aws_s3_bucket" "example" {
  bucket              = var.bucket_name
  acceleration_status = "Enabled"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET"]
    allowed_origins = [""]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }

  grant {
    id          = data.aws_canonical_user_id.current.id
    type        = "CanonicalUser"
    permissions = ["WRITE", "FULL_CONTROL"]
  }

  grant {
    type        = "Group"
    permissions = ["READ_ACP"]
    uri         = "http://acs.amazonaws.com/groups/s3/LogDelivery"
  }

  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "log/"
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"

    rule {
      default_retention {
        mode = "COMPLIANCE"
        days = 3
      }
    }
  }

  policy = <<POLICY
  {
    "Version":"2008-10-17",
    "Statement": [
      {
        "Sid":"AllowPublicRead",
        "Effect":"Allow",
        "Principal": {
          "AWS": "*"
        },
        "Action": "s3:GetObject",
        "Resource": "arn:${data.aws_partition.current.partition}:s3:::${var.bucket_name}/*"
      }
    ]
  }
  POLICY

  request_payer = "Requester"

  versioning {
    enabled = true
  }

  website {
    index_document = "index.html"
    error_document = "error.html"
    routing_rules  = <<EOF
  [
    {
      "Condition": {
        "KeyPrefixEquals": "docs/"
      },
      "Redirect": {
        "ReplaceKeyPrefixWith": "documents/"
      }
    }
  ]
  EOF
  }
}