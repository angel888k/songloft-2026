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
import { delimiter, dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
// Execute the actual workflow step, including its GITHUB_OUTPUT contract.
const workflow = readFileSync(
  join(root, ".github/workflows/release.yml"),
  "utf8",
);
const step = workflow.split("        id: prev-tag\n")[1].split("\n  # ===")[0];
const run = step
  .split("        run: |\n")[1]
  .split("\n")
  .map((line) => line.slice(10))
  .join("\n")
  .replaceAll("${{ steps.parse.outputs.version }}", "${CURRENT_VERSION}");

function fixture(t) {
  const cwd = mkdtempSync(join(tmpdir(), "songloft-previous-release-"));
  t.after(() => rmSync(cwd, { recursive: true, force: true }));
  const git = (...args) => {
    const result = spawnSync("git", args, { cwd, encoding: "utf8" });
    assert.equal(result.status, 0, result.stderr);
    return result.stdout.trim();
  };
  git("init", "-q", "-b", "main");
  git("config", "user.name", "Fixture");
  git("config", "user.email", "fixture@example.invalid");
  const commit = (tag) => {
    git("commit", "-q", "--allow-empty", "-m", `test: ${tag}`);
    git("tag", "-a", tag, "-m", tag);
    return git("rev-parse", "HEAD");
  };
  const first = commit("v2.13.1");
  const bin = join(cwd, "bin");
  mkdirSync(bin);
  writeFileSync(
    join(bin, "gh"),
    '#!/bin/sh\nprintf "%s\\n" "$@" > "$GH_ARGS"\nif [ "$GH_FAIL" = 1 ]; then echo "API unavailable" >&2; exit 1; fi\ncat "$RELEASE_FIXTURE"\n',
    { mode: 0o755 },
  );
  const release = (tag_name, extra = {}) => ({
    tag_name,
    draft: false,
    prerelease: false,
    published_at: "2026-10-15T06:51:49Z",
    ...extra,
  });
  const execute = (pages, current = "v2.13.3", fail = false) => {
    const output = join(cwd, "output");
    writeFileSync(output, "");
    writeFileSync(join(cwd, "releases.json"), JSON.stringify(pages));
    const result = spawnSync(
      "bash",
      [
        "-e",
        "-o",
        "pipefail",
        "-c",
        run.replace(
          "node scripts/previous-release.mjs",
          `node '${join(root, "scripts/previous-release.mjs")}'`,
        ),
      ],
      {
        cwd,
        encoding: "utf8",
        env: {
          ...process.env,
          PATH: `${bin}${delimiter}${process.env.PATH}`,
          CURRENT_VERSION: current,
          CURRENT_TAG: current,
          RELEASE_REPO: "fixture/repo",
          GITHUB_OUTPUT: output,
          RELEASE_FIXTURE: join(cwd, "releases.json"),
          GH_ARGS: join(cwd, "gh-args"),
          GH_FAIL: fail ? "1" : "0",
        },
      },
    );
    return {
      ...result,
      previous: readFileSync(output, "utf8").match(/^previous_tag=(.+)$/m)?.[1],
    };
  };
  return { cwd, git, commit, first, release, execute };
}

test("skips a failed release tag and includes its changes in the next stable release", (t) => {
  const f = fixture(t);
  f.commit("v2.13.2");
  f.commit("v2.13.3");
  const result = f.execute([[f.release("v2.13.1")]]);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.previous, "v2.13.1");
});

test("ignores the current release on reruns and all tags at HEAD in dev builds", (t) => {
  const f = fixture(t);
  f.commit("v2.13.2");
  f.git("tag", "v2.13.3");
  const pages = [
    [f.release("v2.13.3"), f.release("v2.13.2"), f.release("v2.13.1")],
  ];
  for (const current of ["v2.13.3", "dev"]) {
    const result = f.execute(pages, current);
    assert.equal(result.status, 0, result.stderr);
    assert.equal(result.previous, "v2.13.1");
  }
});

test("excludes drafts, prereleases, unpublished releases and non-version tags", (t) => {
  const f = fixture(t);
  for (const tag of [
    "v2.13.2",
    "v2.13.3",
    "v2.13.4",
    "v2.13.5-rc.1",
    "v-next",
    "v2.13.6",
  ])
    f.commit(tag);
  f.git("commit", "-q", "--allow-empty", "-m", "test: dev");
  const result = f.execute(
    [
      [
        f.release("v2.13.1"),
        f.release("v2.13.2", { draft: true }),
        f.release("v2.13.3", { prerelease: true }),
        f.release("v2.13.4", { published_at: null }),
        f.release("v2.13.5-rc.1"),
        f.release("v-next"),
      ],
    ],
    "dev",
  );
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.previous, "v2.13.1");
});

test("paginates releases and orders versions numerically, independently of API order", (t) => {
  const f = fixture(t);
  f.commit("v2.13.9");
  f.commit("v2.13.10");
  f.commit("v2.13.11");
  const result = f.execute(
    [[f.release("v2.13.9")], [f.release("v2.13.1"), f.release("v2.13.10")]],
    "v2.13.11",
  );
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.previous, "v2.13.10");
  const args = readFileSync(join(f.cwd, "gh-args"), "utf8");
  assert.match(args, /--paginate/);
  assert.match(args, /--slurp/);
  assert.match(args, /repos\/fixture\/repo\/releases\?per_page=100/);
});

test("excludes a published release outside the current commit history", (t) => {
  const f = fixture(t);
  f.git("checkout", "-q", "-b", "other");
  f.commit("v9.0.0");
  f.git("checkout", "-q", "main");
  f.commit("v2.13.3");
  const result = f.execute([[f.release("v9.0.0"), f.release("v2.13.1")]]);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.previous, "v2.13.1");
});

test("uses the first commit only when no eligible published release exists", (t) => {
  const f = fixture(t);
  f.commit("v2.13.3");
  const result = f.execute([[]]);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.previous, f.first);
});

test("API failure aborts instead of silently selecting a tag or first commit", (t) => {
  const f = fixture(t);
  f.commit("v2.13.3");
  const result = f.execute([], "v2.13.3", true);
  assert.notEqual(result.status, 0);
  assert.equal(result.previous, undefined);
});
