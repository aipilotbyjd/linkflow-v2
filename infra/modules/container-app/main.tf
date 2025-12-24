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

variable "container_app_environment_id" {
  type = string
}

variable "registry_server" {
  type = string
}

variable "registry_username" {
  type = string
}

variable "registry_password" {
  type      = string
  sensitive = true
}

variable "image_tag" {
  type    = string
  default = "latest"
}

variable "database_host" {
  type = string
}

variable "database_password" {
  type      = string
  sensitive = true
}

variable "redis_host" {
  type = string
}

variable "redis_port" {
  type    = number
  default = 6380
}

variable "redis_password" {
  type      = string
  sensitive = true
}

variable "jwt_secret" {
  type      = string
  sensitive = true
}

variable "environment" {
  type    = string
  default = "production"
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

variable "tags" {
  type    = map(string)
  default = {}
}

locals {
  common_env = [
    { name = "APP_ENVIRONMENT", value = var.environment },
    { name = "APP_DEBUG", value = "false" },
    { name = "DATABASE_HOST", value = var.database_host },
    { name = "DATABASE_PORT", value = "5432" },
    { name = "DATABASE_USER", value = "linkflowadmin" },
    { name = "DATABASE_PASSWORD", secretRef = "db-password" },
    { name = "DATABASE_NAME", value = "linkflow" },
    { name = "DATABASE_SSLMODE", value = "require" },
    { name = "REDIS_HOST", value = var.redis_host },
    { name = "REDIS_PORT", value = tostring(var.redis_port) },
    { name = "REDIS_PASSWORD", secretRef = "redis-password" },
    { name = "REDIS_TLS", value = "true" },
    { name = "JWT_SECRET", secretRef = "jwt-secret" },
  ]

  secrets = [
    { name = "registry-password", value = var.registry_password },
    { name = "db-password", value = var.database_password },
    { name = "redis-password", value = var.redis_password },
    { name = "jwt-secret", value = var.jwt_secret },
  ]
}

# API Container App
resource "azurerm_container_app" "api" {
  name                         = "${var.name_prefix}-api"
  container_app_environment_id = var.container_app_environment_id
  resource_group_name          = var.resource_group_name
  revision_mode                = "Single"
  tags                         = var.tags

  registry {
    server               = var.registry_server
    username             = var.registry_username
    password_secret_name = "registry-password"
  }

  dynamic "secret" {
    for_each = local.secrets
    content {
      name  = secret.value.name
      value = secret.value.value
    }
  }

  ingress {
    external_enabled = true
    target_port      = 8090
    transport        = "http"

    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }

  template {
    min_replicas = var.api_min_replicas
    max_replicas = var.api_max_replicas

    container {
      name   = "api"
      image  = "${var.registry_server}/linkflow-api:${var.image_tag}"
      cpu    = 0.5
      memory = "1Gi"

      dynamic "env" {
        for_each = local.common_env
        content {
          name        = env.value.name
          value       = lookup(env.value, "value", null)
          secret_name = lookup(env.value, "secretRef", null)
        }
      }

      liveness_probe {
        path             = "/api/v1/health/live"
        port             = 8090
        transport        = "HTTP"
        initial_delay    = 10
        interval_seconds = 30
      }

      readiness_probe {
        path             = "/api/v1/health/ready"
        port             = 8090
        transport        = "HTTP"
        initial_delay    = 5
        interval_seconds = 10
      }
    }

    http_scale_rule {
      name                = "http-scaling"
      concurrent_requests = "100"
    }
  }
}

# Worker Container App
resource "azurerm_container_app" "worker" {
  name                         = "${var.name_prefix}-worker"
  container_app_environment_id = var.container_app_environment_id
  resource_group_name          = var.resource_group_name
  revision_mode                = "Single"
  tags                         = var.tags

  registry {
    server               = var.registry_server
    username             = var.registry_username
    password_secret_name = "registry-password"
  }

  dynamic "secret" {
    for_each = local.secrets
    content {
      name  = secret.value.name
      value = secret.value.value
    }
  }

  template {
    min_replicas = var.worker_min_replicas
    max_replicas = var.worker_max_replicas

    container {
      name   = "worker"
      image  = "${var.registry_server}/linkflow-worker:${var.image_tag}"
      cpu    = 0.5
      memory = "1Gi"

      dynamic "env" {
        for_each = local.common_env
        content {
          name        = env.value.name
          value       = lookup(env.value, "value", null)
          secret_name = lookup(env.value, "secretRef", null)
        }
      }
    }

    custom_scale_rule {
      name             = "queue-scaling"
      custom_rule_type = "redis"
      metadata = {
        host       = var.redis_host
        port       = tostring(var.redis_port)
        enableTLS  = "true"
        listName   = "asynq:default:pending"
        listLength = "10"
      }
    }
  }
}

# Scheduler Container App
resource "azurerm_container_app" "scheduler" {
  name                         = "${var.name_prefix}-scheduler"
  container_app_environment_id = var.container_app_environment_id
  resource_group_name          = var.resource_group_name
  revision_mode                = "Single"
  tags                         = var.tags

  registry {
    server               = var.registry_server
    username             = var.registry_username
    password_secret_name = "registry-password"
  }

  dynamic "secret" {
    for_each = local.secrets
    content {
      name  = secret.value.name
      value = secret.value.value
    }
  }

  template {
    min_replicas = 1
    max_replicas = 1

    container {
      name   = "scheduler"
      image  = "${var.registry_server}/linkflow-scheduler:${var.image_tag}"
      cpu    = 0.25
      memory = "0.5Gi"

      dynamic "env" {
        for_each = local.common_env
        content {
          name        = env.value.name
          value       = lookup(env.value, "value", null)
          secret_name = lookup(env.value, "secretRef", null)
        }
      }
    }
  }
}

output "api_url" {
  value = "https://${azurerm_container_app.api.ingress[0].fqdn}"
}

output "api_id" {
  value = azurerm_container_app.api.id
}

output "worker_id" {
  value = azurerm_container_app.worker.id
}

output "scheduler_id" {
  value = azurerm_container_app.scheduler.id
}
