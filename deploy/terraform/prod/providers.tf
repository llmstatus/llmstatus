# AWS — default provider (no alias) is us-west-2, used by the main server.
provider "aws" {
  region = "us-west-2"
}

provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}

provider "aws" {
  alias  = "ap_northeast_1"
  region = "ap-northeast-1"
}

provider "aws" {
  alias  = "ap_southeast_1"
  region = "ap-southeast-1"
}

# Azure — authenticates via `az login` or ARM_* env vars.
provider "azurerm" {
  features {}
}
