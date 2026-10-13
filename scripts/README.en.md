# Songloft Scripts Directory

This directory contains automation scripts for the Songloft backend, covering version management, build helpers, submodule syncing, and more. Multi-platform binary builds, Docker image packaging, and GitHub Release creation are all handled by [`.github/workflows/release.yml`](../.github/workflows/release.yml).

## 📋 Script List

| Script | Purpose |
|------|------|
| `bump-version.sh` | Bumps `VERSION` in `Makefile` and the Swagger `@version` in `main.go`, then commits + tags + pushes (once the tag is pushed, the release workflow takes over the release) |
| `submodule-update.sh` | Batch-syncs all git submodules to the latest main |
| `submodule-push.sh` | Batch-pushes all submodules that have unpushed commits to their remote |
| `docker-entrypoint.sh` | Startup entrypoint inside the Docker image (not invoked directly) |
| `plugin-build.sh` | Builds a single JS plugin, outputting a `.jsplugin.zip` |
| `plugin-release.sh` | Uploads a `.jsplugin.zip` to the corresponding GitHub Release |
| `generate-icons.sh` | Generates all PNG/ICO icons from `favicon.svg` (requires Node.js 18+) |
| `fetch-issues.mjs` / `sync-docs.mjs` | Documentation sync helpers |
| `test_tag.sh` | Manual smoke-test script for the `pkg/tag` command-line tools |
| `ai-review-context.mjs` | Read-only parent/recursive-submodule inventory, open Issues, commits, and worktree fingerprints; writes partial reports and exits 1 for gaps |

## AI Development Evidence

Requires Node.js 22+, Git, and authenticated `gh` for Issues. Run from the parent with unique output filenames outside the worktree; existing files are never overwritten:

```bash
node scripts/ai-review-context.mjs inventory --output /tmp/songloft-inventory-20261013-01.json
node scripts/ai-review-context.mjs issues --output /tmp/songloft-issues-20261013-01.json
node scripts/ai-review-context.mjs commits --since 2026-10-12T00:00:00+08:00 --until 2026-10-13T00:00:00+08:00 --output /tmp/songloft-review-20261013-01.json
node --test scripts/tests/ai-review-context.test.mjs
```

Issues are fully paginated, exclude PRs, retain assignees, and deduplicate repositories. Uninitialized modules still have Issues queried, but unknown nested manifests produce gaps. `complete=false` / exit 1 means incomplete coverage, not an empty backlog. Commits use local history without establishing remote freshness. Collection does not assess fixes, update lessons, or schedule an AI. See the [AI workflow](../docs/en/ai-development.md).

## 🚀 Release Workflow

```bash
# 1. Bump the version locally + create a tag + push
./scripts/bump-version.sh patch        # x.y.z → x.y.(z+1)
# or: make bump TYPE=patch
```

Once the tag is pushed, [`release.yml`](../.github/workflows/release.yml) automatically:

1. Builds lite + full editions of the binaries for 7 platforms (Linux amd64/arm64/armv7, macOS amd64/arm64, Windows amd64/arm64)
2. Builds and pushes multi-architecture Docker images (linux/amd64, linux/arm64, linux/arm/v7) to Docker Hub
3. Generates sha256 checksums
4. Creates a GitHub Release and uploads all artifacts
5. Uses [`requarks/changelog-action`](https://github.com/requarks/changelog-action) to generate a changelog from Conventional Commits, uses it as the Release Notes, and prepends the content to the top of `CHANGELOG.md`, committing it back to `main`

## 🔁 submodule-update.sh

Batch-syncs all submodules (`clients/player` / `plugins/toolchain` / `pkg/tag` / `plugins/src/*` / `jsplugins`, etc.) to their respective main branches:

```bash
./scripts/submodule-update.sh
```

## 🔌 JS Plugin Build Scripts

| Script | Purpose |
|------|------|
| `plugin-build.sh <plugin-name>` | Enters `plugins/src/<plugin-name>/`, runs `pnpm install && pnpm run build`, and copies the artifacts into `jsplugins/` |
| `plugin-release.sh <plugin-name>` | Uploads the corresponding `.jsplugin.zip` to the GitHub Release of that plugin's submodule |

See `plugins/toolchain/README.md` for the detailed plugin development workflow.

## 🧪 test_tag.sh

A manual smoke test for the `cmd/tag`, `cmd/sum`, and `cmd/check` command-line tools provided by `pkg/tag`. Run `go install ./pkg/tag/cmd/...` before executing it.

## 📍 Repository Notes

- **Code & release repository**: https://github.com/songloft-org/songloft
- **Docker image**: [songloft/songloft](https://hub.docker.com/r/songloft/songloft) (keeps the original namespace)
