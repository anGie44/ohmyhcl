resource "aws_s3_bucket" "example" {
  bucket        = "my-example-bucket"
  request_payer = "Requester"
}