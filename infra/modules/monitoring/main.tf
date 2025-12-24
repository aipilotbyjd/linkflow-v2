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

variable "tags" {
  type    = map(string)
  default = {}
}

resource "azurerm_log_analytics_workspace" "main" {
  name                = "${var.name_prefix}-logs"
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = var.tags
}

resource "azurerm_application_insights" "main" {
  name                = "${var.name_prefix}-insights"
  location            = var.location
  resource_group_name = var.resource_group_name
  workspace_id        = azurerm_log_analytics_workspace.main.id
  application_type    = "other"
  tags                = var.tags
}

resource "azurerm_monitor_action_group" "critical" {
  name                = "${var.name_prefix}-critical-alerts"
  resource_group_name = var.resource_group_name
  short_name          = "critical"
  tags                = var.tags
}

resource "azurerm_monitor_metric_alert" "high_cpu" {
  name                = "${var.name_prefix}-high-cpu"
  resource_group_name = var.resource_group_name
  scopes              = [azurerm_application_insights.main.id]
  description         = "Alert when CPU usage is high"
  severity            = 2
  frequency           = "PT5M"
  window_size         = "PT15M"
  tags                = var.tags

  criteria {
    metric_namespace = "microsoft.insights/components"
    metric_name      = "performanceCounters/processCpuPercentage"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 80
  }

  action {
    action_group_id = azurerm_monitor_action_group.critical.id
  }
}

output "log_analytics_workspace_id" {
  value = azurerm_log_analytics_workspace.main.id
}

output "log_analytics_workspace_key" {
  value     = azurerm_log_analytics_workspace.main.primary_shared_key
  sensitive = true
}

output "log_analytics_customer_id" {
  value = azurerm_log_analytics_workspace.main.workspace_id
}

output "app_insights_id" {
  value = azurerm_application_insights.main.id
}

output "app_insights_instrumentation_key" {
  value     = azurerm_application_insights.main.instrumentation_key
  sensitive = true
}

output "app_insights_connection_string" {
  value     = azurerm_application_insights.main.connection_string
  sensitive = true
}
