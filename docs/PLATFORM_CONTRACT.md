# Platform Contract

## v1alpha1 fields

| Field | Owner | Constraint |
| --- | --- | --- |
| `metadata.name` | Developer | Kubernetes DNS label |
| `spec.owner` | Developer | `payments-team` or `orders-team` (approved set) |
| `spec.environment` | Developer | `dev` / `staging` / `prod` |
| `spec.contact` | Developer | optional ownership identifier (≤128 chars) |
| `spec.image` | Developer | digest or non-`latest` tag |
| `spec.port` | Developer | 1–65535 |
| `spec.resources.size` | Developer | small / medium / large |
| `spec.availability.replicas` | Developer | 2–10 |
| `spec.observability.enabled` | Developer | enables `prometheus.io/*` scrape annotations consumed by the platform stack |
| `spec.health.*Path` | Developer | optional HTTP paths |
| `spec.autoscaling` | Developer | optional, bounded values |
| generated Pod security/resources | Platform | not overridable |

The Go decoder rejects unknown YAML fields. JSON Schema supports editor and CI
integration. Helm also has a values schema, while semantic rules such as banning
`latest` are enforced by the Go validator.

## Compatibility

`v1alpha1` indicates that fields can change before a stable version. Breaking
changes require an ADR, migration notes and coordinated CLI/chart/schema updates.

## Escape hatch

No raw PodSpec or arbitrary manifest injection exists. A requested exception is
evaluated as a typed field with a safe bounded domain. Time-limited
PolicyExceptions remain unevaluated by design: every exemption must be a
namespace-scoped platform decision (see POLICIES.md), not a developer
self-grant.

## Responsibilities

The Platform team owns template safety, reconciliation, contract compatibility
and platform component operation. The application team owns image correctness,
application behavior, health endpoint semantics, application telemetry and the
accuracy of its declared owner.
