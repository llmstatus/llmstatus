output "probe_ap_southeast_1_ip" {
  description = "Probe node: ap-southeast-1 (Singapore)"
  value       = module.probe_ap_southeast_1.public_ip
}

output "ansible_inventory_hint" {
  description = "Node IPs for deploy/ansible/inventories/prod/hosts.yml"
  value = {
    # Main server moved off AWS to the operator's OVH host (15.235.186.156)
    # on 2026-08-28 — see deploy/ansible/inventories/prod/hosts.yml.
    main           = "15.235.186.156"
    ap_southeast_1 = module.probe_ap_southeast_1.public_ip
  }
}
