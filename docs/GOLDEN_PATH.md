# Golden Path

## What it creates

The Phase 2 chart emits one Deployment, one ClusterIP Service, one
PodDisruptionBudget, and an HPA only when explicitly enabled. It standardizes
labels, probes, rolling update settings, security context and resource profiles.

## Why profiles instead of raw resources

`small`, `medium`, and `large` keep the developer input simple and make policy
review tractable. They are starter values, not measured production sizing.
Applications still require measurement and profile evolution.

## Deliberate omissions

Ingress, secrets, persistent volumes, cloud dependencies, ServiceMonitor,
NetworkPolicy and policy exceptions are not Phase 2 chart features. Adding them
without a concrete product requirement would make the contract broad but less
coherent.
