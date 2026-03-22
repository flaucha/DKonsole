# Release Process (AI Automation Guide)

This document is a strict step-by-step guide for AI agents to perform a release.

## 1. Preparation
- **Context**: Ensure you are on the `main` branch and it is up to date.
- **Version**: Determine the new version number (e.g., `1.1.11`) based on Semantic Versioning (MAJOR.MINOR.PATCH).

## 2. Update Version Files
Perform the following file updates precisely.

### 2.1 VERSION
- **File**: `VERSION`
- **Action**: Overwrite content with the new version string.
- **Example**: `1.1.11`

### 2.2 CHANGELOG.md
- **File**: `CHANGELOG.md`
- **Action**: Prepend the new version section at the top of the changelog (after the header).
- **Format**:
  ```markdown
  ## [VERSION] - YYYY-MM-DD

  ### Added
  - (List new features)

  ### Changed
  - (List changes)

  ### Fixed
  - (List bug fixes)
  ```

### 2.3 README.md
- **File**: `README.md`
- **Action 1 (Badge)**: Update version badge regex `version-[0-9.]+-green` to `version-NEW_VERSION-green`.
- **Action 2 (Manifest URLs)**: Update raw GitHub URLs that reference `v[0-9.]+` to `vNEW_VERSION`.
- **Action 3 (Image Ref)**: Update image reference regex `dkonsole:[0-9.]+` to `dkonsole:NEW_VERSION`.

### 2.4 Deployment Manifest
- **File**: `deploy/dkonsole.yaml`
- **Action**: Update the image tag and installation examples to the new release.
- **File**: `scripts/render-manifest.sh`
- **Action**: Keep the raw-manifest renderer in sync if manifest structure changes.

### 2.5 Swagger Documentation
- **Task**: Ensure Swagger documentation is updated for the new release.
- **Action**: Generate/Update Swagger docs if necessary and include them in the commit.

### 2.6 Docker Hub README
- **File**: `dockerhub-readme.md`
- **Action**: Update `dockerhub-readme.md` with the same format in each release.
  - Ensure version references in the Tags section are current (if applicable).
  - Keep the format consistent with previous releases.
  - The file will be automatically pushed to Docker Hub by the CI pipeline after the image is built.

## 3. Git Operations
1. **Stage**: `git add VERSION CHANGELOG.md README.md dockerhub-readme.md deploy/dkonsole.yaml scripts/render-manifest.sh backend/main.go .github/workflows/*.yml docs/RELEASE.md`
2. **Commit**: `git commit -m "chore: release vVERSION"`
3. **Push**: `git push origin main`
4. **Tag**: `git tag -a vVERSION -m "Release vVERSION"`
   > [!IMPORTANT]
   > **Pipeline Trigger**: Pushing this tag will trigger the GitHub Release Pipeline.
   > **Pre-requisite**: You MUST confirm that the build works and all tests pass locally or in a previous CI run BEFORE pushing this tag.
5. **Push Tag**: `git push origin vVERSION`

## 4. Docker Image Build and Tagging
After the GitHub Actions pipeline completes (or manually):

1. **Pull the new version image** (if not already local):
   ```bash
   docker pull dkonsole/dkonsole:VERSION
   ```

2. **Tag as latest**:
   ```bash
   docker tag dkonsole/dkonsole:VERSION dkonsole/dkonsole:latest
   ```

3. **Push both tags**:
   ```bash
   docker push dkonsole/dkonsole:VERSION
   docker push dkonsole/dkonsole:latest
   ```

   > [!IMPORTANT]
   > **Latest Tag**: The `latest` tag MUST be created and pushed for every release to ensure users can pull the most recent stable version.

## 5. Verification
- Monitor GitHub Actions for the tag push.
- Verify Docker Hub for the new tag `dkonsole/dkonsole:VERSION`.
- Verify Docker Hub for the updated `dkonsole/dkonsole:latest` tag.
