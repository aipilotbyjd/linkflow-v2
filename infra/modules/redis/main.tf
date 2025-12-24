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

variable "sku_name" {
  type    = string
  default = "Basic"
}

variable "family" {
  type    = string
  default = "C"
}

variable "capacity" {
  type    = number
  default = 0
}

variable "tags" {
  type    = map(string)
  default = {}
}

resource "azurerm_redis_cache" "main" {
  name                          = "${var.name_prefix}-redis"
  location                      = var.location
  resource_group_name           = var.resource_group_name
  capacity                      = var.capacity
  family                        = var.family
  sku_name                      = var.sku_name
  enable_non_ssl_port           = false
  minimum_tls_version           = "1.2"
  public_network_access_enabled = true
  tags                          = var.tags

  redis_configuration {
    maxmemory_policy = "volatile-lru"
  }
}

output "id" {
  value = azurerm_redis_cache.main.id
}

output "hostname" {
  value = azurerm_redis_cache.main.hostname
}

output "port" {
  value = azurerm_redis_cache.main.ssl_port
}

output "primary_access_key" {
  value     = azurerm_redis_cache.main.primary_access_key
  sensitive = true
}

output "connection_string" {
  value     = azurerm_redis_cache.main.primary_connection_string
  sensitive = true
}
