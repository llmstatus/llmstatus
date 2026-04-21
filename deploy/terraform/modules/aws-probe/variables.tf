variable "project" {
  type    = string
  default = "llmstatus"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "node_name" {
  description = "Short name for this probe node, e.g. us-east-1, ap-northeast-1"
  type        = string
}

variable "ssh_public_key" {
  description = "SSH public key content"
  type        = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
