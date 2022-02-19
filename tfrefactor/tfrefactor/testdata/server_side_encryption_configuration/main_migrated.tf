resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

}
resource "aws_s3_bucket_server_side_encryption_configuration" "example_server_side_encryption_configuration" {
  bucket = aws_s3_bucket.example.id
  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.arbitrary.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}
