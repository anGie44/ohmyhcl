resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"
}
resource "aws_s3_bucket_acl" "example_acl" {
  bucket = aws_s3_bucket.example.id
  acl    = "public-read"
}
