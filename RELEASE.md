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
- **Action 2 (Checkout)**: Update git checkout command regex `git checkout v[0-9.]+` to `git checkout vNEW_VERSION`.
- **Action 3 (Docker Tag)**: Update Docker tag regex `tag: "[0-9.]+"` to `tag: "NEW_VERSION"`.
- **Action 4 (Image Ref)**: Update image reference regex `dkonsole:[0-9.]+` to `dkonsole:NEW_VERSION`.
- **Action 5 (Changelog)**:
  - **IMPORTANT**: The README changelog section must contain ONLY the last 3 versions (NEW_VERSION and the 2 previous ones).
  - Add the new changelog entry at the top of the "Changelog" section.
  - Remove the oldest version entry if there are more than 3 versions.
  - Keep the format consistent with CHANGELOG.md entries.

### 2.4 Helm Chart
- **File**: `helm/dkonsole/Chart.yaml`
- **Action**: Update `version` and `appVersion` fields.
- **File**: `helm/dkonsole/values.yaml`
- **Action**: Update `image.tag` field.

### 2.5 Swagger Documentation
- **Task**: Ensure Swagger documentation is updated for the new release.
- **Action**: Generate/Update Swagger docs if necessary and include them in the commit.

## 3. Git Operations
1. **Stage**: `git add VERSION CHANGELOG.md README.md helm/dkonsole/Chart.yaml helm/dkonsole/values.yaml`
2. **Commit**: `git commit -m "chore: release vVERSION"`
3. **Push**: `git push origin main`
4. **Tag**: `git tag -a vVERSION -m "Release vVERSION"`
   > [!IMPORTANT]
   > **Pipeline Trigger**: Pushing this tag will trigger the GitHub Release Pipeline.
   > **Pre-requisite**: You MUST confirm that the build works and all tests pass locally or in a previous CI run BEFORE pushing this tag.
5. **Push Tag**: `git push origin vVERSION`

## 4. Verification
- Monitor GitHub Actions for the tag push.
- Verify Docker Hub for the new tag `dkonsole/dkonsole:VERSION`.
