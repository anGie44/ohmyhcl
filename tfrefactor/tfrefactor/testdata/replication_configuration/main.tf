resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"
      filter {
        prefix = "prefix1"
      }

      delete_marker_replication_status = "Enabled"

      destination {
        access_control_translation {
          owner = "Destination"
        }

        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"

        metrics {}

        replication_time {
          status  = "Enabled"
          minutes = 15
        }
      }

      source_selection_criteria {
        sse_kms_encrypted_objects {
          enabled = true
        }
      }
    }
  }
}