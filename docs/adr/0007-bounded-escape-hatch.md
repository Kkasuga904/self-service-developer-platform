# ADR 0007: Reject raw configuration escape hatches

## Context

A Golden Path must support legitimate variation without becoming a thin wrapper
around arbitrary PodSpecs.

## Decision

Allow only typed, bounded fields. Evaluate repeated needs through contract
changes and ADRs. Do not accept `extraObjects`, arbitrary annotations, raw
security contexts or generic Helm overrides in v1alpha1.

## Alternatives

Raw YAML patches maximize flexibility but bypass invariants and make contract
compatibility meaningless. Per-team charts create drift and duplicate platform
work.

## Consequences

Unusual services may wait for a platform change or choose a separately governed
path. The standard path remains understandable, testable and supportable.
