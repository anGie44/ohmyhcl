resource "aws_s3_bucket" "example" {
  bucket              = "my-example-bucket"
  acceleration_status = "Enabled"
}