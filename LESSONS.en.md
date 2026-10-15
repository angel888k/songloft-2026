# Development Lessons

Evidence-backed lessons for the parent and public host contracts; see the [AI workflow](docs/en/ai-development.md). Subsequent sessions read applicable entries before work. Client, plugin, and device lessons belong in their own repositories.

## Recording and Promotion

- Include rule, scope, failure/fix evidence, successful examples, and independent event dates. Use full repository references and commit SHAs/links.
- `Seen N×` counts distinct dates of independent events, not repeated reviews. Three distinct dates make promotion candidates; mark the authoritative destination. Existing hard rules and serious failures do not wait for three occurrences.
- Never invent counts, resolution status, or validation. Synchronize the [Chinese version](LESSONS.md) and redact sensitive information.

## Entries

### SORT-001: Generate shared pinyin keys before playlist pagination

- Scope: playlist song pages, playback-position ID lists and permanent name sorting, songloft-org/songloft#524.
- Failure evidence: on 2026-10-15, baseline `cffffe39a05f32df6651e3e47acefb4692b8b533` ordered text directly in SQLite's binary order, producing jumps between Chinese initials and grouping English by case. New title/artist/album regressions fail in both directions on the old implementation. An isolated overlay of the old filters.go also fails the permanent name_asc/name_desc regressions. Logs: `/tmp/songloft-524/{before,permanent-before}.log`.
- Fix: this working tree registers a deterministic SQLite pinyin-key function that preserves non-Chinese characters and ignores case and surrounding whitespace. The shared order builder uses it before LIMIT/OFFSET, retaining same-direction position ties. Polyphones use dictionary default readings, without contextual phrase recognition. Both clients rename the playlist “Title” menu to “Song name”.
- Done right: real in-memory Repository regressions cover mixed simplified/traditional Chinese and English, case/homophone ties, four concatenated pages, search, numeric/custom order; two file-database connections verify function registration. Full backend tests/vet, focused race tests and eight real HTTP scenarios pass. A single local sample of 30,000 synthetic songs takes about 344ms for a middle page and 263ms for all IDs, with matching order; this is not a performance guarantee. Evidence: `/tmp/songloft-524/{backend-focused,race,perf}.log`, `api-result.json`.
- Seen 1× (2026-10-15). Status: lesson; Android hardware acceptance and release status require separate verification.

Both clients pass Chinese/English 390px browser checks for name order in both directions, artist sorting, backend preference persistence and re-entry restoration, with no page exceptions. Evidence: `/tmp/songloft-524/{flutter-browser,lynx-browser,validation}.json`. Browser checks do not replace Android hardware acceptance.

### PLUG-001: Check version comparison at every plugin update entry

- Scope: plugin manifest validation, server update checks/install, Flutter/Lynx catalogs, and GitHub discovery.
- Evidence: on 2026-10-14, preparing a patch for [songloft-org/songloft#495](https://github.com/songloft-org/songloft/issues/495), checking only the builder and host led to a proposed `v2026.10.14.1`. The user's request to inspect the complete update chain exposed inconsistent entry rules. At parent `6b050ff34c52f8140063e373319692814b030ec0`, builder/host manifest patterns accept four components and `CompareVersion` compares each component. Real Repository + HTTP fixture checks successfully install `2026.10.14 → 2026.10.14.1` through `CheckUpdate` and non-forced `DownloadUpdate`. Published server v2.13.1 uses the same comparison implementation.
- Limitation: Flutter `a0b2ae17194dd0a1acdaa7c0050ca4681e119c07` and Lynx `9a92929f5f33efdefecc1a01632cb8270ebdc59b` catalog comparisons recognize a fourth component, but GitHub discovery uses `comparePluginVersions` / `compareStableVersions`, which strictly accept three. Four-component comparisons return unknown and omit update prompts. This does not mean every plugin update fails or that every entry supports the format. Tags may start with `v`; manifest `version` must omit it.
- Done right: execute actual Go manifest/check/download/install paths and both clients' source comparison functions. Three-component `2026.10.15` is recognized as newer by these entries and is used for this release without changing native application update protocols. Evidence: `/tmp/songloft-495-20261014/version-upgrade-probe.log`, `version-probe_test.go`, and `client-version-probe.dart`. The temporary Go test is removed; product code is unchanged.
- Rule: check every relevant update entry before introducing a version format. Distinguish accepted packages from UI update prompts; date/semver wording is not actual validation. Do not relax native application update protocols merely to change plugin comparison.
- Follow-up fix: the user requested automatic four-component date releases. Flutter/Lynx GitHub discovery now uses plugin-specific three/four-component comparisons; native application and minimum-host comparisons remain restricted to three components. Regressions cover `.9 → .10`, day rollover, three components equaling `.0`, update buttons, and matching package installation requests. The old Lynx discovery page shows Reinstall in the new interaction test; the fix displays the update version.
- Migration rule: existing `2026.10.15` equals `2026.10.15.0`, so migration begins at `.1`. Retain that prefix until the Beijing date catches up; the next new date begins at `.0`. Builder/Go support does not automatically update discovery behavior in older clients.
- Not run: the complete historical client release matrix and physical-device installation. Interaction regressions do not prove installed older clients prompt for four-component updates.
- Seen 1× (2026-10-14). Status: lesson; current source comparisons for four-component plugins are unified, with historical-client boundaries retained.

### FLOW-001: Proactively deliver an Issue reply draft for approval after pushing its fix

- Scope: development delivery and Issue communication in the parent and child repositories.
- Failure: on 2026-10-13, the fix for songloft-org/songloft#495, [songloft-org/songloft-plugin-miot@1cb1b52](https://github.com/songloft-org/songloft-plugin-miot/commit/1cb1b5214612e9fe611e8d8c61bfa47d5fc2f579), was pushed, v2026.10.14 was published, and the parent was updated at [540bf0c](https://github.com/songloft-org/songloft/commit/540bf0c76f87f08ba54dec1573cf9922126ad68c). Delivery omitted a prepared, visible Issue reply draft. Only after the user asked was the reply explicitly arranged, causing extra exchanges.
- Fix: the current uncommitted workflow update requires AGENTS, AI workflow, and dev-flow to proactively show the complete draft, target Issue, and specific comment approval action after pushing; existing applicable comment authorization permits direct posting. Update drafts as CI/release results change, with separate authorization for comments and closure.
- Done right: after the user explicitly requested the reply, the release update was [posted and read back](https://github.com/songloft-org/songloft/issues/495#issuecomment-6147071915), stating the available version, actual validation, and device/duration acceptance limits. Subsequent practical validation of the new “proactive draft → approval → posting” workflow remains pending; the late reply does not establish that complete flow.
- Seen 1× (2026-10-13). Status: incorporated immediately at the user's explicit request into [Linked Issue Replies](docs/en/ai-development.md#linked-issue-replies), without waiting for the occurrence threshold.

### TEST-001: Check background file writes when temporary directory cleanup fails

- Scope: Go plugin lifecycle, asynchronous bytecode caching, and test temporary directories.
- Rule: do not dismiss `TempDir RemoveAll: directory not empty` as an environment problem or hide it with a delay. Check the ordering of test completion and background disk writes. A successful rerun establishes that attempt's result, not a lifecycle-race fix.
- Failure: on 2026-10-14, the parent gitlink commit [`405e8a0`](https://github.com/songloft-org/songloft/commit/405e8a0f2697bac53c3311567fdca025256335df) for songloft-org/songloft#518 failed during `TestManager_ReloadPlugin` cleanup in the [first CI attempt](https://github.com/songloft-org/songloft/actions/runs/38526078645/attempts/1). Logs show manager closure before `bytecode cached`. This batch changed client gitlinks without changing Go source.
- Investigation and handling: `JSService.Load` compiles and writes bytecode asynchronously, and shutdown does not join that disk write. Evidence points to a race between background writes and temporary directory cleanup. At the same SHA, local `CGO_ENABLED=0 go test -tags dev,lite ./internal/jsplugin -run '^TestManager_ReloadPlugin$' -count=30` passed; rerunning failed jobs passed Quality. Logs are in `/tmp/songloft-518-{parent-ci-failed,reload-repeat}.log`. This batch does not fix that background-task lifecycle; a deterministic regression and shutdown wait remain follow-up work.
- Done right: retain the first failure, actual SHA, and rerun result. A lifecycle fix should verify that no cache writes occur after shutdown returns, then clean the temporary directory. Do not treat one successful rerun as proof of repair.
- Follow-up evidence: v2.13.2 at `c2ecbf9f924528b75106daa56f1cfe087f91377f` failed in `TestManager_ReloadPlugin` on the [first Release attempt](https://github.com/songloft-org/songloft/actions/runs/38605138465/attempts/1) and `TestIdleWakeup_HTTPRequestReloadsEvictedPlugin` on the [second attempt](https://github.com/songloft-org/songloft/actions/runs/38605138465/attempts/2). Both logs show `bytecode cached` after shutdown and temporary directory cleanup failure. Investigation on 2026-10-15 confirmed that this SHA still compiles and writes the cache in an unjoined goroutine in `JSService.Load`; neither `Stop` nor `Manager.Close` waits for the cache task. This is further evidence of the same unresolved lifecycle problem, not another independent occurrence. Logs: `/tmp/songloft-run-38605138465-{attempt1-failed,failed}.log`. This investigation did not implement a fix or rerun CI.
- Fix and validation (2026-10-15): this change registers a cache-completion channel under the `JSService.Load` lock and closes it when the background task exits. `Stop` waits without holding that lock for compilation and disk writes before lifecycle cleanup, so unloading, reloading, and manager shutdown also wait. `TestBytecodeCache_ShutdownWaitsForCompletion` blocks the real cache task with a logging barrier and covers stop/unload/close. Loading the old `service.go` through an isolated Go overlay makes all three cases fail because shutdown returns early; the fix passes all three. Linux amd64 / Go 1.26.0 / Node 22.17.1, baseline `3bbea95fa9eba7519eed1e4e000123923d78e3b4` plus this working-tree change: `CGO_ENABLED=0 go vet -tags dev,lite ./...`, `CGO_ENABLED=0 go test -tags dev,lite ./...`, and `node --test scripts/tests/*.test.mjs` all pass. The new regression and the two previously failing tests pass 30 repetitions, and the same group passes with `CGO_ENABLED=1 go test -race -tags dev,lite`. Evidence: `/tmp/songloft-cache-{before,regression,repeat,all-tests,vet,scripts,race}.log`. Not run: native-device validation; no client or API changes in this fix.
- Seen 1× (2026-10-14). Status: lesson; this commit fixes the cache-task shutdown wait. Verify remote builds separately at the same SHA.

### TEST-002: Synchronize request readiness and callback completion in concurrent-event regressions

- Scope: Go / QuickJS integration tests for inbound WebSocket events during suspended HTTP requests.
- Failure: on 2026-10-15, the pointer commit `8252cb762e8363283a860c3a091748a0aadb4b02` for songloft-org/songloft#520 returned an empty event list in `TestInboundWebSocketHostEventDeliveredDuringHTTPRequest` on the [first CI attempt](https://github.com/songloft-org/songloft/actions/runs/38670932842/attempts/1), without runtime changes in this batch. Local `CGO_ENABLED=0 go test -tags dev,lite ./internal/jsplugin -run '^TestInboundWebSocketHostEventDeliveredDuringHTTPRequest$' -count=30` reproduced the same failure. A fixed 100ms sleep assumed the HTTP handler had reset its list, followed by a 500ms response timer; neither actual request start nor the close callback was synchronized.
- Rule and fix: establish ordering through WebSocket `socket-ready` / `http-started` messages, then resolve the HTTP Promise after the close callback records its event. Go read/request deadlines bound failures; delays do not provide synchronization. Only the test changes, not runtime code.
- Done right: the corrected test passes 30 repetitions, `CGO_ENABLED=0 go test -tags dev,lite ./...` and `go vet -tags dev,lite ./...` pass. An isolated Go overlay removing close-event recording makes the test fail as expected with only `message:ping`. Baseline `9b5ffd96cdd156cfd12b170a6df02eff5aa0f5f1` plus this test working tree, Linux amd64 / Go 1.26.0; logs: `/tmp/songloft-520-followup/parent-{ci-failed,websocket-repeat,websocket-fixed,websocket-negative,tests-fixed,vet-fixed}.log`. Check the final commit and CI separately at the corresponding SHA.
- Seen 1× (2026-10-15). Status: lesson.

### DOCS-001: Fix the automatic sidebar scan root and explicitly exclude dependencies

- Scope: automatic sidebars and LLM documentation indexes configured in `docs/.vitepress/config.mts`.
- Evidence: on 2026-10-15, baseline `8252cb762e8363283a860c3a091748a0aadb4b02` used `path: '.'`; starting from the repository root included submodules and dependencies, returning 7,497,742 bytes from the live `/@siteData` endpoint. Building from `docs` also emitted `No matching file found for sidebar link` because `node_modules` was not excluded by default.
- Rule and fix: plugin 1.7.0 uses `join(process.cwd(), path)`, not `resolve`, so an absolute path cannot be passed directly. Locate docs through the config's `import.meta.url`, then make that path relative to the current directory. Recursion only uses the explicit `ignoreList`, so include dependency, build, static-resource, and original default exclusions instead of suppressing warnings.
- Done right: `node docs/node_modules/vitepress/bin/vitepress.js build docs` from the repository root produces zero missing-sidebar-link warnings; starting dev from `docs` reduces site data to 95,743 bytes. Docker Chrome exercised Chinese/English sidebars, language links, and the mobile menu against both dev and built output; all 66 captured sidebar links returned HTTP 200. Evidence: build logs, runtime data, browser records, and screenshots in `/tmp/songloft-sidebar-WVrz5I/`.
- Limits: the chunk-size warning remains. Config edits triggered a plugin hot-restart failure with duplicate MiniSearch indexes; cold-start verification passed, and this change does not fix hot restart. Not run: unit tests and type checks; validation used builds and actual runtime behavior.
- Seen 1× (2026-10-15). Status: lesson; scan-scope fix is in this uncommitted worktree.

## Entry Format

```text
## Stable ID: one-sentence rule
Scope: repository/module/platform
Failure: owner/repo#number, commit SHA, date, behavior and impact
Fix: commit SHA, validation object and observed outcome
Done right: successful example and evidence; pending if unavailable
Seen N× (YYYY-MM-DD, YYYY-MM-DD)
Status: lesson / promotion candidate / promoted (authoritative destination)
```
