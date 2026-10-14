# API Reference

This document is based on the following source files:

- `internal/app/routers.go` -- Complete route registration (API v1 route group, middleware mounting, JS plugin routes)
- `internal/handlers/response.go` -- respondJSON / respondError response helper functions
- `internal/middleware/auth.go` -- JWT authentication middleware (Bearer Token + access_token query parameter fallback)
- `docs/api_response.md` -- API response format specification

## Table of Contents

1. [Basic Information](#1-basic-information)
   - [API Version and Base URL](#11-api-version-and-base-url)
   - [Authentication](#12-authentication)
   - [Response Format](#13-response-format)
   - [Content-Type Conventions](#14-content-type-conventions)
   - [Pagination](#15-pagination)
2. [Public and Authenticated Endpoints](#2-public-and-authenticated-endpoints)
3. [API Group Index](#3-api-group-index)

---

## 1. Basic Information

### 1.1 API Version and Base URL

**Section source**: `internal/app/routers.go`

All business endpoints are registered uniformly under the `/api/v1` route group. Full URL format:

```
{scheme}://{host}:{port}{base_path}/api/v1/{endpoint}
```

| Component | Default | Description |
|----------|--------|------|
| `scheme` | `http` | Can be `https` behind a reverse proxy |
| `host:port` | `localhost:58091` | Configured at startup via the `-port` parameter |
| `base_path` | empty | Configured for sub-path deployment via `-base-path /xxx` or the environment variable `BASE_PATH=/xxx`; the backend strips the prefix with `http.StripPrefix` |

Example: with the default startup, the song list endpoint is `http://localhost:58091/api/v1/songs`; if `BASE_PATH=/music` is configured, it becomes `http://localhost:58091/music/api/v1/songs`.

The current API version is **v1**, and all endpoints are prefixed with `/api/v1/`. In development mode, the interactive API documentation is available at `/swagger/index.html`.

### 1.2 Authentication

**Section source**: `internal/middleware/auth.go`

Songloft uses a **JWT dual-token mechanism** (Access Token + Refresh Token). The authentication middleware obtains the token in the following order of priority:

1. **Authorization request header** (preferred): `Authorization: Bearer <access_token>`
2. **URL query parameter** (fallback): `?access_token=<token>`, used for scenarios that cannot set a custom header such as `<img>` tags and audio players

On authentication failure, a JSON-formatted error is returned (rather than plain text), keeping it consistent with business endpoints:

```json
{"error": "缺少认证信息"}
{"error": "无效的 token", "detail": "token has expired"}
```

> **Special handling**: Xiaomi (Xiao AI) speaker firmware replaces `&` in URLs with spaces; the middleware automatically splits and restores the swallowed parameters.

### 1.3 Response Format

**Section source**: `internal/handlers/response.go`, `docs/api_response.md`

The project adopts a **RESTful direct-return style** and does not use a unified `{code, data, message}` envelope.

#### Success Response

| Scenario | Format | Example |
|------|------|------|
| Single entity | Returns the model object directly | `{"id":1, "title":"Sample Track", ...}` |
| List (with pagination) | Collection name + pagination metadata | `{"songs":[...], "total":100, "limit":20, "offset":0}` |
| Operation result | `{"message": "..."}` | `{"message": "歌曲已删除"}` |

#### Error Response

Errors for all API endpoints are returned uniformly through `respondError` in a fixed format:

```json
{"error": "人类可读的错误信息", "detail": "可选的技术细节"}
```

- `error` (required): a short, user-facing description
- `detail` (optional): the underlying error message, output only when the internal `err != nil`

Middleware-level errors (such as authentication failures) are also returned in JSON format, using the same `{error, detail}` field structure.

#### Exception: Binary Stream Endpoints

Binary stream endpoints such as playback (`/songs/{id}/play`), resource proxy (`/proxy`), and cover images (`/songs/{id}/cover`) may return plain text errors (`http.Error`), because the client does not expect a JSON body.

### 1.4 Content-Type Conventions

**Section source**: `internal/handlers/response.go`, `internal/app/routers.go`

The default is `application/json`. Exceptions: audio playback (`audio/*`, supports Range), cover images (`image/*`), playlist export and plugin files (`application/octet-stream`), plugin pages (`text/html`), and HLS playlists (`application/vnd.apple.mpegurl`). Upload endpoints (plugin installation, playlist covers) use `multipart/form-data`.

### 1.5 Pagination

**Section source**: `internal/handlers/music.go`

List endpoints that support pagination use the `limit` (default 20, automatically truncated when exceeding the upper limit) and `offset` (default 0) query parameters. The response body contains four fields: the collection name, `total`, `limit`, and `offset`.

---

## 2. Public and Authenticated Endpoints

**Section source**: `internal/app/routers.go`

Route registration is divided into two layers: public endpoints (no authentication required) and authorized endpoints (JWT required).

**Public endpoints** (no Bearer Token required):

| Endpoint | Description |
|------|------|
| `POST /api/v1/auth/login` | User login, obtains a token pair |
| `POST /api/v1/auth/refresh` | Exchanges a refresh token for a new access token |
| `GET /api/v1/version` | Gets service version information |
| `GET /api/v1/health` | Health check |
| `GET /api/v1/jsplugin/{entryPath}` | Plugin static page (HTML) |
| `GET /api/v1/jsplugin/{entryPath}/static/*` | Plugin static resources (CSS/JS/images) |
| `GET /api/v1/jsplugin-assets/*` | Plugin common resources (common.css/common.js/fonts) |
| Paths declared in a plugin's `publicPaths` | API paths declared in a plugin manifest that require no authentication |

**Authenticated endpoints** (Bearer Token required):

All `/api/v1/*` endpoints other than the public endpoints above require authentication. Authentication failure returns `401 Unauthorized`.

---

## 3. API group index

This index was checked against `internal/app/routers.go`, plugin routes, and `docs/swagger.json`. OpenAPI currently lists 140 paths and 191 operations. HEAD is omitted and dynamic plugin routes use placeholders. Tables describe the current source snapshot; Swagger and the corresponding handler define request types and fields.

| Module | Prefix | Operations | Documentation |
| --- | --- | --- | --- |
| Songs | `/api/v1/songs` | 38 | [Songs](歌曲%20API.md) |
| Playlists | `/api/v1/playlists` | 22 | [Playlists](歌单%20API.md) |
| Play history | `/api/v1/play-history` | 3 | [Play history](播放历史%20API.md) |
| Authentication | `/api/v1/auth` | 5 | [Authentication](认证%20API.md) |
| Generic configuration | `/api/v1/configs` | 5 | [Generic configuration](配置与设置%20API.md) |
| Business settings | `/api/v1/settings` | 48 | [Business settings](配置与设置%20API.md) |
| Scanning and fingerprints | `/api/v1/scan` | 10 | [Scanning and fingerprints](扫描管理%20API.md) |
| Cache | `/api/v1/cache-manage` | 5 | [Cache](缓存管理%20API.md) |
| Plugin management | `/api/v1/jsplugins` | 13 | [Plugin management](插件管理%20API.md) |
| Plugin runtime | `/api/v1/jsplugin` | 9 | [Plugin runtime](插件管理%20API.md) |
| Shared plugin assets | `/api/v1/jsplugin-assets` | 2 | [Shared plugin assets](插件管理%20API.md) |
| Plugin health | `/api/v1/plugins` | 1 | [Plugin health](插件管理%20API.md) |
| Custom tags | `/api/v1/song-tags` | 9 | [Custom tags](标签与主题%20API.md) |
| Theme packs | `/api/v1/theme-packs` | 9 | [Theme packs](标签与主题%20API.md) |
| Backend upgrades | `/api/v1/upgrade` | 7 | [Backend upgrades](系统管理%20API.md) |
| Resource and transcode proxy | `/api/v1/proxy` | 2 | [Resource and transcode proxy](系统管理%20API.md) |
| Log export | `/api/v1/logs` | 1 | [Log export](系统管理%20API.md) |
| Health check | `/api/v1/health` | 1 | [Health check](系统管理%20API.md) |
| Version | `/api/v1/version` | 1 | [Version](系统管理%20API.md) |

Interactive documentation in development builds: `/swagger/index.html`. For the generated static reference, see [Swagger API](../../../swagger-api/index.md).
