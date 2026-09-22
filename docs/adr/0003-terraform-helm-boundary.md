# ADR 0003: Separate Terraform and Helm ownership

## Context

Managing both AWS infrastructure and Kubernetes workloads in Terraform creates
coupled lifecycles and overlapping ownership with GitOps.

## Decision

Terraform owns VPC, EKS and IAM. Helm and Argo CD own namespaced platform and
application resources. Argo CD is installed by one documented bootstrap script.

## Alternatives

The Terraform Helm provider could install everything, but would place cluster
state and workload rollouts behind Terraform. Managing AWS through Kubernetes
controllers would expand the contract beyond the MVP.

## Consequences

There is a small bootstrap discontinuity. Once bootstrapped, Git is the desired
state for workloads and Terraform has no competing ownership.
