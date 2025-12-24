#!/bin/bash
set -e

# Bootstrap script to create Terraform state storage
# Run this ONCE before first deployment

RESOURCE_GROUP="linkflow-tfstate-rg"
STORAGE_ACCOUNT="linkflowtfstate"
CONTAINER_NAME="tfstate"
LOCATION="eastus"

echo "=== Terraform Backend Bootstrap ==="

# Check Azure CLI
if ! command -v az &> /dev/null; then
    echo "Error: Azure CLI not installed"
    echo "Install: brew install azure-cli"
    exit 1
fi

# Login check
echo "Checking Azure login..."
az account show > /dev/null 2>&1 || az login

# Create resource group for state
echo "Creating resource group for Terraform state..."
az group create \
    --name $RESOURCE_GROUP \
    --location $LOCATION

# Create storage account
echo "Creating storage account..."
az storage account create \
    --name $STORAGE_ACCOUNT \
    --resource-group $RESOURCE_GROUP \
    --location $LOCATION \
    --sku Standard_LRS \
    --encryption-services blob

# Get storage account key
ACCOUNT_KEY=$(az storage account keys list \
    --resource-group $RESOURCE_GROUP \
    --account-name $STORAGE_ACCOUNT \
    --query '[0].value' -o tsv)

# Create blob container
echo "Creating blob container..."
az storage container create \
    --name $CONTAINER_NAME \
    --account-name $STORAGE_ACCOUNT \
    --account-key $ACCOUNT_KEY

# Enable versioning for state recovery
az storage account blob-service-properties update \
    --account-name $STORAGE_ACCOUNT \
    --resource-group $RESOURCE_GROUP \
    --enable-versioning true

echo ""
echo "=== Bootstrap Complete ==="
echo ""
echo "Terraform backend configured:"
echo "  Resource Group:  $RESOURCE_GROUP"
echo "  Storage Account: $STORAGE_ACCOUNT"
echo "  Container:       $CONTAINER_NAME"
echo ""
echo "Next steps:"
echo "1. cd infra"
echo "2. terraform init -backend-config=environments/dev/backend.tfvars"
echo "3. terraform plan -var-file=environments/dev/terraform.tfvars"
echo "4. terraform apply -var-file=environments/dev/terraform.tfvars"
