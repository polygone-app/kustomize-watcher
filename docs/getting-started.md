# Getting Started

## Prerequisites

- A Kubernetes cluster accessible locally:
    - [kind](https://kind.sigs.k8s.io/) — `kind create cluster`
    - [k3s](https://k3s.io/) — `curl -sfL https://get.k3s.io | sh -`
    - [minikube](https://minikube.sigs.k8s.io/) — `minikube start`
    - Any cluster reachable via your `~/.kube/config`
- Go 1.26+ **or** [mise](https://mise.jdx.dev/) (see below)

## Install

=== "GHCR (Docker)"

    ```sh
    docker pull ghcr.io/polygone-app/kustomize-watcher:latest
    ```

=== "Binary"

    Download the pre-built binary from [GitHub Releases](https://github.com/polygone-app/kustomize-watcher/releases), then:

    ```sh
    tar xzf kustomize-watcher_linux_amd64.tar.gz
    sudo mv kustomize-watcher /usr/local/bin/
    ```

=== "From source"

    ```sh
    go install github.com/polygone-app/kustomize-watcher/cmd/kustomize-watcher@latest
    ```

=== "With mise"

    Clone the repo and run:

    ```sh
    mise install
    go build -o kustomize-watcher ./cmd/kustomize-watcher
    ```

## Write a config file

Create `config.yaml` in your working directory:

```yaml
glob: "./apps/*"
log_level: info
```

The `glob` pattern is matched against the filesystem. Any directory containing a `kustomization.yaml` or `kustomization.yml` at its root is added to the watch list.

See [Configuration](configuration.md) for all options and advanced glob patterns.

## Run

```sh
kustomize-watcher --config config.yaml
```

On startup you'll see one log line per discovered directory:

```json
{"time":"...","level":"INFO","msg":"watching","path":"./apps/nginx"}
{"time":"...","level":"INFO","msg":"initial apply","dir":"./apps/nginx"}
```

## Verify

Check that resources were applied:

```sh
kubectl get all -A
```

## Trigger a change

Edit any file inside a watched directory:

```sh
echo "  replicas: 3" >> ./apps/nginx/deployment.yaml
```

After 500 ms:

```json
{"time":"...","level":"INFO","msg":"applying","dir":"./apps/nginx"}
```

The updated resources are applied via server-side apply. Apply failures are logged at `ERROR` level and do not stop the watcher.

## Stop

Send `SIGINT` (Ctrl+C) or `SIGTERM`. The watcher drains in-flight applies and exits cleanly.
