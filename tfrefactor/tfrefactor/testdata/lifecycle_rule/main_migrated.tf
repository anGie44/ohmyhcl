resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"





}
resource "aws_s3_bucket_lifecycle_configuration" "example_lifecycle_configuration" {
  bucket = aws_s3_bucket.example.id
  rule {
    id     = "id2"
    status = "Enabled"
    filter {
      prefix = "path2/"
    }
    expiration {
      date = "2016-01-12"
    }
  }
  rule {
    id     = "id5"
    status = "Enabled"
    filter {
      and {
        tags = {
          "tagKey"    = "tagValue"
          "terraform" = "hashicorp"
        }
        prefix = ""
      }
    }
    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }
  rule {
    id     = "id6"
    status = "Enabled"
    filter {
      and {
        tags = {
          "tagKey" = "tagValue"
        }
        prefix = ""
      }
    }
    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }
  rule {
    status = "Disabled"
    id     = "id2"
    filter {
      prefix = "path2/"
    }
    noncurrent_version_expiration {
      noncurrent_days = 365
    }
  }
  rule {
    id     = "id3"
    status = "Enabled"
    filter {
      prefix = "path3/"
    }
    noncurrent_version_transition {
      storage_class   = "GLACIER"
      noncurrent_days = 0
    }
  }
}
