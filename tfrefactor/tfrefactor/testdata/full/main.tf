terraform {
  required_providers {
    aws = {
      version = "3.74.0"
    }
  }
}

terraform {
  required_providers {
    aws = "3.74.0"
  }
}

provider "aws" {
  version = "3.74.0"
}

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

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.arbitrary.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }
}

resource "aws_s3_bucket" "log_bucket" {
  bucket = "my-example-log-bucket-44444"
  acl    = "private"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
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

  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = true
    expiration {
      date = "2016-01-12"
    }
  }

  lifecycle_rule {
    id      = "id5"
    enabled = true
    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }
    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id6"
    enabled = true
    tags = {
      "tagKey" = "tagValue"
    }
    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = false
    noncurrent_version_expiration {
      days = 365
    }
  }

  lifecycle_rule {
    id      = "id3"
    prefix  = "path3/"
    enabled = true
    noncurrent_version_transition {
      days          = 0
      storage_class = "GLACIER"
    }
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

  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"
      filter {
        prefix = "prefix1"
      }

      delete_marker_replication_status = "Enabled"

      destination {
        access_control_translation {
          owner = "Destination"
        }

        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"

        metrics {}

        replication_time {
          status  = "Enabled"
          minutes = 15
        }
      }

      source_selection_criteria {
        sse_kms_encrypted_objects {
          enabled = true
        }
      }
    }
    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"
      filter {
        tags = {
          Key2 = "Value2"
        }
      }
      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"

        metrics {
          status  = "Enabled"
          minutes = 15
        }

        replication_time {
          status  = "Enabled"
          minutes = 15
        }
      }
    }
  }

  request_payer = "Requester"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.arbitrary.arn
        sse_algorithm     = "aws:kms"
      }
      bucket_key_enabled = true
    }
  }

  versioning {
    enabled = true
  }

  website {
    index_document = "index.html"
    error_document = "error.html"
    routing_rules = jsonencode([{
      Condition : {
        KeyPrefixEquals : "docs/"
      },
      Redirect : {
        ReplaceKeyPrefixWith : "documents/"
      }
    }])
  }
}