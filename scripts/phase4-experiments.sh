#!/usr/bin/env bash
set -euo pipefail

# Phase 4 failure experiments for the disposable validation cluster.
# Run one action at a time. This script deliberately does not create Git
# commits: Experiment A must travel through the same reviewed GitOps path as a
# developer change.

readonly EXPECTED_CLUSTER="${EXPECTED_CLUSTER:-developer-platform-dev}"
readonly ARGO_NAMESPACE="argocd"
readonly MONITORING_NAMESPACE="monitoring"
readonly PROMETHEUS_STS="prometheus-kube-prometheus-stack-prometheus"

usage() {
  cat <<'EOF'
Usage: scripts/phase4-experiments.sh ACTION

Actions:
  baseline              Require a healthy pre-experiment baseline
  observe-bad-deploy    Capture Experiment A rollout and service signals
  argocd-down           Scale the Argo CD application controller to zero
  argocd-observe        Observe data plane and stalled reconciliation
  argocd-up             Restore the saved Argo CD replica count
  observability-down    Scale Prometheus to zero
  observability-observe Prove the app serves while monitoring is unavailable
  observability-up      Restore the saved Prometheus replica count
  rbac                  Run positive and negative team-boundary checks

Set AWS_PROFILE and AWS_REGION before use. Each mutating action verifies that
the kube context names EXPECTED_CLUSTER. State needed for recovery is stored
under .phase4-state/ (gitignored).
EOF
}

require_tools() {
  local tool
  for tool in kubectl curl; do
    command -v "${tool}" >/dev/null 2>&1 || {
      echo "required executable not found: ${tool}" >&2
      exit 1
    }
  done
}

require_expected_context() {
  local context
  context="$(kubectl config current-context)"
  if [[ "${context}" != *"${EXPECTED_CLUSTER}"* ]]; then
    echo "refusing action: context '${context}' does not contain '${EXPECTED_CLUSTER}'" >&2
    exit 1
  fi
}

timestamp() { date -u +'%Y-%m-%dT%H:%M:%SZ'; }

http_probe() {
  local namespace="$1" service="$2" port="$3"
  kubectl -n "${namespace}" port-forward "service/${service}" "${port}:80" \
    >".phase4-state/${service}-port-forward.log" 2>&1 &
  local pf_pid=$!
  trap 'kill ${pf_pid} 2>/dev/null || true' RETURN
  sleep 2
  curl --silent --show-error --fail --max-time 5 \
    --write-out ' HTTP %{http_code}\n' "http://127.0.0.1:${port}/healthz"
  kill "${pf_pid}" 2>/dev/null || true
  wait "${pf_pid}" 2>/dev/null || true
  trap - RETURN
}

app_state() {
  kubectl -n team-payments get deployment,replicaset,pod,service -l platform.example.io/service=payment-api -o wide
  kubectl -n team-orders get deployment,pod,service -l platform.example.io/service=order-api -o wide
  kubectl -n team-payments get events --sort-by=.lastTimestamp | tail -n 20
  kubectl -n argocd get applications.argoproj.io \
    -o custom-columns='NAME:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision'
}

baseline() {
  require_expected_context
  echo "baseline timestamp: $(timestamp)"
  aws eks describe-cluster --name "${EXPECTED_CLUSTER}" \
    --query 'cluster.[name,status,version]' --output table
  kubectl wait --for=condition=Ready nodes --all --timeout=2m
  kubectl -n argocd wait --for=condition=Available deployment --all --timeout=3m
  kubectl -n kyverno wait --for=condition=Available deployment --all --timeout=3m
  kubectl -n monitoring wait --for=condition=Ready pod \
    -l app.kubernetes.io/name=prometheus --timeout=3m
  kubectl -n team-payments wait --for=condition=Available deployment/payment-api --timeout=3m
  kubectl -n team-orders wait --for=condition=Available deployment/order-api --timeout=3m
  app_state
  http_probe team-payments payment-api 18080
  http_probe team-orders order-api 18081
  kubectl auth can-i get pods -n team-payments \
    --as=phase4-payments-user --as-group=platform-payments-developers | grep -Fx yes
  kubectl auth can-i patch deployments -n team-orders \
    --as=phase4-payments-user --as-group=platform-payments-developers | grep -Fx no
}

observe_bad_deploy() {
  require_expected_context
  echo "bad-deploy observation timestamp: $(timestamp)"
  app_state
  kubectl -n team-payments rollout status deployment/payment-api --timeout=90s || true
  kubectl -n team-payments describe deployment payment-api
  http_probe team-payments payment-api 18080
}

argocd_down() {
  require_expected_context
  mkdir -p .phase4-state
  local deployment="argocd-application-controller"
  local replicas
  replicas="$(kubectl -n "${ARGO_NAMESPACE}" get statefulset "${deployment}" -o jsonpath='{.spec.replicas}')"
  printf '%s\n' "${replicas}" >.phase4-state/argocd-controller-replicas
  kubectl -n "${ARGO_NAMESPACE}" scale statefulset "${deployment}" --replicas=0
  kubectl -n "${ARGO_NAMESPACE}" wait --for=delete pod \
    -l app.kubernetes.io/name=argocd-application-controller --timeout=2m || true
  echo "injected at $(timestamp); saved replicas=${replicas}"
}

argocd_observe() {
  require_expected_context
  echo "Argo CD failure observation timestamp: $(timestamp)"
  kubectl -n "${ARGO_NAMESPACE}" get statefulset,pod -l app.kubernetes.io/name=argocd-application-controller
  app_state
  http_probe team-payments payment-api 18080
}

argocd_up() {
  require_expected_context
  local state=.phase4-state/argocd-controller-replicas
  [[ -f "${state}" ]] || { echo "missing ${state}; refusing to guess replica count" >&2; exit 1; }
  local replicas
  replicas="$(<"${state}")"
  kubectl -n "${ARGO_NAMESPACE}" scale statefulset argocd-application-controller --replicas="${replicas}"
  kubectl -n "${ARGO_NAMESPACE}" rollout status statefulset/argocd-application-controller --timeout=5m
  kubectl -n "${ARGO_NAMESPACE}" get applications.argoproj.io
  echo "recovered at $(timestamp)"
}

observability_down() {
  require_expected_context
  mkdir -p .phase4-state
  local replicas
  replicas="$(kubectl -n "${MONITORING_NAMESPACE}" get statefulset "${PROMETHEUS_STS}" -o jsonpath='{.spec.replicas}')"
  printf '%s\n' "${replicas}" >.phase4-state/prometheus-replicas
  kubectl -n "${MONITORING_NAMESPACE}" scale statefulset "${PROMETHEUS_STS}" --replicas=0
  kubectl -n "${MONITORING_NAMESPACE}" wait --for=delete pod \
    -l app.kubernetes.io/name=prometheus --timeout=2m || true
  echo "injected at $(timestamp); saved replicas=${replicas}"
}

observability_observe() {
  require_expected_context
  echo "observability failure observation timestamp: $(timestamp)"
  kubectl -n "${MONITORING_NAMESPACE}" get statefulset,pod,service \
    -l app.kubernetes.io/name=prometheus
  http_probe team-payments payment-api 18080
  kubectl -n team-payments get deployment,pod -l platform.example.io/service=payment-api
}

observability_up() {
  require_expected_context
  local state=.phase4-state/prometheus-replicas
  [[ -f "${state}" ]] || { echo "missing ${state}; refusing to guess replica count" >&2; exit 1; }
  local replicas
  replicas="$(<"${state}")"
  kubectl -n "${MONITORING_NAMESPACE}" scale statefulset "${PROMETHEUS_STS}" --replicas="${replicas}"
  kubectl -n "${MONITORING_NAMESPACE}" rollout status statefulset/"${PROMETHEUS_STS}" --timeout=5m
  kubectl -n "${MONITORING_NAMESPACE}" get pod -l app.kubernetes.io/name=prometheus
  echo "recovered at $(timestamp)"
}

rbac() {
  require_expected_context
  local as=(--as=phase4-payments-user --as-group=platform-payments-developers)
  kubectl auth can-i get pods -n team-payments "${as[@]}" | grep -Fx yes
  kubectl auth can-i get pods/log -n team-payments "${as[@]}" | grep -Fx yes
  kubectl auth can-i create pods/portforward -n team-payments "${as[@]}" | grep -Fx yes
  kubectl auth can-i get pods -n team-orders "${as[@]}" | grep -Fx no
  kubectl auth can-i patch deployments -n team-orders "${as[@]}" | grep -Fx no
  kubectl auth can-i create clusterroles "${as[@]}" | grep -Fx no

  # Safe real API negative validation: the request is authenticated as the
  # payments group and rejected before mutation. --dry-run=server is an
  # additional guard, not the authorization mechanism being tested.
  if kubectl -n team-orders patch deployment order-api --type=merge \
    -p '{"metadata":{"annotations":{"phase4.invalid":"must-not-apply"}}}' \
    --dry-run=server "${as[@]}"; then
    echo "unexpected authorization success" >&2
    exit 1
  fi
  echo "RBAC boundary observed at $(timestamp)"
}

require_tools
mkdir -p .phase4-state
case "${1:-}" in
  baseline) baseline ;;
  observe-bad-deploy) observe_bad_deploy ;;
  argocd-down) argocd_down ;;
  argocd-observe) argocd_observe ;;
  argocd-up) argocd_up ;;
  observability-down) observability_down ;;
  observability-observe) observability_observe ;;
  observability-up) observability_up ;;
  rbac) rbac ;;
  *) usage; exit 2 ;;
esac
