# ADR 0008: Use a minimal Prometheus/Grafana stack without OTel or Loki

## Context

Phase 3 needs real request/error/latency/availability data for one sample
service on a disposable two-node cluster, plus a place to evaluate the SLO
queries in docs/SLO.md.

## Decision

Install `kube-prometheus-stack` (pinned 91.4.1) with Alertmanager disabled,
6h retention and small resource profiles; scrape Golden Path Pods through the
portable `prometheus.io/*` annotations the chart already renders instead of
adding ServiceMonitor CRDs to the developer path; skip OpenTelemetry and Loki.

## Alternatives

A standalone Prometheus + Grafana pair would be lighter but reassembles
service discovery, dashboards provisioning and kube-state-metrics by hand.
OTel + Tempo/Loki would add collectors, instrumentation SDKs and storage for
signals a single dependency-free `/metrics` endpoint already provides.

## Consequences

One Helm release covers collection, storage, dashboards and baseline alerts.
The Golden Path chart stays operator-agnostic. Alertmanager absence means
alerts are observed, not routed — acceptable for validation, explicitly not a
production alerting story. Log aggregation remains out of scope.
