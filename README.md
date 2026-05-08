# kustomize-watcher

Watch kustomize directories and apply them to the local Kubernetes cluster — no shell exec, native Go.

## How it works

1. On startup, scans directories matching a configurable glob pattern for `kustomization.yaml`
2. Applies each discovered app to the cluster via server-side apply (SSA)
3. Watches those directories for file changes using `fsnotify`
4. On any change, rebuilds the kustomization with the native [krusty](https://pkg.go.dev/sigs.k8s.io/kustomize/api/krusty) Go API and re-applies via the Kubernetes dynamic client — no `kubectl` or `kustomize` binary required

## Prerequisites

- Go 1.26+ (or `mise install` — see [Development](#development))
- A local Kubernetes cluster (kind, k3s, minikube, or in-cluster)

## Installation

=== "Docker / GHCR"

    ```sh
    docker pull ghcr.io/polygone-app/kustomize-watcher:latest
    ```

=== "Binary"

    Download the latest release for your platform from [GitHub Releases](https://github.com/polygone-app/kustomize-watcher/releases).

=== "From source"

    ```sh
    go install github.com/polygone-app/kustomize-watcher/cmd/kustomize-watcher@latest
    ```

## Quick start

Create a `config.yaml`:

```yaml
glob: "./apps/*"
log_level: info
```

Run the watcher:

```sh
kustomize-watcher --config config.yaml
```

The watcher will immediately apply all discovered kustomize apps, then watch for changes and re-apply automatically.

## CLI reference

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--config` | `KUSTOMIZE_WATCHER_CONFIG` | `config.yaml` | Path to the config file |
| `--log-level` | `KUSTOMIZE_WATCHER_LOG_LEVEL` | *(from config)* | Override log level |

Override precedence: `--flag` > env var > config file > built-in default (`info`).

## Config reference

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `glob` | string | yes | — | Glob pattern for directories to scan. Supports `**` for recursive matching. |
| `log_level` | string | no | `info` | Log verbosity: `debug`, `info`, `warn`, `error` |

Example patterns:

```yaml
glob: "./apps/*"          # one level under apps/
glob: "./apps/**"         # all nested directories under apps/
glob: "/manifests/*/base" # fixed-depth absolute path
```

## Log format

All output is JSON (via `log/slog`), written to stdout. Example:

```json
{"time":"2026-05-08T10:00:00Z","level":"INFO","msg":"watching","path":"./apps/nginx"}
{"time":"2026-05-08T10:00:01Z","level":"INFO","msg":"initial apply","dir":"./apps/nginx"}
{"time":"2026-05-08T10:05:32Z","level":"INFO","msg":"applying","dir":"./apps/nginx"}
{"time":"2026-05-08T10:05:32Z","level":"ERROR","msg":"apply resource","name":"nginx","kind":"Deployment","err":"..."}
```

Apply failures are logged at ERROR level and do not stop the watcher.

## Development

This project uses [mise](https://mise.jdx.dev/) to manage tooling.

```sh
mise install          # installs go, goreleaser, properdocs
go build ./...
go test -race ./...

# Local docs preview
properdocs serve --config-file properdocs.yaml
```

### Project layout

```
cmd/kustomize-watcher/   entry point
internal/
  config/               config loading and log-level parsing
  applier/              krusty build + dynamic client SSA apply
  watcher/              fsnotify-based directory watching + debounce
```

## License

MIT
