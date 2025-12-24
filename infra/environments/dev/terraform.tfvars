environment = "dev"
location    = "eastus"

# Scaling (lower for dev)
api_min_replicas    = 1
api_max_replicas    = 3
worker_min_replicas = 1
worker_max_replicas = 2

# Smaller SKUs for cost savings
db_sku    = "B_Standard_B1ms"
redis_sku = "Basic"
