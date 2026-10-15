#!/usr/bin/env node
import { execFileSync } from "node:child_process";

function command(binary, args) {
  return execFileSync(binary, args, {
    encoding: "utf8",
    timeout: 60_000,
    maxBuffer: 32 * 1024 * 1024,
    stdio: ["ignore", "pipe", "pipe"],
  }).trim();
}

try {
  const [currentTag, repository] = process.argv.slice(2);
  if (!currentTag || !/^[\w.-]+\/[\w.-]+$/.test(repository ?? "")) {
    throw new Error(
      "Usage: node scripts/previous-release.mjs <current-tag|dev> <owner/repo>",
    );
  }
  // A tag can survive a failed build. Only published stable Releases are baselines.
  // Fetch every page; API failures must abort instead of silently losing history.
  const pages = JSON.parse(
    command("gh", [
      "api",
      "--paginate",
      "--slurp",
      `repos/${repository}/releases?per_page=100`,
    ]),
  );
  const published = new Set(
    pages
      .flat()
      .filter(
        (release) =>
          !release.draft &&
          !release.prerelease &&
          release.published_at &&
          /^v\d+\.\d+\.\d+$/.test(release.tag_name),
      )
      .map((release) => release.tag_name),
  );
  const head = command("git", ["rev-parse", "HEAD"]);
  const tags = command("git", [
    "tag",
    "--sort=-version:refname",
    "--merged",
    "HEAD",
  ]).split("\n");
  const previous = tags.find(
    (tag) =>
      tag !== currentTag &&
      published.has(tag) &&
      command("git", ["rev-parse", `refs/tags/${tag}^{commit}`]) !== head,
  );
  // No prior stable release: retain the existing first-release fallback.
  console.log(
    previous ??
      command("git", ["rev-list", "--max-parents=0", "--first-parent", "HEAD"]),
  );
} catch (error) {
  console.error(error.message);
  process.exitCode = 1;
}
