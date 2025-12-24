environment = "staging"
location    = "eastus"

# Scaling (moderate for staging)
api_min_replicas    = 2
api_max_replicas    = 5
worker_min_replicas = 1
worker_max_replicas = 3

# Moderate SKUs
db_sku    = "GP_Standard_D2s_v3"
redis_sku = "Standard"
