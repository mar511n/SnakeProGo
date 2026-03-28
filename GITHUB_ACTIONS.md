# GitHub Actions CI/CD Setup

Automated build and release process for SnakeProGo using GitHub Actions.

## How It Works

The workflow (`.github/workflows/build-release.yml`) is triggered when you push a git tag and automatically:

1. **Builds** the game binary for all platforms:
   - Linux (x86_64)
   - macOS ARM (Apple Silicon)
   - Windows (x86_64)

2. **Creates release artifacts** - Each platform gets a zip file

3. **Publishes a GitHub release** - All artifacts are attached to the release

## How to Use

### Create a Release

1. **Tag your commit:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions automatically:**
   - Runs on the tag push
   - Builds for all 4 platforms (parallel)
   - Creates a GitHub release
   - Uploads all binaries

3. **Users can download** the binaries from GitHub Releases page

### View Build Status

Go to your repository → **Actions** tab to see:
- Build progress for each platform
- Logs if anything fails
- Download artifacts before release is created

### What Gets Released

For release `v1.0.0`, GitHub will create:
- `SnakeProGo-linux-amd64.zip`
- `SnakeProGo-darwin-arm64.zip` 
- `SnakeProGo-windows-amd64.zip`

Each is a complete, ready-to-run binary for that platform.

## Push Tags to Release

```bash
# Create and push a version tag
git tag v1.0.0
git push origin v1.0.0

# Or tag an existing commit
git tag v1.0.0 <commit-hash>
git push origin v1.0.0
```

## Modify the Workflow

Edit `.github/workflows/build-release.yml` to:

**Change Go version:**
```yaml
go-version: '1.25'  # Update here
```

**Add macOS 12 support:**
```yaml
- os: macOS (12)
  runs-on: macos-12
  goos: darwin
  goarch: amd64
```

**Add 32-bit builds:**
```yaml
- os: Linux (32-bit)
  runs-on: ubuntu-latest
  goos: linux
  goarch: 386
```

**Add tests before building:**
Add this step after "Set up Go":
```yaml
- name: Run tests
  working-directory: ./src
  run: go test ./...
```

## Requirements

- Public GitHub repository (free)
- No additional setup needed - GitHub Actions is free for public repos
- The workflow uses standard GitHub-hosted runners

## Troubleshooting

**Build fails on a specific platform:**
- Check the logs in Actions tab
- Common issue: Go version too old/new
- Platform-specific code issues

**Release not created:**
- Verify the tag name matches `v*` pattern (e.g., `v1.0.0`)
- Check Actions tab for workflow errors
- Ensure GitHub token has release permissions (default is fine)

**Artifacts not attached:**
- Verify zip creation succeeded in logs
- Check release permissions in workflow

## Next Steps: Assets.zip

To also release `assets.zip`:

1. **Prepare assets** in your repo or a separate repo
2. **Modify the workflow** to zip assets
3. **Add to release** output

Example:
```yaml
- name: Create assets archive
  run: zip -r assets.zip res/assets/

- name: Upload assets
  uses: actions/upload-artifact@v4
  with:
    name: assets
    path: assets.zip
```

Then update the `release` job `files:` to include it.