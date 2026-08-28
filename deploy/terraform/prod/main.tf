locals {
  tags = {
    Project     = var.project
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# ── AWS probe: ap-southeast-1 (Singapore) ───────────────────────────────────
# Sole remaining AWS resource: the platform main server moved to the operator's
# OVH host (15.235.186.156) on 2026-08-28; this probe node still runs the
# live prober until the replacement prober is provisioned there.
module "probe_ap_southeast_1" {
  source = "../modules/aws-probe"

  providers = {
    aws = aws.ap_southeast_1
  }

  project        = var.project
  environment    = var.environment
  node_name      = "ap-southeast-1"
  ssh_public_key = var.ssh_public_key
  tags           = local.tags
}
