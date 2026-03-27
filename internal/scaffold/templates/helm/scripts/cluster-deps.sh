#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

CERT_MANAGER_VERSION="v1.13.0"
EXTERNAL_SECRETS_VERSION="0.9.8"

echo "Cluster Dependencies Manager"
echo "============================"

# Function to check if a release is installed
helm_release_exists() {
  local release=$1
  local namespace=$2
  helm list -n "${namespace}" | grep -q "^${release}" && return 0 || return 1
}

# Install cert-manager if not already installed
if helm_release_exists "cert-manager" "cert-manager"; then
  echo "cert-manager is already installed, skipping..."
else
  echo "Installing cert-manager ${CERT_MANAGER_VERSION}..."
  helm repo add jetstack https://charts.jetstack.io --force-update
  helm repo update

  kubectl create namespace cert-manager --dry-run=client -o yaml | kubectl apply -f -

  helm install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --version "${CERT_MANAGER_VERSION}" \
    --set installCRDs=true \
    --wait

  echo "cert-manager installed successfully."
fi

# Install external-secrets if not already installed
if helm_release_exists "external-secrets" "external-secrets-system"; then
  echo "external-secrets is already installed, skipping..."
else
  echo "Installing external-secrets ${EXTERNAL_SECRETS_VERSION}..."
  helm repo add external-secrets https://charts.external-secrets.io --force-update
  helm repo update

  kubectl create namespace external-secrets-system --dry-run=client -o yaml | kubectl apply -f -

  helm install external-secrets external-secrets/external-secrets \
    --namespace external-secrets-system \
    --version "${EXTERNAL_SECRETS_VERSION}" \
    --wait

  echo "external-secrets installed successfully."
fi

echo "All cluster dependencies are ready!"
