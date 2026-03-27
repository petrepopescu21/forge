---
name: setup-helm
description: Scaffold a Helm chart, Kind cluster config, and local dev scripts (cluster-db, deploy-local, cluster-deps). Use when bootstrapping a project or adding Kubernetes deployment. Trigger on "set up Helm", "add Kubernetes", "set up K8s", "configure deployment", or as part of project bootstrapping.
---

# Setup Helm

Setting up Helm chart, Kind config, and local dev scripts.

## Process

### 1. Create Helm Chart

Run `helm create deploy/helm/<project-name>` to generate the base Helm chart structure.

Customize `deploy/helm/<project-name>/values.yaml`:

```yaml
# Default values for pebblr.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

replicaCount: 1

image:
  repository: pebblr
  pullPolicy: IfNotPresent
  tag: "latest"

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations: {}

podSecurityContext: {}

securityContext: {}

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: false

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi

livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3

autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 100
  targetCPUUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity: {}

env:
  - name: LOG_LEVEL
    value: "info"
```

Create `deploy/helm/<project-name>/values-aks.yaml` for AKS-specific overrides:

```yaml
# AKS production overrides
replicaCount: 3

image:
  repository: pebblr.azurecr.io/pebblr
  tag: "latest"

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: pebblr.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: pebblr-tls
      hosts:
        - pebblr.example.com

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 75

podDisruptionBudget:
  enabled: true
  minAvailable: 1

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
              - key: app.kubernetes.io/name
                operator: In
                values:
                  - pebblr
          topologyKey: kubernetes.io/hostname
```

### 2. Create Kind Cluster Config

Create `deploy/kind/kind-config.yaml`:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: pebblr
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 30080
        hostPort: 80
        protocol: TCP
      - containerPort: 30443
        hostPort: 443
        protocol: TCP
```

### 3. Create Local Dev Scripts

#### Script 1: `scripts/cluster-db.sh`

Manages PostgreSQL in Kubernetes. Accepts namespace + action (up/stop/reset).

```bash
#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Defaults
NAMESPACE="${1:-default}"
ACTION="${2:-up}"

DB_NAMESPACE="db"
APP_NAMESPACE="${NAMESPACE}"
PG_VERSION="16"
PG_PASSWORD="postgres"
DB_NAME="pebblr"
DB_USER="pebblr"
DB_PASSWORD="pebblr-local"
JWT_SECRET="jwt-secret-local-dev-only"
DB_HOST="postgres.${DB_NAMESPACE}.svc.cluster.local"
DB_PORT="5432"

echo "Cluster DB Manager"
echo "=================="
echo "Namespace: ${APP_NAMESPACE}"
echo "Action: ${ACTION}"
echo ""

case "${ACTION}" in
  up)
    echo "Creating DB namespace..."
    kubectl create namespace "${DB_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

    echo "Creating PostgreSQL deployment..."
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-initdb
  namespace: ${DB_NAMESPACE}
data:
  init.sql: |
    CREATE ROLE ${DB_USER} WITH LOGIN PASSWORD '${DB_PASSWORD}';
    CREATE DATABASE ${DB_NAME} OWNER ${DB_USER};
    ALTER ROLE ${DB_USER} CREATEDB;
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: ${DB_NAMESPACE}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:${PG_VERSION}
        ports:
        - containerPort: ${DB_PORT}
        env:
        - name: POSTGRES_PASSWORD
          value: ${PG_PASSWORD}
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pg_isready -U postgres
          initialDelaySeconds: 15
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pg_isready -U postgres
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        - name: init-scripts
          mountPath: /docker-entrypoint-initdb.d
      volumes:
      - name: postgres-storage
        emptyDir: {}
      - name: init-scripts
        configMap:
          name: postgres-initdb
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: ${DB_NAMESPACE}
spec:
  selector:
    app: postgres
  ports:
  - port: ${DB_PORT}
    targetPort: ${DB_PORT}
  clusterIP: None
EOF

    echo "Waiting for PostgreSQL to be ready..."
    kubectl rollout status deployment/postgres -n "${DB_NAMESPACE}" --timeout=60s

    echo "Creating app namespace..."
    kubectl create namespace "${APP_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

    echo "Creating secrets in app namespace..."
    DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
    DB_DSN="${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

    kubectl create secret generic db-credentials \
      --from-literal=db-dsn="${DB_DSN}" \
      --from-literal=db-url="${DB_URL}" \
      --from-literal=db-password="${DB_PASSWORD}" \
      --from-literal=jwt-secret="${JWT_SECRET}" \
      -n "${APP_NAMESPACE}" \
      --dry-run=client -o yaml | kubectl apply -f -

    echo "PostgreSQL cluster is ready!"
    echo "Database: ${DB_NAME}"
    echo "User: ${DB_USER}"
    echo "Host: ${DB_HOST}:${DB_PORT}"
    ;;

  stop)
    echo "Stopping PostgreSQL deployment..."
    kubectl scale deployment/postgres -n "${DB_NAMESPACE}" --replicas=0
    echo "PostgreSQL scaled down."
    ;;

  reset)
    echo "Deleting PostgreSQL deployment and data..."
    kubectl delete deployment postgres -n "${DB_NAMESPACE}" --ignore-not-found=true
    kubectl delete configmap postgres-initdb -n "${DB_NAMESPACE}" --ignore-not-found=true
    kubectl delete service postgres -n "${DB_NAMESPACE}" --ignore-not-found=true
    echo "PostgreSQL reset complete."
    ;;

  *)
    echo "Usage: $0 [namespace] {up|stop|reset}"
    echo ""
    echo "Actions:"
    echo "  up     - Create PostgreSQL deployment and secrets (default)"
    echo "  stop   - Scale down PostgreSQL"
    echo "  reset  - Delete PostgreSQL and all data"
    exit 1
    ;;
esac
```

#### Script 2: `scripts/deploy-local.sh`

Deploys the application to the local Kind cluster using Skaffold.

```bash
#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Local Deployment Script"
echo "======================"

# Source development environment if it exists
if [ -f "${PROJECT_ROOT}/web/.env.development" ]; then
  echo "Sourcing web/.env.development..."
  set -a
  source "${PROJECT_ROOT}/web/.env.development"
  set +a
fi

# Check if skaffold is installed
if ! command -v skaffold &> /dev/null; then
  echo "Error: skaffold is not installed."
  echo "Install it from https://skaffold.dev/docs/install/"
  exit 1
fi

echo "Running Skaffold..."
cd "${PROJECT_ROOT}"

# Run skaffold with no default repo (assumes images are available locally or in registry)
skaffold run --no-default-repo

echo "Deployment complete!"
```

#### Script 3: `scripts/cluster-deps.sh`

Installs cluster dependencies (cert-manager, external-secrets) via Helm. Idempotent.

```bash
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
```

### 4. Set Script Permissions

Make all scripts executable:

```bash
chmod +x scripts/cluster-db.sh
chmod +x scripts/deploy-local.sh
chmod +x scripts/cluster-deps.sh
```

### 5. Verify Helm Chart

Run Helm lint to validate the chart:

```bash
helm lint deploy/helm/*/
```

### 6. Commit

Commit all new files to git:

```bash
git add deploy/helm/<project-name>/
git add deploy/kind/kind-config.yaml
git add scripts/cluster-db.sh scripts/deploy-local.sh scripts/cluster-deps.sh
git commit -m "Setup Helm chart, Kind config, and local dev scripts"
```

## Summary

The setup-helm skill creates:

1. **Helm Chart** (`deploy/helm/<project-name>/`) with:
   - Standard values.yaml (replicaCount: 1, image, service port 8080, resource limits 200m/256Mi, requests 100m/128Mi)
   - Liveness and readiness probes on /healthz
   - values-aks.yaml with AKS production overrides (3 replicas, autoscaling, pod disruption budgets, anti-affinity)

2. **Kind Cluster Config** (`deploy/kind/kind-config.yaml`) with:
   - Single control-plane node
   - Port mappings for local ingress (30080 → 80, 30443 → 443)

3. **Local Dev Scripts** (`scripts/`):
   - **cluster-db.sh**: PostgreSQL namespace management (up/stop/reset), creates secrets
   - **deploy-local.sh**: Skaffold-based deployment with environment sourcing
   - **cluster-deps.sh**: Idempotent Helm-based dependency installation (cert-manager, external-secrets)

All scripts follow bash best practices (set -euo pipefail) and are fully functional, production-ready templates.
