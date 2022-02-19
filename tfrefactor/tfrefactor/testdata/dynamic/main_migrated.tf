resource "aws_s3_bucket" "a" {
  count = var.create_bucket ? 1 : 0

  bucket        = var.bucket
  bucket_prefix = var.bucket_prefix

}

resource "aws_s3_bucket" "b" {
  count = var.create_bucket ? 1 : 0

  bucket        = var.bucket
  bucket_prefix = var.bucket_prefix

}

resource "aws_s3_bucket" "c" {
  count = var.create_bucket ? 1 : 0

  bucket        = var.bucket
  bucket_prefix = var.bucket_prefix

}


resource "aws_s3_bucket_website_configuration" "a_website_configuration" {
  for_each = length(keys(var.website)) == 0 ? [] : [var.website]

  bucket = aws_s3_bucket.a[each.key].id
  error_document {
    key = lookup(each.value, "error_document", null)
  }
  redirect_all_requests_to {
    host_name = lookup(each.value, "redirect_all_requests_to", null)
  }
  # TODO: Replace with your 'routing_rule' configuration
  index_document {
    suffix = lookup(each.value, "index_document", null)
  }
}

resource "aws_s3_bucket_logging" "b_logging" {
  for_each = length(keys(var.logging)) == 0 ? [] : [var.logging]

  bucket        = aws_s3_bucket.b[each.key].id
  target_bucket = each.value.target_bucket
  target_prefix = lookup(each.value, "target_prefix", null)
}

resource "aws_s3_bucket_cors_configuration" "c_cors_configuration" {
  # TODO: Replace with your intended 'for_each' value
  # for_each = try(jsondecode(var.cors_rule), var.cors_rule)

  # TODO: Replace 'bucket' argument value with correct instance index
  # e.g. aws_s3_bucket.c[count.index].id
  bucket = aws_s3_bucket.c.id
  dynamic "cors_rule" {
    for_each = try(jsondecode(var.cors_rule), var.cors_rule)

    content {
      allowed_methods = each.value.allowed_methods
      allowed_origins = each.value.allowed_origins
      allowed_headers = lookup(each.value, "allowed_headers", null)
      expose_headers  = lookup(each.value, "expose_headers", null)
      max_age_seconds = lookup(each.value, "max_age_seconds", null)
    }
  }
}
