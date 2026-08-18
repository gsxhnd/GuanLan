# GuanLan Backend Services - goreman Procfile
# Usage: goreman start
#
# api: HTTP + DuckDB + orchestrator + cron (:8080)
# crawler: stateless Python gRPC (:50061)
# prediction: lazy-loaded serving gRPC (:50062)

api: air -c .air.api.toml
crawler: uv run python -m quant.crawler.server --addr :50061
prediction: uv run python -m quant.ml.serving --addr :50062 --models data/models
