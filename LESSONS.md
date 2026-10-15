# 开发经验

保存主仓库与公共宿主契约的证据型经验；执行见 [AI 开发流程](docs/ai-development.md)。后续会话在相关任务前读取适用条目。客户端、插件和设备经验放所属子仓库。

## 记录与升级

- 每条含规则、范围、失败/修复证据、正确示例和独立事件日期；使用完整仓库引用与 commit SHA/链接。
- `Seen N×` 统计独立事件的不同日期，同一事故反复复盘不递增；三天后提出规范升级并标记权威位置。已有铁律、严重问题不等待三次。
- 不编造次数、修复状态或验证结果，与 [英文版](LESSONS.en.md) 同步，敏感信息脱敏。

## 条目

### SORT-001：歌单文字排序必须在分页前统一生成拼音键

- 范围：歌单歌曲分页、播放定位 ID 列表及永久名称排序，songloft-org/songloft#524。
- 失败证据：2026-10-15，基线 `cffffe39a05f32df6651e3e47acefb4692b8b533` 的 `applyPlaylistSongOrder` 直接按 SQLite 文本编码排列，中文首字母跳动、英文大小写分组；新回归在旧实现下的 title/artist/album 双方向均失败。隔离旧 filters.go 后，永久 name_asc/name_desc 回归也因错误顺序失败，日志 `/tmp/songloft-524/{before,permanent-before}.log`。
- 修复：本次工作树注册确定性 SQLite 拼音键函数，保留非中文字符、忽略大小写及首尾空白，由同一个排序构造器在 LIMIT/OFFSET 前使用；等值键继续按 position 同方向排序。多音字使用字典默认读音，不承诺词组语境识别。两个客户端将「标题」菜单明确为「歌曲名称」。
- Done right：真实内存 Repository 回归覆盖简繁中文/英文混排、大小写与同音等值、四页拼接、搜索、数字/自定义顺序；文件数据库双连接验证函数注册。后端完整测试/vet、定向 race 和 8 组真实 HTTP 场景通过；30,000 首合成曲目的中间页约 344ms、完整 ID 约 263ms，二者顺序一致（单次本机采样，非性能保证）。证据 `/tmp/songloft-524/{backend-focused,race,perf}.log`、`api-result.json`。
- Seen 1× (2026-10-15)。状态：经验；Android 原机验收与发布状态另行核对。

双客户端中英文 390px 浏览器实测名称升降序、艺术家排序、后端偏好保存及重新进入恢复通过，页面异常为零；证据 `/tmp/songloft-524/{flutter-browser,lynx-browser,validation}.json`。浏览器验证不替代 Android 真机验收。

### PLUG-001：插件版本兼容性需要核对更新入口的比较规则

- 范围：插件清单校验、服务端更新检查与安装、Flutter/Lynx 插件商店及 GitHub 发现页。
- 证据：2026-10-14，为 [songloft-org/songloft#495](https://github.com/songloft-org/songloft/issues/495) 准备补丁版时，先核对构建器与服务端便拟使用 `v2026.10.14.1`；用户要求继续核对完整升级链路后，发现各入口并不统一。主仓库 `6b050ff34c52f8140063e373319692814b030ec0` 的构建器/宿主清单正则接受四段，`CompareVersion` 逐段比较，真实 Repository + HTTP fixture 的 `CheckUpdate`/非强制 `DownloadUpdate` 完成 `2026.10.14 → 2026.10.14.1` 安装；已发布 v2.13.1 的服务端比较实现相同。
- 限制：Flutter `a0b2ae17194dd0a1acdaa7c0050ca4681e119c07`、Lynx `9a92929f5f33efdefecc1a01632cb8270ebdc59b` 的普通插件源比较可识别第四段，但 GitHub 发现页分别调用 `comparePluginVersions` / `compareStableVersions`，严格只接受三段，四段比较返回未知，不显示更新提示。这不是所有插件升级都失败，也不能宣称所有入口均兼容。tag 可带 `v`，清单中的 `version` 必须不带。
- Done right：实际执行 Go 清单/更新/下载/安装链路和两客户端的源码比较函数；三段 `2026.10.15` 在这些比较入口均识别为新版本，本次发布采用该版本，不修改客户端原生升级协议。证据在 `/tmp/songloft-495-20261014/version-upgrade-probe.log`、`version-probe_test.go`、`client-version-probe.dart`。临时 Go 测试已清理，未修改产品代码。
- 规则：发布新版本格式前核对所有相关入口，区分打包/安装可接受与 UI 能否提示更新；不以日期/semver 文案代替实际校验。不能为插件比较直接放宽客户端原生应用升级协议。
- 后续修复：按用户要求改为自动四段日期版本发布，Flutter/Lynx GitHub 发现页采用插件专用三/四段比较，原生应用与最低宿主版本比较仍只接受三段。回归覆盖 `.9 → .10`、跨日、三段与 `.0` 相等，以及更新按钮和对应包安装请求；旧 Lynx 发现页在新增交互用例中显示 Reinstall，修复后显示更新版本。
- 迁移规则：既有 `2026.10.15` 与 `2026.10.15.0` 数值相等，首次迁移从 `.1` 开始；北京时间日期追上前沿用该前缀，下一新日期从 `.0` 开始。构建器、Go 安装链路接受四段不代表旧客户端发现页自动获得该能力。
- Not run：各历史客户端发布包的完整矩阵、真机安装；交互回归不证明已安装旧客户端能提示四段更新。
- Seen 1× (2026-10-14)。状态：经验，当前源码的四段插件比较已统一；历史客户端边界保留。

### FLOW-001：关联 Issue 的修复推送后，主动交付待确认的回复草稿

- 范围：主仓库与子仓库的开发交付及 Issue 沟通。
- 失败：2026-10-13，songloft-org/songloft#495 的修复 [songloft-org/songloft-plugin-miot@1cb1b52](https://github.com/songloft-org/songloft-plugin-miot/commit/1cb1b5214612e9fe611e8d8c61bfa47d5fc2f579) 已推送，v2026.10.14 已发布并同步主仓库 [540bf0c](https://github.com/songloft-org/songloft/commit/540bf0c76f87f08ba54dec1573cf9922126ad68c)，交付却未准备并展示 Issue 回复草稿。用户追问后才明确安排回复，造成额外往返。
- 修复：当前未提交的流程更新，在 AGENTS、AI 流程及 dev-flow 中要求推送后主动展示完整草稿、目标 Issue 和具体评论确认动作；已有适用评论授权直接执行。草稿跟随 CI/发布结果更新，评论与关闭分别授权。
- Done right：用户明确要求回复后，发布说明已[发表并回读验证](https://github.com/songloft-org/songloft/issues/495#issuecomment-6147071915)，说明可用版本、实际验证和真机/时长差异的验收边界。新增「主动草稿 → 确认 → 发表」流程的后续实用验证待补，不能把本次补发视为已验证该完整流程。
- Seen 1× (2026-10-13)。状态：按用户明确要求立即纳入 [关联 Issue 回复](docs/ai-development.md#关联-issue-回复)，不等待经验次数门槛。

### TEST-001：临时目录清理失败要核对后台文件写入

- 范围：Go 插件生命周期、异步字节码缓存与测试临时目录。
- 规则：`TempDir RemoveAll: directory not empty` 不能直接归为环境问题或用延时掩盖；检查测试结束和后台写盘的先后顺序。重跑通过只证明该次检查通过，不代表生命周期竞态已修复。
- 失败：2026-10-14，songloft-org/songloft#518 的父指针提交 [`405e8a0`](https://github.com/songloft-org/songloft/commit/405e8a0f2697bac53c3311567fdca025256335df) 在 [CI 首轮](https://github.com/songloft-org/songloft/actions/runs/38526078645/attempts/1) 的 `TestManager_ReloadPlugin` 清理阶段报错。日志先出现管理器关闭，再出现 `bytecode cached`；本批只改客户端指针，未改 Go 源码。
- 调查与处理：`JSService.Load` 异步执行字节码编译及缓存写入，关闭流程未等待该写盘；证据指向后台写盘与临时目录清理竞态。同一 SHA 本地 `CGO_ENABLED=0 go test -tags dev,lite ./internal/jsplugin -run '^TestManager_ReloadPlugin$' -count=30` 通过；失败作业重跑后 Quality 通过。日志在 `/tmp/songloft-518-{parent-ci-failed,reload-repeat}.log`。本批未修复该后台任务生命周期，后续应补确定性回归与退出等待。
- Done right：保留首轮失败、实际 SHA 与重跑结果；若修复生命周期，验证关闭返回后不再写缓存，再清理临时目录，不把单次重跑当成修复证明。
- 后续证据：v2.13.2 的 `c2ecbf9f924528b75106daa56f1cfe087f91377f` 在 [Release 首轮](https://github.com/songloft-org/songloft/actions/runs/38605138465/attempts/1) 的 `TestManager_ReloadPlugin`、[第二轮](https://github.com/songloft-org/songloft/actions/runs/38605138465/attempts/2) 的 `TestIdleWakeup_HTTPRequestReloadsEvictedPlugin` 均出现关闭后 `bytecode cached` 与临时目录清理失败。2026-10-15 排查确认该 SHA 的 `JSService.Load` 仍用未等待的 goroutine 编译并写缓存，`Stop` / `Manager.Close` 未等待缓存任务；属于同一未修复生命周期问题的后续证据，不增加独立事件次数。日志在 `/tmp/songloft-run-38605138465-{attempt1-failed,failed}.log`；本次只调查，未修复或重跑 CI。
- 修复与验证（2026-10-15）：本次修复在 `JSService.Load` 持锁登记缓存完成信号，后台任务退出时关闭；`Stop` 在无锁等待整个编译及写盘完成后才执行生命周期清理，因此卸载、重载和管理器关闭也会等待。新增 `TestBytecodeCache_ShutdownWaitsForCompletion`，通过日志屏障阻塞真实缓存任务，覆盖 stop/unload/close；用 Go overlay 隔离加载旧 `service.go` 时三项均因提前返回失败，修复后三项通过。Linux amd64 / Go 1.26.0 / Node 22.17.1，基线 `3bbea95fa9eba7519eed1e4e000123923d78e3b4` + 本次工作树：`CGO_ENABLED=0 go vet -tags dev,lite ./...`、`CGO_ENABLED=0 go test -tags dev,lite ./...`、`node --test scripts/tests/*.test.mjs` 全通过；新回归与上述两项原失败测试重复 30 轮通过，同组 `CGO_ENABLED=1 go test -race -tags dev,lite` 通过。证据在 `/tmp/songloft-cache-{before,regression,repeat,all-tests,vet,scripts,race}.log`。Not run：各原生平台实机验证；本次未修改客户端或 API。
- Seen 1× (2026-10-14)。状态：经验，本次提交修复缓存任务的停止等待；远端构建结果另行核对同一 SHA。

### TEST-002：并发事件回归必须同步请求就绪与回调完成

- 范围：Go / QuickJS 入站 WebSocket 与挂起 HTTP 请求的集成测试。
- 失败：2026-10-15，songloft-org/songloft#520 的指针提交 `8252cb762e8363283a860c3a091748a0aadb4b02` 在 [CI 首轮](https://github.com/songloft-org/songloft/actions/runs/38670932842/attempts/1) 的 `TestInboundWebSocketHostEventDeliveredDuringHTTPRequest` 返回空事件列表；本批没有修改运行时代码。本地 `CGO_ENABLED=0 go test -tags dev,lite ./internal/jsplugin -run '^TestInboundWebSocketHostEventDeliveredDuringHTTPRequest$' -count=30` 也复现同一失败。测试用固定 100ms 推断 HTTP 已清空列表，再以 500ms 定时器返回，未同步真正的请求开始与 close 回调。
- 规则与修复：使用 WebSocket `socket-ready` / `http-started` 消息建立先后顺序，HTTP Promise 由 close 回调记录事件后完成；Go 侧设置读与请求截止时间，只用于失败收敛，不把延时当同步。仅调整测试，未改运行时。
- Done right：修复后目标用例重复 30 轮、`CGO_ENABLED=0 go test -tags dev,lite ./...` 和 `go vet -tags dev,lite ./...` 通过；Go overlay 隔离删去 close 事件记录后，用例按预期因只收到 `message:ping` 而失败。基线 `9b5ffd96cdd156cfd12b170a6df02eff5aa0f5f1` + 本次测试工作树，Linux amd64 / Go 1.26.0，日志 `/tmp/songloft-520-followup/parent-{ci-failed,websocket-repeat,websocket-fixed,websocket-negative,tests-fixed,vet-fixed}.log`。最终提交与 CI 另按对应 SHA 核对。
- Seen 1× (2026-10-15)。状态：经验。

### DOCS-001：自动侧栏必须固定扫描根并显式排除依赖目录

- 范围：`docs/.vitepress/config.mts` 的自动侧栏及 LLM 文档索引。
- 证据：2026-10-15，基线 `8252cb762e8363283a860c3a091748a0aadb4b02` 的 `path: '.'` 从仓库根启动时混入子模块与依赖，运行中的 `/@siteData` 返回 7,497,742 字节；从 `docs` 构建也会因 `node_modules` 未被默认排除产生 `No matching file found for sidebar link`。
- 规则与修复：插件 1.7.0 用 `join(process.cwd(), path)`，不是 `resolve`，不能直接传绝对路径；用配置文件的 `import.meta.url` 定位 docs，再转为当前目录相对路径。递归仅使用显式 `ignoreList`，因此同时列出依赖、构建、静态资源及插件原有排除目录，不屏蔽告警。
- Done right：根目录执行 `node docs/node_modules/vitepress/bin/vitepress.js build docs`，缺失侧栏链接警告归零；在 `docs` 启动 dev 后站点数据降至 95,743 字节。Docker Chrome 实际点击开发版及构建产物的中英文侧栏、语言链接和移动菜单，采集到的 66 个侧栏链接经 HTTP 均返回 200。证据：`/tmp/songloft-sidebar-WVrz5I/` 中的构建日志、运行时数据、浏览器记录及截图。
- 边界：构建包体积警告仍存在；编辑配置触发插件热重启时遇到 MiniSearch 重复索引，冷启动验证正常，本次未修复热重启。Not run：单元测试和类型检查，本次采用构建及真实运行验证。
- Seen 1× (2026-10-15)。状态：经验；扫描范围修复在本次未提交工作树。

## 条目格式

```text
## 稳定编号：一句话规则
范围：仓库/模块/平台
失败：owner/repo#编号、commit SHA、日期、行为与后果
修复：commit SHA、验证对象与实际结果
Done right：正确示例及证据；没有时写待补
Seen N× (YYYY-MM-DD, YYYY-MM-DD)
状态：经验 / 升级候选 / 已升级（规范位置）
```
