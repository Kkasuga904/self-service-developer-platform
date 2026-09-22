# Golden Path

## What it creates

The chart emits one Deployment, one ClusterIP Service, one
PodDisruptionBudget, and an HPA only when explicitly enabled. It standardizes
labels (including `platform.example.io/owner|service|environment` from the
contract), probes, rolling update settings, security context and resource profiles.
It deliberately depends on no operator CRDs: metrics reach Prometheus through
the portable `prometheus.io/*` annotations, so the developer path works on any
cluster and the operator-specific wiring stays platform-owned (see ADR 0008).

## Why profiles instead of raw resources

`small`, `medium`, and `large` keep the developer input simple and make policy
review tractable. They are starter values, not measured production sizing.
Applications still require measurement and profile evolution.

## Deliberate omissions

Ingress, secrets, persistent volumes, cloud dependencies, ServiceMonitor,
NetworkPolicy and policy exceptions are not chart features. Adding them
without a concrete product requirement would make the contract broad but less
coherent. Two teams (`payments-team`, `orders-team`) are served by the same
chart; the ApplicationSet derives each service's namespace from `spec.owner`.
