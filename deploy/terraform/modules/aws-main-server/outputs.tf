output "public_ip" {
  description = "Elastic IP of the main server"
  value       = aws_eip.this.public_ip
}

output "instance_id" {
  value = aws_instance.this.id
}

output "availability_zone" {
  value = aws_instance.this.availability_zone
}
