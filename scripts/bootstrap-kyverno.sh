#!/usr/bin/env bash
# Installs Kyverno (pinned chart) and applies the platform ClusterPolicies
# from policies/kyverno with validationFailureAction=Enforce as stored.
# The same files are tested in CI with `kyverno test`, so admission enforces
# exactly what CI previews. Run after Argo CD bootstrap; policies exclude
# platform namespaces (kube-system, argocd, kyverno, monitoring).
set -euo pipefail

readonly KYVERNO_CHART_VERSION="3.9.1"
readonly KYVERNO_APP_VERSION="v1.19.1"

kubectl create namespace kyverno --dry-run=client -o yaml | kubectl apply -f -
helm repo add kyverno https://kyverno.github.io/kyverno/
helm repo update kyverno
helm upgrade --install kyverno kyverno/kyverno \
  --namespace kyverno \
  --version "${KYVERNO_CHART_VERSION}" \
  --set admissionController.replicas=1 \
  --set backgroundController.replicas=1 \
  --set reportsController.replicas=1 \
  --set cleanupController.replicas=1 \
  --wait \
  --timeout 10m

installed_app="$(kubectl -n kyverno get deployment kyverno-admission-controller \
  -o jsonpath='{.metadata.labels.app\.kubernetes\.io/version}' 2>/dev/null || true)"
if [[ "${installed_app}" != "${KYVERNO_APP_VERSION}" ]]; then
  echo "expected Kyverno ${KYVERNO_APP_VERSION}, found '${installed_app}'" >&2
  exit 1
fi

kubectl apply -f policies/kyverno/
kubectl get clusterpolicies.kyverno.io
