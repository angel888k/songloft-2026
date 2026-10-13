# Development Lessons

Evidence-backed lessons for the parent and public host contracts; see the [AI workflow](docs/en/ai-development.md). Subsequent sessions read applicable entries before work. Client, plugin, and device lessons belong in their own repositories.

## Recording and Promotion

- Include rule, scope, failure/fix evidence, successful examples, and independent event dates. Use full repository references and commit SHAs/links.
- `Seen N×` counts distinct dates of independent events, not repeated reviews. Three distinct dates make promotion candidates; mark the authoritative destination. Existing hard rules and serious failures do not wait for three occurrences.
- Never invent counts, resolution status, or validation. Synchronize the [Chinese version](LESSONS.md) and redact sensitive information.

## Entries

### FLOW-001: Proactively deliver an Issue reply draft for approval after pushing its fix

- Scope: development delivery and Issue communication in the parent and child repositories.
- Failure: on 2026-10-13, the fix for songloft-org/songloft#495, [songloft-org/songloft-plugin-miot@1cb1b52](https://github.com/songloft-org/songloft-plugin-miot/commit/1cb1b5214612e9fe611e8d8c61bfa47d5fc2f579), was pushed, v2026.10.14 was published, and the parent was updated at [540bf0c](https://github.com/songloft-org/songloft/commit/540bf0c76f87f08ba54dec1573cf9922126ad68c). Delivery omitted a prepared, visible Issue reply draft. Only after the user asked was the reply explicitly arranged, causing extra exchanges.
- Fix: the current uncommitted workflow update requires AGENTS, AI workflow, and dev-flow to proactively show the complete draft, target Issue, and specific comment approval action after pushing; existing applicable comment authorization permits direct posting. Update drafts as CI/release results change, with separate authorization for comments and closure.
- Done right: after the user explicitly requested the reply, the release update was [posted and read back](https://github.com/songloft-org/songloft/issues/495#issuecomment-6147071915), stating the available version, actual validation, and device/duration acceptance limits. Subsequent practical validation of the new “proactive draft → approval → posting” workflow remains pending; the late reply does not establish that complete flow.
- Seen 1× (2026-10-13). Status: incorporated immediately at the user's explicit request into [Linked Issue Replies](docs/en/ai-development.md#linked-issue-replies), without waiting for the occurrence threshold.

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
