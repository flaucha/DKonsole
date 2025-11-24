# Release Process Guide

This document describes the complete process for creating a new release and pushing a Docker image to Docker Hub.

## Overview

The release process uses:
- **VERSION file**: Centralized version declaration (`VERSION`)
- **GitHub Actions**: Automated Docker build and push on tag creation
- **Semantic Versioning**: Follows `MAJOR.MINOR.PATCH` format (e.g., `1.1.7`)

## Prerequisites

- Git configured with push access to the repository
- Access to create and push tags
- GitHub Actions enabled (automatically builds on tag push)

## Release Workflow

### Step 1: Create a Feature/Hotfix Branch

```bash
# For a new feature
git checkout main
git pull origin main
git checkout -b feature/your-feature-name

# For a hotfix
git checkout main
git pull origin main
git checkout -b hotfix/version-number
# Example: git checkout -b hotfix/1.1.7
```

### Step 2: Make Your Changes

Implement your feature, bug fixes, or improvements in the branch.

### Step 3: Update Version Files

#### 3.1 Update `VERSION` file

```bash
# Edit VERSION file (in project root)
echo "1.1.8" > VERSION
```

Or manually edit the file:
```
1.1.8
```

#### 3.2 Update `CHANGELOG.md`

Add a new section at the top of the changelog:

```markdown
## [1.1.8] - YYYY-MM-DD

### Added
- New feature description

### Changed
- Change description

### Fixed
- Bug fix description
```

**Format guidelines:**
- Use [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format
- Use semantic versioning
- Include date in `YYYY-MM-DD` format
- Group changes by type: Added, Changed, Deprecated, Removed, Fixed, Security

#### 3.3 Update `README.md`

Update version references:

1. **Version badge** (line ~5):
   ```markdown
   ![Version](https://img.shields.io/badge/version-1.1.8-green.svg)
   ```

2. **Git checkout command** (line ~35):
   ```markdown
   git checkout v1.1.8
   ```

3. **Docker image tag** (line ~103):
   ```yaml
   tag: "1.1.8"
   ```

4. **Docker image reference** (line ~110):
   ```markdown
   - **Unified**: `dkonsole/dkonsole:1.1.8`
   ```

5. **Changelog section** (add new entry at line ~116):
   ```markdown
   ### v1.1.8 (YYYY-MM-DD)
   **Brief description of release**
   
   - Change 1
   - Change 2
   ```

#### 3.4 Update Helm Charts

**`helm/dkonsole/Chart.yaml`:**
```yaml
version: 1.1.8
appVersion: "1.1.8"
```

**`helm/dkonsole/values.yaml`:**
```yaml
image:
  repository: dkonsole/dkonsole
  tag: "1.1.8"
```

### Step 4: Commit and Push Branch

```bash
# Stage all changes
git add -A

# Commit with descriptive message
git commit -m "feat: your feature description"
# or
git commit -m "fix: bug fix description"
# or
git commit -m "chore: release v1.1.8 - release description"

# Push branch to remote
git push origin feature/your-feature-name
# or
git push origin hotfix/1.1.8
```

### Step 5: Merge to Main

```bash
# Switch to main
git checkout main
git pull origin main

# Merge your branch
git merge feature/your-feature-name -m "Merge feature/your-feature-name: description"
# or
git merge hotfix/1.1.8 -m "Merge hotfix/1.1.8: description"

# Push main
git push origin main
```

### Step 6: Create Release Commit (if needed)

If you haven't committed the version updates yet:

```bash
# Make sure you're on main
git checkout main

# Stage version files
git add VERSION CHANGELOG.md README.md helm/dkonsole/Chart.yaml helm/dkonsole/values.yaml

# Commit version bump
git commit -m "chore: release v1.1.8 - release description"

# Push
git push origin main
```

### Step 7: Create and Push Git Tag

The tag triggers the GitHub Actions workflow to build and push the Docker image.

```bash
# Create annotated tag
git tag -a v1.1.8 -m "Release v1.1.8: release description"

# Push tag to remote
git push origin v1.1.8
```

**Important:** The tag name must follow semantic versioning:
- ✅ Valid: `v1.1.8`, `1.1.8`
- ❌ Invalid: `v1.1`, `1.1.8-beta`, `test`

### Step 8: Verify GitHub Actions Build

1. Go to: `https://github.com/flaucha/DKonsole/actions`
2. Find the workflow run triggered by the tag push
3. Verify that:
   - Tests pass (test-backend, test-frontend)
   - Build completes successfully
   - Docker image is built and pushed

The workflow will:
- Extract version from tag (`v1.1.8` → `1.1.8`)
- Validate version format (must be `x.y.z`)
- Build Docker image
- Tag image as: `dkonsole/dkonsole:1.1.8` and `dkonsole/dkonsole:1.1`
- Push to Docker Hub

## Quick Reference: Complete Release Example

```bash
# 1. Create branch
git checkout main
git pull origin main
git checkout -b feature/new-feature

# 2. Make changes, then update version
echo "1.1.8" > VERSION
# Edit CHANGELOG.md, README.md, helm files...

# 3. Commit and push branch
git add -A
git commit -m "feat: add new feature"
git push origin feature/new-feature

# 4. Merge to main
git checkout main
git merge feature/new-feature -m "Merge feature/new-feature"
git push origin main

# 5. Create release commit (if version files not committed)
git add VERSION CHANGELOG.md README.md helm/dkonsole/Chart.yaml helm/dkonsole/values.yaml
git commit -m "chore: release v1.1.8 - new feature"
git push origin main

# 6. Create and push tag
git tag -a v1.1.8 -m "Release v1.1.8: new feature"
git push origin v1.1.8

# 7. Verify build in GitHub Actions
# https://github.com/flaucha/DKonsole/actions
```

## Hotfix Release Example

For urgent fixes on the current version:

```bash
# 1. Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/1.1.8

# 2. Make fix, update version to 1.1.9
echo "1.1.9" > VERSION
# Update CHANGELOG.md, README.md, helm files...

# 3. Commit and push
git add -A
git commit -m "fix: critical bug fix"
git push origin hotfix/1.1.9

# 4. Merge to main
git checkout main
git merge hotfix/1.1.9 -m "Merge hotfix/1.1.9: critical fix"
git push origin main

# 5. Create tag
git tag -a v1.1.9 -m "Release v1.1.9: critical fix"
git push origin v1.1.9
```

## Version Numbering Guidelines

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.1.0): New features, backward compatible
- **PATCH** (0.0.1): Bug fixes, backward compatible

Examples:
- `1.1.7` → `1.1.8` (patch: bug fix)
- `1.1.7` → `1.2.0` (minor: new feature)
- `1.1.7` → `2.0.0` (major: breaking change)

## CI/CD Behavior

### What Triggers What

| Event | Tests | Build | Docker Build |
|-------|-------|-------|--------------|
| Push to `main` | ✅ | ❌ | ❌ |
| Push tag `v*` | ✅ | ✅ | ✅ |
| Pull Request | ✅ | ❌ | ❌ |

### Docker Image Tags

When a tag is pushed, the image is tagged as:
- `dkonsole/dkonsole:{version}` (e.g., `1.1.8`)
- `dkonsole/dkonsole:{major.minor}` (e.g., `1.1`)

## Troubleshooting

### Tag validation fails

**Error:** `Tag 'v1.1' doesn't match semver pattern`

**Solution:** Use full semantic version: `v1.1.0` or `1.1.0`

### Docker build doesn't trigger

**Check:**
1. Tag format is correct (`v1.1.8` or `1.1.8`)
2. Tag was pushed: `git push origin v1.1.8`
3. GitHub Actions is enabled for the repository
4. Check Actions tab for workflow runs

### Version mismatch

**Problem:** Docker image has wrong version

**Solution:** 
- Ensure `VERSION` file matches tag version
- Tag must be in format `v1.1.8` or `1.1.8`
- GitHub Actions extracts version from tag automatically

### Multiple pipelines running

**Problem:** Both main push and tag push trigger builds

**Solution:** This is expected behavior. The workflow is configured to:
- Run tests on push to main
- Run full build (tests + Docker) only on tag push

## Files to Update for Each Release

1. ✅ `VERSION` - Version number
2. ✅ `CHANGELOG.md` - Release notes
3. ✅ `README.md` - Version badge, checkout command, image tags, changelog
4. ✅ `helm/dkonsole/Chart.yaml` - Chart version and appVersion
5. ✅ `helm/dkonsole/values.yaml` - Docker image tag

## Verification Checklist

Before pushing the tag, verify:

- [ ] `VERSION` file updated
- [ ] `CHANGELOG.md` updated with new section
- [ ] `README.md` version badge updated
- [ ] `README.md` git checkout command updated
- [ ] `README.md` Docker image references updated
- [ ] `README.md` changelog section added
- [ ] `helm/dkonsole/Chart.yaml` version updated
- [ ] `helm/dkonsole/values.yaml` image tag updated
- [ ] All changes committed and pushed to main
- [ ] Tag name follows semantic versioning (`v1.1.8` or `1.1.8`)

## Additional Notes

- The `VERSION` file is used by `vite.config.js` to inject the version into the frontend build
- The Dockerfile copies the `VERSION` file to make it available during build
- GitHub Actions automatically validates tag format before building
- Only tags trigger Docker builds (not pushes to main)
- The workflow fails fast if tag format is invalid

---

**Last Updated:** 2025-11-24  
**Current Version:** 1.1.8

