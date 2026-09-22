# ADR 0001: Use EKS as the target runtime

## Context

The portfolio must demonstrate a Kubernetes platform product without repeating
the existing ECS infrastructure portfolio or local reliability lab.

## Decision

Use one disposable EKS cluster as the eventual validation runtime. Treat the
Service Definition and developer journey—not cluster creation—as the product.

## Alternatives

ECS has less operational overhead but does not exercise the Kubernetes GitOps
contract. kind remains the preferred local validation environment but cannot
prove AWS IAM, networking or managed control-plane integration.

## Consequences

Real validation has an hourly cost and requires explicit creation/destruction.
EKS-specific claims remain NOT RUN until observed.
