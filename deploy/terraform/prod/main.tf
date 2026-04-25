locals {
  tags = {
    Project     = var.project
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# ── Main server (us-west-2) ─────────────────────────────────────────────────
# Runs: PostgreSQL, InfluxDB, Go API, ingest, detector, Next.js frontend.
# Also acts as the us-west-2 probe node (prober process runs locally).
module "main_server" {
  source = "../modules/aws-main-server"

  project             = var.project
  environment         = var.environment
  ssh_public_key      = var.ssh_public_key
  data_volume_size_gb = 80
  probe_node_cidrs = [
    "${module.probe_us_east_1.public_ip}/32",
    "${module.probe_ap_northeast_1.public_ip}/32",
    "${module.probe_ap_southeast_1.public_ip}/32",
    "${module.probe_eu_central_1.public_ip}/32",
  ]
  tags = local.tags
}

# ── AWS probe: us-east-1 (near Google / AWS Bedrock) ────────────────────────
module "probe_us_east_1" {
  source = "../modules/aws-probe"

  providers = {
    aws = aws.us_east_1
  }

  project        = var.project
  environment    = var.environment
  node_name      = "us-east-1"
  ssh_public_key = var.ssh_public_key
  tags           = local.tags
}

# ── AWS probe: ap-northeast-1 (Tokyo) ───────────────────────────────────────
module "probe_ap_northeast_1" {
  source = "../modules/aws-probe"

  providers = {
    aws = aws.ap_northeast_1
  }

  project        = var.project
  environment    = var.environment
  node_name      = "ap-northeast-1"
  ssh_public_key = var.ssh_public_key
  tags           = local.tags
}

# ── AWS probe: ap-southeast-1 (Singapore) ───────────────────────────────────
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

# ── AWS probe: eu-central-1 (Frankfurt, EU) ─────────────────────────────────
module "probe_eu_central_1" {
  source = "../modules/aws-probe"

  providers = {
    aws = aws.eu_central_1
  }

  project        = var.project
  environment    = var.environment
  node_name      = "eu-central-1"
  ssh_public_key = var.ssh_public_key
  tags           = local.tags
}
