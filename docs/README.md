# ARCA: Asset Resolution for AI Assistants

ARCA is a decentralized standard for distributing, versioning, and consuming agentic assets (prompts, rules, skills, instructions).

## What is ARCA?

ARCA provides a unified way for AI assistants—including coding agents (Copilot, Cursor), web agents (Manus), and general LLMs (ChatGPT, Gemini)—to discover and integrate specialized assets. By using Git-based manifests and deterministic locking, ARCA ensures that your agents always have the right version of the instructions they need.

## Core Features

- **Decentralized Registry**: Host your assets in any Git repository or local folder.
- **Deterministic Locking**: Reproducible environments with `.arca-assets.lock`.
- **High Performance**: Zero-dependency Go CLI for fast resolution and syncing.
- **Multi-Assistant Projections**: Sync one asset to multiple locations (e.g., `.cursor/prompts` and `.github/prompts`).
- **Strict Integrity**: SHA-256 verification with mandatory LF-normalization for cross-platform consistency.
- **Mobile Friendly**: Built to run on Windows, macOS, Linux, Android, and iOS.

## Documentation Index

1. [Purpose & Benefits](./purpose.md) - Why we built ARCA.
2. [Getting Started](./getting-started.md) - How to install and use the CLI.
3. [Protocol Deep-Dive](./protocol.md) - How the manifests and resolution flows work.
4. [Contribution Guide](./contributing.md) - How to improve ARCA.
