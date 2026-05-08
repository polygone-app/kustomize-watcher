# syntax=docker/dockerfile:1
FROM alpine:3.22

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/kustomize-watcher /kustomize-watcher

ENTRYPOINT ["/kustomize-watcher"]
