environment = "prod"
location    = "eastus"

# Scaling (production ready)
api_min_replicas    = 3
api_max_replicas    = 10
worker_min_replicas = 2
worker_max_replicas = 5

# Production SKUs
db_sku    = "GP_Standard_D4s_v3"
redis_sku = "Premium"
