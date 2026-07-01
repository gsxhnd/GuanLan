# GuanLan Backend Services - goreman Procfile
# Usage: goreman start
#
# Each service uses air for hot-reload during development.
# For production, replace with direct binary execution:
#   api: ./tmp/api
#   gateway: ./tmp/gateway

# gRPC API Server (port :50051)
api: air -c .air.api.toml

# HTTP Gateway / gRPC-Gateway (port :8080)
gateway: air -c .air.gateway.toml
