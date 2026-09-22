# ADR 0004: Use a small YAML Service Definition

## Context

Developers need a durable API that hides repetitive Kubernetes syntax without
concealing application-relevant choices.

## Decision

Define `platform.example.io/v1alpha1 Service` as strict YAML backed by Go types,
semantic validation, JSON Schema and a Helm values schema.

## Alternatives

Raw Helm values expose implementation details. A Kubernetes CRD and controller
would add reconciliation code already supplied by Argo CD. A web portal would
add presentation surface before the API is stable.

## Consequences

The contract is testable and Git-reviewable. v1alpha1 changes require coordinated
updates to CLI, schemas, chart and migration documentation.
