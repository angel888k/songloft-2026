import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import {
  mkdtempSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import {
  collectCommits,
  collectIssues,
  discover,
  githubRepo,
  main,
  snapshot,
} from "../ai-review-context.mjs";

function git(path, ...args) {
  const result = spawnSync("git", args, { cwd: path, encoding: "utf8" });
  assert.equal(result.status, 0, result.stderr);
  return result.stdout.trim();
}

function repo(path, remote) {
  mkdirSync(path, { recursive: true });
  git(path, "init", "-b", "main");
  git(path, "config", "user.name", "Fixture");
  git(path, "config", "user.email", "fixture@example.invalid");
  git(path, "remote", "add", "origin", remote);
  writeFileSync(join(path, "tracked.txt"), "original\n");
  git(path, "add", "tracked.txt");
  git(path, "commit", "-m", "fixture: initial");
  return git(path, "rev-parse", "HEAD");
}

function fixture(t) {
  const directory = mkdtempSync(join(tmpdir(), "songloft-ai-context-"));
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  const root = join(directory, "root");
  repo(root, "git@github.com:example/main.git");
  const child = join(root, "child");
  const childHead = repo(child, "https://github.com/example/client.git");
  const nested = join(child, "nested");
  const nestedHead = repo(nested, "https://github.com/example/nested.git");
  writeFileSync(
    join(child, ".gitmodules"),
    '[submodule "nested"]\n path = nested\n url = ../nested.git\n',
  );
  git(child, "add", ".gitmodules");
  git(
    child,
    "update-index",
    "--add",
    "--cacheinfo",
    `160000,${nestedHead},nested`,
  );
  git(child, "commit", "-m", "fixture: nested module");
  writeFileSync(
    join(root, ".gitmodules"),
    '[submodule "child"]\n path = child\n url = https://github.com/example/client.git\n[submodule "missing"]\n path = missing\n url = git@github.com:example/missing.git\n[submodule "duplicate"]\n path = duplicate\n url = https://github.com/example/client.git\n',
  );
  git(root, "add", ".gitmodules");
  for (const path of ["child", "missing", "duplicate"])
    git(
      root,
      "update-index",
      "--add",
      "--cacheinfo",
      `160000,${childHead},${path}`,
    );
  git(root, "commit", "-m", "fixture: modules");
  return { root, child, nested, directory };
}

test("discovers nested and uninitialized modules, preserving committed pointers separately from checkout HEAD", (t) => {
  const { root, child } = fixture(t);
  const inventory = discover(root);
  assert.deepEqual(
    inventory.repositories.map((r) => r.repo),
    [
      "example/main",
      "example/client",
      "example/nested",
      "example/missing",
      "example/client",
    ],
  );
  const client = inventory.repositories.find((r) => r.path === "child");
  assert.notEqual(client.head, client.pinnedSha);
  assert.equal(client.head, git(child, "rev-parse", "HEAD"));
  assert.equal(
    inventory.repositories.find((r) => r.path === "missing").initialized,
    false,
  );
  assert.equal(inventory.discoveryErrors.length, 2);
});

test("normalizes GitHub SSH/HTTPS and relative URLs, rejecting other hosts", () => {
  for (const url of [
    "https://github.com/owner/repo.git",
    "git@github.com:owner/repo.git",
    "ssh://git@github.com/owner/repo.git",
  ])
    assert.equal(githubRepo(url), "owner/repo");
  assert.equal(githubRepo("../other.git", "owner/parent"), "owner/other");
  assert.equal(
    githubRepo("../../other/repo.git", "owner/parent"),
    "other/repo",
  );
  assert.equal(githubRepo("https://other.example/owner/repo.git"), null);
});

test("paginates every unique repo, excludes PRs, retains assigned issues, and continues after failures", (t) => {
  const inventory = discover(fixture(t).root);
  const calls = [];
  const api = (endpoint, paginated) => {
    calls.push([endpoint, paginated]);
    if (endpoint === "repos/example/missing") throw new Error("HTTP 403");
    if (endpoint.endsWith("/commits/main")) return { sha: "remote-tip" };
    if (!endpoint.includes("/issues?"))
      return { default_branch: "main", has_issues: true };
    assert.equal(paginated, true);
    const issue = (number) => ({
      number,
      title: `Issue ${number}`,
      html_url: `https://github.com/example/main/issues/${number}`,
      updated_at: "2026-10-13T00:00:00Z",
      labels: [],
      assignees: [{ login: "owner" }],
    });
    return [
      Array.from({ length: 100 }, (_, i) => issue(i + 1)),
      [issue(101), { pull_request: {} }],
    ];
  };
  const report = collectIssues(inventory, api);
  assert.equal(report.repositories.length, 4);
  assert.equal(
    calls.filter(([endpoint]) => endpoint === "repos/example/client").length,
    1,
  );
  assert.equal(report.repositories[0].issues.length, 101);
  assert.deepEqual(report.repositories[0].issues[100].assignees, ["owner"]);
  assert.equal(report.repositories[0].issues[0].ref, "example/main#1");
  assert.equal(report.repositories[0].issues[0].assessment, "needs_triage");
  assert.equal(report.repositories[3].status, "not_checked");
  assert.equal(report.complete, false);
});

test("distinguishes Issues disabled from an empty successfully scanned queue", () => {
  const inventory = {
    repositories: [{ repo: "example/disabled" }, { repo: "example/empty" }],
    discoveryErrors: [],
  };
  const api = (endpoint) => {
    if (endpoint.endsWith("/commits/main")) return { sha: "tip" };
    if (endpoint.includes("/issues?")) return [[]];
    return {
      default_branch: "main",
      has_issues: !endpoint.endsWith("/disabled"),
    };
  };
  const report = collectIssues(inventory, api);
  assert.equal(report.complete, true);
  assert.deepEqual(
    report.repositories.map((r) => r.status),
    ["issues_disabled", "scanned"],
  );
});

test("evidence fingerprints change for tracked edits, staging, and untracked content", (t) => {
  const { root } = fixture(t);
  const before = snapshot(root);
  writeFileSync(join(root, "tracked.txt"), "modified\n");
  const edited = snapshot(root);
  assert.notEqual(edited.trackedDiffSha256, before.trackedDiffSha256);
  git(root, "add", "tracked.txt");
  assert.equal(snapshot(root).trackedDiffSha256, edited.trackedDiffSha256);
  assert.notEqual(snapshot(root).stagedDiffSha256, before.stagedDiffSha256);
  writeFileSync(join(root, "tracked.txt"), "original\n");
  const stagedOnly = snapshot(root);
  assert.equal(stagedOnly.trackedDiffSha256, before.trackedDiffSha256);
  assert.notEqual(
    stagedOnly.stagedDiffSha256,
    before.stagedDiffSha256,
    "staged changes remain identifiable even if working files match HEAD",
  );
  writeFileSync(join(root, "new.txt"), "new\n");
  const added = snapshot(root);
  writeFileSync(join(root, "new.txt"), "changed\n");
  assert.notEqual(
    snapshot(root).untrackedFiles.find((r) => r.path === "new.txt").sha256,
    added.untrackedFiles.find((r) => r.path === "new.txt").sha256,
  );
});

test("collects commit bodies per checkout and explicitly reports missing histories", (t) => {
  const inventory = discover(fixture(t).root);
  const report = collectCommits(
    inventory,
    "2000-01-01T00:00:00Z",
    "2099-01-01T00:00:00Z",
  );
  assert.equal(report.complete, false);
  assert.equal(report.repositories[0].commits.length, 2);
  assert.equal(
    report.repositories[0].commits[0].sha,
    inventory.repositories[0].head,
  );
  assert.equal(report.repositories[0].commits[0].message, "fixture: modules");
  assert.equal(
    report.repositories.find((r) => r.path === "missing").status,
    "not_checked",
  );
});

test("refuses stale commit evidence when checkout HEAD changes after discovery", (t) => {
  const { root } = fixture(t);
  const inventory = discover(root);
  git(
    root,
    "commit",
    "--allow-empty",
    "-m",
    "fixture: changed after discovery",
  );
  const report = collectCommits(
    inventory,
    "2000-01-01T00:00:00Z",
    "2099-01-01T00:00:00Z",
  );
  assert.equal(report.repositories[0].status, "not_checked");
  assert.match(report.repositories[0].error, /HEAD changed/);
  assert.equal(report.complete, false);
});

test("writes incomplete reports before exiting nonzero and refuses overwrite or worktree outputs", (t) => {
  const { root, directory } = fixture(t);
  const output = join(directory, "report.json");
  assert.equal(main(["inventory", "--root", root, "--output", output]), 1);
  assert.equal(JSON.parse(readFileSync(output)).repositories.length, 5);
  assert.throws(
    () => main(["inventory", "--root", root, "--output", output]),
    /EEXIST/,
  );
  assert.throws(
    () =>
      main([
        "inventory",
        "--root",
        root,
        "--output",
        join(root, "report.json"),
      ]),
    /outside the worktree/,
  );
  assert.throws(
    () => main(["commits", "--root", root, "--since", "wrong"]),
    /Invalid time window/,
  );
});
