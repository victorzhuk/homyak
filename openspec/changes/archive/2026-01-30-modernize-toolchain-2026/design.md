# Design Document: Toolchain Modernization 2026

## Architecture Decisions

### ADR-001: Go Version Selection

**Status**: Accepted

**Context**: Need to upgrade from Go 1.23 (EOL August 2025) to a supported version.

**Options Considered**:

| Option | Version | Pros | Cons |
|--------|---------|------|------|
| A | 1.24.12 | Conservative, well-tested | Older features |
| B | **1.25.6** | **Latest, 6 recent CVE fixes** | Slightly newer |

**Decision**: Use Go 1.25.6

**Rationale**:
- Released January 15, 2026 (2 weeks ago)
- Contains 6 security fixes (CVE patches)
- `go tool` directive support for better tooling management
- Project is not in production yet, can afford latest
- 5 months of real-world usage since August 2025

**Consequences**:
- Need to verify all dependencies support Go 1.25
- CI/CD may need Go version update

---

### ADR-002: Node.js Version Selection

**Status**: Accepted

**Context**: Current Node 22.12 is in Maintenance LTS, ending April 2027.

**Options Considered**:

| Option | Version | Status | Support Ends |
|--------|---------|--------|--------------|
| A | 22.22 | Maintenance LTS | Apr 2027 |
| B | **24.13.0** | **Active LTS** | **Oct 2028** |
| C | 25.5.0 | Current | Jun 2026 |

**Decision**: Use Node 24.13.0 LTS "Krypton"

**Rationale**:
- Active LTS with longer support window
- "Krypton" is the codename for v24 LTS
- Stable and production-ready
- Better alignment with React 19 requirements

**Consequences**:
- May need to update some build tools
- Longer support reduces future upgrade burden

---

### ADR-003: Container Runtime Selection

**Status**: Accepted

**Context**: Currently using Alpine 3.21 with manual security setup.

**Options Considered**:

| Option | Image | Pros | Cons |
|--------|-------|------|------|
| A | Alpine 3.21 | Small, familiar | Manual hardening, shell present |
| B | **Distroless Debian 13** | **Minimal attack surface, Google-hardened** | No shell for debugging |
| C | Scratch | Smallest | No CA certs, completely manual |

**Decision**: Use Google Distroless Debian 13

**Rationale**:
- Smaller attack surface than Alpine (no shell, no package manager)
- Pre-hardened by Google security team
- Built-in nonroot user
- Includes CA certificates
- Industry best practice for Go applications

**Consequences**:
- Cannot `docker exec` into container for debugging
- Need external logging/monitoring for troubleshooting
- Must ensure all runtime dependencies are statically linked

**Mitigation**:
- Use structured logging extensively
- Add health check endpoints
- Keep debug images in separate tag for emergencies

---

### ADR-004: Golangci-lint v2 Migration Strategy

**Status**: Accepted

**Context**: Currently using v1.x which is deprecated. v2 has new config format.

**Options Considered**:

| Option | Approach | Pros | Cons |
|--------|----------|------|------|
| A | Stay on v1 | No changes needed | Deprecated, missing features |
| B | **v2 with migration tool** | **Latest features, maintained** | Config file changes required |
| C | v2 with manual config | Full control | Time-consuming |

**Decision**: Use golangci-lint v2.8.0 with migration tool

**Rationale**:
- v1 is deprecated and will not receive updates
- v2 has improved performance and new linters
- Built-in migration tool: `golangci-lint migrate`
- `go tool` integration for better dependency management

**Consequences**:
- `.golangci.yml` needs format changes
- Some linter configurations may need adjustment
- Makefile needs update

**Migration Steps**:
1. Install golangci-lint v2 locally
2. Run `golangci-lint migrate` to convert config
3. Review and adjust migrated config
4. Update Makefile to use `go tool`
5. Test with `make lint`

---

### ADR-005: React 19 Migration Approach

**Status**: Accepted

**Context**: React 18.3 reached EOL in December 2024.

**Options Considered**:

| Option | Version | Pros | Cons |
|--------|---------|------|------|
| A | Stay on 18.3 | No migration work | No security updates, technical debt |
| B | **19.2.4** | **Latest, supported** | Some breaking changes |

**Decision**: Upgrade to React 19.2.4

**Rationale**:
- React 18 is end-of-life
- React 19 has been stable since December 2024
- 19.2.4 is latest patch (January 26, 2026)
- New features: `use()` hook, improved refs, better SSR

**Breaking Changes to Address**:
1. **Refs**: Can now pass ref as regular prop (no `forwardRef` needed)
2. **Context**: `use()` hook for reading context
3. **Cleanup**: Ref callbacks now support cleanup functions
4. **Deprecated APIs**: `ReactDOM.render` removed (use `createRoot`)

**Migration Strategy**:
- Update package.json versions
- Run codemod if available
- Test all components manually
- Check for console warnings

---

## Technical Approach

### Phase Dependencies

```
Phase 1: Go 1.25.6
    │
    ▼
Phase 2: Golangci-lint v2
    │
    ▼
Phase 3: Dockerfile (depends on Phases 1 & 2)
    │
    ▼
Phase 4: Node 24 + React 19 (independent, can parallel)
    │
    ▼
Phase 5: Validation
```

### Rollback Strategy

Each phase is in a separate commit. If issues arise:

1. **Go upgrade issues**: Revert go.mod changes, stay on 1.23 temporarily
2. **Lint issues**: Keep v1 config temporarily, fix issues incrementally
3. **Docker issues**: Revert to Alpine-based Dockerfile
4. **React issues**: Downgrade to React 18.3, keep Node 24

### Testing Strategy

| Layer | Tool | Coverage |
|-------|------|----------|
| Go unit | `go test` | All packages |
| Go lint | `golangci-lint` | All files |
| Frontend build | `yarn build` | Production bundle |
| Frontend lint | `yarn lint` | All source files |
| Integration | Docker | Full application |
| Security | Trivy/Snyk | Container image |

## Files Modified

| File | Phase | Description |
|------|-------|-------------|
| `go.mod` | 1 | Go version, dependencies, tool directive |
| `go.sum` | 1 | Regenerated |
| `Makefile` | 2 | Update lint target |
| `.golangci.yml` | 2 | Migrate to v2 format |
| `build/prod.Dockerfile` | 3 | New base images, distroless |
| `web/package.json` | 4 | Update all versions |
| `web/yarn.lock` | 4 | Regenerated |

## Success Metrics

- Build time: Should remain similar or improve
- Image size: Should decrease (distroless is smaller)
- Security scan: Zero critical/high vulnerabilities
- Test pass rate: 100%
- Lint pass rate: 100%
