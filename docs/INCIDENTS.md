# Unexpected incidents

This file is for unplanned failures encountered while building or validating
the platform. Deliberate Phase 4 injections are kept separately in
[FAILURE_EXPERIMENTS.md](FAILURE_EXPERIMENTS.md); the two evidence classes must
not be mixed.

Unexpected incidents already preserved in the detailed validation record
include EKS node join/security-group routing, Argo CD source files hidden by
`.gitignore`, an invalid Trivy Action tag, GitHub OIDC custom-subject mismatch,
a stale Terraform plan, Kyverno webhook access on port 9443, bootstrap version
validation, and CRD client dry-run discovery behavior. See the Phase 2/3
incident entries in [VALIDATION.md](VALIDATION.md) for symptoms, diagnosis,
fixes, and rerun evidence.

At the start of Phase 4 on 2026-09-23 JST, the `portfolio` AWS SSO session was
expired. This was an access prerequisite, not an injected platform failure; no
AWS mutation occurred before authentication was restored.

Other Phase 4 setup incidents were also outside the intentional experiments:

- Docker Desktop initially crashed because a local `sailor-ingest.sock` could
  not be renamed. Workspace policy correctly prevented deletion outside the
  repository. A temporary pinned Go image builder pushed the bootstrap image;
  Docker later recovered and the final Docker build plus `/healthz` passed.
- The Linux bash runtime had an isolated HOME and no Windows kubeconfig, so the
  first Argo namespace apply failed against localhost before any cluster
  mutation. The same pinned bootstrap commands were then run with Windows
  kubectl/Helm and succeeded.
- The Resource Groups Tagging API retained a deleted NAT ARN after direct NAT
  API and every other residual check returned empty. Residual status is kept
  PARTIAL rather than treating the tag-index tombstone as a live resource.
