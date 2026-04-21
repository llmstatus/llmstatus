output "main_server_ip" {
  description = "Main server Elastic IP — set as Cloudflare A record for llmstatus.io + www"
  value       = module.main_server.public_ip
}

output "probe_us_east_1_ip" {
  description = "Probe node: us-east-1"
  value       = module.probe_us_east_1.public_ip
}

output "probe_ap_northeast_1_ip" {
  description = "Probe node: ap-northeast-1 (Tokyo)"
  value       = module.probe_ap_northeast_1.public_ip
}

output "probe_ap_southeast_1_ip" {
  description = "Probe node: ap-southeast-1 (Singapore)"
  value       = module.probe_ap_southeast_1.public_ip
}

output "probe_eu_ip" {
  description = "Probe node: Germany West Central (EU)"
  value       = module.probe_eu.public_ip
}

output "cloudflare_records" {
  description = "Add these A records in Cloudflare (proxy OFF for direct IP access)"
  value = {
    "llmstatus.io"     = module.main_server.public_ip
    "www.llmstatus.io" = module.main_server.public_ip
  }
}

output "ansible_inventory_hint" {
  description = "Node IPs for deploy/ansible/inventories/prod/hosts.yml"
  value = {
    main        = module.main_server.public_ip
    us_east_1   = module.probe_us_east_1.public_ip
    ap_ne_1     = module.probe_ap_northeast_1.public_ip
    ap_se_1     = module.probe_ap_southeast_1.public_ip
    eu_west     = module.probe_eu.public_ip
  }
}
