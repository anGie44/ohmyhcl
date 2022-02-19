resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

}
resource "aws_s3_bucket_replication_configuration" "example_replication_configuration" {
  bucket = aws_s3_bucket.example.id
  role   = aws_iam_role.role.arn
  rule {
    id       = "rule1"
    priority = 1
    status   = "Enabled"
    delete_marker_replication {
      status = "Enabled"
    }
    filter {
      prefix = "prefix1"
    }
    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
      access_control_translation {
        owner = "Destination"
      }
      metrics {
      }
      replication_time {
        status = "Enabled"
        time {
          minutes = 15
        }
      }
    }
    source_selection_criteria {
      sse_kms_encrypted_objects {
        status = "Enabled"
      }
    }
  }
}
