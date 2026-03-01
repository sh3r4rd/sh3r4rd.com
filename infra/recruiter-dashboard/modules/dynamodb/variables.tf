variable "table_name" {
  description = "Name of the DynamoDB table."
  type        = string
}

variable "read_capacity" {
  description = "Provisioned read capacity units for the table."
  type        = number
  default     = 5
}

variable "write_capacity" {
  description = "Provisioned write capacity units for the table."
  type        = number
  default     = 5
}

variable "gsi_read_capacity" {
  description = "Provisioned read capacity units for each GSI."
  type        = number
  default     = 5
}

variable "gsi_write_capacity" {
  description = "Provisioned write capacity units for each GSI."
  type        = number
  default     = 5
}
