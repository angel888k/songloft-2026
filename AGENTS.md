# AGENTS.md

本文件是 Songloft 的 AI 开发入口：保留通用铁律、常用命令与任务导航，业务细节按需读取。目录、依赖、API 和表结构以代码为事实来源。

修改前先读适用的 `LESSONS.md`、子模块本地规范及文末[业务约束索引](#business-constraints)中的相关章节；无需为无关任务加载全部专题。英文版见 [AGENTS.en.md](AGENTS.en.md)。

> **详细文档**：
> - AI 开发：[流程与复盘](docs/ai-development.md) · [经验记录](LESSONS.md)
> - 业务约束：[按主题查阅](docs/pitfalls.md)（扫描、歌词、指纹、插件商店、代理、缓存与播放）
> - 架构：[整体](docs/architecture.md) · [后端](docs/architecture_backend.md) · [前端](docs/architecture_frontend.md)
> - 专题：[数据库操作](docs/database_migrations.md) · [颜色系统](docs/color_system.md) · [API 响应格式](docs/api_response.md) · [快速上手](docs/quick-start.md) · [前端踩坑与铁律](docs/frontend_gotchas.md)
> - 插件开发：见 `plugins/toolchain/README.md`（独立仓库）
> - 插件源制作：[插件源制作指南](docs/plugin_registry.md)
> - API：开发模式启动后访问 `/swagger/index.html`

---

## 项目概述

Songloft 是自托管本地音乐服务器，支持**服务器部署**和**Bundle 本地模式**（将 Go 后端嵌入客户端，无需单独部署服务器）。多仓库结构：

| 目录 | 技术 | 说明 |
|------|------|------|
| `/` | Go 1.26 + Chi v5 + SQLite | 后端 API 服务（默认端口 58091，账号 admin/admin） |
| `/mobile` | Go + gomobile | Go 后端的移动端绑定入口（gomobile bind 用，导出 Start/Stop/IsRunning/GetPort） |
| `/clients/player` ([独立仓库](https://github.com/songloft-org/songloft-player)) | Flutter 3.41+ / Dart 3.7+ | 跨平台前端（6 平台），支持 Bundle 本地模式 |
| `/clients/player-lynx` ([独立仓库](https://github.com/songloft-org/songloft-player-lynx)) | TS + ReactLynx | Android/iOS/HarmonyOS/Web 预览客户端，独立 main dev 构建与 v* 发版（子模块，无桌面/Bundle） |
| `/clients/tv` ([独立仓库](https://github.com/songloft-org/songloft-tv)) | Kotlin | Android TV 客户端（子模块） |
| `/plugins/toolchain` ([独立仓库](https://github.com/songloft-org/plugin-toolchain)) | TS + pnpm | JS 插件开发工具链（SDK / Builder / 脚手架） |
| `/plugins/src` | TS | JS 插件源码（子模块集合，每个插件在自己仓库下分发 release） |
| `/pkg/tag` | Go | 音频元数据**读写**库（基于上游 tag 库扩展 MP3/FLAC 写入） |
| `/integrations/home-assistant` ([独立仓库](https://github.com/songloft-org/home-assistant-addon)) | HA add-on | Home Assistant 加载项（薄层复用 Docker 镜像，子模块）。**清单 `repository.yaml` 必须在那个仓库的根目录**，这就是拆出独立仓库的原因（#340）。设计/踩坑/版本同步见 [integrations/home-assistant/README.md](integrations/home-assistant/README.md) |
| `/tools/ffmpeg-builder` ([独立仓库](https://github.com/hanxi/ffmpeg-builder)) | Docker | 静态编译 ffmpeg/ffprobe 最小镜像构建器（子模块），供下载转码 / 音频指纹用 |
| `/tools/tracely` ([独立仓库](https://github.com/hanxi/tracely)) | Go + Vue | 自托管前端监控后端（安装/升级追踪），后端经其 Go SDK 上报；子模块，SDK 依赖走 go.mod |

---

## 仓库规范归属

- 主仓库记录后端实现、公共宿主能力、接口契约及跨仓库通用约定。
- 子模块的业务实现、设备兼容、历史问题、回归矩阵与专属验证命令写在各自仓库根目录的 `AGENTS.md`，同步维护 `AGENTS.en.md`。
- 修改子模块前先读取其本地规范；涉及宿主接口时同时核对主仓库的通用规范。子模块提交完成后，再在主仓库更新指针。

## AI 开发流程

- 任务交办后连续完成调查、实现、验证、自审和交付，不逐阶段确认；关键缺失信息、产品决定或扩大范围才询问。提交、推送、正式发布和 Issue 写操作分别按用户授权，已有授权不重复询问。
- 关联 Issue 的修复推送后，主动准备并展示回复草稿和目标 Issue，等待评论确认；已有适用的评论授权则直接回复。草稿随 CI、发布状态更新，不等用户追问；详见 [关联 Issue 回复](docs/ai-development.md#关联-issue-回复)。
- 开始前读适用的 `LESSONS.md`，结束记录有证据的新经验并同步英文。三个不同日期的独立事件提出规范升级，同一事故重复复盘不递增；已有铁律立即执行。
- 排查未解决 Issue 包含主仓库和全部子模块，递归发现、去重、完整分页。统一 `owner/repo#编号`，缺口明确报告；结合正文、全部评论、最新代码与验证判断，不把 open 当未修复证明。
- 验证绑定仓库、SHA 或未提交工作树标识、环境、命令和结果；相关未跑项写 `Not run`。代码/子模块变化重跑受影响检查；追踪同一 SHA 的必要 CI，取消/跳过不算通过。详见 [AI 开发流程](docs/ai-development.md)。

## 常用命令

```bash
# 后端
make run            # 启动（dev 模式，含 Swagger）
make build          # 编译开发版（完整版，嵌入前端）
make build-lite     # 编译开发版（精简版，不嵌入前端）
make build-prod     # 编译生产版（完整版，嵌入前端）
make build-prod-lite # 编译生产版（精简版，不含前端）
make test           # 测试
make check          # fmt + vet + test
make sqlc           # 重新生成 sqlc 代码（改了 queries/*.sql 后必跑）
make swagger        # 重新生成 API 文档

# 前端构建（产物落到 clients/player-build/，供后端嵌入或独立部署）
make build-frontend-web-embedded   # 嵌入 Go 二进制用（隐藏 API 地址 UI）
make build-frontend-web            # 独立部署 web
make build-frontend-{linux,windows,macos,android,ios,all}

# Bundle 本地模式（Go 后端编译为移动端库 / 桌面端可执行文件）
make build-go-mobile-android       # Android .aar（gomobile bind，arm64 + arm + x86_64）
make build-go-mobile-ios           # iOS .xcframework（gomobile bind，arm64，仅 macOS）
make build-go-desktop-linux        # Linux 可执行文件
make build-go-desktop-windows      # Windows .exe
make build-go-desktop-macos        # macOS x86_64
make build-go-desktop-macos-arm64  # macOS ARM64

# 前端开发
cd clients/player && flutter run -d chrome          # standalone
cd clients/player && flutter run -d chrome --dart-define=DEPLOY_MODE=embedded
```

---

## 代码格式化（铁律）

每次修改代码后**必须**格式化，提交前确认无格式差异：

- **Go**：在项目根目录执行 `gofmt -w .`
- **Dart**：在 `clients/player/` 目录执行 `dart format lib/ test/`

---

## 数据库规范（铁律）

> 完整操作步骤见 [docs/database_migrations.md](docs/database_migrations.md)。

访问栈：**goose 迁移 + sqlc 固定 SQL + squirrel 动态 SQL + Repository + UnitOfWork**。

- **改 schema** → `internal/database/migrations/000N_xxx.sql`，启动时 `goose.Up` 自动执行；**禁止**手动 `ALTER data/songloft.db`
- **加固定 SQL** → `internal/database/queries/{table}.sql` + `make sqlc`；生成产物 `internal/database/sqlc/` 必须入库
- **动态 SQL（变长 WHERE/SET）** → 在 `*_repository.go` 内用 squirrel，禁止拼字符串
- **跨表写** → `db.RunInTx(ctx, func(ctx, uow))` 拿同一 `*sql.Tx` 下的 `uow.Songs/Playlists/...`；**禁止** service 层手 `BeginTx`，否则会 SQLITE_BUSY
- **错误语义** → 仓储未命中统一 `database.ErrNotFound`；service 用 `errors.Is` 判别
- **测试** → `testutil.OpenMemoryDB(t)` 跑真实 `:memory:` + 真实 Repository；**禁止**手写 mockDB
- **内置数据** → 迁移预置歌单 id=1「收藏」、id=2「电台收藏」（`labels=["built_in"]`），及 `music_path / jwt_secret / source_*` 默认 config。测试行数断言记得扣掉
- **`queries/*.sql` 里 `-- name:` 之后不要写中文注释** → sqlc 会把 query 注释当 doc comment 按**字节**嵌进生成代码，中文被截断成非法 UTF-8，`make sqlc` 直接失败（`error generating code: illegal UTF-8 encoding`）。注释写在 Go 仓储方法上
- **同表自引用子查询要给外层列限定表名** → `DELETE FROM t WHERE col = ? AND id NOT IN (SELECT x.id FROM t x ...)` 里外层裸 `col` 会被 sqlc 判为 `column reference is ambiguous`，写成 `t.col` 才过

---

## 后端编码约定

- 标准 Go layout（`internal/` 防外部依赖），Chi v5 路由，JWT 双 Token
- 依赖注入：普通读写通过 Repository 接口；跨表事务复用既有 `RunInTx + UnitOfWork` 注入（如 `PlayHistoryService` 接收 `database.DB`），**禁止** service 自行 `BeginTx` 或绕过仓储写 SQL。
- 日志：标准库 `slog`；HTTP 错误：统一 `respondError`
- **API 响应格式**：RESTful 直返，**禁止** `{code, data, message}` 信封；错误统一 `{"error","detail"}`。完整规范见 [docs/api_response.md](docs/api_response.md)
- 不用 ORM：固定 SQL → sqlc，动态 SQL → squirrel，跨表写 → `RunInTx + UnitOfWork`
- 测试文件 `*_test.go` 与源码同目录

---

## API 文档规范（铁律）

**所有在 `internal/app/routers.go`（含 `RegisterStaticRoutes` / `RegisterAPIRoutes` 等子注册函数）里注册的 handler 方法，必须有 swag 注释**。后端 API 文档由 [swaggo/swag](https://github.com/swaggo/swag) 从注释生成，是前端开发与外部集成的唯一来源。

### 必填字段（每个 handler 至少有这 7 项）

```go
// @Summary <一行中文摘要>
// @Description <详细描述，可多行；说清楚副作用 / 默认值 / 错误码触发条件>
// @Tags <业务分组，中文>
// @Produce json
// @Success 200 {object} <返回类型> "<说明>"
// @Security BearerAuth
// @Router /<path> [<method>]
func (h *XxxHandler) Method(w http.ResponseWriter, r *http.Request) { ... }
```

- 有请求体的接口额外加 `@Accept json` 和 `@Param request body <type> true "<说明>"`
- 错误路径明显的接口加 `@Failure 400/404/500 {object} map[string]string "..."`
- 路径参数 / 查询参数用 `@Param <name> path/query <type> true/false "<说明>"`
- **公开端点**（无需 token，如健康检查）省略 `@Security BearerAuth`
- **业务 tag 命名**：复用现有 tag（「歌曲管理」「歌单管理」「电台与 HLS」「扫描管理」「配置管理」「缓存管理」「JS 插件」「JS插件管理」「数据备份」「设置」「系统升级」「认证管理」「系统管理」「资源代理」「主题包」「播放历史」「标签管理」），不要随手造新 tag
- **`@Router` 路径禁止包含 `/api/v1` 前缀**：`main.go` 已声明 `// @BasePath /api/v1`，swag 会自动给所有 `@Router` 路径加上该前缀。如果写了 `/api/v1`，生成的文档路径会变成 `/api/v1/api/v1/...`。正确写法是相对路径（如 `/songs/{id}/tags`），而非 `/api/v1/songs/{id}/tags`。

### 多别名 / catch-all 路由

- 一个 handler 注册了多条 alias 路径（如 `/songs/{id}/play` 与 `/songs/{id}/play.m3u8`）→ 每条 alias 单写一行 `@Router`
- HEAD 是 GET 的子集，**不单独列**；OpenAPI 不强制
- `r.HandleFunc(...)` 这种接受 ANY HTTP 方法的 catch-all → 列出所有实际可能的方法（`[get] [post] [put] [delete]`），每个一行 `@Router`
- 动态路径（`{entryPath}` 由运行时按已安装插件决定的）→ 在 `@Description` 里注明「动态路由，{xxx} 由运行时决定，OpenAPI 仅作占位」

### 改完必跑

修改 / 新增 handler 注释后必须跑 `make swagger`：会重新生成 `docs/swagger.json`、`docs/swagger.yaml`、`docs/docs.go`，**这些产物必须入库**。否则 `/swagger/index.html` 与代码不同步，前端按旧文档对接会踩坑。

### 验证

- `make swagger` 输出里搜索新加的 `@Router` 路径，确认 `Generating <Type>` 包含你新写的请求/响应类型
- `grep '<your-new-path>' docs/swagger.json` 应有命中
- 启动 `make run`，访问 `http://localhost:58091/swagger/index.html` 在 UI 里点开新端点目测

### 没有豁免

「凡 routers 注册即必注释」是绝对规则。哪怕是动态路由 catch-all、静态资源 handler、反代端点，也要写 swag——`@Description` 里把"它是什么、为什么 OpenAPI schema 不精确"说清楚即可。

---

## 配置接口规范（铁律）

项目里有两类配置接口，**用户可见的功能开关一律走业务端点**，通用 KV 仅作 admin 入口。

### `/api/v1/settings/<name>` — 孤立配置端点（前端业务功能默认走这里）

- 路径风格：`/settings/<kebab-case-name>`（如 `/settings/hls-proxy`、`/settings/music-path`、`/settings/http-proxy`、`/settings/library-browse`、`/settings/proxy-private-allowlist`）
- 数据形态：**强类型** JSON（如 `{enabled: bool}` 或聚合对象），不是 `{value: string}`
- 默认值：handler 内部承担（配置缺失时 GET 返回业务默认，PUT 时直接写入即可，**前端无需先 POST 创建**）
- 副作用：在 PUT 内部直接触发（如 `music_path` PUT 完异步 `onMusicPathChanged` 重建 Scanner）
- 归属：放进对应业务模块的 handler（如 hls-proxy 在 `HLSHandler`，music-path 在 `ScanHandler`），handler 同时持有 `*services.ConfigService` 完成读写
- 命名套路：`Is<Name>Enabled() / Set<Name>Enabled(bool)` 业务方法 + `Get<Name>Setting / Update<Name>Setting` HTTP handler + `/settings/<name>` 路由

### `/api/v1/<module>/*` — 业务模块聚合端点（含配置）

某些业务模块自带"动作端点+配置端点"组合（典型例子 `/cache-manage/{stats,clean,config}`），此时配置端点**保留在模块前缀下**，不强行拆到 `/settings/`。

- 适用场景：配置与该模块的其他动作端点强相关（如 cache 的 `config` 跟 `stats/clean` 共用同一个 `CacheService`）
- 选择依据：业界主流（AWS、GitHub、Discord）都是业务模块聚合；GitLab 那种"全局集中、模块分散"的混合模式同样接受
- 已有的例子：`/api/v1/cache-manage/config`（GET/PUT）
- **判定准则**：
  - **孤立**配置（不属于任何业务模块、或跨模块共享）→ `/settings/<name>`
  - **模块内**配置（与该模块动作端点强相关）→ `/<module>/config` 或 `/<module>/<sub-name>`

### `/api/v1/configs/{key}` — 通用 KV（admin 编辑器专用）

- 仅供前端 `config_manager.dart` 这种**通用配置编辑器**使用，让管理员手编任意 key/value 调试
- **新加业务功能不要直调** `/configs/{key}`：通用 PUT 在 key 不存在时返回 404，且没有强类型、没有副作用、没有默认值
- 业务化封装后，通用接口仍可改同一 key（保留双入口），但副作用必须同时挂在 `configHandler.SetOnConfigChanged` 回调里（参考 `routers.go` 里 `musicPathChanged`），保证两条入口语义一致

### 客户端约定

- `SettingsApi`（`clients/player/lib/features/settings/data/settings_api.dart`）封装所有 `/settings/*` 调用，业务功能 Provider 一律走它
- `ConfigApi` 只在 `config_manager.dart` 与「列出所有配置」这类 admin UI 里使用

### 历史决策记录

- 该规范在 2026-06 引入，背景：`hls_proxy_enabled` 默认未预置导致 PUT `/configs/{key}` 返回 404，发现项目里 `/configs` + `/settings/*` + `/cache-manage/config` 三种风格并存
- 选定方向：业务端点是用户可见入口的**唯一来源**，通用 KV 退化为 admin 后门

---

## 文档双语同步规范（铁律）

项目文档为**中英双语并存**，改任一语言版本时**必须同步改另一版本**，禁止只改一边导致中英内容漂移。

- **映射关系**：
  - `README.md` ↔ `README.en.md`
  - `AGENTS.md` ↔ `AGENTS.en.md`
  - `docs/<name>.md` ↔ `docs/en/<name>.md`（同名文件，英文版在 `docs/en/` 下）
- **判定准则**：凡是新增、修改、删除文档中的**内容/结构/链接**（正文、章节、表格、导航链接等），中英两版都要落地对应改动；仅英文表述本地化，结构保持一致
- **改前先确认对应文件是否存在**：`docs/en/` 下有同名文件则必须同步；README 一律有 `.en.md` 对应
- **例外**：某些内容天然只属于单一语言版本（如仅中文版的社区说明），无对应版本时不强制镜像，但应确保是有意为之，而非遗漏

---

## 文档站结构（docs/ — VitePress 自定义主题）

Songloft 文档站（`docs/`）用 **VitePress + 自定义主题**（`docs/.vitepress/theme/`），**不是默认主题**。改文档站前先判断页面来源，不能直接编辑生成产物：

- **自定义落地页（改数据，不改 markdown）**：首页 `docs/index.md` 仅一行 `<Landing />`，内容由结构化数据 `docs/.vitepress/data/*.ts`（安装方式 `downloads.ts`、功能 `features.ts`、文案 `landing-i18n.ts`）驱动，由 `docs/.vitepress/theme/components/landing/*.vue` 渲染。改落地页 → 改 `data/*.ts`（双语 `{zh,en}` 字段）；图标要对齐组件里的映射表（如 `LandingInstaller.vue` 的 `ICONS`）。
- **自动生成页（禁止手改）**：`docs/quick-start.md`、`docs/en/quick-start.md`、`docs/changelog.md` 由 `scripts/sync-docs.mjs` 从根 `README.md` / `README.en.md` / `CHANGELOG.md` 生成，已被 `docs/.gitignore` 忽略。要改正文 → 改源 `README` / `CHANGELOG`，`docs:dev` / `docs:build` 会先跑 `sync` 重新生成。**手改会被覆盖且不入库**。
- **子模块同步页（同样禁止手改，但源在别的仓库）**：`docs/addon/`、`docs/player/`、`docs/player-lynx/`（及英文 `docs/en/player-lynx/`）、`docs/plugin-toolchain/` 由 `sync-docs.mjs` 分别从 `integrations/home-assistant/`、`clients/player/docs/cn/`、`clients/player-lynx/`、`plugins/toolchain/` **子模块**同步，也都被 `docs/.gitignore` 忽略。要改正文 → **去对应子模块仓库改，再回主仓库 bump 子模块指针**（先检查子仓库工作树与目标提交，再按授权提交并更新父仓库 gitlink），否则文档站永远显示旧内容。两个要点：① `to:` 目标路径与源路径**刻意解耦**（如 `integrations/home-assistant/README.md` → `docs/addon/index.md`），因为 `/addon/` 这类对外 URL 已进 sitemap，不能跟着源仓库改名；② 子模块未 checkout 时 `sync` 只 warn 不 fail，页面会**静默消失**，所以 `static.yml` 的 submodule init 列表必须包含它们。
- **repowiki（`docs/repowiki/` — AI 手动生成，手动维护）**：这套 wiki（系统架构/数据模型/后端/前端/API/插件/部署/故障排除/开发指南等）是 **AI 基于源文件手动生成的**，仓内**没有生成脚本**、也不是外部服务自动产物，**入库的 markdown 即唯一真实来源**。因此任何工具（AI/人）直接编辑并 commit 即可，**不要等"重新生成"**。改代码相关内容时按需同步这些页面，与其他源文档一样对照代码保持准确（例如 `internal/jsplugin/assets/` 的公共资源改名后，repowiki 里引用旧文件名的地方也要一并改）。

---

## Git 提交约定

- **直接提交到 `main` 分支**，不新建功能分支、不走 PR 流程（本仓库约定）
- 提交信息**禁止**添加 `Co-Authored-By` 尾部标记
- 遵循 Conventional Commits 格式：`type(scope): description`，description 和 body 尽量用中文
- 关联 GitHub issue 的提交信息必须带 issue 引用
- issue 引用规则：短写 `#123` 永远指向**当前 commit 所在仓库**的 issue；只要引用的不是当前仓库的 issue，就必须写完整 `owner/repo#123`
  - 父仓库 `songloft-org/songloft` 的 commit 引用父仓库 issue：可写 `#155`，也可写 `songloft-org/songloft#155`
  - 子仓库（如 `pkg/tag`、`clients/player`、`plugins/toolchain`、`plugins/src/*`）的 commit 引用自身仓库 issue：可写 `#14`，也可写完整仓库路径
  - 子仓库的 commit 引用父仓库 issue：必须写完整路径，如 `songloft-org/songloft#155`，不能只写 `#155`（否则 GitHub 会解析为子仓库自身的 issue）
  - 任意跨仓库引用一律写完整路径，如 `songloft-org/songloft-player#14`

---

## 构建与部署

- 构建标签：`dev`（含 Swagger + pprof） / `lite`（精简版，不嵌前端） / 无标签（完整版，嵌 Flutter Web）
- `VERSION=dev` 时 Makefile 自动启用 `-tags dev`（无需手动传 `EXTRA_TAGS=dev`）
- 两个正交维度：**VERSION**（`dev` / `X.Y.Z`）控制是否为开发版；**BUILD_TYPE**（`lite` / 空即 `full`）控制是否嵌入前端。**禁止** `BUILD_TYPE=dev` 等混合值
- 嵌入路径是 `clients/player-build/web-embedded`（**不是** `clients/player/build/web-embedded`）
- SPA 回退：`internal/app/embed.go` 仅对不存在且末段不含 `.` 的路径返回 `index.html`；缺失的静态资源返回 404，不能返回 HTML 冒充 JS/JSON。
- 部署模式由 `--dart-define=DEPLOY_MODE=embedded|standalone` 切换，`AppConfig.isEmbedded` 是编译时常量，tree-shaking 会移除嵌入模式下的 API 地址 UI
- 子路径部署：启动时通过 `-base-path /xxx` 或 `BASE_PATH=/xxx` 配置；后端用 `http.StripPrefix` 在最外层剥离前缀，`embed.go` 运行时将 `<base href="/">` 替换为 `<base href="/xxx/">`；前端嵌入模式从 `Uri.base.path` 自动检测子路径

### Bundle 本地模式（v2.9.0+）

将 Go 后端嵌入 Flutter 客户端，用户无需单独部署服务器即可使用。编译时 `--dart-define=HAS_BACKEND=true` 启用。

- **移动端（Android/iOS）**：通过 `gomobile bind` 将 Go 后端编译为原生库（`.aar` / `.xcframework`），Flutter 通过 `MethodChannel('com.songloft/backend')` 调用
- **桌面端（macOS/Windows/Linux）**：Go 后端编译为独立可执行文件 `songloft-server`，Flutter 启动时作为子进程运行
- **Web**：不支持 Bundle 模式（仅远程服务器）
- 运行模式：`RunMode.local`（本地）/ `RunMode.remote`（远程），持久化到 SharedPreferences，启动时自动恢复
- 本地模式启动流程：申请存储权限 → 启动嵌入后端（`127.0.0.1:<port>`）→ 健康检查轮询（最多 10 次 × 300ms）→ 自动使用 `admin/admin` 登录
- `BackendLifecycle`（WidgetsBindingObserver）：App 前台恢复时自动重启后端，detached 时停止
- 关键入口：`mobile/mobile.go`（gomobile 绑定）、`clients/player/lib/core/backend/`（Flutter 侧抽象层）
- CI 产物统一以 `songloft-bundled-` 开头；Android/Linux/Windows 带架构，Apple 分别为 `songloft-bundled-macos.zip`、`songloft-bundled-ios-nosign.ipa`。4 个并行 Job（Android/Linux/Apple/Windows），失败不阻塞主 Release。

### Docker 热替换规则（`scripts/docker-entrypoint.sh`）

Docker 镜像内含底包 `/app/songloft`，持久化 data 卷存放实际运行的 `/app/data/songloft`。容器启动时 entrypoint 决定是否用底包覆盖 data 目录：

**核心原则：底包代表用户意图；dev/正式或 full/lite 不一致时用底包覆盖。只有「同通道 + 同 BUILD_TYPE」时才比较新旧：dev 按 Build Time，release 按版本号。**

前置分支：目标文件不存在时直接初始化；已有目标时，底包通道未知则保留目标，目标通道未知则用已识别底包替换。同为 stable 但任一版本未通过脚本的 `is_semver` 检查（如自定义 tag）时按底包覆盖，不强行比较新旧。

| 场景 | 行为 | 原因 |
|------|------|------|
| dev ↔ release 通道不同 | 替换 | 用户换了镜像通道 |
| BUILD_TYPE 不同（full↔lite） | 替换 | 用户换了镜像变体 |
| 同为 dev + 同类型 + 底包 Build Time > data Build Time | 替换 | dev 滚动构建按构建时间选最新 |
| 同为 dev + 同类型 + data Build Time >= 底包 Build Time | 不替换 | data 可能通过 API 在线升级过 |
| 同为 release + 同类型 + 底包版本 > data 版本 | 替换 | 正式版升级 |
| 同为 release + 同类型 + data 版本 >= 底包 | 不替换 | data 可能通过 API 在线升级过 |

### Docker 非 root 运行（PUID/PGID，songloft-org/songloft#380）

- **默认不设置 = 保持 root 运行**，与旧版本行为完全一致，零迁移风险。只有显式设置 `PUID` 或 `PGID`（任一即可，另一个默认补 `1000`）才启用降权，entrypoint 末尾用 Alpine 自带的 `su-exec`（比 `gosu` 轻，官方仓库自带无需额外下载）切到该 uid:gid 再 `exec` 主程序
- **`/app/data` 每次启动都递归 `chown`，`/app/music` 默认只 chown 顶层目录，不递归**：`/app/data` 体量小（db、封面、缓存等）且必须修复旧 root 运行遗留的属主，否则新用户打不开旧数据库；`/app/music` 可能是几十万文件/数 TB 的个人曲库，每次启动递归扫描的 IO 代价不可接受。顶层目录可写之后，新下载/新写入的文件本身就会以目标 uid:gid 创建，天然正确，不需要事后再修一遍
- **升级前遗留在 `/app/music` 内的历史 root 属主文件不会被自动修复**（如插入/覆盖标签写入过的旧文件），这是刻意的性能取舍而非遗漏。需要时设置 `FIX_MUSIC_PERMISSIONS=true` 显式触发一次递归修复，仅推荐在切换为非 root 运行后手动跑一次，不要做成默认行为
- **`integrations/home-assistant` 不受影响**：其 `run.sh` 直接覆盖了镜像的 `ENTRYPOINT`，完全绕开 `docker-entrypoint.sh`，权限模型由 HA supervisor 另行管理

---

## 前端 UI 验证（Docker 无头浏览器）

需要**真的把改动在界面上跑一遍**（而不是只跑 `flutter test`）时，走 Docker 里的无头 Chrome：
本仓库开发机上宿主的 `google-chrome --headless` 会 core dump，CDP 端口也起不来，别在那上面浪费时间。

```bash
# 1. 前端编成嵌入模式，由 Go 后端同源提供（省掉 standalone 的 API 地址配置步骤）
make build-frontend-web-embedded
go build -tags dev -o /tmp/songloft-full .
/tmp/songloft-full -port 58191 -db <tmpdir>/test.db -music <musicdir>

# 2. 起浏览器容器。--network host 才能让容器里的 Chrome 访问宿主 127.0.0.1:58191
docker run -d --name uichrome --network host browserless/chrome:latest

# 3. 用 /function 端点跑 puppeteer 脚本（Content-Type 必须是 application/javascript）
curl -s -X POST http://127.0.0.1:3000/function \
  -H 'Content-Type: application/javascript' --data-binary @script.js
```

脚本形如 `module.exports = async ({ page }) => { ...; return { data: {...}, type: 'application/json' } }`，
截图用 `page.screenshot({ encoding: 'base64' })` 塞进 `data` 里带回来再本地 base64 解码成 png。

### 驱动 Flutter Web 的踩坑

- **Flutter Web 是 canvas 渲染，DOM 里没有按钮**。可交互元素只以 `<flt-semantics>` 无障碍节点暴露：
  用 `getBoundingClientRect()` 拿到中心点后 `page.mouse.click(x, y)`。语义树默认已开启；
  若首屏出现 `<flt-semantics-placeholder>` 需先 `.click()` 它
- **语义 label 会被合并**：一屏内容常聚合成一个长 aria-label 节点，按钮文字未必是独立节点。
  找不到时**先截图看坐标再点**，不要在 label 匹配上死磕
- **文本框**：`<input aria-label="Username">` 是真实 DOM，但用坐标点击 + `keyboard.type` 更稳；
  点击与输入之间要留 ~800ms，否则 Flutter 焦点还没建立、输入会丢
- **验证双语**：用 `evaluateOnNewDocument` 覆写 `navigator.language`/`languages` 为 `zh-CN`，
  Flutter l10n 会跟着切
- **深链**：`page.goto('/settings/category/2')` 是整页刷新，会在 boot 到一半时产生假的
  "Failed to load"。用应用内点击导航，别用 goto 跳嵌套路由

### 断言要落在后端可观测状态上

截图只能证明「渲染对了」。交互是否真的生效必须另找证据，例如点完开关 `curl` 对应
`/settings/<name>` 端点看值有没有变、点完「停止计算」用 `pgrep -x ffmpeg` 看子进程有没有归零。

- **数进程别用 `ps -ef | grep <关键字> | wc -l`**：当前 shell 自己的命令行里就含那个关键字，
  会稳定多算 1~2 个，很容易得出「没停掉」的错误结论。用 `pgrep -x <可执行名>`

---

## 平台适配踩坑

- 升级检查 (`/api/v1/upgrade/check`) 非 Docker 也可调用，并返回 `is_docker`；执行升级、回退仍仅 Docker 可用。
- Flutter `secure_storage` 在 macOS 未签名沙盒下自动降级到 SharedPreferences
- Android 构建前需 `sdkmanager --licenses`；Android 13+ 需运行时申请通知权限
- 所有原生平台（Win/Linux/macOS/Android/iOS）音频后端统一走 media_kit/libmpv（经 `just_audio_media_kit` / 自定义 `SongloftJustAudioPlatform`），无回退原生、无 kill-switch
- HyperOS3 等需 `androidStopForegroundOnPause: false` 防后台回收
- **Bundle 模式 Android**：CWD 是 `/`，covers 目录路径必须相对于 `DBPath` 而非 CWD 解析（`da65db1` 修复）
- **Bundle 模式原生桥接**：Android 用 `Class.forName("mobile.Mobile")` 反射调用 gomobile 生成类，未打包 `.aar` 时 `isAvailable()` 返回 false（优雅降级）；iOS 同理用 Swift 调用 `MobileStart` 等 Objective-C 函数
- **Bundle 桌面子进程**：`DesktopBackendService` 在 Flutter 可执行文件**同目录**（macOS 在 `Contents/Resources/`）查找 `songloft-server`，通过 stdout 解析实际监听端口

---

## JS 插件

- 源码 `plugins/src/<name>/`，构建产物在各插件仓库的 GitHub Releases
- 新建插件：`npx create-songloft-plugin@latest`（交互式脚手架，支持 WebView / Lynx 两种渲染引擎模板，详见 `plugins/toolchain/README.md`）
- 沙盒：QuickJS。`fetch` 是运行时内置的全局函数（走 `__go_fetch_async`），`console` 系列（`songloft.log` 等）是日志包装——两者均非桥接调用。桥接由 `internal/jsplugin/api_bridge.go` 的 `HandleBridgeCall` 分发，namespace 包括：`storage`、`persistent-storage`、`songs`、`playlists`、`tags`、`comm`、`plugin`、`command`、`jsenv`、`fs`、`net`、`websocket`
- 路由：`/api/v1/jsplugin/{entry_path}/...`
- 公共资源：`/api/v1/jsplugin-assets/*` 提供嵌入在 Go 二进制中的 `theme.css`/`components.css`/`common.js`/字体，`injectHTMLHead` 自动注入到所有插件 HTML 页面（注入顺序：theme.css → components.css → common.js）
- 主题同步：`common.js` 内含 embed 检测 + 主题桥接（URL `?theme=` 参数 + `postMessage` 实时更新 + `data-theme` 属性 + `songloft-theme-change` 事件），暴露 `window.SongloftPlugin` 全局 API（`getTheme`/`onThemeChange`/`apiGet`/`apiPost`/`getCookies` 等）
- **客户端宿主桥接（WebView）**（`@songloft/client-sdk`，`common.js` 内含）：插件前端页面通过 `window.SongloftPlugin.host` / `player` / `getCookies` / `invokeHost` 等调用 Flutter 客户端宿主能力。native 平台走 `flutter_inappwebview.callHandler('songloftHost', {ns, method, params})`，Web/iframe 走 `postMessage` 到父窗口。分发逻辑在 `clients/player/lib/features/home/presentation/plugin_host_dispatch.dart`（传输无关、web-safe），native 桥接在 `plugin_host_bridge.dart`（mixin，注册 callHandler + 注入平台相关回调）。已注册的 namespace：`host`（getInfo）、`player`（播放控制）、`cookies`（Cookie 读取）、`favorite`（收藏状态同步，`refresh` 方法，传 `{songId, isFavorited}` 增量更新 Flutter 侧 FavoriteNotifier 缓存，不传参则全量重载）。**`window.SongloftPlugin` 的公开成员以 `common.js` 末尾那个对象字面量为唯一真实来源**。**插件不要手写宿主 API 的类型声明**，用 `@songloft/client-sdk` 的 `SongloftPluginGlobal`；非要手写就先去 `common.js` 核对那个字面量
- **Cookie 读取桥**（`window.SongloftPlugin.getCookies(origin)`）：读取宿主 WebView Cookie Store 中指定 origin 的 Cookie（含 HttpOnly），返回 `{name: value}` 映射。**仅原生客户端可用**（Android/iOS/macOS/Windows；Linux 桌面无 WebView 实现），Web 端因浏览器同源策略无法实现，调用会 reject。实现路径：`common.js getCookies()` → `invokeHost('cookies', 'get', {origin})` → Flutter `PluginHostDispatcher` → `cookieProvider` 回调 → `CookieManager.instance().getCookies(url: WebUri(origin))`。origin 必须含协议+主机（如 `https://example.com`），无效格式会被校验拒绝。典型用途：FN Connect 等第三方网关的会话复用（用户在应用内 WebView 登录后，插件读取 Cookie 用于后续 API 调用）
- **Lynx 原生渲染桥接**（`@songloft/lynx-plugin-sdk`，`renderEngine: "lynx"` 专用）：声明 `renderEngine: "lynx"` 的插件使用 ReactLynx 编写 UI，编译为 `.lynx.bundle`，宿主通过 Lynx `<frame>` 元素原生加载。通信走 `NativeModules.SongloftPluginBridge`（三端原生模块：Android `SongloftPluginBridgeModule.kt` / iOS `SongloftPluginBridgeModule.swift` / HarmonyOS `SongloftPluginBridgeModule.ets`），按 `frameId` 路由父子 frame 间 RPC + 事件推送。SDK 提供 `invokeHost(ns, method, params)` / `onPlayerState` / `onThemeChange` / `onPush` 等 API。**不使用** `window.SongloftPlugin` / `@songloft/client-sdk`（那些是 WebView 专用）。构建时自动生成 `index.html` + `.web.bundle` 供 Flutter 客户端 WebView 回退
- `theme.css` 定义 `--md-*` CSS 变量（亮/暗双主题）、字体声明与 reset 样式；`components.css` 提供组件库样式。所有使用这些变量的插件自动跟随主题切换
- 权限：manifest 中 `permissions: ["net", "storage", "fs:music", ...]`，运行时由 `internal/jsplugin` 校验
- **fetch 内部控制头**：`X-Fetch-No-Redirect`（不跟随重定向，收集中间跳 `Set-Cookie` 必需）、`X-Fetch-Timeout-Ms`（单次超时 100–30000ms）、`X-Fetch-Insecure`（跳过 TLS 校验，**需 `net:insecure-tls` 权限**）。三者一律不转发给上游——`X-Fetch-Insecure` 曾因未实现而被当普通头透给上游（songloft-org/songloft#401）
- **新增权限常量要同步三处，否则插件根本用不上**：① 宿主 `internal/jsplugin/permissions.go` 的 `AllPermissions`；② `plugins/toolchain/packages/plugin-builder/src/manifest.ts` 的 `VALID_PERMISSIONS`——**它在 build 时硬失败**（`unknown permission: xxx`），漏了这处插件连编译产物都出不来；③ `plugins/toolchain/packages/create-songloft-plugin/src/index.ts` 的 `AVAILABLE_PERMISSIONS`（脚手架勾选列表）。另外 `ValidatePermissions`（宿主侧）**刻意对未知权限只 warn 不拒绝**：它挂在安装/更新的必经路径上，拒绝会让声明了新权限的插件在**旧宿主**上直接装不上，砸掉的还是插件原本正常的其他功能；宽容是安全中立的（`CheckPermission` 只认它认识的字符串）。**但旧宿主的严格校验已经发布出去了**，所以新权限落地后仍需等宿主版本铺开才能发插件——`minHostVersion` **纯属信息字段、宿主从不强制**，别指望它拦住谁
- **`net:insecure-tls` 权限（跳过 TLS 证书校验）**：`jsruntime` 层**没有权限概念**，`fetch` 本身也不需要 `net` 权限；该权限靠 `service.go` 在建环境后调 `jsManager.SetAllowInsecureTLS(envID, ...)` 投影成 `JSEnv.allowInsecureTLS`（默认 false），是权限在运行时层的唯一体现。几个要点：① **`CheckPermission` 无冒号前缀匹配，`net` 不覆盖 `net:insecure-tls`**——刻意如此，否则所有已声明 `net` 的存量插件升级后会静默获得跳过校验的能力，而用户在权限列表里看不到任何变化；② 独立的 `insecureTransport`，**不能**复用 `sharedTransport` 再改单次请求的 `TLSClientConfig`——连接池按 Transport 隔离，那样会污染所有插件的连接复用；③ 子环境（`songloft.jsenv` 创建的）必须继承父插件的策略，否则「插件在子 worker 里发请求」会莫名校验失败；④ 无权限却带了该头时打 `slog.Warn` 而非静默丢弃——插件作者普遍误以为「宿主不识别就零副作用」，一条日志能省掉一整轮排查。**存在理由**：自建 NAS（飞牛 fnOS 5667）默认自签，且插件按**裸 IP** 访问，即便装了 CA 正式证书主体也是域名，按 IP 连必然 hostname mismatch——这类目标结构性无解
- 健康检查 + 文件指纹热更新均自动进行
- **UDP Socket API**（`songloft.net`，需 `net` 权限）：Go 侧托管 UDP socket + 消息推送模式。`udpBind` 创建 socket 并启动 reader goroutine，收到的 UDP 包通过 scheduler 队列异步推送到 JS 回调（`onData`）。支持多播组（`udpJoinMulticast/udpLeaveMulticast`），典型用途：SSDP 设备发现（DLNA/UPnP）。每插件最多 8 个 socket，有活跃 socket 的插件不会被空闲驱逐，插件卸载时自动清理。实现在 `internal/jsplugin/api_bridge_net.go`
- **TCP Socket API**（`songloft.net.tcpConnect`，需 `net` 权限）：出站 TCP 连接，`tcpConnect(host, port, options?)` 返回带 `send()/onData()/onClose()/close()` 的 socket 句柄。数据接收复用 UDP 的 Go readLoop + host event 队列推送模式（`postHostEvent("tcp_data")` → JS `__dispatchHostEvent`）。**data 为 base64 编码的原始字节**（send 侧 `btoa`、onData 侧 `atob`，同 UDP）：TCP 是字节流，一次读取可能在多字节 UTF-8 字符中间截断，raw string 会被 `json.Marshal` 替换为 U+FFFD 而永久损坏，故必须 base64；插件需累积字节跨 chunk 拼接后再 UTF-8 解码。**仅允许连接私有 / 回环 / 链路本地地址**（`isPrivateHostAllowed`，防 SSRF）；与 UDP 各自独立计每插件 8 个 socket 配额；有活跃 TCP 连接的插件不被空闲驱逐；卸载时自动清理。典型用途：控制本机 MPD（6600 端口 idle 事件推送）。实现在 `internal/jsplugin/api_bridge_tcp.go`
- **私有源认证**：`RegistryConfig` 支持 `token` 字段，拉取该源下所有资源时自动携带 `Authorization: Bearer <token>` 头，兼容 GitHub 私有仓库 PAT 和自托管私有源。详见 [插件源制作指南 · 私有源认证](docs/plugin_registry.md#私有源认证)
- **歌词/封面提供者**（`songloft.lyrics` / `songloft.covers`，无需权限）：插件通过 `registerProvider()` 注册为歌词或封面提供者。歌曲无歌词/封面时，宿主自动遍历已注册的提供者，调用 `/lyric-search` 或 `/cover-search` 端点（15s 超时，first-match-wins）。搜索参数包含 `title/artist/album`，歌词额外带 `duration`，两者均可选带 `fingerprint`（Chromaprint）和 `isrc`。搜到的歌词存入 DB（`scraped`），本地歌曲还嵌入文件标签；封面对本地歌曲下载到 `cover_path` 并嵌入标签，远程歌曲存 `cover_url`。提供者注册在空闲驱逐时不丢失，仅禁用插件时清理。实现在 `manager.go`（`SearchLyrics/SearchCover`）、`api_bridge.go`（JS API）、`handlers/music.go`（fallback 调用）。详见 [插件开发指南 · 歌词/封面提供者](docs/js-plugin-development-guide.md#songloft-lyrics-—-歌词提供者)

---

<a id="business-constraints"></a>

## 业务约束（按任务必读）

修改下列模块前，**必须先读对应章节**；涉及多个模块时一并阅读。这里只保留提醒，完整约束与历史依据见 [业务约束与踩坑](docs/pitfalls.md)。新经验遵循 `LESSONS.md` 流程，专题细节维护在对应文档，不再追加到入口。

| 修改范围 | 关键提醒 | 必读章节 |
|---|---|---|
| 插件商店身份 | 冲突必须在写入前返回；安装态按 entry_path + identity 匹配。 | [详情](docs/pitfalls.md#plugin-identity) |
| 插件商店缓存 | 仅缓存源数据，不缓存安装态；token 哈希入键，proxyDown 按次隔离。 | [详情](docs/pitfalls.md#registry-cache) |
| 扫描标题与验证目录 | 默认标签优先，尊重 filename 配置；验证曲库不能放在 /tmp 下。 | [详情](docs/pitfalls.md#scan-title) |
| 流式扫描入库 | 按目录定稿；取消保留已提取成果，写库使用独立 flushCtx。 | [详情](docs/pitfalls.md#scan-batches) |
| 目录遍历 | 避免逐项 Stat；保留目录名缓存、单飞与脱离请求取消的语义。 | [详情](docs/pitfalls.md#directory-walk) |
| DSF / DSD | 扫描默认值与迁移同步；不支持 DSD 的播放器需显式转码。 | [详情](docs/pitfalls.md#dsf) |
| 旁挂歌词 | 固定候选路径探测；不覆盖 manual，不把旁挂歌词回写音频标签。 | [详情](docs/pitfalls.md#sidecar-lyrics) |
| 标签写入 | 同目录临时文件原子写入；不支持的格式降级，不阻塞主流程。 | [详情](docs/pitfalls.md#tag-writing) |
| 音频指纹 | 默认不自动计算；采样 ≤120 秒、失败落库、可取消、并发随 CPU 收敛。 | [详情](docs/pitfalls.md#fingerprints) |
| 播放历史 | 按上下文仅记录 play；upsert 与 50 条裁剪同事务，不把序数当稳定位置。 | [详情](docs/pitfalls.md#play-history) |
| HLS 代理 | 默认关闭；保留同源校验与 SSRF 防护，改写 URL 使用相对路径。 | [详情](docs/pitfalls.md#hls-proxy) |
| 外发 HTTP 代理 | 使用代理感知客户端；ffmpeg 仅支持 HTTP/HTTPS 代理。 | [详情](docs/pitfalls.md#http-proxy) |
| 私网代理白名单 | 空白名单维持私网封禁；仅影响资源及转码代理，不放宽 HLS。 | [详情](docs/pitfalls.md#private-allowlist) |
| 远程 URL 转码 | 不落盘、不入库；与 seek/均衡流共享 4 个槽位，满则 503。 | [详情](docs/pitfalls.md#url-transcode) |
| 音乐缓存 | 流式代理不等整首下载；落盘转码失败保留原码，不改变显式下载语义。 | [详情](docs/pitfalls.md#music-cache) |
| 均衡与实时转码 | 共用可配置响度及缓存键；保留 Range、精确字节数、延迟 206 与 Flush。 | [详情](docs/pitfalls.md#normalization) |
| 服务端 seek | seek 与 Range 可组合；保留 CBR、尾部 3 秒护栏和未知时长超时。 | [详情](docs/pitfalls.md#seek-stream) |
| 歌曲持久化 | 仅用于用户合法拥有的音乐；转码或歌词获取失败不阻塞持久化。 | [详情](docs/pitfalls.md#song-download) |
| 跨设备文件搬移 | 使用 internal/fileutil.MoveFile；不要用裸 os.Rename 搬临时文件到其他挂载点。 | [详情](docs/pitfalls.md#file-move) |
