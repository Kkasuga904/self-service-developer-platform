# ADR 0002: Use Argo CD ApplicationSet

## Context

Each Service Definition needs an independently visible deployment without asking
developers to author Argo Application resources.

## Decision

Use a Git file generator to create one Application per service definition. Use a
separate root Application for platform-owned resources.

## Alternatives

App of Apps would require a repeated child Application file. A custom controller
would add code and operational burden without an unmet capability.

## Consequences

ApplicationSet template correctness becomes platform-owned. Each service has a
separate sync/health view and a bad service should not block unrelated services.
