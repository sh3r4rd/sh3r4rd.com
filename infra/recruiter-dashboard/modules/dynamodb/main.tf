# COST WARNING: Using PROVISIONED billing mode (not PAY_PER_REQUEST).
# On-demand has NO free tier for requests.
# Provisioned 25 RCU/WCU is always free (not time-limited).
# Total capacity: table (5/5) + GSI1 (5/5) + GSI2 (5/5) = 15/15 (under 25 limit).

resource "aws_dynamodb_table" "recruiter_emails" {
  name         = var.table_name
  billing_mode = "PROVISIONED"

  read_capacity  = var.read_capacity
  write_capacity = var.write_capacity

  hash_key  = "id"
  range_key = "received_at"

  # Primary key attributes
  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "received_at"
    type = "S"
  }

  # GSI key attributes
  attribute {
    name = "company"
    type = "S"
  }

  attribute {
    name = "date_month"
    type = "S"
  }

  # GSI 1: Filter by company
  global_secondary_index {
    name            = "company-index"
    hash_key        = "company"
    range_key       = "received_at"
    projection_type = "ALL"
    read_capacity   = var.gsi_read_capacity
    write_capacity  = var.gsi_write_capacity
  }

  # GSI 2: Filter by month
  global_secondary_index {
    name            = "date-index"
    hash_key        = "date_month"
    range_key       = "received_at"
    projection_type = "ALL"
    read_capacity   = var.gsi_read_capacity
    write_capacity  = var.gsi_write_capacity
  }

  deletion_protection_enabled = true

  point_in_time_recovery {
    enabled = false
  }

  server_side_encryption {
    enabled = true
  }
}
