# ARCA v1.0.0: Formal Specification

ARCA (Asset Resolution for AI Assistants) is a decentralized standard for distributing, versioning, and consuming agentic assets (prompts, skills, instructions).

## 1. Actors

| Actor | Responsibility |
| :--- | :--- |
| **Maintainer** | Publishes assets to a **Source Repository** (hosting the `arca-manifest.yaml`). |
| **Consumer** | Developers and tools using assets via `.arca-assets.yaml`. |
| **Source Repository** | A Git-based or local host containing asset files and the manifest. |
| **Provider** | Hosting services (GitHub, Azure DevOps, Local FS). |
| **ARCA CLI** | The Go binary used to resolve, fetch, and verify assets. |
| **AI Assistant** | The consumer (ChatGPT, Gemini, Cursor, Manus) that utilizes the projected assets. |

## 2. Technical Specification

### 2.1 The Manifest (`arca-manifest.yaml`)

```yaml
schema: 1.0
assets:
  <asset-id>:
    kind: prompt | skill | instruction
    description: "Brief description"
    versions:
      <version-string>:
        path: "path/to/file.md"
        ref: "v1.0.0" # Optional. Git tag/branch/commit.
```

### 2.2 The Configuration (`.arca-assets.yaml`)

```yaml
schema: 1.0
sources:
  my-org:
    type: git | local
    url: "https://github.com/my-org/agent-assets"
    path: "~/local-assets" # if type: local
assets:
  - id: refactor-logic
    source: my-org
    version: "^1.2.0"
    projections:
      default: ".github/prompts/refactor.md"
      cursor: ".cursor/prompts/refactor.md"
```

### 2.3 The Lockfile (`.arca-assets.lock`)

```json
{
  "assets": [
    {
      "id": "refactor-logic",
      "version": "1.2.5",
      "source": "my-org",
      "commit": "abc12345",
      "sha256": "df7a8b9c...",
      "manifestHash": "...",
      "resolvedAt": "2026-02-17T..."
    }
  ]
}
```

## 3. Resolution and Integration

### 3.1 LF-Normalization
To ensure platform-independent hashes, all text-based assets MUST be normalized to LF (`\n`) before computing the SHA-256 hash.

### 3.2 Projection System
Assets are stored in a central cache (e.g., `~/.arca-cache`) and symlinked (projected) into the workspace. This allows one asset to be adopted by multiple AI assistants simultaneously without duplication.

### 3.3 CLI Interface
The ARCA CLI provides machine-readable output (`--json`) to allow IDE extensions and other tools to integrate seamlessly.
