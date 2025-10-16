# Example: Strong Account with AWS IAM Authentication
# This example demonstrates creating a strong account using AWS IAM authentication
# for an Amazon RDS MySQL database target with IAM authentication enabled.

# Provider configuration
terraform {
  required_providers {
    cyberark_sia = {
      source = "local/cyberark-sia"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "cyberark_sia" {
  # Authentication credentials should be set via environment variables:
  # export CYBERARK_CLIENT_ID="your-client-id"
  # export CYBERARK_CLIENT_SECRET="your-client-secret"
  # export CYBERARK_TENANT_SUBDOMAIN="your-subdomain"
  # Optional (for GovCloud or custom deployments):
  # export CYBERARK_IDENTITY_URL="https://your-tenant.cyberarkgov.cloud"
}

provider "aws" {
  region = "us-east-1"
}

# Get AWS account ID for reference
data "aws_caller_identity" "current" {}

# Database target for RDS MySQL with IAM authentication
resource "cyberark_sia_database_workspace" "rds_mysql" {
  name                  = "production-rds-mysql"
  database_type         = "mysql"
  address               = aws_db_instance.mysql.endpoint
  port                  = 3306
  authentication_method = "rds_iam_authentication"
  cloud_provider        = "aws"
  region                = "us-east-1" # Required for RDS IAM auth
  description           = "Production RDS MySQL database with IAM authentication"

  tags = {
    environment = "production"
    team        = "platform"
    auth_type   = "iam"
  }

  depends_on = [aws_db_instance.mysql]
}

# Strong account with AWS IAM authentication
# Note: This uses an IAM user with database access permissions
resource "cyberark_sia_secret" "rds_iam_user" {
  name               = "rds-mysql-iam-account"
  # NOTE: Secrets are standalone - no workspace_id needed = cyberark_sia_database_workspace.rds_mysql.id
  authentication_type = "aws_iam"

  # AWS IAM credentials
  aws_access_key_id     = var.iam_access_key_id     # Sensitive - should be stored securely
  aws_secret_access_key = var.iam_secret_access_key # Sensitive - should be stored securely

  description = "IAM user account for RDS MySQL ephemeral access provisioning"

  # Rotation settings
  # Note: AWS IAM key rotation must be coordinated with IAM policies
  rotation_enabled      = false
  rotation_interval_days = 90

  tags = {
    purpose       = "ephemeral-access"
    managed_by    = "terraform"
    cloud         = "aws"
    auth_method   = "iam"
  }
}

# Example RDS MySQL instance with IAM authentication enabled
resource "aws_db_instance" "mysql" {
  identifier     = "production-mysql"
  engine         = "mysql"
  engine_version = "8.0"
  instance_class = "db.t3.micro"

  allocated_storage = 20
  storage_type      = "gp2"

  username = "admin"
  password = var.db_master_password

  # Enable IAM database authentication
  iam_database_authentication_enabled = true

  vpc_security_group_ids = [aws_security_group.rds.id]
  db_subnet_group_name   = aws_db_subnet_group.rds.name

  skip_final_snapshot = true

  tags = {
    Name        = "production-mysql"
    Environment = "production"
    ManagedBy   = "terraform"
  }
}

# Security group for RDS (example)
resource "aws_security_group" "rds" {
  name_prefix = "rds-mysql-"
  description = "Security group for RDS MySQL instance"

  # Add appropriate ingress/egress rules

  tags = {
    Name = "rds-mysql-sg"
  }
}

# DB subnet group (example)
resource "aws_db_subnet_group" "rds" {
  name       = "rds-mysql-subnet-group"
  subnet_ids = var.subnet_ids # Should reference VPC subnets

  tags = {
    Name = "rds-mysql-subnet-group"
  }
}

# Variables for sensitive data
variable "iam_access_key_id" {
  description = "AWS IAM access key ID for database authentication"
  type        = string
  sensitive   = true
}

variable "iam_secret_access_key" {
  description = "AWS IAM secret access key for database authentication"
  type        = string
  sensitive   = true
}

variable "db_master_password" {
  description = "Master password for RDS database"
  type        = string
  sensitive   = true
}

variable "subnet_ids" {
  description = "List of subnet IDs for RDS subnet group"
  type        = list(string)
}

# Outputs (safe - does not expose sensitive data)
output "strong_account_id" {
  description = "ID of the created strong account"
  value       = cyberark_sia_secret.rds_iam_user.id
}

output "# NOTE: Secrets are standalone - no workspace_id needed" {
  description = "ID of the database target"
  value       = cyberark_sia_database_workspace.rds_mysql.id
}

output "rds_endpoint" {
  description = "RDS MySQL endpoint"
  value       = aws_db_instance.mysql.endpoint
}
