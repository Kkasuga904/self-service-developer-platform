# ADR 0006: Use team namespaces as an administrative boundary

## Context

Teams need clear ownership, resource destinations and blast-radius controls.

## Decision

Map an approved owner to a team namespace and restrict Argo Projects to approved
destinations. Phase 2 includes only payments-team; the second team and its RBAC,
quota and network policy belong to Phase 3.

## Alternatives

A namespace per service provides finer quotas but more policy objects. A cluster
per team provides stronger isolation at materially higher cost and operation.

## Consequences

Multiple services share a team namespace. A namespace is not claimed as a hard
security boundary because nodes, kernel, CNI and cluster controllers are shared.
