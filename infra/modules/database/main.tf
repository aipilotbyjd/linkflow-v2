terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.80"
    }
  }
}

variable "name_prefix" {
  type = string
}

variable "location" {
  type = string
}

variable "resource_group_name" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "dns_zone_id" {
  type = string
}

variable "admin_password" {
  type      = string
  sensitive = true
}

variable "sku_name" {
  type    = string
  default = "B_Standard_B1ms"
}

variable "storage_mb" {
  type    = number
  default = 32768
}

variable "tags" {
  type    = map(string)
  default = {}
}

resource "azurerm_postgresql_flexible_server" "main" {
  name                          = "${var.name_prefix}-postgres"
  resource_group_name           = var.resource_group_name
  location                      = var.location
  version                       = "15"
  delegated_subnet_id           = var.subnet_id
  private_dns_zone_id           = var.dns_zone_id
  administrator_login           = "linkflowadmin"
  administrator_password        = var.admin_password
  zone                          = "1"
  storage_mb                    = var.storage_mb
  sku_name                      = var.sku_name
  backup_retention_days         = 7
  geo_redundant_backup_enabled  = false
  auto_grow_enabled             = true
  public_network_access_enabled = false
  tags                          = var.tags

  lifecycle {
    ignore_changes = [
      zone,
      high_availability[0].standby_availability_zone
    ]
  }
}

resource "azurerm_postgresql_flexible_server_database" "linkflow" {
  name      = "linkflow"
  server_id = azurerm_postgresql_flexible_server.main.id
  collation = "en_US.utf8"
  charset   = "UTF8"
}

resource "azurerm_postgresql_flexible_server_configuration" "extensions" {
  name      = "azure.extensions"
  server_id = azurerm_postgresql_flexible_server.main.id
  value     = "UUID-OSSP,PGCRYPTO"
}

output "server_id" {
  value = azurerm_postgresql_flexible_server.main.id
}

output "server_fqdn" {
  value = azurerm_postgresql_flexible_server.main.fqdn
}

output "database_name" {
  value = azurerm_postgresql_flexible_server_database.linkflow.name
}

output "admin_username" {
  value = azurerm_postgresql_flexible_server.main.administrator_login
}
