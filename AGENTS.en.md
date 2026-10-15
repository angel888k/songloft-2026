# AGENTS.md

This is Songloft's AI development entry point: shared hard rules, common commands, and task navigation, with business details loaded on demand. Code is the source of truth for directories, dependencies, APIs, and schemas.

Before editing, read applicable `LESSONS.md` entries, local submodule instructions, and relevant sections from the [business-constraint index](#business-constraints) below; unrelated tasks need not load every topic. See the [Chinese version](AGENTS.md).

> **Detailed docs**:
> - AI development: [Workflow and retrospectives](docs/en/ai-development.md) · [Lessons](LESSONS.en.md)
> - Business constraints: [Browse by topic](docs/en/pitfalls.md) (scanning, lyrics, fingerprints, plugin store, proxies, caching, and playback)
> - Architecture: [Overview](docs/en/architecture.md) · [Backend](docs/en/architecture_backend.md) · [Frontend](docs/en/architecture_frontend.md)
> - Topics: [Database operations](docs/en/database_migrations.md) · [Color system](docs/en/color_system.md) · [API response format](docs/en/api_response.md) · [Quick start](docs/en/quick-start.md) · [Frontend gotchas](docs/en/frontend_gotchas.md)
> - Plugin development: see `plugins/toolchain/README.md` (separate repo)
> - Plugin registry authoring: [Plugin registry authoring guide](docs/en/plugin_registry.md)
> - API: after starting in dev mode, visit `/swagger/index.html`

---

## Project Overview

Songloft is a self-hosted local music server that supports both **server deployment** and **Bundle local mode** (embedding the Go backend into the client, so no separate server deployment is required). It has a multi-repo structure:

| Directory | Tech | Description |
|------|------|------|
| `/` | Go 1.26 + Chi v5 + SQLite | Backend API service (default port 58091, account admin/admin) |
| `/mobile` | Go + gomobile | Mobile binding entry point for the Go backend (for gomobile bind; exports Start/Stop/IsRunning/GetPort) |
| `/clients/player` ([separate repo](https://github.com/songloft-org/songloft-player)) | Flutter 3.41+ / Dart 3.7+ | Cross-platform frontend (6 platforms), supports Bundle local mode |
| `/clients/player-lynx` ([separate repo](https://github.com/songloft-org/songloft-player-lynx)) | TS + ReactLynx | Android/iOS/HarmonyOS/Web preview client, separate main dev builds and v* releases (submodule; no desktop/Bundle) |
| `/clients/tv` ([separate repo](https://github.com/songloft-org/songloft-tv)) | Kotlin | Android TV client (submodule) |
| `/plugins/toolchain` ([separate repo](https://github.com/songloft-org/plugin-toolchain)) | TS + pnpm | JS plugin development toolchain (SDK / Builder / scaffolding) |
| `/plugins/src` | TS | JS plugin source code (a collection of submodules; each plugin distributes releases from its own repo) |
| `/pkg/tag` | Go | Audio metadata **read/write** library (extends the upstream tag library with MP3/FLAC writing) |
| `/integrations/home-assistant` ([separate repo](https://github.com/songloft-org/home-assistant-addon)) | HA add-on | Home Assistant add-on (thin layer reusing the Docker image; submodule). **The `repository.yaml` manifest must sit at that repo's root** — which is exactly why it was split out (#340). Design/pitfalls/version sync: see [integrations/home-assistant/README.md](integrations/home-assistant/README.md) |
| `/tools/ffmpeg-builder` ([separate repo](https://github.com/hanxi/ffmpeg-builder)) | Docker | Minimal-image builder for statically compiled ffmpeg/ffprobe (submodule); used for download transcode / audio fingerprinting |
| `/tools/tracely` ([separate repo](https://github.com/hanxi/tracely)) | Go + Vue | Self-hosted frontend monitoring backend (install/upgrade tracking); the backend reports via its Go SDK. Submodule; the SDK dependency comes from go.mod |

---

## Ownership of Repository Instructions

- The parent repository documents backend implementation, shared host capabilities, interface contracts, and conventions used across repositories.
- Each submodule keeps its business implementation, device compatibility, historical issues, regression matrix, and dedicated validation commands in its own root `AGENTS.md`, with a synchronized `AGENTS.en.md`.
- Read local instructions before editing a submodule, and check the parent instructions when host interfaces are involved. Commit the submodule before updating its pointer in the parent.

## AI Development Workflow

- Once assigned a task, investigate, implement, validate, review, and deliver continuously without separate stage approvals. Ask for essential missing information, product decisions, or scope expansion. Commits, pushes, formal releases, and Issue writes follow distinct authorization; do not ask again for already authorized actions.
- After pushing an Issue-related fix, proactively prepare and show the reply draft and target Issue for comment approval; post directly when applicable comment authorization already exists. Update drafts as CI and release status change without waiting for the user to ask. See [Linked Issue Replies](docs/en/ai-development.md#linked-issue-replies).
- Read applicable `LESSONS.md` entries before work and record evidence-backed lessons afterward, synchronizing English/Chinese. Three independent events on distinct dates make promotion candidates; re-reviewing one incident does not increase counts. Existing hard rules apply immediately.
- Unresolved Issue triage includes parent and all submodules, with recursive discovery, deduplication, and full pagination. Use `owner/repo#number` and report gaps; assess bodies, all comments, current code, and validation rather than treating open status as proof of an unfixed bug.
- Bind validation to repository, SHA or uncommitted-worktree fingerprint, environment, commands, and results; list relevant `Not run` checks. Changes require rerunning affected checks; follow required CI for the same SHA, and never count cancelled/skipped as passed. See the [AI workflow](docs/en/ai-development.md).

## Common Commands

```bash
# Backend
make run            # Start (dev mode, with Swagger)
make build          # Build dev version (full, embeds frontend)
make build-lite     # Build dev version (slim, no embedded frontend)
make build-prod     # Build production version (full, embeds frontend)
make build-prod-lite # Build production version (slim, no frontend)
make test           # Test
make check          # fmt + vet + test
make sqlc           # Regenerate sqlc code (must run after editing queries/*.sql)
make swagger        # Regenerate API docs

# Frontend build (artifacts land in clients/player-build/, for backend embedding or standalone deployment)
make build-frontend-web-embedded   # For embedding in the Go binary (hides the API-address UI)
make build-frontend-web            # Standalone web deployment
make build-frontend-{linux,windows,macos,android,ios,all}

# Bundle local mode (Go backend compiled into a mobile library / desktop executable)
make build-go-mobile-android       # Android .aar (gomobile bind, arm64 + arm + x86_64)
make build-go-mobile-ios           # iOS .xcframework (gomobile bind, arm64, macOS only)
make build-go-desktop-linux        # Linux executable
make build-go-desktop-windows      # Windows .exe
make build-go-desktop-macos        # macOS x86_64
make build-go-desktop-macos-arm64  # macOS ARM64

# Frontend development
cd clients/player && flutter run -d chrome          # standalone
cd clients/player && flutter run -d chrome --dart-define=DEPLOY_MODE=embedded
```

---

## Code Formatting (Hard Rule)

After every code change, you **must** format the code before committing:

- **Go**: Run `gofmt -w .` from the project root
- **Dart**: Run `dart format lib/ test/` from the `clients/player/` directory

---

## Database Conventions (Hard Rules)

> For the complete procedure see [docs/en/database_migrations.md](docs/en/database_migrations.md).

Access stack: **goose migrations + sqlc fixed SQL + squirrel dynamic SQL + Repository + UnitOfWork**.

- **Changing the schema** → `internal/database/migrations/000N_xxx.sql`, executed automatically by `goose.Up` at startup; **do not** manually `ALTER data/songloft.db`
- **Adding fixed SQL** → `internal/database/queries/{table}.sql` + `make sqlc`; the generated output `internal/database/sqlc/` must be committed
- **Dynamic SQL (variable-length WHERE/SET)** → use squirrel inside `*_repository.go`; do not concatenate strings
- **Cross-table writes** → `db.RunInTx(ctx, func(ctx, uow))` to obtain `uow.Songs/Playlists/...` under the same `*sql.Tx`; **do not** call `BeginTx` manually in the service layer, or you'll hit SQLITE_BUSY
- **Error semantics** → repository misses uniformly return `database.ErrNotFound`; services distinguish with `errors.Is`
- **Testing** → use `testutil.OpenMemoryDB(t)` to run a real `:memory:` DB + real Repository; **do not** hand-write a mockDB
- **Built-in data** → migrations preload playlists id=1 "Favorites" and id=2 "Radio Favorites" (`labels=["built_in"]`), plus default config for `music_path / jwt_secret / source_*`. Remember to subtract these when asserting row counts in tests
- **Never write CJK comments after `-- name:` in `queries/*.sql`** → sqlc embeds query comments into generated code as doc comments **byte-wise**; multi-byte characters get truncated into invalid UTF-8 and `make sqlc` fails outright (`error generating code: illegal UTF-8 encoding`). Put such comments on the Go repository method instead
- **Qualify outer columns in self-referencing subqueries** → in `DELETE FROM t WHERE col = ? AND id NOT IN (SELECT x.id FROM t x ...)`, a bare outer `col` is rejected by sqlc as `column reference is ambiguous`; write `t.col`

---

## Backend Coding Conventions

- Standard Go layout (`internal/` guards against external dependencies), Chi v5 routing, JWT dual-token
- Dependency injection: ordinary reads/writes use Repository interfaces; cross-table transactions reuse existing `RunInTx + UnitOfWork` injection (for example, `PlayHistoryService` receives `database.DB`). Services **must not** call `BeginTx` themselves or bypass repositories with SQL.
- Logging: standard library `slog`; HTTP errors: uniformly `respondError`
- **API response format**: return RESTful data directly, **no** `{code, data, message}` envelope; errors are uniformly `{"error","detail"}`. For the full spec see [docs/en/api_response.md](docs/en/api_response.md)
- No ORM: fixed SQL → sqlc, dynamic SQL → squirrel, cross-table writes → `RunInTx + UnitOfWork`
- Test files `*_test.go` live in the same directory as the source

---

## API Documentation Conventions (Hard Rules)

**Every handler method registered in `internal/app/routers.go` (including sub-registration functions such as `RegisterStaticRoutes` / `RegisterAPIRoutes`) must have swag annotations.** The backend API docs are generated by [swaggo/swag](https://github.com/swaggo/swag) from these annotations and are the single source of truth for frontend development and external integration.

### Required fields (every handler has at least these 7)

```go
// @Summary <one-line summary in Chinese>
// @Description <detailed description, may span multiple lines; clarify side effects / defaults / error-code trigger conditions>
// @Tags <business group, in Chinese>
// @Produce json
// @Success 200 {object} <return type> "<description>"
// @Security BearerAuth
// @Router /<path> [<method>]
func (h *XxxHandler) Method(w http.ResponseWriter, r *http.Request) { ... }
```

- Endpoints with a request body additionally add `@Accept json` and `@Param request body <type> true "<description>"`
- Endpoints with obvious error paths add `@Failure 400/404/500 {object} map[string]string "..."`
- Path/query parameters use `@Param <name> path/query <type> true/false "<description>"`
- **Public endpoints** (no token required, e.g. health checks) omit `@Security BearerAuth`
- **Business tag naming**: reuse existing tags (「歌曲管理」「歌单管理」「电台与 HLS」「扫描管理」「配置管理」「缓存管理」「JS 插件」「JS插件管理」「数据备份」「设置」「系统升级」「认证管理」「系统管理」「资源代理」「主题包」「播放历史」「标签管理」); do not casually invent new tags
- **`@Router` paths must NOT contain the `/api/v1` prefix**: `main.go` already declares `// @BasePath /api/v1`, and swag automatically prepends this prefix to every `@Router` path. Including `/api/v1` in the annotation will cause the generated docs to have paths like `/api/v1/api/v1/...`. Always use relative paths (e.g. `/songs/{id}/tags`), never `/api/v1/songs/{id}/tags`.

### Multi-alias / catch-all routes

- A handler registered under multiple alias paths (e.g. `/songs/{id}/play` and `/songs/{id}/play.m3u8`) → write one `@Router` line per alias
- HEAD is a subset of GET; **do not list it separately**; OpenAPI does not require it
- A catch-all like `r.HandleFunc(...)` that accepts ANY HTTP method → list all methods actually possible (`[get] [post] [put] [delete]`), one `@Router` line each
- Dynamic paths (`{entryPath}` determined at runtime by the installed plugins) → note in `@Description`: "dynamic route, {xxx} is determined at runtime, OpenAPI serves only as a placeholder"

### Must run after changes

After modifying/adding handler annotations you must run `make swagger`: it regenerates `docs/swagger.json`, `docs/swagger.yaml`, and `docs/docs.go`, and **these outputs must be committed**. Otherwise `/swagger/index.html` will be out of sync with the code, and the frontend will hit pitfalls integrating against the stale docs.

### Verification

- Search the `make swagger` output for your newly added `@Router` path, and confirm that `Generating <Type>` includes the request/response types you just wrote
- `grep '<your-new-path>' docs/swagger.json` should return a hit
- Start `make run`, visit `http://localhost:58091/swagger/index.html`, and open the new endpoint in the UI to eyeball it

### No exemptions

"Everything registered in routers must be annotated" is an absolute rule. Even dynamic-route catch-alls, static-resource handlers, and reverse-proxy endpoints must have swag — just make the `@Description` clear about "what it is and why the OpenAPI schema is imprecise."

---

## Configuration Endpoint Conventions (Hard Rules)

The project has two kinds of configuration endpoints. **User-visible feature toggles always go through business endpoints**, while the generic KV store is only an admin entry point.

### `/api/v1/settings/<name>` — Standalone config endpoints (frontend business features default here)

- Path style: `/settings/<kebab-case-name>` (e.g. `/settings/hls-proxy`, `/settings/music-path`, `/settings/http-proxy`, `/settings/library-browse`, `/settings/proxy-private-allowlist`)
- Data shape: **strongly typed** JSON (e.g. `{enabled: bool}` or an aggregate object), not `{value: string}`
- Default values: handled inside the handler (when config is missing, GET returns the business default; PUT just writes directly, **the frontend need not POST-create first**)
- Side effects: triggered directly inside PUT (e.g. after a `music_path` PUT, asynchronously `onMusicPathChanged` rebuilds the Scanner)
- Ownership: placed in the corresponding business module's handler (e.g. hls-proxy in `HLSHandler`, music-path in `ScanHandler`), which also holds `*services.ConfigService` to do the read/write
- Naming pattern: `Is<Name>Enabled() / Set<Name>Enabled(bool)` business methods + `Get<Name>Setting / Update<Name>Setting` HTTP handlers + `/settings/<name>` route

### `/api/v1/<module>/*` — Business module aggregate endpoints (config included)

Some business modules come with an "action endpoint + config endpoint" combo (the canonical example being `/cache-manage/{stats,clean,config}`); in this case the config endpoint **stays under the module prefix** rather than being forcibly split out into `/settings/`.

- When applicable: the config is strongly related to the module's other action endpoints (e.g. cache's `config` shares the same `CacheService` as `stats/clean`)
- Rationale: the industry mainstream (AWS, GitHub, Discord) all aggregate by business module; the GitLab-style hybrid of "globally centralized, module-dispersed" is also acceptable
- Existing example: `/api/v1/cache-manage/config` (GET/PUT)
- **Decision criteria**:
  - **Standalone** config (belongs to no business module, or is shared across modules) → `/settings/<name>`
  - **In-module** config (strongly related to the module's action endpoints) → `/<module>/config` or `/<module>/<sub-name>`

### `/api/v1/configs/{key}` — Generic KV (admin editor only)

- Only for use by frontend **generic config editors** like `config_manager.dart`, letting admins hand-edit arbitrary key/value for debugging
- **New business features must not call `/configs/{key}` directly**: the generic PUT returns 404 when the key doesn't exist, and it has no strong typing, no side effects, no default values
- After a business wrapper is in place, the generic endpoint can still modify the same key (dual entry points are retained), but the side effect must also be hooked into the `configHandler.SetOnConfigChanged` callback (see `musicPathChanged` in `routers.go`), ensuring both entry points have consistent semantics

### Client conventions

- `SettingsApi` (`clients/player/lib/features/settings/data/settings_api.dart`) wraps all `/settings/*` calls; business-feature Providers always go through it
- `ConfigApi` is used only in `config_manager.dart` and admin UIs like "list all configs"

### Historical decision record

- This convention was introduced in 2026-06. Background: `hls_proxy_enabled` was not preloaded by default, causing PUT `/configs/{key}` to return 404, which revealed that the project had three coexisting styles — `/configs` + `/settings/*` + `/cache-manage/config`
- Chosen direction: business endpoints are the **single source** for user-visible entry points, and the generic KV degrades to an admin back door

---

## Bilingual Documentation Sync (Hard Rule)

Project docs are **maintained in both Chinese and English**. When you change either language version, you **must** apply the corresponding change to the other version — never edit one side only and let the two drift apart.

- **Mappings**:
  - `README.md` ↔ `README.en.md`
  - `AGENTS.md` ↔ `AGENTS.en.md`
  - `docs/<name>.md` ↔ `docs/en/<name>.md` (same filename; English version lives under `docs/en/`)
- **Criterion**: any add / modify / delete of documentation **content, structure, or links** (body text, sections, tables, navigation links, etc.) must land in both language versions; only the wording is localized, the structure stays consistent
- **Check the counterpart exists first**: if a same-named file exists under `docs/en/`, sync it; README and AGENTS always have an `.en.md` counterpart
- **Exception**: some content is inherently single-language (e.g. a community note only in the Chinese version) — no mirror is required, but make sure that is intentional rather than an omission

---

## Docs Site Structure (docs/ — VitePress Custom Theme)

The Songloft docs site (`docs/`) uses **VitePress + a custom theme** (`docs/.vitepress/theme/`), **not the default theme**. Identify each page's source before editing; do not edit generated output directly:

- **Custom landing page (edit data, not markdown)**: the home page `docs/index.md` is a single line `<Landing />`; its content is driven by structured data in `docs/.vitepress/data/*.ts` (install methods `downloads.ts`, features `features.ts`, copy `landing-i18n.ts`) and rendered by `docs/.vitepress/theme/components/landing/*.vue`. To change the landing page → edit `data/*.ts` (bilingual `{zh,en}` fields); align icons with the mapping table inside the component (e.g. `ICONS` in `LandingInstaller.vue`).
- **Auto-generated pages (do NOT hand-edit)**: `docs/quick-start.md`, `docs/en/quick-start.md`, and `docs/changelog.md` are generated by `scripts/sync-docs.mjs` from the root `README.md` / `README.en.md` / `CHANGELOG.md`, and are ignored via `docs/.gitignore`. To change the body → edit the source `README` / `CHANGELOG`; `docs:dev` / `docs:build` runs `sync` first to regenerate. **Manual edits get overwritten and are never committed.**
- **Submodule-synced pages (also do NOT hand-edit — but the source lives in another repo)**: `docs/addon/`, `docs/player/`, `docs/player-lynx/` (and English `docs/en/player-lynx/`), and `docs/plugin-toolchain/` are synced by `sync-docs.mjs` from the `integrations/home-assistant/`, `clients/player/docs/cn/`, `clients/player-lynx/`, and `plugins/toolchain/` **submodules** respectively, and are likewise ignored via `docs/.gitignore`. To change the body → **edit it in the corresponding submodule repo, then come back and bump the submodule pointer** (inspect the submodule worktree and target commit first, then commit and update the parent gitlink as authorized), otherwise the docs site keeps showing stale content. Two things to note: (1) the `to:` target path is **deliberately decoupled** from the source path (e.g. `integrations/home-assistant/README.md` → `docs/addon/index.md`), because public URLs like `/addon/` are already in the sitemap and must not follow source-repo renames; (2) when a submodule is not checked out, `sync` only warns instead of failing and the page **silently disappears** — which is why the submodule init list in `static.yml` must include them.
- **repowiki (`docs/repowiki/` — AI-generated by hand, manually maintained)**: this wiki (system architecture / data model / backend / frontend / API / plugins / deployment / troubleshooting / dev guide, etc.) is **generated by hand from source files by an AI** — there is **no generator script** in the repo and it is not an external-service artifact; the **committed markdown is the single source of truth**. Any tool (AI or human) edits it directly and commits — **do not wait for a "regeneration"**. Keep these pages in sync with code changes as needed, verifying against the code just like any other source doc (e.g. after renaming public assets in `internal/jsplugin/assets/`, update any repowiki references to the old file name too).

---

## Git Commit Conventions

- **Commit directly to the `main` branch** — do not create feature branches or open PRs (this repo's convention)
- Commit messages **must not** include a `Co-Authored-By` trailer
- Follow the Conventional Commits format: `type(scope): description`, prefer Chinese for description and body
- Commit messages that reference a GitHub issue must include the issue reference
- Issue reference rules: the short form `#123` always points to an issue in **the repo where the commit lives**; whenever the referenced issue is not in the current repo, you must write the full `owner/repo#123`
  - A commit in the parent repo `songloft-org/songloft` referencing a parent-repo issue: may write `#155`, or `songloft-org/songloft#155`
  - A commit in a submodule repo (such as `pkg/tag`, `clients/player`, `plugins/toolchain`, `plugins/src/*`) referencing an issue in its own repo: may write `#14`, or the full repo path
  - A commit in a submodule repo referencing a parent-repo issue: must write the full path, e.g. `songloft-org/songloft#155`, not just `#155` (otherwise GitHub resolves it to an issue in the submodule's own repo)
  - Any cross-repo reference always uses the full path, e.g. `songloft-org/songloft-player#14`

---

## Build and Deployment

- Build tags: `dev` (includes Swagger + pprof) / `lite` (slim version, no embedded frontend) / no tag (full version, embeds Flutter Web)
- When `VERSION=dev`, the Makefile automatically enables `-tags dev` (no need to manually pass `EXTRA_TAGS=dev`)
- Two orthogonal dimensions: **VERSION** (`dev` / `X.Y.Z`) controls whether it's a dev build; **BUILD_TYPE** (`lite` / empty i.e. `full`) controls whether the frontend is embedded. **Do not** use mixed values like `BUILD_TYPE=dev`
- The embed path is `clients/player-build/web-embedded` (**not** `clients/player/build/web-embedded`)
- SPA fallback: `internal/app/embed.go` returns `index.html` only for missing paths whose final segment contains no `.`; missing static assets return 404, never HTML masquerading as JS/JSON.
- Deployment mode is switched via `--dart-define=DEPLOY_MODE=embedded|standalone`; `AppConfig.isEmbedded` is a compile-time constant, and tree-shaking removes the API-address UI in embedded mode
- Sub-path deployment: configured at startup via `-base-path /xxx` or `BASE_PATH=/xxx`; the backend strips the prefix at the outermost layer with `http.StripPrefix`, and `embed.go` replaces `<base href="/">` with `<base href="/xxx/">` at runtime; in embedded mode the frontend auto-detects the sub-path from `Uri.base.path`

### Bundle Local Mode (v2.9.0+)

Embeds the Go backend into the Flutter client so users can use it without deploying a separate server. Enabled at compile time with `--dart-define=HAS_BACKEND=true`.

- **Mobile (Android/iOS)**: `gomobile bind` compiles the Go backend into a native library (`.aar` / `.xcframework`), and Flutter calls it via `MethodChannel('com.songloft/backend')`
- **Desktop (macOS/Windows/Linux)**: the Go backend is compiled into a standalone executable `songloft-server`, which Flutter runs as a subprocess at startup
- **Web**: Bundle mode is not supported (remote server only)
- Run mode: `RunMode.local` / `RunMode.remote`, persisted to SharedPreferences and auto-restored at startup
- Local-mode startup flow: request storage permission → start the embedded backend (`127.0.0.1:<port>`) → health-check polling (up to 10 × 300ms) → auto-login with `admin/admin`
- `BackendLifecycle` (WidgetsBindingObserver): auto-restarts the backend when the app returns to the foreground, stops it on detached
- Key entry points: `mobile/mobile.go` (gomobile binding), `clients/player/lib/core/backend/` (Flutter-side abstraction layer)
- CI artifacts start with `songloft-bundled-`; Android/Linux/Windows include architecture, while Apple uses `songloft-bundled-macos.zip` and `songloft-bundled-ios-nosign.ipa`. There are 4 parallel jobs (Android/Linux/Apple/Windows); failures do not block the main Release.

### Docker Hot-Swap Rules (`scripts/docker-entrypoint.sh`)

The Docker image contains a base package `/app/songloft`, while the persistent data volume holds the actually-running `/app/data/songloft`. On container startup the entrypoint decides whether to overwrite the data directory with the base package:

**Core principle: the base package represents the user's intent; when dev/release or full/lite differ, overwrite with the base package. Only compare old vs. new when "same channel + same BUILD_TYPE": dev by Build Time, release by version number.**

Earlier branches: initialize directly if the target file is missing. For an existing target, an unknown base channel preserves it; an unknown target channel is replaced by a recognized base. If both are stable but either version fails the script's `is_semver` check (such as a custom tag), use the base rather than attempting version comparison.

| Scenario | Behavior | Reason |
|------|------|------|
| dev ↔ release channel differs | Replace | The user switched image channels |
| BUILD_TYPE differs (full↔lite) | Replace | The user switched image variants |
| Both dev + same type + base package Build Time > data Build Time | Replace | Dev rolling builds pick the newest by build time |
| Both dev + same type + data Build Time >= base package Build Time | Don't replace | The data may have been upgraded online via the API |
| Both release + same type + base package version > data version | Replace | Release upgrade |
| Both release + same type + data version >= base package | Don't replace | The data may have been upgraded online via the API |

### Docker Non-Root Operation (PUID/PGID, songloft-org/songloft#380)

- **Unset by default = keeps running as root**, identical to prior behavior with zero migration risk. Non-root mode only activates when `PUID` or `PGID` is explicitly set (either one is enough; the unset one defaults to `1000`); the entrypoint drops privileges to that uid:gid via Alpine's built-in `su-exec` (lighter than `gosu`, ships in the official repo with no extra download) right before `exec`ing the main program
- **`/app/data` is recursively `chown`ed on every startup, `/app/music` only has its top-level directory `chown`ed, never recursively**: `/app/data` is small (db, covers, cache, etc.) and must be fixed up to remove ownership left over from a prior root-based run, otherwise the new user can't open the old database; `/app/music` can be a personal library with hundreds of thousands of files / multiple TB, and the IO cost of recursively walking it on every startup is unacceptable. Once the top-level directory is writable, newly downloaded/written files are already created with the correct target uid:gid — no follow-up fix needed
- **Pre-existing root-owned files left inside `/app/music` from before the upgrade are not auto-fixed** (e.g. old files that had tags written into them) — this is a deliberate performance trade-off, not an oversight. Set `FIX_MUSIC_PERMISSIONS=true` when you need a one-time recursive fixup; this is meant to be run manually once after switching to non-root, not as a default behavior
- **`integrations/home-assistant` is unaffected**: its `run.sh` overrides the image's `ENTRYPOINT` entirely, bypassing `docker-entrypoint.sh`; its permission model is managed separately by the HA supervisor

---

## Frontend UI Verification (Dockerized Headless Browser)

When you need to **actually exercise a change in the UI** (rather than only running `flutter test`), use a
headless Chrome inside Docker: on this repo's dev machine the host's `google-chrome --headless` core dumps
and never binds its CDP port, so don't waste time on it.

```bash
# 1. Build the frontend in embedded mode so the Go backend serves it same-origin
#    (skips the standalone API-address configuration step)
make build-frontend-web-embedded
go build -tags dev -o /tmp/songloft-full .
/tmp/songloft-full -port 58191 -db <tmpdir>/test.db -music <musicdir>

# 2. Start the browser container. --network host is what lets the containerized Chrome
#    reach the host's 127.0.0.1:58191
docker run -d --name uichrome --network host browserless/chrome:latest

# 3. Drive it through the /function endpoint (Content-Type must be application/javascript)
curl -s -X POST http://127.0.0.1:3000/function \
  -H 'Content-Type: application/javascript' --data-binary @script.js
```

Scripts look like `module.exports = async ({ page }) => { ...; return { data: {...}, type: 'application/json' } }`.
Take screenshots with `page.screenshot({ encoding: 'base64' })`, return them inside `data`, and base64-decode
them into PNGs locally.

### Pitfalls when driving Flutter Web

- **Flutter Web renders to a canvas, so there are no buttons in the DOM.** Interactive elements are only
  exposed as `<flt-semantics>` accessibility nodes: read `getBoundingClientRect()`, then
  `page.mouse.click(x, y)` on the center. The semantics tree is usually already enabled; if a
  `<flt-semantics-placeholder>` shows up on first paint, `.click()` it first
- **Semantics labels get merged**: a whole screen often collapses into one long aria-label node, and a
  button's text is not necessarily its own node. When you can't find it, **screenshot first and click by
  coordinates** instead of fighting label matching
- **Text fields**: `<input aria-label="Username">` is real DOM, but clicking by coordinates plus
  `keyboard.type` is more reliable; leave ~800ms between the click and the typing, otherwise Flutter has not
  established focus yet and the input is dropped
- **Verifying both languages**: override `navigator.language`/`languages` to `zh-CN` via
  `evaluateOnNewDocument`; Flutter l10n follows
- **Deep links**: `page.goto('/settings/category/2')` is a full reload and can produce a bogus
  "Failed to load" while the app is mid-boot. Navigate by clicking inside the app instead of using goto for
  nested routes

### Anchor assertions on backend-observable state

A screenshot only proves "it rendered correctly". Whether the interaction actually took effect needs separate
evidence — e.g. after flipping a toggle, `curl` the matching `/settings/<name>` endpoint and check the value
changed; after clicking "Stop computing", use `pgrep -x ffmpeg` to confirm the child processes hit zero.

- **Do not count processes with `ps -ef | grep <keyword> | wc -l`**: the current shell's own command line
  contains that keyword, so you consistently over-count by 1–2 and can easily conclude "it didn't stop".
  Use `pgrep -x <executable>`

---

## Platform Adaptation Pitfalls

- Upgrade checks (`/api/v1/upgrade/check`) also work outside Docker and return `is_docker`; upgrade execution and rollback remain Docker-only.
- Flutter `secure_storage` automatically falls back to SharedPreferences under an unsigned macOS sandbox
- Before an Android build you need `sdkmanager --licenses`; Android 13+ requires requesting notification permission at runtime
- All native platforms (Win/Linux/macOS/Android/iOS) uniformly use media_kit/libmpv as the audio backend (via `just_audio_media_kit` / the custom `SongloftJustAudioPlatform`), with no native fallback and no kill-switch
- HyperOS3 and similar need `androidStopForegroundOnPause: false` to prevent background reclamation
- **Bundle mode Android**: the CWD is `/`, so the covers directory path must be resolved relative to `DBPath` rather than the CWD (fixed in `da65db1`)
- **Bundle mode native bridging**: Android uses `Class.forName("mobile.Mobile")` reflection to call the gomobile-generated class; when the `.aar` isn't bundled, `isAvailable()` returns false (graceful degradation); iOS likewise uses Swift to call the Objective-C functions like `MobileStart`
- **Bundle desktop subprocess**: `DesktopBackendService` looks for `songloft-server` in the **same directory** as the Flutter executable (on macOS, `Contents/Resources/`), and parses the actual listening port from stdout

---

## JS Plugins

- Source at `plugins/src/<name>/`; build artifacts are in each plugin repo's GitHub Releases
- Create a new plugin: `npx create-songloft-plugin@latest` (interactive scaffolding; supports WebView / Lynx rendering engine templates; see `plugins/toolchain/README.md` for details)
- Sandbox: QuickJS. `fetch` is a built-in global function in the runtime (uses `__go_fetch_async`), and the `console` family (`songloft.log`, etc.) are log wrappers — neither are bridge calls. The bridge is dispatched by `HandleBridgeCall` in `internal/jsplugin/api_bridge.go`; namespaces include: `storage`, `persistent-storage`, `songs`, `playlists`, `tags`, `comm`, `plugin`, `command`, `jsenv`, `fs`, `net`, `websocket`
- Routing: `/api/v1/jsplugin/{entry_path}/...`
- Common assets: `/api/v1/jsplugin-assets/*` serves the `theme.css`/`components.css`/`common.js`/fonts embedded in the Go binary, which `injectHTMLHead` automatically injects into all plugin HTML pages (injection order: theme.css → components.css → common.js)
- Theme sync: `common.js` contains embed detection + theme bridging (URL `?theme=` parameter + real-time `postMessage` updates + `data-theme` attribute + `songloft-theme-change` event), and exposes the `window.SongloftPlugin` global API (`getTheme`/`onThemeChange`/`apiGet`/`apiPost`/`getCookies`, etc.)
- **Client host bridge (WebView)** (`@songloft/client-sdk`, built into `common.js`): Plugin frontend pages call Flutter client host capabilities via `window.SongloftPlugin.host` / `player` / `getCookies` / `invokeHost`. Native platforms use `flutter_inappwebview.callHandler('songloftHost', {ns, method, params})`, Web/iframe uses `postMessage` to parent window. Dispatch logic in `clients/player/lib/features/home/presentation/plugin_host_dispatch.dart` (transport-agnostic, web-safe), native bridge in `plugin_host_bridge.dart` (mixin, registers callHandler + injects platform-specific callbacks). Registered namespaces: `host` (getInfo), `player` (playback control), `cookies` (cookie reading), `favorite` (favorite state sync — `refresh` method, pass `{songId, isFavorited}` for incremental update of Flutter-side FavoriteNotifier cache, or omit params for full reload). **The object literal at the end of `common.js` is the single source of truth for what `window.SongloftPlugin` exposes**. **Plugins must not hand-write type declarations for host APIs**; use `SongloftPluginGlobal` from `@songloft/client-sdk`, or if you must hand-write them, check that literal in `common.js` first
- **Cookie reading bridge** (`window.SongloftPlugin.getCookies(origin)`): Reads cookies for a specified origin from the host WebView Cookie Store (including HttpOnly), returns `{name: value}` map. **Native clients only** (Android/iOS/macOS/Windows; Linux desktop has no WebView implementation) — Web platform cannot implement this due to browser same-origin policy; calls will reject. Implementation path: `common.js getCookies()` → `invokeHost('cookies', 'get', {origin})` → Flutter `PluginHostDispatcher` → `cookieProvider` callback → `CookieManager.instance().getCookies(url: WebUri(origin))`. Origin must include protocol+host (e.g. `https://example.com`); invalid formats are rejected by validation. Typical use case: session reuse for third-party gateways like FN Connect (user logs in within the app's WebView, plugin reads cookies for subsequent API calls)
- **Lynx native rendering bridge** (`@songloft/lynx-plugin-sdk`, for `renderEngine: "lynx"` only): plugins declaring `renderEngine: "lynx"` use ReactLynx to build UI, compiled into `.lynx.bundle`; the host loads them natively via Lynx `<frame>` elements. Communication goes through `NativeModules.SongloftPluginBridge` (three-platform native modules: Android `SongloftPluginBridgeModule.kt` / iOS `SongloftPluginBridgeModule.swift` / HarmonyOS `SongloftPluginBridgeModule.ets`), routing parent–child frame RPC + event push by `frameId`. The SDK provides `invokeHost(ns, method, params)` / `onPlayerState` / `onThemeChange` / `onPush` etc. **Does not use** `window.SongloftPlugin` / `@songloft/client-sdk` (those are WebView-only). At build time an `index.html` + `.web.bundle` is auto-generated as a WebView fallback for the Flutter client
- `theme.css` defines `--md-*` CSS variables (dual light/dark theme), font declarations and reset styles; `components.css` provides component library styles. Any plugin using these variables automatically follows theme switches
- Permissions: `permissions: ["net", "storage", "fs:music", ...]` in the manifest, validated at runtime by `internal/jsplugin`
- **fetch internal control headers**: `X-Fetch-No-Redirect` (don't follow redirects; required to collect `Set-Cookie` from intermediate hops), `X-Fetch-Timeout-Ms` (per-request timeout 100–30000ms), `X-Fetch-Insecure` (skip TLS verification, **requires the `net:insecure-tls` permission**). None of the three are ever forwarded upstream — `X-Fetch-Insecure` used to leak upstream as an ordinary header because it was unimplemented (songloft-org/songloft#401)
- **A new permission constant must be synced in three places, or plugins can't use it at all**: (1) `AllPermissions` in the host's `internal/jsplugin/permissions.go`; (2) `VALID_PERMISSIONS` in `plugins/toolchain/packages/plugin-builder/src/manifest.ts` — **this one hard-fails at build time** (`unknown permission: xxx`), so missing it means the plugin can't even produce a build artifact; (3) `AVAILABLE_PERMISSIONS` in `plugins/toolchain/packages/create-songloft-plugin/src/index.ts` (the scaffold's checklist). Separately, `ValidatePermissions` (host side) **deliberately only warns on unknown permissions rather than rejecting**: it sits on the mandatory install/update path, and rejecting would make a plugin that declares a new permission uninstallable on **older hosts**, taking down the plugin's otherwise-working features too; leniency is security-neutral (`CheckPermission` only honors strings it knows). **But older hosts with the strict check are already released**, so even after a new permission lands you must wait for host adoption before shipping the plugin — `minHostVersion` is **purely informational and never enforced by the host**, so don't expect it to gate anyone
- **`net:insecure-tls` permission (skip TLS certificate verification)**: the `jsruntime` layer **has no concept of permissions**, and `fetch` itself doesn't require `net`; this permission is projected by `service.go` calling `jsManager.SetAllowInsecureTLS(envID, ...)` after env creation into `JSEnv.allowInsecureTLS` (default false) — that flag is the permission's only runtime-layer manifestation. Key points: (1) **`CheckPermission` does no colon-prefix matching, so `net` does not cover `net:insecure-tls`** — deliberately, otherwise every existing plugin that already declares `net` would silently gain verification-skipping on upgrade while the user sees no change in the permission list; (2) a dedicated `insecureTransport` — do **not** reuse `sharedTransport` and mutate `TLSClientConfig` per request, since connection pools are per-Transport and that would poison connection reuse for every plugin; (3) child environments (created via `songloft.jsenv`) must inherit the parent plugin's policy, or "plugin issues a request from a child worker" fails verification for no visible reason; (4) when the header is present without the permission, emit `slog.Warn` rather than dropping it silently — plugin authors commonly assume "if the host doesn't recognize it, it's a no-op", and one log line saves a whole debugging round. **Why it exists**: self-hosted NAS devices (fnOS port 5667) ship self-signed certs by default and plugins reach them by **bare IP**; even with a properly CA-signed cert the subject is a hostname, so connecting by IP always mismatches — structurally unsolvable for these targets
- Health checks + file-fingerprint hot updates both happen automatically
- **UDP Socket API** (`songloft.net`, requires `net` permission): the Go side hosts the UDP socket + a message-push model. `udpBind` creates a socket and starts a reader goroutine; received UDP packets are pushed asynchronously to the JS callback (`onData`) via the scheduler queue. Supports multicast groups (`udpJoinMulticast/udpLeaveMulticast`), typical use: SSDP device discovery (DLNA/UPnP). Each plugin gets at most 8 sockets; a plugin with active sockets won't be idle-evicted, and sockets are cleaned up automatically on plugin unload. Implemented in `internal/jsplugin/api_bridge_net.go`
- **TCP Socket API** (`songloft.net.tcpConnect`, requires `net` permission): outbound TCP connections. `tcpConnect(host, port, options?)` returns a socket handle with `send()/onData()/onClose()/close()`. Data reception reuses UDP's Go readLoop + host event queue push model (`postHostEvent("tcp_data")` → JS `__dispatchHostEvent`). **`data` is base64-encoded raw bytes** (`btoa` on send, `atob` on onData, same as UDP): TCP is a byte stream and a single read may split in the middle of a multi-byte UTF-8 character; a raw string would be replaced with U+FFFD by `json.Marshal` and permanently corrupted, so base64 is mandatory. Plugins must accumulate bytes across chunks before UTF-8 decoding. **Only private / loopback / link-local addresses are allowed** (`isPrivateHostAllowed`, anti-SSRF); the 8-sockets-per-plugin quota is counted independently from UDP; a plugin with active TCP connections won't be idle-evicted; sockets are cleaned up on unload. Typical use: controlling a local MPD (idle event push on port 6600). Implemented in `internal/jsplugin/api_bridge_tcp.go`
- **Private registry authentication**: `RegistryConfig` supports a `token` field; when fetching any resource under that registry it automatically carries an `Authorization: Bearer <token>` header, compatible with GitHub private-repo PATs and self-hosted private registries. See [Plugin registry authoring guide · Private registry authentication](docs/en/plugin_registry.md#private-registry-authentication)
- **Lyrics/Cover providers** (`songloft.lyrics` / `songloft.covers`, no permission required): plugins call `registerProvider()` to register as a lyrics or cover provider. When a song has no lyrics/cover, the host iterates registered providers and calls `/lyric-search` or `/cover-search` endpoints (15s timeout, first-match-wins). Search params include `title/artist/album`; lyrics also carry `duration`; both optionally carry `fingerprint` (Chromaprint) and `isrc`. Found lyrics are cached to DB (`scraped`); local songs also get lyrics embedded into file tags. Found covers are downloaded to `cover_path` for local songs (and embedded into tags); remote songs store `cover_url`. Provider registration survives idle eviction; only cleared on plugin disable. Implementation in `manager.go` (`SearchLyrics/SearchCover`), `api_bridge.go` (JS API), `handlers/music.go` (fallback calls). See [Plugin development guide · Lyrics/Cover providers](docs/en/js-plugin-development-guide.md#songloft-lyrics-—-lyrics-provider)

---

<a id="business-constraints"></a>

## Business Constraints (Required by Task)

Before editing these modules, **read the linked sections**; read all applicable sections for cross-module changes. This table is only a reminder; full constraints and historical evidence live in [Business Constraints and Pitfalls](docs/en/pitfalls.md). Follow the `LESSONS.md` workflow for new lessons and maintain topic details in the relevant documents, not the entry point.

| Change area | Key reminder | Required reading |
|---|---|---|
| Plugin store identity | Return on conflicts before writing; match install state by entry_path + identity. | [Details](docs/en/pitfalls.md#plugin-identity) |
| Plugin store cache | Cache registry data, not install state; hash tokens in keys and isolate proxyDown per call. | [Details](docs/en/pitfalls.md#registry-cache) |
| Scan titles and test directories | Prefer tags by default, honoring filename mode; keep test music outside /tmp. | [Details](docs/en/pitfalls.md#scan-title) |
| Streaming scan imports | Finalize by directory; preserve extracted work on cancel with a separate flushCtx. | [Details](docs/en/pitfalls.md#scan-batches) |
| Directory traversal | Avoid per-entry Stat; preserve directory-name caching, single-flight, and cancellation detachment. | [Details](docs/en/pitfalls.md#directory-walk) |
| DSF / DSD | Keep scan defaults and migrations aligned; players without DSD support need explicit transcoding. | [Details](docs/en/pitfalls.md#dsf) |
| Sidecar lyrics | Probe fixed candidate paths; never overwrite manual lyrics or embed sidecar lyrics into audio tags. | [Details](docs/en/pitfalls.md#sidecar-lyrics) |
| Tag writing | Write atomically through sibling temp files; unsupported formats must not block the main flow. | [Details](docs/en/pitfalls.md#tag-writing) |
| Audio fingerprints | Off by default; cap samples at 120 seconds, persist failures, support cancellation and CPU-scaled concurrency. | [Details](docs/en/pitfalls.md#fingerprints) |
| Play history | Record only play events per context; upsert and trim to 50 in one transaction, never trust ordinals as stable positions. | [Details](docs/en/pitfalls.md#play-history) |
| HLS proxy | Off by default; preserve same-origin and SSRF checks, using relative rewritten URLs. | [Details](docs/en/pitfalls.md#hls-proxy) |
| Outbound HTTP proxy | Use proxy-aware clients; ffmpeg supports only HTTP/HTTPS proxies. | [Details](docs/en/pitfalls.md#http-proxy) |
| Private proxy allowlist | An empty allowlist keeps private networks blocked; it affects resource/transcode proxies, not HLS. | [Details](docs/en/pitfalls.md#private-allowlist) |
| Remote URL transcoding | No disk or DB persistence; share 4 slots with seek/normalization streams and return 503 when full. | [Details](docs/en/pitfalls.md#url-transcode) |
| Music cache | Stream without waiting for full downloads; failed cache transcodes keep originals and do not change explicit downloads. | [Details](docs/en/pitfalls.md#music-cache) |
| Normalization and live transcoding | Share configured loudness and cache keys; preserve Range, exact byte counts, deferred 206 and Flush. | [Details](docs/en/pitfalls.md#normalization) |
| Server-side seek | Combine seek with Range; preserve CBR, the 3-second end guard and the unknown-duration timeout. | [Details](docs/en/pitfalls.md#seek-stream) |
| Song persistence | Only for music the user legally owns; transcode or lyric-fetch failures must not block persistence. | [Details](docs/en/pitfalls.md#song-download) |
| Cross-device file moves | Use internal/fileutil.MoveFile, not bare os.Rename, to move temp files across mounts. | [Details](docs/en/pitfalls.md#file-move) |
