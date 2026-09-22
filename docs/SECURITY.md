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

## Not implemented

Kyverno admission rules, team RBAC, NetworkPolicy, secret delivery, image
signature verification, vulnerability admission, dedicated break-glass access,
and hostile tenant isolation are Phase 3 or production considerations.

## CI vs Admission

CI gives actionable feedback before merge. Admission is the final protection
against direct API use or a bypassed pipeline. Phase 2 has contract validation
and secure generated defaults, but no admission controller; this limitation is
explicit rather than implied away.
