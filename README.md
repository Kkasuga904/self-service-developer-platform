# Self-Service Developer Platform on EKS

> Phase 2 MVP. Local validation is recorded in
> [docs/VALIDATION.md](docs/VALIDATION.md). No AWS resource has been created and
> no real EKS or Argo CD reconciliation is claimed yet.

This personal portfolio explores a Platform Engineering product: a small,
reviewable contract that lets an application developer request a safe standard
workload without authoring Kubernetes objects directly.

## Problem

Without a platform contract, every developer must repeatedly understand and
assemble Kubernetes Deployments, Services, probes, resource settings, disruption
budgets, security contexts, labels, and deployment tooling. The result is slow
onboarding and inconsistent safety.

## Platform users and product

The user is an application developer. The product is a Golden Path plus a
self-service interface:

```text
small Service Definition -> reviewed Git change -> standard Helm workload -> Argo CD reconciliation
```

The platform absorbs repeatable infrastructure choices. Developers retain the
application image, port, ownership, size, replicas, health paths, and bounded
autoscaling choices.

## Current status

| Capability | Implementation | Validation |
| --- | --- | --- |
| Service Definition v1alpha1 | Implemented | See validation record |
| Go CLI: create-service / validate / doctor | Implemented | See validation record |
| Golden Path Helm chart | Implemented | See validation record |
| Terraform VPC/EKS/IAM/ECR foundation | Implemented as code | Local validation only; AWS NOT RUN |
| Argo CD root Application/ApplicationSet | Implemented as manifests | Rendering only; reconciliation NOT RUN |
| Sample application | Implemented | Local tests/build recorded separately |
| GitHub Actions | Implemented | Hosted workflow NOT RUN until pushed |
| Kyverno admission policy | Not implemented; Phase 3 | NOT RUN |
| Observability stack | Not implemented; Phase 3 | NOT RUN |
| Second team | Not implemented; Phase 3 | NOT RUN |
| Reliability experiments | Not implemented; Phase 4 | NOT RUN |

## Architecture

```mermaid
flowchart LR
  D[Application developer] --> C[Go platform CLI]
  C --> S[Service Definition]
  S --> PR[GitHub pull request]
  PR --> CI[Contract, Go, Terraform, Helm, manifest and security checks]
  CI --> G[main branch desired state]
  G --> AS[Argo CD ApplicationSet]
  AS --> H[Golden Path Helm chart]
  H --> NS[team-payments namespace]
  TF[Terraform] --> AWS[VPC / IAM / EKS]
  AWS --> NS
  AR[Argo CD] --> AS
```

Terraform owns the AWS foundation, including one sample-image ECR repository. Helm and Argo CD own cluster configuration
and workloads. Terraform deliberately does not manage developer Deployments,
Services, HPAs, PDBs, or Argo Applications.

## Developer journey

Build the CLI and create a contract document:

```bash
make build
./platform create-service \
  --name payment-api \
  --owner payments-team \
  --image ghcr.io/acme/payment-api:v1.2.3 \
  --port 8080
./platform validate services/payments-team/payment-api/service.yaml
```

The CLI refuses unknown owners, invalid names, `latest` or untagged images,
invalid sizes, ports, and replica counts. It never overwrites an existing file.
The developer reviews the generated YAML, commits it, and opens a PR. CI renders
the Golden Path. After merge, Argo CD is intended to create one Application per
service and reconcile it; that end-to-end path remains unvalidated until the
explicit EKS validation step.

The committed sample uses `ghcr.io/example/payment-api:v0.1.0` as a contract
example, not as a claimed deployable artifact. Before EKS validation it must be
replaced with the pushed sample image's immutable reference.

## Platform contract

The v1alpha1 contract is intentionally small. Platform-owned defaults include:

- Deployment, ClusterIP Service, PDB and optional HPA;
- CPU/memory requests and limits selected by `small`, `medium`, or `large`;
- readiness and liveness probes;
- owner and managed-by labels;
- non-root execution, RuntimeDefault seccomp, no privilege escalation, all
  Linux capabilities dropped, read-only root filesystem;
- no service-account token mounted into the workload.

Arbitrary PodSpec fragments and `extraObjects` are not supported. A new typed
field is preferred over a raw escape hatch. See
[PLATFORM_CONTRACT.md](docs/PLATFORM_CONTRACT.md).

## Repository structure

```text
cmd/platform/             three-command CLI entrypoint
internal/contract/        Service Definition types and validation
internal/command/         small command implementations
schemas/                  published JSON Schema
charts/golden-path/       standard workload templates
services/                 developer-owned definitions
sample-app/               dependency-free HTTP example
gitops/bootstrap/         imperative Argo CD boundary and root Application
gitops/platform/          AppProject, team namespace, ApplicationSet
infra/                    AWS foundation only
docs/                     decisions, boundaries, runbooks and evidence
```

## Local use

Prerequisites are Go, Terraform, Helm, kubectl, Bash, and optionally Docker and
kubeconform. `platform doctor` reports the core tools.

```bash
make validate
```

The validation script does not contact AWS or create a cluster. Missing optional
tools are explicitly reported rather than silently counted as PASS.

## EKS lifecycle

EKS creation is intentionally not part of CI and has not been run. After a
separate cost and resource review:

```bash
terraform -chdir=infra init
cp infra/terraform.tfvars.example infra/terraform.tfvars
# Replace the documentation-only API CIDR with your current public IP /32.
terraform -chdir=infra plan -out=tfplan
terraform -chdir=infra show tfplan
# terraform -chdir=infra apply tfplan  # only after explicit approval
```

Destruction is plan-first and requires typing `destroy`:

```bash
make destroy
```

## Cost

The design is disposable, not always-on. Principal charges are the EKS control
plane, two `t3.medium` nodes, EBS, one NAT Gateway, public IPv4, CloudWatch and
data processing. One NAT is an explicit cost/availability compromise for this
personal validation environment. The final estimate will be refreshed before
any apply. Leaving this stack running continuously can exceed USD 100/month.

## Key trade-offs

- **EKS vs ECS:** EKS is justified here by the platform API and GitOps control
  plane, not as a universally better scheduler. ECS was covered separately.
- **Terraform vs Helm:** Terraform stops at AWS/EKS; Helm owns repeatable
  workloads. This avoids dual ownership and Terraform state full of workloads.
- **Push vs GitOps:** Git supplies review, audit, rollback and drift visibility,
  at the cost of an indirect deployment path.
- **Abstraction vs flexibility:** the typed API makes the safe path easy but
  does not expose every PodSpec feature.
- **Central control vs autonomy:** teams own application intent; the platform
  owns cluster-wide defaults and allowed destinations.
- **Standardization vs escape hatch:** new typed fields require deliberate
  platform evolution; raw manifest injection is excluded.
- **Security vs velocity:** Phase 2 provides secure generated defaults and CI
  validation. Admission enforcement intentionally waits for Phase 3.

See [TRADE_OFFS.md](docs/TRADE_OFFS.md) for consequences and rejected options.

## Non-goals

This MVP does not provision databases or queues, manage application secrets,
provide a portal, implement a custom operator, run a service mesh, support
hostile multi-tenancy, deliver progressive deployment, or claim production HA.

## Honest scope

This is a personal project, not production operation at an employer. The Go and
sample-app tests, static Terraform checks, Helm rendering, and other commands in
the validation record are the only locally validated claims. EKS creation,
Argo CD reconciliation, GitHub-hosted CI, admission policy, two-team isolation,
observability, SLO measurement and failure scenarios remain unvalidated or
unimplemented as marked above.

It is not described as production-ready.
