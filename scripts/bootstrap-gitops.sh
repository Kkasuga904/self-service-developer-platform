#!/usr/bin/env bash
set -euo pipefail

readonly ARGO_CD_CHART_VERSION="8.3.0"

kubectl apply -f gitops/bootstrap/namespace.yaml
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update argo
helm upgrade --install argocd argo/argo-cd \
  --namespace argocd \
  --version "${ARGO_CD_CHART_VERSION}" \
  --wait \
  --timeout 10m
kubectl apply -f gitops/bootstrap/root-application.yaml
