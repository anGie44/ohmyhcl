resource "aws_s3_bucket" "log_bucket" {
  bucket = "my-example-log-bucket-44444"
}
resource "aws_s3_bucket_accelerate_configuration" "log_bucket_accelerate_configuration" {
  bucket = aws_s3_bucket.log_bucket.id
  status = "Enabled"
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  bucket = aws_s3_bucket.log_bucket.id
  acl    = "public-read"
}
