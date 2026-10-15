import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const script = join(repoRoot, "scripts/bump-version.sh");

function git(cwd, ...args) {
  const result = spawnSync("git", args, { cwd, encoding: "utf8" });
  assert.equal(result.status, 0, `git ${args.join(" ")}: ${result.stderr}`);
  return result.stdout.trim();
}

function fixture(t, version = "1.2.3") {
  const directory = mkdtempSync(join(tmpdir(), "songloft-bump-parent-"));
  t.after(() => spawnSync("rm", ["-rf", directory], { timeout: 30000 }));
  const repo = join(directory, "repo");
  const remote = join(directory, "remote.git");
  mkdirSync(repo);
  git(directory, "init", "-q", "--bare", remote);
  git(repo, "init", "-q", "-b", "main");
  git(repo, "config", "user.name", "Fixture");
  git(repo, "config", "user.email", "fixture@example.invalid");
  writeFileSync(join(repo, "Makefile"), `VERSION ?= ${version}\n\nswagger:\n\t@echo VERSION=$(VERSION)\n`);
  writeFileSync(join(repo, "main.go"), `// @version ${version}\npackage main\n`);
  mkdirSync(join(repo, "docs"));
  writeFileSync(join(repo, "docs", "swagger.json"), `{"version":"${version}"}`);
  git(repo, "add", ".");
  git(repo, "commit", "-q", "-m", "test: baseline");
  git(repo, "remote", "add", "origin", remote);
  git(repo, "push", "-q", "-u", "origin", "main");
  git(repo, "fetch", "-q", "origin", "main");
  const run = (args) =>
    spawnSync("bash", [script, ...args], {
      cwd: repo,
      encoding: "utf8",
      env: { ...process.env, GIT_AUTHOR_NAME: "Fixture", GIT_AUTHOR_EMAIL: "f@e.invalid", GIT_COMMITTER_NAME: "Fixture", GIT_COMMITTER_EMAIL: "f@e.invalid" },
    });
  return { repo, remote, run, git: (...args) => git(repo, ...args) };
}

test("patch release commits Makefile/main.go/docs, tags, and atomically pushes main + tag", (t) => {
  const f = fixture(t);
  const result = f.run(["patch", "--yes"]);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout.trim().split("\n").pop(), "v1.2.4");
  assert.match(readFileSync(join(f.repo, "Makefile"), "utf8"), /VERSION \?= 1\.2\.4/);
  assert.match(readFileSync(join(f.repo, "main.go"), "utf8"), /@version 1\.2\.4/);
  assert.equal(f.git("rev-parse", "v1.2.4^{}"), f.git("rev-parse", "HEAD"));
  assert.equal(f.git("status", "--porcelain"), "");
  // 远程 main 与 tag 均已更新
  const pushed = spawnSync("git", ["--git-dir", f.remote, "show-ref"], { encoding: "utf8" });
  assert.match(pushed.stdout, /refs\/tags\/v1\.2\.4/);
  assert.match(pushed.stdout, /refs\/heads\/main/);
});

test("--no-push keeps commit and tag local; --dry-run writes nothing and prints the planned tag", (t) => {
  const f = fixture(t);
  const before = f.git("rev-parse", "HEAD");
  const dry = f.run(["minor", "--dry-run"]);
  assert.equal(dry.status, 0, dry.stderr);
  assert.equal(dry.stdout.trim().split("\n").pop(), "v1.3.0");
  assert.equal(f.git("rev-parse", "HEAD"), before);
  assert.equal(f.git("status", "--porcelain"), "");
  assert.equal(f.git("tag"), "");
  const local = f.run(["minor", "--yes", "--no-push"]);
  assert.equal(local.status, 0, local.stderr);
  assert.equal(f.git("rev-parse", "v1.3.0^{}"), f.git("rev-parse", "HEAD"));
  const pushed = spawnSync("git", ["--git-dir", f.remote, "show-ref"], { encoding: "utf8" });
  assert.doesNotMatch(pushed.stdout, /v1\.3\.0/);
});

test("release strips prerelease; explicit versions are validated and must be newer", (t) => {
  const f = fixture(t, "1.3.0-rc.2");
  const result = f.run(["release", "--yes", "--no-push"]);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout.trim().split("\n").pop(), "v1.3.0");
  assert.equal(f.run(["release", "--dry-run"]).status, 1); // 1.3.0 -> 1.3.0 不前进
  assert.equal(f.run(["1.2.0", "--dry-run"]).status, 1); // 低于当前
  assert.equal(f.run(["1.2", "--dry-run"]).status, 1); // 非法格式
  assert.equal(f.run(["1.4.0-rc.3", "--dry-run"]).stdout.trim().split("\n").pop(), "v1.4.0-rc.3");
});

test("preflight refuses dirty trees and existing local tags without sweeping changes", (t) => {
  const f = fixture(t);
  const before = f.git("rev-parse", "HEAD");
  writeFileSync(join(f.repo, "stray.txt"), "keep me\n");
  assert.equal(f.run(["patch", "--yes"]).status, 1);
  assert.equal(f.git("rev-parse", "HEAD"), before);
  assert.match(f.git("status", "--porcelain"), /stray\.txt/); // 未被裹带提交
  rmSync(join(f.repo, "stray.txt"));
  assert.equal(f.run(["patch", "--yes", "--no-push"]).status, 0);
  f.git("tag", "v1.2.5"); // 预置同名 tag，模拟已发布版本
  assert.equal(f.run(["patch", "--yes", "--no-push"]).status, 1);
  assert.equal(f.git("log", "-1", "--format=%s"), "chore: release version 1.2.4");
});
