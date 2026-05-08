# Deployment

## Running locally (kubeconfig)

The watcher falls back to your `~/.kube/config` (or `KUBECONFIG`) when not running inside a pod. This is the simplest way to test locally:

```sh
kustomize-watcher --config config.yaml
```

## Running in-cluster

### Docker image

```
ghcr.io/polygone-app/kustomize-watcher:latest
```

Multi-arch manifest: `linux/amd64` and `linux/arm64`.

### Config as a ConfigMap

Mount the config file into the pod via a ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kustomize-watcher-config
  namespace: kustomize-watcher
data:
  config.yaml: |
    glob: /manifests/*
    log_level: info
```

### RBAC

The watcher needs permission to read, create, patch, and delete any resource type present in your kustomize apps. The exact set depends on what you deploy.

!!! warning "Security note"
    `cluster-admin` is shown below for simplicity. In production, restrict the `ClusterRole` rules to only the API groups and resource types your apps actually use.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kustomize-watcher
  namespace: kustomize-watcher
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kustomize-watcher
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["get", "list", "create", "update", "patch", "delete", "apply"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kustomize-watcher
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kustomize-watcher
subjects:
  - kind: ServiceAccount
    name: kustomize-watcher
    namespace: kustomize-watcher
```

### Deployment

Mount your manifests directory as a volume (e.g. from a git-sync sidecar or a PVC) and the config as a ConfigMap:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kustomize-watcher
  namespace: kustomize-watcher
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kustomize-watcher
  template:
    metadata:
      labels:
        app: kustomize-watcher
    spec:
      serviceAccountName: kustomize-watcher
      containers:
        - name: kustomize-watcher
          image: ghcr.io/polygone-app/kustomize-watcher:latest
          args:
            - --config
            - /config/config.yaml
          volumeMounts:
            - name: config
              mountPath: /config
            - name: manifests
              mountPath: /manifests
      volumes:
        - name: config
          configMap:
            name: kustomize-watcher-config
        - name: manifests
          # Replace with your actual manifests source (gitRepo, PVC, etc.)
          emptyDir: {}
```

### Namespace

Create the namespace first:

```sh
kubectl create namespace kustomize-watcher
```

Then apply all manifests:

```sh
kubectl apply -f rbac.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
```

### Verify

```sh
kubectl logs -n kustomize-watcher deploy/kustomize-watcher -f
```

Expected output on a healthy start:

```json
{"level":"INFO","msg":"watching","path":"/manifests/nginx"}
{"level":"INFO","msg":"initial apply","dir":"/manifests/nginx"}
```

## Minimal RBAC (known resource types)

If you know exactly which resource types your kustomize apps create, restrict the ClusterRole:

```yaml
rules:
  - apiGroups: [""]
    resources: ["configmaps", "services", "serviceaccounts"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
```
