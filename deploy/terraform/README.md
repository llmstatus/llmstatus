# Terraform — llmstatus.io production infrastructure

Provisions 1 main server (AWS us-west-2) + 3 AWS probe nodes + 1 Azure probe node.

## Prerequisites

- Terraform >= 1.9
- AWS CLI configured (`aws configure`) with credentials that have EC2/S3/DynamoDB permissions
- Azure CLI logged in (`az login`) with an active subscription

## First-time setup

### 1. Bootstrap remote state (once only)

```bash
cd deploy/terraform/bootstrap
terraform init
terraform apply
```

This creates the S3 bucket and DynamoDB lock table used by all subsequent runs.

### 2. Configure variables

```bash
cd deploy/terraform/prod
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars — add your SSH public key
```

Generate a key if needed:
```bash
ssh-keygen -t ed25519 -f ~/.ssh/llmstatus-prod -C "llmstatus-prod"
```

### 3. Init and apply

```bash
cd deploy/terraform/prod
terraform init
terraform plan
terraform apply
```

### 4. Note the outputs

```
terraform output
```

- `main_server_ip` → add as Cloudflare A record for `llmstatus.io` + `www` (proxy OFF)
- `ansible_inventory_hint` → paste IPs into `deploy/ansible/inventories/prod/hosts.yml`

## Directory layout

```
bootstrap/      one-time S3 + DynamoDB for TF state
modules/
  aws-main-server/   EC2 t3.small + 80GB EBS + EIP (us-west-2)
  aws-probe/         EC2 t3.micro + EIP (reused for 3 regions)
  azure-probe/       Azure VM Standard_B1s + static IP (Germany West Central)
prod/           root module — instantiates all modules
```

## Cost estimate (~$61/mo)

| Node | Type | Region | $/mo |
|---|---|---|---|
| Main server | EC2 t3.small + 80GB EBS | us-west-2 | ~$23 |
| Probe | EC2 t3.micro | us-east-1 | ~$8 |
| Probe | EC2 t3.micro | ap-northeast-1 | ~$9 |
| Probe | EC2 t3.micro | ap-southeast-1 | ~$9 |
| Probe EU | Azure Standard_B1s + SSD | Germany West Central | ~$12 |

China nodes (Aliyun Shanghai + Tencent Guangzhou) are provisioned separately.
