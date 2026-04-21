variable "project" {
  type    = string
  default = "llmstatus"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "node_name" {
  description = "Short name for this probe node"
  type        = string
  default     = "eu-west"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "Germany West Central"
}

variable "vm_size" {
  type    = string
  default = "Standard_B1s"
}

variable "ssh_public_key" {
  description = "SSH public key content"
  type        = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
