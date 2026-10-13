#!/usr/bin/env node
// Read-only evidence collection. Conclusions and permission to act remain with the agent/user.
import { spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import {
  existsSync,
  lstatSync,
  readFileSync,
  readlinkSync,
  realpathSync,
  writeFileSync,
} from "node:fs";
import { dirname, isAbsolute, relative, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

function command(binary, args, cwd) {
  const result = spawnSync(binary, args, {
    cwd,
    encoding: "utf8",
    timeout: 60_000,
    maxBuffer: 32 * 1024 * 1024,
  });
  if (result.error) throw result.error;
  if (result.status !== 0)
    throw new Error(
      `${binary} failed (${result.status}): ${result.stderr.trim()}`,
    );
  return result.stdout;
}

function git(root, ...args) {
  return command("git", args, root);
}

export function githubRepo(url, parentRepo) {
  const match = url.match(
    /^(?:https:\/\/github\.com\/|(?:ssh:\/\/)?git@github\.com[:/])([^/]+)\/([^/?#]+?)(?:\.git)?\/?$/i,
  );
  if (match) return `${match[1]}/${match[2]}`;
  if (parentRepo && /^\.\.?\//.test(url)) {
    const parts = parentRepo.split("/");
    for (const part of url.replace(/\.git$/, "").split("/")) {
      if (part === "..") parts.pop();
      else if (part !== ".") parts.push(part);
    }
    if (parts.length === 2 && parts.every(Boolean)) return parts.join("/");
  }
  return null;
}

function localCheckout(path) {
  try {
    return (
      realpathSync(git(path, "rev-parse", "--show-toplevel").trim()) ===
      realpathSync(path)
    );
  } catch {
    return false;
  }
}

function modules(path) {
  const manifest = resolve(path, ".gitmodules");
  if (!existsSync(manifest)) return [];
  const result = spawnSync(
    "git",
    [
      "config",
      "--null",
      "--file",
      manifest,
      "--get-regexp",
      "^submodule\..*\.(path|url)$",
    ],
    { encoding: "utf8" },
  );
  if (result.error) throw result.error;
  if (result.status === 1 && !result.stdout) return [];
  if (result.status !== 0)
    throw new Error(`Cannot parse ${manifest}: ${result.stderr.trim()}`);
  const entries = new Map();
  for (const item of result.stdout.split("\0").filter(Boolean)) {
    const split = item.indexOf("\n");
    const key = item.slice(0, split).match(/^submodule\.(.*)\.(path|url)$/);
    if (!key) continue;
    const entry = entries.get(key[1]) ?? {};
    entry[key[2]] = item.slice(split + 1);
    entries.set(key[1], entry);
  }
  return [...entries.values()];
}

export function discover(root) {
  root = realpathSync(root);
  const repositories = [];
  const discoveryErrors = [];
  let origin;
  try {
    origin = git(root, "remote", "get-url", "origin").trim();
  } catch {
    origin = "";
  }
  const mainRepo = githubRepo(origin);
  if (!mainRepo)
    discoveryErrors.push("Root origin is not a supported GitHub URL.");
  const visited = new Set();
  function visit(path, repo, pinnedSha, parent) {
    const pathLabel = relative(root, path) || ".";
    const initialized = localCheckout(path);
    const record = {
      repo,
      path: pathLabel,
      parent,
      pinnedSha,
      initialized,
      head: null,
    };
    repositories.push(record);
    if (!repo)
      discoveryErrors.push(`${pathLabel}: unsupported repository URL.`);
    if (!initialized) {
      discoveryErrors.push(
        `${pathLabel}: not initialized; its Issues can be scanned, but nested submodules cannot be discovered locally.`,
      );
      return;
    }
    const realPath = realpathSync(path);
    if (visited.has(realPath)) {
      discoveryErrors.push(`${pathLabel}: repeated checkout path.`);
      return;
    }
    visited.add(realPath);
    try {
      record.head = git(path, "rev-parse", "HEAD").trim();
      for (const entry of modules(path)) {
        if (!entry.path || !entry.url) {
          discoveryErrors.push(
            `${pathLabel}: submodule is missing path or URL.`,
          );
          continue;
        }
        const childPath = resolve(path, entry.path);
        const rel = relative(path, childPath);
        if (
          isAbsolute(entry.path) ||
          !rel ||
          rel === ".." ||
          rel.startsWith(`..${sep}`) ||
          (existsSync(childPath) &&
            !realpathSync(childPath).startsWith(`${realPath}${sep}`))
        ) {
          discoveryErrors.push(
            `${pathLabel}: submodule path escapes its checkout: ${entry.path}`,
          );
          continue;
        }
        const tree = git(path, "ls-tree", "HEAD", "--", entry.path);
        const pinned = tree.match(/^160000 commit ([0-9a-f]+)\t/);
        if (!pinned)
          discoveryErrors.push(
            `${relative(root, childPath)}: no committed gitlink; pinned SHA is unknown.`,
          );
        visit(
          childPath,
          githubRepo(entry.url, repo),
          pinned?.[1] ?? null,
          repo,
        );
      }
    } catch (error) {
      discoveryErrors.push(`${pathLabel}: ${error.message}`);
    }
  }
  visit(root, mainRepo, null, null);
  return { root, repositories, discoveryErrors };
}

function hash(value) {
  return createHash("sha256").update(value).digest("hex");
}

export function snapshot(path) {
  const head = git(path, "rev-parse", "HEAD").trim();
  const trackedDiffSha256 = hash(
    git(
      path,
      "diff",
      "--binary",
      "--no-ext-diff",
      "--no-textconv",
      "HEAD",
      "--",
    ),
  );
  const stagedDiffSha256 = hash(
    git(
      path,
      "diff",
      "--cached",
      "--binary",
      "--no-ext-diff",
      "--no-textconv",
      "--",
    ),
  );
  const untrackedFiles = git(
    path,
    "ls-files",
    "--others",
    "--exclude-standard",
    "-z",
  )
    .split("\0")
    .filter(Boolean)
    .map((file) => {
      const fullPath = resolve(path, file);
      const stat = lstatSync(fullPath);
      const content = stat.isSymbolicLink()
        ? readlinkSync(fullPath)
        : readFileSync(fullPath);
      return { path: file, mode: stat.mode, sha256: hash(content) };
    });
  return {
    head,
    status: git(path, "status", "--short"),
    trackedDiffSha256,
    stagedDiffSha256,
    untrackedFiles,
  };
}

export function githubApi(endpoint, paginate = false) {
  const args = ["api", "--method", "GET", endpoint];
  if (paginate) args.push("--paginate", "--slurp");
  return JSON.parse(command("gh", args));
}

export function collectIssues(inventory, api = githubApi) {
  const grouped = new Map();
  for (const record of inventory.repositories) {
    if (!record.repo) continue;
    const key = record.repo.toLowerCase();
    if (!grouped.has(key))
      grouped.set(key, {
        repo: record.repo,
        checkouts: [],
        status: "not_checked",
        issues: [],
      });
    grouped.get(key).checkouts.push(record);
  }
  const repositories = [...grouped.values()];
  for (const record of repositories) {
    try {
      const metadata = api(`repos/${record.repo}`);
      record.defaultBranch = metadata.default_branch;
      record.remoteHead = api(
        `repos/${record.repo}/commits/${encodeURIComponent(metadata.default_branch)}`,
      ).sha;
      if (!record.remoteHead)
        throw new Error("Remote default-branch SHA is missing.");
      if (metadata.has_issues === false) {
        record.status = "issues_disabled";
        continue;
      }
      const pages = api(
        `repos/${record.repo}/issues?state=open&per_page=100`,
        true,
      );
      if (!Array.isArray(pages) || !pages.every(Array.isArray))
        throw new Error("Invalid paginated Issues response.");
      record.issues = pages
        .flat()
        .filter((issue) => !issue.pull_request)
        .map((issue) => ({
          ref: `${record.repo}#${issue.number}`,
          number: issue.number,
          title: issue.title,
          url: issue.html_url,
          updatedAt: issue.updated_at,
          labels: issue.labels.map((label) => label.name),
          assignees: issue.assignees.map((user) => user.login),
          assessment: "needs_triage",
        }));
      record.status = "scanned";
    } catch (error) {
      record.status = "not_checked";
      record.error = error.message;
    }
  }
  return {
    repositories,
    complete:
      inventory.discoveryErrors.length === 0 &&
      repositories.every((record) => record.status !== "not_checked"),
  };
}

export function collectCommits(inventory, since, until) {
  const repositories = inventory.repositories.map((record) => {
    const result = { ...record, status: "not_checked", commits: [] };
    if (!record.initialized) return result;
    try {
      const path = resolve(inventory.root, record.path);
      result.snapshot = snapshot(path);
      if (result.snapshot.head !== record.head)
        throw new Error("HEAD changed during collection; recollect evidence.");
      // Use committer dates and --since-as-filter so backdated commits do not truncate traversal.
      const log = git(
        path,
        "log",
        record.head,
        `--since-as-filter=${since}`,
        `--until=${until}`,
        "--format=%H%x00%cI%x00%B%x00",
      );
      const fields = log.split("\0");
      for (let i = 0; i + 2 < fields.length; i += 3) {
        result.commits.push({
          sha: fields[i].trim(),
          committedAt: fields[i + 1],
          message: fields[i + 2].trim(),
        });
      }
      result.status = "collected";
      result.historySource =
        "local checkout; remote freshness has not been verified";
    } catch (error) {
      result.error = error.message;
    }
    return result;
  });
  return {
    repositories,
    complete:
      inventory.discoveryErrors.length === 0 &&
      repositories.every((record) => record.status === "collected"),
  };
}

export function main(args = process.argv.slice(2)) {
  if (args.includes("--help")) {
    process.stdout.write(
      "Usage: node scripts/ai-review-context.mjs inventory|issues|commits [--root PATH] [--since ISO] [--until ISO] [--output FILE]\nRead-only; issues needs gh authentication. Exit 1 means incomplete coverage; JSON is still written.\n",
    );
    return 0;
  }
  const mode = args.shift();
  if (!["inventory", "issues", "commits"].includes(mode))
    throw new Error("Choose inventory, issues, or commits; see --help.");
  const options = {};
  while (args.length) {
    const key = args.shift();
    if (
      !["--root", "--since", "--until", "--output"].includes(key) ||
      !args.length
    )
      throw new Error(`Invalid argument: ${key}`);
    options[key.slice(2)] = args.shift();
  }
  const root = git(
    resolve(options.root ?? "."),
    "rev-parse",
    "--show-toplevel",
  ).trim();
  const inventory = discover(root);
  const generatedAt = new Date().toISOString();
  let report = {
    schemaVersion: 1,
    mode,
    generatedAt,
    ...inventory,
    complete: inventory.discoveryErrors.length === 0,
  };
  if (mode === "issues") report = { ...report, ...collectIssues(inventory) };
  if (mode === "commits") {
    const until = new Date(options.until ?? generatedAt);
    const since = new Date(options.since ?? until.getTime() - 86_400_000);
    if (
      !Number.isFinite(since.getTime()) ||
      !Number.isFinite(until.getTime()) ||
      since > until
    )
      throw new Error("Invalid time window.");
    report = {
      ...report,
      since: since.toISOString(),
      until: until.toISOString(),
      ...collectCommits(inventory, since.toISOString(), until.toISOString()),
    };
  }
  const output = `${JSON.stringify(report, null, 2)}\n`;
  if (options.output) {
    const target = resolve(options.output);
    const outputDir = realpathSync(dirname(target));
    const existing = existsSync(target) ? realpathSync(target) : target;
    if (
      existing === root ||
      existing.startsWith(`${root}${sep}`) ||
      outputDir === root ||
      outputDir.startsWith(`${root}${sep}`)
    ) {
      throw new Error(
        "Write evidence outside the worktree (for example /tmp), so collection cannot invalidate its own snapshot.",
      );
    }
    writeFileSync(target, output, { flag: "wx", mode: 0o600 });
    process.stderr.write(`Written ${target}; complete=${report.complete}\n`);
  } else process.stdout.write(output);
  return report.complete ? 0 : 1;
}

if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  try {
    process.exitCode = main();
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exitCode = 1;
  }
}
