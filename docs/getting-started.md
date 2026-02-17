# Getting Started with ARCA

## Installation

### Windows (Winget)

```powershell
winget install ARCA.CLI
```

### macOS/Linux (Homebrew)

```bash
brew install arca-cli/tap/arca
```

### From Source

```bash
go install github.com/adryledo/arca-cli/cmd/arca@latest
```

## Basic Usage

### 1. Initialize a project

Create a `.arca-assets.yaml` file in your project root, or use the `install` command.

### 2. Install an asset

```bash
# Add an asset from a GitHub repository
arca install https://github.com/org/assets my-asset --target .github/prompts/my-asset.md
```

### 3. Sync existing assets

If you've cloned a project that already has an ARCA configuration:

```bash
# Downloads missed assets and restores symlinks
arca sync
```

### 4. Direct tool projections

Map an asset to specific AI assistants:

```bash
# Map to both Copilot and Cursor locations
arca install https://github.com/org/assets my-asset --name cursor --target .cursor/prompts/my-asset.md
```

### 5. Listing and Browsing

```bash
# List assets in a remote manifest
arca list-remote https://github.com/org/assets

# List currently installed assets in the project
arca list

# Output in JSON for tool integration
arca list --json
```

### 6. Maintainer Workflow

```bash
# Add a new version of an asset to the local arca-manifest.yaml
arca publish my-asset 1.2.0 prompt prompts/my-asset.md
```
