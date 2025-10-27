# Complete Workflow Example: AWS RDS + CyberArk SIA Integration
#
# This example demonstrates the complete lifecycle for onboarding an AWS RDS database
# to CyberArk SIA, including:
# 1. AWS RDS PostgreSQL database provisioning (using AWS provider)
# 2. SIA database target registration (using CyberArk SIA provider)
# 3. Strong account creation for SIA ephemeral access
# 4. Update operations (changing database port, rotating credentials)
# 5. Proper dependency management and cleanup
#
# Prerequisites:
# - AWS credentials configured (for RDS provisioning)
# - CyberArk Identity/ISPSS credentials configured (for SIA integration)
# - Network connectivity between SIA and RDS (security groups, VPC, etc.)

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    cyberark_sia = {
      source  = "aaearon/cyberarksia" # Replace with registry path after publication
      version = "~> 1.0"
    }
  }
}

# ============================================================================
# Provider Configuration
# ============================================================================

provider "aws" {
  region = var.aws_region
}

provider "cyberark_sia" {
  client_id                 = var.cyberark_client_id
  client_secret             = var.cyberark_client_secret
  identity_tenant_subdomain = var.cyberark_tenant_subdomain

  # Optional: Override identity URL for GovCloud or custom deployments
  # identity_url = var.cyberark_identity_url
}

# ============================================================================
# Variables
# ============================================================================

variable "aws_region" {
  description = "AWS region for RDS deployment"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (dev, staging, production)"
  type        = string
  default     = "dev"
}

variable "cyberark_client_id" {
  description = "CyberArk Identity ISPSS client ID"
  type        = string
  sensitive   = true
}

variable "cyberark_client_secret" {
  description = "CyberArk Identity ISPSS client secret"
  type        = string
  sensitive   = true
}

variable "cyberark_identity_url" {
  description = "CyberArk Identity tenant URL (optional - only needed for GovCloud or custom deployments)"
  type        = string
  default     = ""
}

variable "cyberark_tenant_subdomain" {
  description = "CyberArk Identity tenant subdomain (e.g., 'abc123' from abc123.cyberark.cloud)"
  type        = string
}

variable "db_master_password" {
  description = "Master password for RDS database"
  type        = string
  sensitive   = true
}

variable "sia_strong_account_password" {
  description = "Password for SIA strong account"
  type        = string
  sensitive   = true
}

# ============================================================================
# Data Sources
# ============================================================================

data "aws_caller_identity" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

# ============================================================================
# Networking (VPC, Subnets, Security Groups)
# ============================================================================

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.environment}-sia-vpc"
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name        = "${var.environment}-sia-subnet-${count.index + 1}"
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

resource "aws_db_subnet_group" "main" {
  name       = "${var.environment}-sia-db-subnet"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name        = "${var.environment}-sia-db-subnet-group"
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

resource "aws_security_group" "rds" {
  name        = "${var.environment}-sia-rds-sg"
  description = "Security group for RDS PostgreSQL database"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "PostgreSQL from VPC"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.environment}-sia-rds-sg"
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# ============================================================================
# Step 1: AWS RDS PostgreSQL Database Provisioning
# ============================================================================

resource "aws_db_instance" "postgresql" {
  identifier = "${var.environment}-sia-postgres-db"

  # Engine configuration
  engine               = "postgres"
  engine_version       = "14.7"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
  storage_encrypted    = true

  # Database configuration
  db_name  = "appdb"
  username = "postgres"
  password = var.db_master_password

  # Network configuration
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false

  # Backup and maintenance
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"

  # Monitoring and logging
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  monitoring_interval             = 60

  # Lifecycle
  skip_final_snapshot = true # For dev/testing; set to false in production
  deletion_protection = false # For dev/testing; set to true in production

  tags = {
    Name        = "${var.environment}-sia-postgresql"
    Environment = var.environment
    ManagedBy   = "Terraform"
    SIAManaged  = "true"
  }
}

# ============================================================================
# Step 2: SIA Database Target Registration
# ============================================================================

resource "cyberarksia_database_workspace" "postgresql" {
  # Database identification
  # IMPORTANT: 'name' is the actual database name on the server (from RDS db_name),
  # NOT a workspace label. SIA uses this to connect (postgres://host:port/DATABASE_NAME)
  name          = aws_db_instance.postgresql.db_name  # "appdb" - actual database name
  database_type = "postgresql"

  # Connection details (from AWS RDS)
  address = aws_db_instance.postgresql.address
  port    = aws_db_instance.postgresql.port

  # Authentication method
  authentication_method = "local_ephemeral_user"

  # Cloud provider metadata
  cloud_provider = "aws"
  region         = var.aws_region

  # Dependency: Wait for RDS to be available
  depends_on = [aws_db_instance.postgresql]
}

# ============================================================================
# Step 3: SIA Strong Account Creation
# ============================================================================

resource "cyberarksia_secret" "postgres_admin" {
  # Basic identification
  name               = "${var.environment}-postgres-admin-account"

  # Authentication credentials
  authentication_type = "local"
  username            = "sia_service_account"
  password            = var.sia_strong_account_password

  # NOTE: Secrets are standalone - no dependency on workspace required
  # Optionally depend on workspace if you want sequential creation
  # depends_on = [cyberarksia_database_workspace.postgresql]
}

# ============================================================================
# Outputs
# ============================================================================

output "rds_endpoint" {
  description = "RDS database endpoint"
  value       = aws_db_instance.postgresql.endpoint
}

output "rds_address" {
  description = "RDS database address"
  value       = aws_db_instance.postgresql.address
}

output "rds_port" {
  description = "RDS database port"
  value       = aws_db_instance.postgresql.port
}

output "sia_database_workspace_id" {
  description = "SIA database workspace ID"
  value       = cyberarksia_database_workspace.postgresql.id
}

output "sia_secret_id" {
  description = "SIA secret ID"
  value       = cyberarksia_secret.postgres_admin.id
  sensitive   = true
}

output "integration_status" {
  description = "Integration status summary"
  value = {
    rds_identifier      = aws_db_instance.postgresql.identifier
    sia_workspace_name  = cyberarksia_database_workspace.postgresql.name
    sia_secret_name     = cyberarksia_secret.postgres_admin.name
    environment         = var.environment
    integration_complete = true
  }
}

# ============================================================================
# Example Update Scenarios
# ============================================================================

# To update the database workspace (e.g., change authentication method):
# 1. Modify an attribute above (e.g., authentication_method)
# 2. Run: terraform plan
# 3. Run: terraform apply
# Expected: In-place update of the database workspace

# To rotate strong account credentials:
# 1. Update var.sia_strong_account_password variable
# 2. Run: terraform plan
# 3. Run: terraform apply
# Expected: SIA immediately updates credentials (per FR-015a)

# To change database type (forces replacement):
# 1. Change database_type attribute (e.g., "postgresql" to "mysql")
# 2. Run: terraform plan
# Expected: Plan shows destroy + create (ForceNew behavior)
# NOTE: This would require changing the RDS instance as well

# ============================================================================
# Cleanup Instructions
# ============================================================================

# To destroy all resources:
# 1. Run: terraform destroy
# 2. Confirm the plan
#
# Destruction order (handled automatically by Terraform):
# 1. SIA strong account deleted
# 2. SIA database target deleted
# 3. AWS RDS instance deleted
# 4. AWS networking resources deleted
#
# Note: Ensure no active SIA sessions or policies reference these resources
# before destroying.
