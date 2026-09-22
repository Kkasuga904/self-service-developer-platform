# ADR 0005: Stage policy enforcement

## Context

CI provides early feedback but cannot prevent direct cluster writes. Admission
provides enforcement but later feedback and another component to operate.

## Decision

Phase 2 implements strict contract validation and secure generated defaults.
Phase 3 will use Kyverno policies in both CI and Admission after policy tests and
failure messages are designed.

## Alternatives

Custom validation alone cannot protect the Kubernetes API. Gatekeeper is capable
but would introduce Rego for a small policy set. Installing Kyverno now would
violate the approved Phase 2 scope.

## Consequences

Phase 2 is explicitly not protected against cluster-side bypass. This limitation
is documented rather than presenting generated defaults as admission control.
