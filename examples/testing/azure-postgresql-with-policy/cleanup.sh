#!/bin/bash
set -e

# ============================================================================
# Azure PostgreSQL + SIA Policy Test Cleanup Script
# ============================================================================

echo "🧹 Azure PostgreSQL + SIA Policy Test Cleanup"
echo ""

# ============================================================================
# Confirmation
# ============================================================================

echo "⚠️  WARNING: This will DESTROY all test resources!"
echo ""
echo "Resources to be destroyed:"
terraform state list 2>/dev/null || echo "  (Run 'terraform state list' to see resources)"
echo ""
read -p "Are you sure you want to destroy everything? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "❌ Cleanup aborted by user"
    exit 0
fi
echo ""

# ============================================================================
# Terraform Destroy
# ============================================================================

echo "🗑️  Destroying resources..."
echo "   Terraform output: /tmp/tf-destroy.log"
echo ""

# Run destroy in background with progress tracking
terraform destroy -auto-approve > /tmp/tf-destroy.log 2>&1 &
TF_PID=$!

# Show spinner while waiting
spin='-\|/'
i=0
while kill -0 $TF_PID 2> /dev/null; do
    i=$(( (i+1) %4 ))
    printf "\r   ${spin:$i:1} Destroying resources... "
    sleep 0.5
done
echo ""

# Check if destroy succeeded
wait $TF_PID
if [ $? -eq 0 ]; then
    echo "✅ All resources destroyed successfully!"
else
    echo "❌ Destroy failed! Check /tmp/tf-destroy.log"
    tail -30 /tmp/tf-destroy.log
    exit 1
fi
echo ""

# ============================================================================
# Verification
# ============================================================================

echo "🔍 Verifying cleanup..."
echo ""

# Check Terraform state
REMAINING=$(terraform state list 2>/dev/null | wc -l)
if [ "$REMAINING" -eq 0 ]; then
    echo "✅ Terraform state is clean (no resources remaining)"
else
    echo "⚠️  WARNING: $REMAINING resources still in state"
    terraform state list
    echo ""
fi

# Check Azure resources
echo "Checking Azure resources..."
if [ -f "terraform.tfstate" ]; then
    RESOURCE_GROUP=$(terraform output -raw azure_resource_group 2>/dev/null || echo "")
    if [ -n "$RESOURCE_GROUP" ]; then
        if az group show --name "$RESOURCE_GROUP" &> /dev/null; then
            echo "⚠️  WARNING: Resource group '$RESOURCE_GROUP' still exists!"
            echo "   Manual cleanup: az group delete --name '$RESOURCE_GROUP' --yes"
        else
            echo "✅ Azure resource group deleted"
        fi
    else
        echo "✅ No Azure resources found"
    fi
else
    echo "✅ No state file found"
fi
echo ""

# ============================================================================
# Summary
# ============================================================================

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Cleanup Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 Verification Checklist:"
echo "   1. Check SIA UI - no orphaned resources"
echo "   2. Check Azure Portal - resource group deleted"
echo "   3. Verify costs stopped accumulating"
echo ""
echo "📊 Terraform logs available at:"
echo "   /tmp/tf-init.log"
echo "   /tmp/tf-plan.log"
echo "   /tmp/tf-apply.log"
echo "   /tmp/tf-destroy.log"
echo ""
