# Remote state in S3. Run bootstrap/ first to create this bucket.
terraform {
  backend "s3" {
    bucket         = "llmstatus-tfstate-prod"
    key            = "prod/terraform.tfstate"
    region         = "us-west-2"
    dynamodb_table = "llmstatus-tfstate-lock"
    encrypt        = true
  }
}
