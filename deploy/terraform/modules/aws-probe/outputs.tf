output "public_ip" {
  description = "Elastic IP of this probe node"
  value       = aws_eip.this.public_ip
}

output "instance_id" {
  value = aws_instance.this.id
}
