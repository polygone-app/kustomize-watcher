# Configuration

## Config file

The config file is YAML. By default the watcher looks for `config.yaml` in the working directory.

```yaml
glob: "./apps/*"
log_level: info
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `glob` | string | **yes** | — | Glob pattern for directories to scan |
| `log_level` | string | no | `info` | Log verbosity |

### `glob`

Matched against the filesystem at startup. Any directory in the match set that contains a `kustomization.yaml` or `kustomization.yml` file is watched.

The pattern is evaluated by [doublestar](https://github.com/bmatcuk/doublestar), which supports `**` for recursive matching:

| Pattern | Matches |
|---------|---------|
| `./apps/*` | One level under `apps/` |
| `./apps/**` | All nested directories under `apps/` |
| `/manifests/*/base` | `base` subdirectory inside any single-level match under `/manifests/` |
| `../../other-repo/apps/*` | Relative paths outside the working directory |

!!! note
    Patterns are relative to the **working directory** when the binary is run, not the config file's location.

### `log_level`

Controls the minimum log level emitted. Valid values:

| Value | Description |
|-------|-------------|
| `debug` | All messages including internal state transitions |
| `info` | Normal operation: watch events, apply success |
| `warn` | Non-fatal problems: directory cannot be watched |
| `error` | Apply failures and watcher errors |

## CLI flags

| Flag | Description |
|------|-------------|
| `--config PATH` | Path to the config file |
| `--log-level LEVEL` | Override the log level from config |

## Environment variables

| Variable | Equivalent flag |
|----------|----------------|
| `KUSTOMIZE_WATCHER_CONFIG` | `--config` |
| `KUSTOMIZE_WATCHER_LOG_LEVEL` | `--log-level` |

## Override precedence

```
--flag  >  env var  >  config file  >  built-in default
```

The built-in default for `log_level` is `info`.

## Log format

All output is JSON (via `log/slog`), written to **stdout**. Startup errors (e.g. config file not found) go to **stderr**.

Example log lines:

```json
{"time":"2026-05-08T10:00:00.000Z","level":"INFO","msg":"watching","path":"./apps/nginx"}
{"time":"2026-05-08T10:00:00.001Z","level":"INFO","msg":"initial apply","dir":"./apps/nginx"}
{"time":"2026-05-08T10:05:32.100Z","level":"INFO","msg":"applying","dir":"./apps/nginx"}
{"time":"2026-05-08T10:05:32.300Z","level":"ERROR","msg":"apply resource","name":"nginx","kind":"Deployment","err":"..."}
{"time":"2026-05-08T10:05:32.301Z","level":"WARN","msg":"cannot watch dir","path":"./apps/broken","err":"no such file"}
```

Fields present on every line: `time`, `level`, `msg`. Additional context fields vary by message.
