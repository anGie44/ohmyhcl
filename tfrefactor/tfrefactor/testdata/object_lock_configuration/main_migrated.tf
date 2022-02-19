resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

}
resource "aws_s3_bucket_object_lock_configuration" "example_object_lock_configuration" {
  bucket              = aws_s3_bucket.example.id
  object_lock_enabled = "Enabled"
  rule {
    default_retention {
      mode = "COMPLIANCE"
      days = 3
    }
  }
}
