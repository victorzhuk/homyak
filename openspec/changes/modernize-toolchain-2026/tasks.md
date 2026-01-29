# Implementation Tasks

## Phase 1: Go 1.25.6 Upgrade

**Objective**: Update Go toolchain and dependencies
**Estimated Time**: 2 hours
**Dependencies**: None

### Task 1.1: Update Go Version
- [x] Run `go mod edit -go=1.25.6`
- [x] Verify `go.mod` shows `go 1.25.6`
- [x] Run `go mod tidy`
- [ ] Verify build: `go build ./...`
- [ ] Verify tests: `go test ./...`

### Task 1.2: Update Dependencies
- [ ] Run `go get -u ./...`
- [ ] Run `go mod tidy`
- [ ] Review updated `go.sum` for unexpected changes
- [ ] Verify build: `go build ./...`
- [ ] Verify tests: `go test ./...`

### Task 1.3: Add Go Tool Directive
- [x] Run `go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`
- [x] Verify `go.mod` contains `tool` directive
- [x] Run `go mod tidy`

**Phase 1 Validation**:
```bash
go version  # Should show 1.25.6
go build ./...
go test ./...
```

---

## Phase 2: Golangci-lint v2 Migration

**Objective**: Migrate from v1 to v2 with new config format
**Estimated Time**: 2 hours
**Dependencies**: Phase 1

### Task 2.1: Backup Current Config
- [x] Copy `.golangci.yml` to `.golangci.yml.v1.backup`

### Task 2.2: Install and Run Migration
- [x] Install golangci-lint v2 locally (if not already)
- [x] Run `golangci-lint migrate`
- [x] Review migrated `.golangci.yml`
- [x] Compare with v1 backup

### Task 2.3: Update Makefile
- [x] Change: `go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ...`
- [x] To: `go tool golangci-lint run ...`
- [x] Update both `lint` and `lint-changes` targets

### Task 2.4: Verify Linting
- [ ] Run `make lint`
- [ ] Fix any new linting errors
- [ ] Run `make lint-changes`

**Phase 2 Validation**:
```bash
make lint
```

---

## Phase 3: Dockerfile Modernization

**Objective**: Migrate to distroless Debian 13
**Estimated Time**: 2 hours
**Dependencies**: Phases 1, 2

### Task 3.1: Update Build Stages
- [x] Change node stage: `node:22.12-alpine3.21` → `node:24.13.0-slim`
- [x] Change go stage: `golang:1.23.4-alpine3.21` → `golang:1.25.6-bookworm`
- [x] Update package manager: `apk` → `apt-get`

### Task 3.2: Migrate to Distroless
- [x] Change runtime: `alpine:3.21` → `gcr.io/distroless/static-debian13:nonroot`
- [x] Remove manual user creation (distroless has built-in nonroot)
- [x] Remove `ca-certificates` install (included in distroless)
- [x] Remove `chown` (not needed with distroless)

### Task 3.3: Test Docker Build
- [ ] Run `docker build -f build/prod.Dockerfile -t homyak:test .`
- [ ] Verify build succeeds
- [ ] Check image size: `docker images homyak:test`
- [ ] Run container: `docker run -p 8080:8080 homyak:test`
- [ ] Verify app responds on port 8080

### Task 3.4: Security Verification
- [ ] Try `docker exec -it <container> sh` (should fail - no shell)
- [ ] Check user: `docker run --rm homyak:test id` (should show nonroot)

**Phase 3 Validation**:
```bash
docker build -f build/prod.Dockerfile -t homyak:test .
docker run -d -p 8080:8080 --name homyak-test homyak:test
curl http://localhost:8080/health  # or appropriate endpoint
docker stop homyak-test && docker rm homyak-test
```

---

## Phase 4: Node.js 24 + React 19

**Objective**: Update frontend toolchain
**Estimated Time**: 4 hours
**Dependencies**: None (can run parallel with Phases 1-3)

### Task 4.1: Update Node.js Runtime
- [x] Update `build/prod.Dockerfile` node stage (if not done in Phase 3)
- [ ] Verify local Node version: `node --version` → v24.13.0

### Task 4.2: Update package.json
- [x] Update React: `^18.3.1` → `^19.2.4`
- [x] Update React-DOM: `^18.3.1` → `^19.2.4`
- [x] Update all devDependencies to latest

### Task 4.3: Install and Build
- [ ] Run `yarn install`
- [ ] Run `yarn build`
- [ ] Fix any build errors

### Task 4.4: Address React 19 Changes
- [ ] Check for `ReactDOM.render` usage (deprecated) → use `createRoot`
- [ ] Check ref usage patterns
- [ ] Run `yarn lint`
- [ ] Fix any linting errors

### Task 4.5: Manual Testing
- [ ] Run `yarn dev` (if available)
- [ ] Verify UI renders correctly
- [ ] Check browser console for warnings/errors
- [ ] Test key user flows

**Phase 4 Validation**:
```bash
cd web
yarn install
yarn build
yarn lint
```

---

## Phase 5: Integration & Validation

**Objective**: Full system validation
**Estimated Time**: 2 hours
**Dependencies**: Phases 1-4

### Task 5.1: Full Docker Build
- [ ] Run complete build: `docker build -f build/prod.Dockerfile -t homyak:latest .`
- [ ] Verify no errors

### Task 5.2: Integration Testing
- [ ] Start container: `docker run -d -p 8080:8080 --name homyak-latest homyak:latest`
- [ ] Test health endpoint
- [ ] Test main application endpoints
- [ ] Check logs: `docker logs homyak-latest`

### Task 5.3: Security Scan
- [ ] Run Trivy scan: `trivy image homyak:latest`
- [ ] Review results for critical/high vulnerabilities
- [ ] Document any accepted risks

### Task 5.4: Final Verification
- [ ] Run `make test` (Go tests)
- [ ] Run `make lint` (Go lint)
- [ ] Run `make build` (Go build)
- [ ] Verify frontend build in container context

### Task 5.5: Documentation
- [ ] Update README with new version requirements
- [ ] Document any breaking changes for developers
- [ ] Update CI/CD docs if needed

**Phase 5 Validation**:
```bash
# Go
make test
make lint
make build

# Frontend
cd web && yarn build && yarn lint

# Docker
docker build -f build/prod.Dockerfile -t homyak:latest .
docker run -d -p 8080:8080 homyak:latest
# Test endpoints...
```

---

## Summary

| Phase | Tasks | Est. Time | Status |
|-------|-------|-----------|--------|
| 1 | Go 1.25.6 | 2h | 🟡 In Progress |
| 2 | Golangci-lint v2 | 2h | ✅ Complete |
| 3 | Dockerfile | 2h | ✅ Complete |
| 4 | Node 24 + React 19 | 4h | 🟡 In Progress |
| 5 | Validation | 2h | ⬜ |
| **Total** | | **12h** | **~60%** |

## Notes

- Phases 1-3 are sequential (Go → Lint → Docker)
- Phase 4 (Node/React) can run in parallel with Phases 1-3
- Phase 5 requires all previous phases complete
- Each phase should be a separate commit for easy rollback
