# Observability

Minimal, honest monitoring for a disposable validation cluster. The goal is
not "tools installed" but: a developer can see request rate, error rate,
latency and availability of their Golden Path service with real data, and the
SLI queries in docs/SLO.md evaluate against that data.

## Stack (pinned, minimal)

- `kube-prometheus-stack` chart 91.4.1 (`scripts/bootstrap-observability.sh`,
  values in `observability/prometheus-values.yaml`): Prometheus (6h
  retention, small requests/limits), Grafana with dashboard sidecar,
  kube-state-metrics and node-exporter. Alertmanager is **disabled**: there is
  no paging route in a personal validation environment; baseline alerts exist
  as PrometheusRules and are observed in the Prometheus/Grafana UIs.
- No OpenTelemetry, no Loki: unjustified for one sample service with a
  dependency-free `/metrics` endpoint. The chart stays operator-agnostic by
  scraping the portable `prometheus.io/*` annotations it already renders
  (see ADR 0008); no ServiceMonitor CRD leaks into the developer path.

## Platform-provided signals

The sample app exposes dependency-free Prometheus exposition:

- `http_requests_total{method,path,code}` — request rate, availability numerator/denominator;
- `http_request_errors_total{method,path}` — 5xx error rate;
- `http_request_duration_seconds_{bucket,sum,count}` — latency histogram (p95 via `histogram_quantile`);
- plus `up{job="golden-path"}` and `kube_deployment_status_replicas_available` for target/workload health.

Label values use a normalized path (`/other` fallback) so cardinality is bounded.

## Dashboard

`observability/dashboards/golden-path-overview.json` (Git-managed, loaded via
a labeled ConfigMap created by the bootstrap script): request rate, error
rate, p95 latency, 5m availability, ready replicas. Structural validity
(panels, queries, known metrics only) is asserted in Go
(`internal/guardrails`); real data is verified on the EKS cluster.

## Responsibility boundary

Platform team: metrics infrastructure, scrape configuration, standard
dashboard, baseline `PrometheusRule` alerts, collection mechanism.
Application team: business metrics, service-specific SLI definitions,
meaningful thresholds, incident response. Instrumenting a new framework means
exposing the same four metric families; the dashboard and alerts keep working.
