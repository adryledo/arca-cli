AGENTS.md

Purpose
- Give humans and automation (agents) a clear, consistent workflow for contributing to and maintaining arca-cli.
- Ensure changes are tested, linted, formatted, and documented before merge.

Scope
- Applies to local developers, CI pipelines, and automated agents making changes or PRs.
- Covers formatting, vetting, testing, building, docs, PR metadata, and agent behavior rules.

Quick start — canonical commands
- Format:
  - `gofmt -w .`
  - or `gofumpt -w .` (recommended for stricter formatting)
- Vet:
  - `go vet ./...`
- Test:
  - `go test ./...`
- Build CLI:
  - `go build -o bin/arca ./cmd/arca`
  - or run: `go run ./cmd/arca <args>`
- Lint (recommended):
  - `golangci-lint run` (if installed)

Local pre-PR checklist (must pass before opening PR)
- [ ] Run `gofmt -w .` or `gofumpt -w .`
- [ ] Run `go vet ./...`
- [ ] Run `go test ./...` (fix or document any test failures)
- [ ] Run `golangci-lint run` (if available; fix or document issues)
- [ ] Update `docs/` for any user-visible changes
- [ ] Update `TASK.md` or `CHANGELOG` for behavioral changes
- [ ] Include the commands you ran in the PR description

Documentation requirement
- Any change that affects behavior, CLI flags, manifest format, or user-facing instructions must include a docs update in `docs/`.
- Small wording or clarifying docs changes are allowed in the same PR.
- For major feature or behavior changes, add a short “Migration notes” subsection in the affected docs.

PR metadata and template expectations
- Provide a concise description, motivation, and testing steps.
- Include the checklist (see above) and mark completed items.
- Attach failing/passing test outputs if applicable.

Agent-specific rules (for automated changes)
- Agents must run the full Local pre-PR checklist before creating a PR.
- Agents must include exact commands and attachments for test output in the PR body.
- Agents must not push or merge to `main`/`master` directly. Create a PR and request human review.
- Agents must label PRs clearly (e.g., `automation`, `docs`, `fix`) and add a short rationale.
- If an agent cannot fix a lint or test failure deterministically, it must create the PR with a clear instruction: `Action required: human intervention to resolve <issue>`.

Branching and merge rules
- Feature work: branch from `main` named `feat/<short-desc>` or `fix/<short-desc>`.
- Small docs or chore changes: `docs/<short-desc>` or `chore/<short-desc>`.
- All merges to `main` must use PRs and have at least one human review approving.

CI guidance (recommended)
- Steps to run:
  1. Setup Go (respect `go.mod`)
  2. Run `gofmt -l .` (fail if any files are not formatted)
  3. Run `go vet ./...`
  4. Run `golangci-lint run` (optional but recommended)
  5. Run `go test ./...` (with `-v` for logs)
  6. Build `bin/arca` to verify compilation
- Suggested GitHub Actions workflow: checkout, setup Go, format-check, vet, lint, test, build.

Tooling & recommended configs
- Formatter: `gofumpt` or built-in `gofmt`
- Linter: `golangci-lint` with a conservative config
- Optional release tooling: `goreleaser` for binary releases
- Pre-commit: use `pre-commit` or provide a simple installable `.git/hooks/pre-commit` script in `.github/hooks/` (see repository copy)

Commit message guidance
- Use imperative, concise subject lines (e.g., "cmd: improve installer error handling")
- Include a short body if needed to explain rationale
- Close issues with `Fixes #<issue>` only when the PR fully addresses the issue

Troubleshooting & notes
- Long-running tests: keep unit tests fast; CI can run expensive integration tests in a separate job.
- If CI fails formatting/linting, run the canonical commands locally and update the PR.
- For docs-only changes: keep changes limited to `docs/`, mark PR with `docs` label, and note that code CI steps still run to ensure repository integrity.

Contact & escalation
- For unclear failures, tag a maintainer in the PR and add the `help wanted` label.