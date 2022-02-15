resource "aws_s3_bucket" "this" {
  count = var.create_bucket ? 1 : 0

  bucket        = var.bucket
  bucket_prefix = var.bucket_prefix


  tags          = var.tags
  force_destroy = var.force_destroy

}
resource "aws_s3_bucket_accelerate_configuration" "this_accelerate_configuration" {
  bucket = aws_s3_bucket.this.id
  status = var.acceleration_status
}

resource "aws_s3_bucket_request_payment_configuration" "this_request_payment_configuration" {
  bucket = aws_s3_bucket.this.id
  payer  = var.request_payer
}

resource "aws_s3_bucket_acl" "this_acl" {
  bucket = aws_s3_bucket.this.id
  acl    = var.acl != "null" ? var.acl : null
}

resource "aws_s3_bucket_website_configuration" "this_website_configuration" {
  for_each = length(keys(var.website)) == 0 ? [] : [var.website]

  bucket = aws_s3_bucket.this[each.key].id
  redirect_all_requests_to {
    host_name = lookup(each.value, "redirect_all_requests_to", null)
  }
  # TODO: Replace with your 'routing_rule' configuration
  index_document {
    suffix = lookup(each.value, "index_document", null)
  }
  error_document {
    key = lookup(each.value, "error_document", null)
  }
}
