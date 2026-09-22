# SLO

## Objectives (design targets, not measured commitments)

For the sample payment API over a 30-day window:

- availability SLI: `sum(rate(http_requests_total{code!~"5.."}[5m])) / sum(rate(http_requests_total[5m]))`;
- availability SLO: **99.9%**;
- latency SLI: `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))`;
- latency SLO: **99% of valid requests within 500 ms**.

## Why these numbers (personal-project assumptions)

- 99.9% (~43m error budget/month) is the conventional entry-level
  "three nines" for an internal platform service: strict enough to matter,
  loose enough that a disposable two-node validation cluster with no
  multi-AZ redundancy can plausibly stay inside it during a short run. It is
  not derived from production traffic.
- The 500 ms threshold is generous headroom, deliberately: a local
  sequential probe (300 requests, Windows localhost, per-request curl spawn
  overhead included) measured p50 ≈ 30 ms, p99 ≈ 73 ms, max ≈ 84 ms on
  2026-09-22. True in-handler latency is a fraction of that. 500 ms leaves
  >6x headroom for in-cluster networking and shared-node noise, so a breach
  during validation would indicate a real problem rather than threshold
  noise.
- These are hypotheses to evaluate, not commitments met in production. The
  Phase 3 validation checks that the SLI queries *evaluate against real
  data* (Prometheus + Grafana on the disposable cluster), not that a 30-day
  burn stays within budget — a two-hour cluster cannot prove a 30-day SLO.

## Collection (Phase 3, implemented)

The sample app exposes `http_requests_total{method,path,code}`,
`http_request_errors_total` and `http_request_duration_seconds_*`; the
stack scrapes Golden Path Pods via the chart's `prometheus.io/*`
annotations; `observability/dashboards/golden-path-overview.json` graphs
rate, error ratio, p95 and 5m availability; baseline alerts live in
`observability/prometheus-rules.yaml` (observed, not routed — no
Alertmanager in this environment).

Ownership split: the platform owns collection plumbing, standard labels,
dashboard and baseline alerts; the application team owns request
classification, instrumentation correctness and service-specific targets.
