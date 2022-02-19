resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "log/"
  }
}