# Example: Onboard AWS RDS PostgreSQL database to CyberArk SIA
#
# This example demonstrates:
# - Creating an RDS PostgreSQL instance using AWS provider
# - Registering the database with SIA using database_target resource
# - Using Terraform references to avoid hardcoding values
# - AWS-specific configuration (region, account_id)

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    cyberarksia = {
      source  = "aaearon/cyberarksia"
      version = "~> 1.0"
    }
  }
}

# Configure AWS Provider
provider "aws" {
  region = "us-east-1"
}

# Configure CyberArk SIA Provider
provider "cyberarksia" {
  username      = var.cyberark_username
  client_secret = var.cyberark_client_secret

  # Optional: Override identity URL for GovCloud or custom deployments
  # identity_url = var.cyberark_identity_url
}

# Get current AWS account information
data "aws_caller_identity" "current" {}

# Example: Reference existing RDS instance (created elsewhere)
data "aws_db_instance" "existing" {
  db_instance_identifier = "production-postgres"
}

# OR: Create new RDS instance
resource "aws_db_instance" "postgres" {
  identifier     = "sia-managed-postgres"
  engine         = "postgres"
  engine_version = "14.7"
  instance_class = "db.t3.micro"

  allocated_storage = 20
  storage_type      = "gp2"
  storage_encrypted = true

  db_name  = "appdb"
  username = "dbadmin"
  password = var.db_master_password # Sensitive

  # Enable IAM authentication for enhanced security
  iam_database_authentication_enabled = false # Set to true for AWS IAM auth

  # Network configuration
  db_subnet_group_name   = aws_db_subnet_group.postgres.name
  vpc_security_group_ids = [aws_security_group.postgres.id]
  publicly_accessible    = false

  # Backup and maintenance
  backup_retention_period = 7
  skip_final_snapshot     = true # Set to false in production

  tags = {
    Environment = "production"
    ManagedBy   = "Terraform"
    SIAManaged  = "true"
  }
}

# Create secret for database access
resource "cyberarksia_secret" "postgres_admin" {
  name                = "postgres-admin-secret"
  authentication_type = "local"

  username = "dbadmin"
  password = var.db_master_password

  rotation_enabled       = false
  rotation_interval_days = 90
}

# Register the RDS instance with CyberArk SIA
resource "cyberarksia_database_workspace" "postgres" {
  # Database name on the server (actual database that SIA connects to)
  name          = aws_db_instance.postgres.db_name # "appdb"
  database_type = "postgres-aws-rds"

  # Use Terraform reference to get endpoint dynamically
  address = aws_db_instance.postgres.address
  port    = aws_db_instance.postgres.port

  # Required: Reference to secret for ephemeral access provisioning
  secret_id = cyberarksia_secret.postgres_admin.id

  # Authentication method
  authentication_method = "local_ephemeral_user" # Use "rds_iam_authentication" for RDS IAM auth

  # AWS-specific metadata
  cloud_provider = "aws"
  region         = "us-east-1"

  # Ensure RDS instance and secret are created before SIA registration
  depends_on = [
    aws_db_instance.postgres,
    cyberarksia_secret.postgres_admin
  ]
}

# Supporting resources (simplified for example)
resource "aws_db_subnet_group" "postgres" {
  name       = "postgres-subnet-group"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "PostgreSQL DB subnet group"
  }
}

resource "aws_security_group" "postgres" {
  name        = "postgres-sg"
  description = "Security group for PostgreSQL RDS instance"
  vpc_id      = var.vpc_id

  ingress {
    description = "PostgreSQL from VPC"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "postgres-sg"
  }
}

# Variables
variable "cyberark_username" {
  description = "CyberArk service account username (e.g., user@cyberark.cloud.12345)"
  type        = string
  sensitive   = true
}

variable "cyberark_client_secret" {
  description = "CyberArk ISPSS client secret"
  type        = string
  sensitive   = true
}

variable "cyberark_identity_url" {
  description = "CyberArk Identity URL (optional - only needed for GovCloud or custom deployments)"
  type        = string
  default     = ""
}


variable "db_master_password" {
  description = "Master password for RDS instance"
  type        = string
  sensitive   = true
}

variable "vpc_id" {
  description = "VPC ID where RDS will be deployed"
  type        = string
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs for RDS"
  type        = list(string)
}

# Outputs
output "# NOTE: Secrets are standalone - no workspace_id needed" {
  description = "SIA database target ID"
  value       = cyberarksia_database_workspace.postgres.id
}

output "rds_endpoint" {
  description = "RDS instance endpoint"
  value       = aws_db_instance.postgres.endpoint
}
