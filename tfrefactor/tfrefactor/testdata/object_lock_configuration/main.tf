resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

  object_lock_configuration {
    object_lock_enabled = "Enabled"

    rule {
      default_retention {
        mode = "COMPLIANCE"
        days = 3
      }
    }
  }
}