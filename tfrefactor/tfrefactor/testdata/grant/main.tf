resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

  grant {
    id          = data.aws_canonical_user_id.current.id
    type        = "CanonicalUser"
    permissions = ["WRITE", "FULL_CONTROL"]
  }

  grant {
    type        = "Group"
    permissions = ["READ_ACP"]
    uri         = "http://acs.amazonaws.com/groups/s3/LogDelivery"
  }
}