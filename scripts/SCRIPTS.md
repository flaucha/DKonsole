# Build & Release Scripts

This directory (`scripts/`) contains scripts for building and releasing DKonsole.

## Scripts Overview

### 🔨 `build.sh` - Build & Push Docker Images
Builds and pushes Docker images to Docker Hub **without** creating git tags.

**Use this when:**
- You want to rebuild images for the current version
- Testing changes before creating a release
- Updating images without bumping version

**Usage:**
```bash
cd scripts
chmod +x build.sh
./build.sh
```

**What it does:**
1. ✅ Builds backend Docker image
2. ✅ Builds frontend Docker image
3. ✅ Pushes both images to Docker Hub

---

### 🚀 `release.sh` - Full Release Process
Builds, pushes Docker images, **and** creates/pushes git tags.

**Use this when:**
- Creating a new version release
- You want to tag the release in git
- Publishing a new version

**Usage:**
```bash
cd scripts
chmod +x release.sh
./release.sh
```

**What it does:**
1. ✅ Checks for uncommitted changes (warns if found)
2. ✅ Builds backend Docker image
3. ✅ Builds frontend Docker image
4. ✅ Pushes both images to Docker Hub
5. ✅ Creates annotated git tag (e.g., `v1.0.4`)
6. ✅ Pushes git tag to remote repository

**Features:**
- Checks if tag already exists and asks to recreate
- Warns about uncommitted changes
- Creates detailed annotated tags with release notes
- Interactive prompts for safety

---

### ⚠️ `deploy.sh` - DEPRECATED
Legacy script kept for backward compatibility. Use `build.sh` instead.

---

## Version Management

The version is defined at the top of each script:

```bash
VERSION="1.0.4"
```

**To release a new version:**

1. Update the version in scripts:
   - `scripts/build.sh`
   - `scripts/release.sh`

2. Update version in:
   - `helm/dkonsole/Chart.yaml` (version and appVersion)
   - `helm/dkonsole/values.yaml` (image tags)
   - `gitops/apps/dkonsole/values.yaml` (image tags)
   - `README.md` (version badge and references)

3. Commit your changes:
   ```bash
   git add .
   git commit -m "chore: bump version to X.Y.Z"
   ```

4. Run the release script:
   ```bash
   ./release.sh
   ```

---

## Examples

### Rebuild current version images
```bash
cd scripts
./build.sh
```

### Create a new release
```bash
# 1. Update version in scripts and files
# 2. Commit changes
git add .
git commit -m "chore: bump version to 1.0.5"

# 3. Run release
cd scripts
./release.sh
```

### Manual git tag (if needed)
```bash
git tag -a v1.0.4 -m "Release v1.0.4"
git push origin v1.0.4
```

---

## Prerequisites

- Docker installed and logged in (`docker login`)
- Git configured with push access to the repository
- Bash shell (Linux, macOS, or WSL on Windows)

---

## Troubleshooting

**"Permission denied" error:**
```bash
cd scripts
chmod +x build.sh release.sh
```

**Tag already exists:**
The `release.sh` script will ask if you want to delete and recreate it.

**Uncommitted changes:**
The `release.sh` script will warn you and ask if you want to continue.
