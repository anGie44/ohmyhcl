resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"
}
resource "aws_s3_bucket_request_payment_configuration" "example_request_payment_configuration" {
  bucket = aws_s3_bucket.example.id
  payer  = "Requester"
}
