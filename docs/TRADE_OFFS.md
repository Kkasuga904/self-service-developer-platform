# Trade-offs

## EKS instead of ECS

EKS exposes Kubernetes and GitOps extension points appropriate to this
portfolio's Platform Contract. It also adds cost and operational surface. ECS is
the better choice for teams that do not need this platform ecosystem.

## Terraform and Helm boundary

Terraform owns AWS resources whose lifecycle and identity live in AWS. Helm and
Argo CD own Kubernetes desired state. The cost is a documented bootstrap step;
the benefit is one owner per resource and visible GitOps drift.

## GitOps instead of push deployment

GitOps preserves review and audit history and continuously compares desired and
actual state. It introduces reconciliation delay and makes Argo CD availability
part of deployment operations.

## Abstraction and escape hatches

A narrow typed API gives a consistent, testable Golden Path. It cannot express
all Kubernetes behavior. Raw overrides are rejected because they turn the
contract into an unreviewable pass-through. The platform should add a bounded
field only after repeated evidence of a legitimate need.

## Central control and autonomy

The platform centrally owns resource profiles and security settings. Teams own
images and application behavior. This slows unusual workloads but prevents each
team from rebuilding undifferentiated deployment machinery.

## Security and velocity

Early CI feedback is valuable but cannot protect the cluster from bypasses.
Phase 3 adds Kyverno admission reusing the exact CI-tested files, so the two
layers cannot drift. Legacy `ClusterPolicy` (rather than CEL
`ValidatingPolicy`) is used because Pod autogen covers Deployments without
extra plumbing for eight rules; upstream deprecates it in v1.19, so migration
to `ValidatingPolicy` is the recorded follow-up, not ignored debt.

## Disposable cost and availability

One NAT Gateway serves two private worker subnets. This reduces portfolio cost
but creates a cross-AZ dependency and is not the recommended production HA
shape. Two On-Demand nodes favor predictable validation over Spot interruption
experiments.
