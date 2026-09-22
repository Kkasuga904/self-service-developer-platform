# Multi-tenancy

Two fictitious teams share one disposable cluster: `payments-team`
(`team-payments`, service `payment-api`) and `orders-team` (`team-orders`,
service `order-api`). Both services reuse the sample application; the point
is ownership, isolation and contract — not application count.

## What a namespace is here

A team namespace is an **administrative boundary** for:

- ownership: every namespace carries `platform.example.io/owner`, and the
  ApplicationSet derives the destination namespace from `spec.owner`, so a
  service cannot land in another team's namespace through the normal path;
- RBAC scope: each team's `Role`/`RoleBinding` (`team-developer`) lives
  inside its own namespace and grants read + logs + port-forward only.
  Workload writes are withheld because GitOps owns desired state;
- policy scope: Kyverno exemptions and (future) NetworkPolicy/Quota attach
  per namespace;
- blast radius: a bad Service Definition affects one generated Application in
  one namespace (plus the shared control plane, see limits).

## What a namespace is not

A namespace is **not** a complete security boundary. Nodes, kernel, CNI,
kubelet, the EKS control plane and cluster controllers (Argo CD, Kyverno,
monitoring) are shared. A privileged container escape, a cluster-scoped
permission grant, or a noisy neighbor can cross namespaces. This platform
does not claim hostile-tenant isolation; NetworkPolicy, ResourceQuota and
admission already narrow the gap but do not close it.

## RBAC model

| Principal | Scope | Permissions |
| --- | --- | --- |
| `platform-payments-developers` (IdP group) | `team-payments` | get/list/watch Deployments, Services, Pods, HPA, PDB; get/list pod logs; create port-forward |
| `platform-orders-developers` (IdP group) | `team-orders` | same, orders only |
| Platform administrator | cluster | break-glass via IdP-mapped access, outside Git by design |

Group names are placeholders mapped to the corporate identity provider out
of band; no user credentials live in this repository. The negative property
— payments cannot write (or read, beyond what the API server exposes) in
`team-orders` and vice versa — is asserted structurally in Go
(`internal/gitops`: one Role + one RoleBinding per namespace, no write verbs
except port-forward `create`, distinct groups per team) and verified live
with `kubectl auth can-i --as/--as-group` during EKS validation.

## Ownership model

Every service carries `spec.owner`, `spec.environment` and optional
`spec.contact`; the chart renders them as
`platform.example.io/owner|service|environment` labels plus the standard
`app.kubernetes.io/managed-by`, and the ownership-labels Kyverno policy
rejects workloads without them. There is no "unknown owner" state on the
Golden Path.
