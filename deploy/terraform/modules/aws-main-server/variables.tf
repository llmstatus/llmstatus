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

variable "probe_node_cidrs" {
  description = "CIDR blocks for probe nodes allowed to reach Postgres port 15432"
  type        = list(string)
  default     = []
}

variable "tags" {
  type    = map(string)
  default = {}
}
