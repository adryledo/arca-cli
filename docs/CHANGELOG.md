# 📋 Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.0.1](https://github.com/adryledo/arca-cli/releases/tag/v0.0.1) - 2026-03-07

### ✨ Added
- **Core ARCA engine** — asset resolution pipeline: discovery, version matching, download, validation, caching, and projection
- **`arca install`** — end-to-end asset installation from Git or local sources into `.arca-assets.yaml`
- **`arca sync`** — reconciles missing or stale assets and restores symlinks
- **`arca list`** — lists currently installed assets; supports `--json` for tool integration
- **`arca list-remote`** — browses a remote manifest without installing
- **`arca publish`** — maintainer command to register a new asset version in `arca-manifest.yaml`
- **`arca init`** — initializes a manifest by scanning instruction and skill folders
- **Multi-assistant projections** — single asset projected to multiple AI tool paths simultaneously
- **Directory-based skills** — support for multi-file skill assets
- **Recursive dependency resolution** — resolves transitive asset dependencies
- **SemVer constraint matching** — resolves `^`, `~`, and exact version constraints to Git tags/refs
- **SHA-256 integrity verification** with mandatory LF-normalization for cross-platform consistency
- **Global `~/.arca-cache`** — centralized content-addressable cache
- **`--json` output flag** — machine-readable output for IDE and CI integration
- **Advanced Git auth** — reads `ARCA_GIT_TOKEN`, `GITHUB_TOKEN`, or `AZURE_DEVOPS_EXTTOKEN` environment variables
- **Sample assets** — consumer/maintainer samples in `samples/`
- **`AGENTS.md`** — project-level guidelines for AI coding agents
- **`docs/PULL_REQUEST_TEMPLATE.md`** — standard PR template for contributors
- **`.github/hooks/pre-commit`** — pre-commit hook enforcing `go test`, `go vet`, and formatting checks
- **MIT License**
- **GitHub Actions pipeline** — automates release artifact creation and submits to `winget-pkgs`
- **Open source documentation** — `docs/README.md`, `docs/getting-started.md`, `docs/purpose.md`, `docs/protocol.md`, `docs/CONTRIBUTING.md`, `docs/CODE_OF_CONDUCT.md`, `docs/ROADMAP.md`
- **Unit tests** for all internal packages (`hasher`, `resolver`, `config`, `downloader`, `auth`, `cmd/arca`)

### 🗑️ Removed
- All references to deprecated `prompts` asset kind

### 🔒 Security
- SHA-256 content verification on every asset fetch
- Token-based Git authentication via environment variables (no hardcoded credentials)

---

[Documentation Index](./README.md)
