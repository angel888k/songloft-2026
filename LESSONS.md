# 开发经验

保存主仓库与公共宿主契约的证据型经验；执行见 [AI 开发流程](docs/ai-development.md)。后续会话在相关任务前读取适用条目。客户端、插件和设备经验放所属子仓库。

## 记录与升级

- 每条含规则、范围、失败/修复证据、正确示例和独立事件日期；使用完整仓库引用与 commit SHA/链接。
- `Seen N×` 统计独立事件的不同日期，同一事故反复复盘不递增；三天后提出规范升级并标记权威位置。已有铁律、严重问题不等待三次。
- 不编造次数、修复状态或验证结果，与 [英文版](LESSONS.en.md) 同步，敏感信息脱敏。

## 条目

### PLUG-001：插件版本兼容性需要核对更新入口的比较规则

- 范围：插件清单校验、服务端更新检查与安装、Flutter/Lynx 插件商店及 GitHub 发现页。
- 证据：2026-10-14，为 [songloft-org/songloft#495](https://github.com/songloft-org/songloft/issues/495) 准备补丁版时，先核对构建器与服务端便拟使用 `v2026.10.14.1`；用户要求继续核对完整升级链路后，发现各入口并不统一。主仓库 `6b050ff34c52f8140063e373319692814b030ec0` 的构建器/宿主清单正则接受四段，`CompareVersion` 逐段比较，真实 Repository + HTTP fixture 的 `CheckUpdate`/非强制 `DownloadUpdate` 完成 `2026.10.14 → 2026.10.14.1` 安装；已发布 v2.13.1 的服务端比较实现相同。
- 限制：Flutter `a0b2ae17194dd0a1acdaa7c0050ca4681e119c07`、Lynx `9a92929f5f33efdefecc1a01632cb8270ebdc59b` 的普通插件源比较可识别第四段，但 GitHub 发现页分别调用 `comparePluginVersions` / `compareStableVersions`，严格只接受三段，四段比较返回未知，不显示更新提示。这不是所有插件升级都失败，也不能宣称所有入口均兼容。tag 可带 `v`，清单中的 `version` 必须不带。
- Done right：实际执行 Go 清单/更新/下载/安装链路和两客户端的源码比较函数；三段 `2026.10.15` 在这些比较入口均识别为新版本，本次发布采用该版本，不修改客户端原生升级协议。证据在 `/tmp/songloft-495-20261014/version-upgrade-probe.log`、`version-probe_test.go`、`client-version-probe.dart`。临时 Go 测试已清理，未修改产品代码。
- 规则：发布新版本格式前核对所有相关入口，区分打包/安装可接受与 UI 能否提示更新；不以日期/semver 文案代替实际校验。不能为插件比较直接放宽客户端原生应用升级协议。
- Not run：真实客户端页面上的四段版本交互；各历史客户端发布包的完整矩阵。四段仅是当前实现下部分入口接受的行为，尚无统一版本契约。
- Seen 1× (2026-10-14)。状态：经验，统一插件版本契约待另行处理。

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
- Seen 1× (2026-10-14)。状态：经验，生命周期修复待处理。

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
