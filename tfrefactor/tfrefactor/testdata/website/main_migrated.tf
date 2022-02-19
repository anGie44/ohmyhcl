resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

}
resource "aws_s3_bucket_website_configuration" "example_website_configuration" {
  bucket = aws_s3_bucket.example.id
  routing_rule {
    condition {
      key_prefix_equals = "docs/"
    }
    redirect {
      replace_key_prefix_with = "documents/"
    }
  }
  index_document {
    suffix = "index.html"
  }
  error_document {
    key = "error.html"
  }
}
