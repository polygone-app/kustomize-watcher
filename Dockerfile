# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/kustomize-watcher \
    ./cmd/kustomize-watcher

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/kustomize-watcher /kustomize-watcher

ENTRYPOINT ["/kustomize-watcher"]
