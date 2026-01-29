# Docker & Container Specification

## Overview

Modernize container images for security hardening: migrate from Alpine to distroless and update to Debian 13 "Trixie".

## Requirements

### FR-1: Go Builder Image Update
**Given** the current builder uses `golang:1.23.4-alpine3.21`
**When** the update is applied
**Then** the builder uses `golang:1.25.6-bookworm`

### FR-2: Node.js Builder Image Update
**Given** the current frontend builder uses `node:22.12-alpine3.21`
**When** the update is applied
**Then** the frontend builder uses `node:24.13.0-slim`

### FR-3: Runtime Image Migration
**Given** the current runtime uses `alpine:3.21` with manual user creation
**When** the update is applied
**Then** the runtime uses `gcr.io/distroless/static-debian13:nonroot`

### FR-4: Security Hardening
**Given** the current image has shell and package manager
**When** using distroless
**Then** the image has no shell, no package manager, and runs as non-root

### FR-5: Build Compatibility
**Given** the current Dockerfile builds successfully
**When** built with updated images
**Then** the build succeeds and application runs correctly

## Technical Details

### Current Dockerfile

```dockerfile
FROM node:22.12-alpine3.21 AS nodejs
WORKDIR /web
COPY web .
RUN yarn install
RUN yarn build

FROM golang:1.23.4-alpine3.21 AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /app
RUN apk add --update make git
COPY . .
COPY --from=nodejs /web/dist /app/web/dist
RUN make build

FROM alpine:3.21
ENV USER=zhuk
ENV GROUPNAME=$USER
WORKDIR /app
RUN addgroup "$GROUPNAME" && \
    adduser --disabled-password --gecos "" --home "$(pwd)" \
    --ingroup "$GROUPNAME" --no-create-home $USER
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bin/homyaksrv .
RUN chown -R $USER:$GROUPNAME /app
USER $USER
EXPOSE 8080/tcp
CMD ["/app/homyaksrv", "run"]
```

### Target Dockerfile

```dockerfile
# Stage 1: Build frontend
FROM node:24.13.0-slim AS nodejs
WORKDIR /web
COPY web .
RUN yarn install
RUN yarn build

# Stage 2: Build Go binary
FROM golang:1.25.6-bookworm AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends make git && rm -rf /var/lib/apt/lists/*
COPY . .
COPY --from=nodejs /web/dist /app/web/dist
RUN make build

# Stage 3: Runtime - distroless for minimal attack surface
FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app
COPY --from=builder /app/bin/homyaksrv .
# Note: distroless:nonroot already runs as uid:nonroot (65532)
EXPOSE 8080/tcp
CMD ["/app/homyaksrv", "run"]
```

### Security Comparison

| Aspect | Alpine 3.21 | Distroless Debian 13 |
|--------|-------------|----------------------|
| Base size | ~5 MB | ~2 MB |
| Shell | `/bin/sh` | None |
| Package manager | `apk` | None |
| User setup | Manual (`adduser`) | Built-in (`nonroot`) |
| CA certificates | Manual install | Included |
| Attack surface | Medium | Minimal |
| Google hardened | No | Yes |

### Files to Modify

| File | Changes |
|------|---------|
| `build/prod.Dockerfile` | Complete rewrite with new base images |

### Commands

```bash
# Build the image
docker build -f build/prod.Dockerfile -t homyak:latest .

# Verify it runs
docker run -p 8080:8080 homyak:latest

# Check image size
docker images homyak:latest

# Security scan (optional)
trivy image homyak:latest
```

## Acceptance Criteria

- [ ] `docker build` succeeds without errors
- [ ] Container starts and serves traffic on port 8080
- [ ] Image size is smaller or comparable to previous
- [ ] `docker exec` cannot get shell (no `/bin/sh`)
- [ ] Container runs as non-root user
- [ ] No critical/high vulnerabilities in security scan

## Dependencies

- Docker daemon
- Access to Docker Hub (for node, golang images)
- Access to gcr.io (for distroless images)

## References

- [Distroless GitHub](https://github.com/GoogleContainerTools/distroless)
- [Debian 13 Trixie](https://www.debian.org/releases/stable/)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
