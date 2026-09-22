# Security

## Implemented in Phase 2

- private worker subnets and no inbound node access;
- IMDSv2 required with hop limit 1;
- standard EKS cluster and node IAM policies;
- EKS API authentication plus bootstrap creator access;
- optional GitHub OIDC role restricted to main and `eks:DescribeCluster`;
- AppProject destination and resource-kind allowlist;
- non-root containers, RuntimeDefault seccomp, read-only root filesystem,
  privilege escalation disabled, all capabilities dropped;
- no service-account token mounted into Golden Path Pods;
- no secret values committed.

The EKS public API endpoint is enabled for disposable remote validation, but the
caller must supply trusted operator CIDRs and validation rejects `0.0.0.0/0`.
IAM authentication still applies. EKS 1.28 and later envelope-encrypts all API
data with an AWS-owned KMS key by default; a customer-managed KMS key is omitted
to avoid unjustified cost in this disposable environment. The CNI policy is
attached to the node role in this MVP; a dedicated Pod Identity/IRSA role is a
future hardening item.

## Implemented in Phase 3 (added)

- Kyverno admission (8 ClusterPolicies, Enforce) reusing the exact files CI
  tests; exemptions are namespace-scoped to platform namespaces only;
- per-team `Role`/`RoleBinding` (`team-developer`): read, pod logs and
  port-forward inside the team's own namespace; no workload writes, no
  cross-team access (GitOps owns desired state);
- GitHub OIDC trust hardened from `StringLike` to `StringEquals` on the
  exact custom subject; wildcard-free, guarded by a static CI test;
- ownership labels (`owner`, `service`, `environment`, `managed-by`)
  required by admission, rendered from the contract by the chart.

## Not implemented

NetworkPolicy, secret delivery, image signature verification, vulnerability
admission, dedicated break-glass access, and hostile tenant isolation remain
production considerations. See MULTI_TENANCY.md for why namespaces are an
administrative boundary, not a complete security boundary.

## CI vs Admission

CI gives actionable feedback before merge. Admission is the final protection
against direct API use or a bypassed pipeline. Phase 3 implements both from
the same Kyverno files: `kyverno test` in CI is fast feedback, Enforce in the
cluster is the final boundary. See docs/POLICIES.md.
