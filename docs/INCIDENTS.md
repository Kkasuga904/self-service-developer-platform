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
