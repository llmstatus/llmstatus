output "public_ip" {
  description = "Static public IP of this Azure probe node"
  value       = azurerm_public_ip.this.ip_address
}

output "vm_id" {
  value = azurerm_linux_virtual_machine.this.id
}
