resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"
}
resource "aws_s3_bucket_accelerate_configuration" "example_accelerate_configuration" {
  bucket = aws_s3_bucket.example.id
  status = "Enabled"
}
