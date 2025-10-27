#!/bin/bash
# Import existing certificate from CyberArk SIA
#
# Prerequisites:
#   - Certificate exists in SIA (created via UI or API)
#   - Know the certificate ID (numeric string, e.g., "1761251731882561")
#
# Usage:
#   ./import.sh <certificate_id>
#
# Example:
#   ./import.sh 1761251731882561
#
# After Import:
#   1. Certificate appears in terraform state
#   2. All fields populated (including cert_body from API)
#   3. You MUST add cert_body to your Terraform config for subsequent updates to work
#      - API requires cert_body for ALL update operations (even metadata-only changes)
#      - Without cert_body in config, next `terraform apply` will fail
#
# Finding Certificate ID:
#   - SIA UI: Navigate to Certificates page, copy ID from URL or table
#   - API: GET /api/certificates (lists all certificates with IDs)

if [ -z "$1" ]; then
  echo "Usage: $0 <certificate_id>"
  echo "Example: $0 1761251731882561"
  exit 1
fi

CERTIFICATE_ID=$1

# Import certificate into Terraform state
terraform import cyberark_sia_certificate.imported "$CERTIFICATE_ID"

echo ""
echo "Certificate imported successfully!"
echo ""
echo "IMPORTANT: Add certificate configuration to your Terraform files:"
echo ""
echo "resource \"cyberark_sia_certificate\" \"imported\" {"
echo "  cert_name        = \"<certificate-name-from-state>\""
echo "  cert_body        = file(\"path/to/certificate.pem\")  # REQUIRED for updates"
echo "  cert_description = \"<description-from-state>\""
echo "  domain_name      = \"<domain-from-state>\""
echo "  labels = {"
echo "    environment = \"production\""
echo "  }"
echo "}"
echo ""
echo "Run 'terraform state show cyberark_sia_certificate.imported' to see current values"
