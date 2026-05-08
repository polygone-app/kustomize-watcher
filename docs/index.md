# kustomize-watcher

Watch kustomize directories and apply them to the local Kubernetes cluster — no shell exec, native Go.

## Features

- **Zero external binaries** — uses the [krusty](https://pkg.go.dev/sigs.k8s.io/kustomize/api/krusty) Go API and Kubernetes dynamic client directly
- **Server-side apply** — field-manager aware, conflict-safe
- **File watching** — reacts to any file change in a watched directory within 500 ms
- **Initial reconcile** — applies all discovered apps on startup, before any file event
- **Fault-tolerant** — apply failures are logged and skipped; the watcher never panics
- **Structured logging** — JSON output via `log/slog`, configurable level
- **Multi-arch Docker image** — `linux/amd64` and `linux/arm64` on GHCR
- **Glob discovery** — supports `**` patterns via [doublestar](https://github.com/bmatcuk/doublestar)

## Next steps

- [Getting Started](getting-started.md) — install, configure, and run
- [Configuration](configuration.md) — full config and CLI reference
- [Architecture](architecture.md) — how the watcher works internally
- [Deployment](deployment.md) — running in-cluster with RBAC
