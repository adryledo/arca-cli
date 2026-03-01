# Pull Request Template

## Description
- Short description of changes (1-2 sentences):

## Motivation and Context
- Why is this change needed? What problem does it solve?

## How to test
- Commands run locally (copy/paste the exact commands you ran):
  - `gofmt -w .`
  - `go vet ./...`
  - `go test ./...`
  - `golangci-lint run` (if used)
  - build: `go build -o bin/arca ./cmd/arca`

## Checklist (required)
- [ ] Code formatted: `gofmt -w .` or `gofumpt -w .`
- [ ] `go vet ./...` completed
- [ ] `go test ./...` passed
- [ ] Linting: `golangci-lint run` passed or documented exceptions
- [ ] Docs updated in `docs/` if user-facing behavior changed
- [ ] `TASK.md` or `CHANGELOG` updated if behavior changed

## Labels (maintainer / agent use)
- Suggested labels: `bug`, `feat`, `chore`, `docs`, `automation`

## Additional notes
- Attach test output or CI logs if relevant.
- If this PR was created by an automation agent, include the exact command list the agent executed and any non-deterministic decisions made.