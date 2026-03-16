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
    name = "recruiter_email"
    type = "S"
  }

  attribute {
    name = "date_year"
    type = "S"
  }

  attribute {
    name = "date_day"
    type = "S"
  }

  # GSI 1: Filter by recruiter email
  global_secondary_index {
    name            = "recruiter-index"
    projection_type = "ALL"
    read_capacity   = var.gsi_read_capacity
    write_capacity  = var.gsi_write_capacity

    key_schema {
      attribute_name = "recruiter_email"
      key_type       = "HASH"
    }

    key_schema {
      attribute_name = "received_at"
      key_type       = "RANGE"
    }
  }

  # GSI 2: Filter by date — supports year, month, day, and range queries.
  # Hash key: date_year (e.g. "2026") for natural partitioning.
  # Range key: date_day (e.g. "2026-03-15") for begins_with / BETWEEN queries.
  global_secondary_index {
    name            = "date-index"
    projection_type = "ALL"
    read_capacity   = var.gsi_read_capacity
    write_capacity  = var.gsi_write_capacity

    key_schema {
      attribute_name = "date_year"
      key_type       = "HASH"
    }

    key_schema {
      attribute_name = "date_day"
      key_type       = "RANGE"
    }
  }

  deletion_protection_enabled = true

  point_in_time_recovery {
    enabled = false
  }

  server_side_encryption {
    enabled = true
  }
}
