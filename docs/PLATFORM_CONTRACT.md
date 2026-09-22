# Platform Contract

## v1alpha1 fields

| Field | Owner | Constraint |
| --- | --- | --- |
| `metadata.name` | Developer | Kubernetes DNS label |
| `spec.owner` | Developer | Phase 2: `payments-team` only |
| `spec.image` | Developer | digest or non-`latest` tag |
| `spec.port` | Developer | 1–65535 |
| `spec.resources.size` | Developer | small / medium / large |
| `spec.availability.replicas` | Developer | 2–10 |
| `spec.observability.enabled` | Developer | annotation only in Phase 2 |
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
evaluated as a typed field with a safe bounded domain. Time-limited policy
exceptions are a Phase 3 design topic, not an implemented capability.

## Responsibilities

The Platform team owns template safety, reconciliation, contract compatibility
and platform component operation. The application team owns image correctness,
application behavior, health endpoint semantics, application telemetry and the
accuracy of its declared owner.
