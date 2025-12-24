terraform {
  required_version = ">= 1.5.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.80"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }

  backend "azurerm" {
    # Configure in backend.tfvars or via -backend-config
    # resource_group_name  = "linkflow-tfstate-rg"
    # storage_account_name = "linkflowtfstate"
    # container_name       = "tfstate"
    # key                  = "terraform.tfstate"
  }
}

provider "azurerm" {
  features {
    key_vault {
      purge_soft_delete_on_destroy    = true
      recover_soft_deleted_key_vaults = true
    }
  }
}

# Variables
variable "environment" {
  type        = string
  description = "Environment name (dev, staging, prod)"
}

variable "location" {
  type        = string
  default     = "eastus"
  description = "Azure region"
}

variable "image_tag" {
  type        = string
  default     = "latest"
  description = "Container image tag"
}

variable "api_min_replicas" {
  type    = number
  default = 1
}

variable "api_max_replicas" {
  type    = number
  default = 10
}

variable "worker_min_replicas" {
  type    = number
  default = 1
}

variable "worker_max_replicas" {
  type    = number
  default = 5
}

variable "db_sku" {
  type    = string
  default = "B_Standard_B1ms"
}

variable "redis_sku" {
  type    = string
  default = "Basic"
}

locals {
  name_prefix = "linkflow-${var.environment}"
  tags = {
    Environment = var.environment
    Project     = "linkflow"
    ManagedBy   = "terraform"
  }
}

# Resource Group
resource "azurerm_resource_group" "main" {
  name     = "${local.name_prefix}-rg"
  location = var.location
  tags     = local.tags
}

# Generate secrets
resource "random_password" "db_password" {
  length  = 24
  special = false
}

resource "random_password" "jwt_secret" {
  length  = 32
  special = false
}

# Container Registry
resource "azurerm_container_registry" "main" {
  name                = replace("${local.name_prefix}registry", "-", "")
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Basic"
  admin_enabled       = true
  tags                = local.tags
}

# Key Vault for secrets
resource "azurerm_key_vault" "main" {
  name                       = "${local.name_prefix}-kv"
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7
  purge_protection_enabled   = false
  tags                       = local.tags

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    secret_permissions = [
      "Get", "List", "Set", "Delete", "Purge"
    ]
  }
}

data "azurerm_client_config" "current" {}

# Store secrets in Key Vault
resource "azurerm_key_vault_secret" "db_password" {
  name         = "db-password"
  value        = random_password.db_password.result
  key_vault_id = azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "jwt_secret" {
  name         = "jwt-secret"
  value        = random_password.jwt_secret.result
  key_vault_id = azurerm_key_vault.main.id
}

# Networking
module "networking" {
  source              = "./modules/networking"
  name_prefix         = local.name_prefix
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tags                = local.tags
}

# Monitoring
module "monitoring" {
  source              = "./modules/monitoring"
  name_prefix         = local.name_prefix
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tags                = local.tags
}

# Database
module "database" {
  source              = "./modules/database"
  name_prefix         = local.name_prefix
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  subnet_id           = module.networking.database_subnet_id
  dns_zone_id         = module.networking.postgres_dns_zone_id
  admin_password      = random_password.db_password.result
  sku_name            = var.db_sku
  tags                = local.tags
}

# Redis
module "redis" {
  source              = "./modules/redis"
  name_prefix         = local.name_prefix
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku_name            = var.redis_sku
  tags                = local.tags
}

# Container Apps Environment
resource "azurerm_container_app_environment" "main" {
  name                       = "${local.name_prefix}-env"
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  log_analytics_workspace_id = module.monitoring.log_analytics_workspace_id
  infrastructure_subnet_id   = module.networking.container_apps_subnet_id
  tags                       = local.tags
}

# Container Apps
module "container_apps" {
  source                       = "./modules/container-app"
  name_prefix                  = local.name_prefix
  location                     = azurerm_resource_group.main.location
  resource_group_name          = azurerm_resource_group.main.name
  container_app_environment_id = azurerm_container_app_environment.main.id
  registry_server              = azurerm_container_registry.main.login_server
  registry_username            = azurerm_container_registry.main.admin_username
  registry_password            = azurerm_container_registry.main.admin_password
  image_tag                    = var.image_tag
  database_host                = module.database.server_fqdn
  database_password            = random_password.db_password.result
  redis_host                   = module.redis.hostname
  redis_port                   = module.redis.port
  redis_password               = module.redis.primary_access_key
  jwt_secret                   = random_password.jwt_secret.result
  environment                  = var.environment
  api_min_replicas             = var.api_min_replicas
  api_max_replicas             = var.api_max_replicas
  worker_min_replicas          = var.worker_min_replicas
  worker_max_replicas          = var.worker_max_replicas
  tags                         = local.tags
}

# Outputs
output "api_url" {
  value       = module.container_apps.api_url
  description = "API endpoint URL"
}

output "registry_login_server" {
  value       = azurerm_container_registry.main.login_server
  description = "Container registry login server"
}

output "resource_group_name" {
  value       = azurerm_resource_group.main.name
  description = "Resource group name"
}

output "key_vault_name" {
  value       = azurerm_key_vault.main.name
  description = "Key Vault name for secrets"
}

output "database_host" {
  value       = module.database.server_fqdn
  description = "PostgreSQL server hostname"
  sensitive   = true
}
