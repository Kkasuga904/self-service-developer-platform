# Interview guide

This is a map from interview questions to decisions and repository evidence,
not a script to memorize. Say what was actually observed, distinguish design
from validation, and use [VALIDATION.md](VALIDATION.md) and
[FAILURE_EXPERIMENTS.md](FAILURE_EXPERIMENTS.md) as the evidence boundary.

## 1. Why EKS instead of ECS?

- **Key concepts:** portable Kubernetes workload API, admission policy,
  GitOps controllers, ecosystem cost, operational burden.
- **Repository evidence:** [ADR 0001](adr/0001-use-eks-for-platform-contract.md),
  Helm chart, ApplicationSet, Kyverno policies, real EKS lifecycle evidence.
- **Trade-off:** EKS enables the platform-control experiments this project is
  about, but ECS would be simpler for many application-only teams.
- **Follow-up questions:** What organization size justifies EKS? Which EKS
  responsibilities disappear with ECS? When would you reverse the decision?

## 2. Why Terraform for infrastructure but Helm/GitOps for workloads?

- **Key concepts:** ownership boundaries, lifecycle cadence, state, drift,
  declarative reconciliation.
- **Repository evidence:** `infra/` owns VPC/EKS/IAM/ECR;
  `charts/golden-path/` and `gitops/` own Kubernetes workloads; ADR 0003.
- **Trade-off:** two toolchains and bootstrap ordering in exchange for keeping
  high-churn application objects out of Terraform state.
- **Follow-up questions:** Who owns namespaces and CRDs? How do you prevent
  dual ownership? How is bootstrap recovered?

## 3. Why Argo CD?

- **Key concepts:** pull-based deployment, Git audit trail, drift detection,
  reconciliation, credential direction.
- **Repository evidence:** root Application, AppProject, ApplicationSet,
  automated prune/self-heal, recorded drift recovery.
- **Trade-off:** deployment becomes indirect and depends on a shared control
  plane; a Git commit can be accepted even when runtime rollout fails.
- **Follow-up questions:** Why not GitHub Actions pushing with kubectl? How is
  Argo access scoped? What is the disaster-recovery path?

## 4. Why ApplicationSet?

- **Key concepts:** fleet templating, discovery from service definitions,
  uniform destination and project policy.
- **Repository evidence:** `gitops/platform/applicationset.yaml` discovers
  `services/*/*/service.yaml` and derives release, values, and namespace.
- **Trade-off:** concise onboarding but templating errors or an overly broad
  generator affect multiple teams.
- **Follow-up questions:** How do deletions behave? How would multiple
  environments be modeled? What prevents a destination escape?

## 5. Why Kyverno?

- **Key concepts:** Kubernetes-native policy, admission-time enforcement,
  policy-as-code, readable denial feedback.
- **Repository evidence:** eight ClusterPolicies, cases A–E, live admission
  rejections, same policy files in CI and cluster.
- **Trade-off:** another critical webhook/control-plane component; policy scope
  must cover all relevant workload kinds and container classes.
- **Follow-up questions:** Why not Gatekeeper? What happens if the webhook is
  unavailable? How do exemptions and policy upgrades work?

## 6. Why enforce policy in both CI and admission?

- **Key concepts:** fast feedback versus authoritative enforcement,
  non-Git entry points, defense in depth.
- **Repository evidence:** Kyverno CLI tests in CI and Enforce-mode policies in
  EKS; intentional invalid fixtures prove rejection.
- **Trade-off:** duplicated execution surfaces and version-parity risk; sharing
  policy files reduces semantic drift.
- **Follow-up questions:** Which result is authoritative? How is policy version
  pinned? What if CI passes and admission rejects?

## 7. What happens if Argo CD goes down?

- **Key concepts:** management/control plane versus application data plane,
  reconciliation lag, stale status.
- **Repository evidence:** automated sync/self-heal configuration and Phase 4
  Experiment B ledger. Do not claim measured continuity until it is PASS.
- **Trade-off:** existing Kubernetes resources should serve independently, but
  deploys, drift detection, and self-heal stop.
- **Follow-up questions:** Which Argo component was stopped? How was traffic
  measured? What happens to a pending commit after recovery?

## 8. What happens if Prometheus goes down?

- **Key concepts:** service health versus visibility, missing samples, SLI
  unavailability, monitoring the monitor.
- **Repository evidence:** minimal stack and queries in `observability/`, ADR
  0008, and Phase 4 Experiment C ledger.
- **Trade-off:** application should keep serving, but detection and historical
  evidence have a blind spot; current setup has no external alert route.
- **Follow-up questions:** Are missing samples failures or unknown? What data is
  recovered? How would independent monitoring work?

## 9. How does GitOps drift/self-heal work?

- **Key concepts:** desired versus live state, refresh, diff, reconciliation,
  prune, field ownership.
- **Repository evidence:** `automated.prune: true`, `selfHeal: true`, and live
  Phase 2/3 drift experiment in VALIDATION.
- **Trade-off:** strong convergence can undo emergency changes; break-glass
  operations need a documented route back into Git.
- **Follow-up questions:** What drift is ignored? How quickly is it detected?
  What happens while Argo is unavailable?

## 10. How did you debug the EKS/Kyverno webhook networking failure?

- **Key concepts:** API-server-to-node webhook path, security groups, service
  target port 9443, symptom-driven isolation.
- **Repository evidence:** incident record and Terraform security-group rule
  added for the managed control-plane/node path; admission retest afterward.
- **Trade-off:** narrowly opening the actual webhook port restores admission
  without broad node ingress.
- **Follow-up questions:** Why did Pods appear healthy? Which source SG was
  correct? How did you prove the fix rather than bypassing the webhook?

## 11. Why is namespace not a complete security boundary?

- **Key concepts:** shared kernel/nodes/CNI/controllers, cluster-scoped APIs,
  resource contention, RBAC versus network isolation.
- **Repository evidence:** [MULTI_TENANCY.md](MULTI_TENANCY.md), namespaced
  Roles, security contexts, and documented absence of NetworkPolicy/Quota.
- **Trade-off:** sufficient administrative ownership for a personal platform,
  not hostile multi-tenant isolation.
- **Follow-up questions:** What threats cross namespaces? What would dedicated
  clusters solve? What should be added first?

## 12. How is GitHub OIDC restricted?

- **Key concepts:** short-lived federation, audience and subject claims,
  least-privilege IAM, no stored cloud keys.
- **Repository evidence:** `infra/modules/github-oidc`, exact repository/main or
  configured custom subject, STS positive and negative validation.
- **Trade-off:** custom subject templates complicate portability; account-level
  providers require careful ownership.
- **Follow-up questions:** What prevents a fork or feature branch? What can the
  role actually do? How are pull_request subjects handled?

## 13. What does the Platform Contract abstract?

- **Key concepts:** product API, paved road, safe defaults, ownership metadata.
- **Repository evidence:** schema, Go validator, Service definitions, chart;
  generated Deployment/Service/PDB and optional HPA.
- **Trade-off:** developers choose image, port, owner, size, replicas, health,
  and autoscaling bounds; platform owns repetitive Kubernetes structure.
- **Follow-up questions:** How is compatibility versioned? How are defaults
  evolved? How does a developer discover invalid input?

## 14. What does it intentionally NOT abstract?

- **Key concepts:** leaky abstractions, bounded surface, escape hatches.
- **Repository evidence:** no arbitrary PodSpec or `extraObjects`; no database,
  secrets platform, operator, mesh, or progressive delivery.
- **Trade-off:** legitimate novel needs require platform API evolution rather
  than a raw override.
- **Follow-up questions:** What is the exception process? When should a team
  leave the Golden Path? Which next typed field would you add?

## 15. How would you evolve this for 50+ teams?

- **Key concepts:** tenancy tiers, platform SLOs, ownership automation, policy
  rollout, capacity, chargeback, support model.
- **Repository evidence:** current two-team ApplicationSet/RBAC model is a
  small demonstrator, not proof of scale.
- **Trade-off:** first add remote state/locking, real identity integration,
  quotas/network controls, HA controllers, alert routing, staged policy and
  chart rollouts; only then consider broader product features.
- **Follow-up questions:** One cluster or many? Who owns upgrades? How do you
  limit generator and controller blast radius?

## 16. What would you change before production?

- **Key concepts:** reliability, security, operability, lifecycle and support.
- **Repository evidence:** Future Production Extensions in README/TRADE_OFFS;
  limitations explicitly list remote state, NetworkPolicy, ResourceQuota,
  signing, long SLO evaluation and alert routing.
- **Trade-off:** each control adds operational cost; prioritize from threat and
  failure evidence, not a technology checklist.
- **Follow-up questions:** Which three are blocking? What RTO/RPO drives HA?
  How would secrets be handled?

## 17. What did real AWS reveal that local validation did not?

- **Key concepts:** cloud identity, managed control-plane networking, async
  readiness, real controller behavior, cleanup convergence.
- **Repository evidence:** node join/SG and webhook 9443 incidents, OIDC custom
  subject, real admission/GitOps/Prometheus behavior, destroy/residual checks.
- **Trade-off:** disposable cloud validation costs money and time but exposes
  integration failures static/render tests cannot.
- **Follow-up questions:** Which issue was most surprising? Which local test was
  added afterward? What remained unobservable?

## 18. Which parts are validated versus only designed?

- **Key concepts:** evidence levels, repeatability, temporal limits.
- **Repository evidence:** status tables in VALIDATION; Phase 3 infrastructure,
  admission, RBAC, metrics, OIDC and destroy are recorded; Phase 4 rows remain
  NOT RUN until fresh evidence exists; 30-day SLO is only a design target.
- **Trade-off:** destroyed environments make evidence honest but not currently
  re-observable; commands and decisive excerpts preserve reproducibility.
- **Follow-up questions:** Why are some results PARTIAL? What would upgrade
  them? Which claim relies only on structural tests?

## 19. What are the largest blast-radius risks?

- **Key concepts:** shared controllers, cluster-wide policy, generator scope,
  networking, Terraform state and IAM.
- **Repository evidence:** one cluster, shared Argo/Kyverno/Prometheus, cluster
  policies, ApplicationSet across all service definitions, local state.
- **Trade-off:** a small portfolio keeps cost low but concentrates failures;
  namespaces reduce administrative scope, not shared-control-plane risk.
- **Follow-up questions:** How would policy changes be canaried? What protects
  Terraform apply? Which component can block all new workload admissions?

## 20. What did AI generate, and what must the candidate understand?

- **Key concepts:** authorship versus accountability, verification, provenance.
- **Repository evidence:** Git history and validation output show changes and
  tests, not who understood them. State honestly which drafts or commands used
  AI assistance; never imply generated material is independent proof.
- **Trade-off:** AI accelerates scaffolding and review, but the candidate must
  personally explain every trust boundary, Terraform resource, controller
  loop, failure hypothesis, observed result, and recovery decision.
- **Follow-up questions:** Show one AI suggestion you rejected. Reproduce one
  diagnosis without the transcript. Which claim would you withdraw if evidence
  were missing?

## Safe interview claims

- Built a personal, disposable EKS platform prototype around a typed service
  contract, Helm Golden Path, Argo CD, Kyverno, RBAC, and Prometheus/Grafana.
- Ran and documented the specific real-AWS validations marked PASS in
  VALIDATION, including full destroy and residual checks.
- Used layered CI/admission policy and short-lived GitHub OIDC, including
  recorded negative tests.
- Can explain limitations: this is not employer production operation, not
  hostile multi-tenancy, and short runs do not prove a 30-day SLO.

Do not claim production-proven, enterprise-ready, battle-tested, production
Kubernetes operating experience, long-term SLO compliance, or scale beyond the
two-team disposable validation that the evidence supports.
