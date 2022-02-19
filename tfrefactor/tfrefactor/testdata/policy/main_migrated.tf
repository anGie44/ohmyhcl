resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"

}
resource "aws_s3_bucket_policy" "example_policy" {
  bucket = aws_s3_bucket.example.id
  policy = <<POLICY
  {
    "Version":"2008-10-17",
    "Statement": [
      {
        "Sid":"AllowPublicRead",
        "Effect":"Allow",
        "Principal": {
          "AWS": "*"
        },
        "Action": "s3:GetObject",
        "Resource": "arn:${data.aws_partition.current.partition}:s3:::${var.bucket_name}/*"
      }
    ]
  }
  POLICY
}
