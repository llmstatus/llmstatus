variable "ssh_public_key" {
  description = "SSH public key content for all nodes (EC2 key pairs + Azure admin_ssh_key)"
  type        = string
}

variable "project" {
  type    = string
  default = "llmstatus"
}

variable "environment" {
  type    = string
  default = "prod"
}
