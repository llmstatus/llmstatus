variable "project" {
  type    = string
  default = "llmstatus"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "ssh_public_key" {
  description = "SSH public key content"
  type        = string
}

variable "data_volume_size_gb" {
  description = "EBS data volume size in GB (PostgreSQL + InfluxDB)"
  type        = number
  default     = 80
}

variable "tags" {
  type    = map(string)
  default = {}
}
