resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

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
}