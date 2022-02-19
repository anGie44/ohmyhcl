resource "aws_s3_bucket" "log_bucket" {
  bucket              = "my-example-log-bucket-44444"
  acceleration_status = "Enabled"
  acl                 = "public-read"
}