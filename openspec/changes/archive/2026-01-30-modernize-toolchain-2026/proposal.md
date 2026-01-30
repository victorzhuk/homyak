# Proposal: Modernize Toolchain to Latest Versions (2026)

## Problem Statement

The project is running on outdated toolchain versions with security and support implications:

| Component | Current | Status | Risk |
|-----------|---------|--------|------|
| Go | 1.23 | **End of Life** (Aug 2025) | No security patches |
| Node.js | 22.12 | Maintenance LTS | Ending April 2027 |
| React | 18.3 | **End of Life** (Dec 2024) | No security patches |
| Alpine | 3.21 | Current | Larger attack surface than distroless |
| Golangci-lint | v1.x | Deprecated | Missing new linters, v2 available |

## Goals

1. **Security**: Move to supported versions with active security patches
2. **Modern Tooling**: Adopt latest Go toolchain features (`go tool`)
3. **Hardening**: Replace Alpine with distroless for minimal attack surface
4. **Maintainability**: Update all dependencies to latest stable

## Non-Goals

- No application logic changes
- No API changes
- No database schema changes
- No new features

## Success Criteria

- [ ] All builds pass (`make build`, `docker build`)
- [ ] All tests pass (`make test`)
- [ ] Linting passes with new golangci-lint v2 (`make lint`)
- [ ] Frontend builds successfully (`yarn build`)
- [ ] Container runs and serves traffic
- [ ] Security scan shows no critical vulnerabilities

## Proposed Changes

### Go Toolchain
- **From**: Go 1.23
- **To**: Go 1.25.6 (released Jan 15, 2026)
- **Rationale**: Latest stable with 6 recent CVE fixes

### Node.js Runtime
- **From**: Node 22.12
- **To**: Node 24.13.0 LTS "Krypton"
- **Rationale**: Active LTS until October 2026

### Frontend Dependencies
- **React**: 18.3 → 19.2.4 (latest)
- **All devDependencies**: Update to latest

### Container Base Images
- **Build stage**: `golang:1.25.6-bookworm`
- **Frontend stage**: `node:24.13.0-slim`
- **Runtime**: `gcr.io/distroless/static-debian13:nonroot`
- **Rationale**: Debian 13 "Trixie" is current stable (released Aug 2025, updated Jan 2026)

### Linting
- **From**: golangci-lint v1.x
- **To**: golangci-lint v2.8.0
- **Config**: Migrate to v2 format
- **Integration**: Use `go tool` instead of `go run`

## Effort Estimate

| Phase | Description | Estimated Effort |
|-------|-------------|------------------|
| 1 | Go 1.25.6 + dependency updates | 2 hours |
| 2 | Golangci-lint v2 migration | 2 hours |
| 3 | Dockerfile distroless migration | 2 hours |
| 4 | Node 24 + React 19 migration | 4 hours |
| 5 | Validation & testing | 2 hours |
| **Total** | | **~12 hours** |

## Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| React 19 breaking changes | Medium | High | Test thoroughly, review migration guide |
| Distroless missing runtime deps | Low | High | Test full application flow in container |
| Golangci-lint v2 config issues | Low | Medium | Use built-in migration tool |
| Go 1.25 subtle behavior changes | Low | Low | Run full test suite |

## Rollback Plan

1. Feature branches for each phase
2. Docker image tags for each successful build
3. Revert commits if issues detected in staging

## References

- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Go 1.25.6 Security Fixes](https://seclists.org/oss-sec/2026/q1/68)
- [Node.js 24 LTS](https://nodejs.org/en/about/previous-releases)
- [React 19 Migration](https://react.dev/blog/2024/12/05/react-19)
- [Distroless Images](https://github.com/GoogleContainerTools/distroless)
- [Golangci-lint v2 Migration](https://golangci-lint.run/docs/product/migration-guide/)
