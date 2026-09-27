# RULES.md - Operational Guidelines for AI Agents & Maintainers

## 1. Product Direction & Core Philosophy

### Product Vision
Compressor is a dense, high-performance, and secure command-line zip archive engine. It prioritizes:
- **Terminal Ergonomics**: Inspired by the refined, dense design of `tw93/Mole`. Clean alignment, stable column widths, single-screen summaries, and optional drill-down logging.
- **Cross-Platform Parity**: Identical operational semantics and security guarantees across macOS, Linux, and Windows.
- **Zero-Compromise Security**: Defenses against Zip-Slip, symlink escapes, path traversal attacks, and corrupted partial writes.

### What the CLI Should Do
- Provide ultra-fast streaming compression and decompression of folders and files.
- Deliver granular compression tuning (levels 0 through 9).
- Accurately inspect archive headers, CRC32, sizes, and permissions without extraction overhead.
- Safeguard file systems by strictly enforcing boundary validations.
- Provide built-in benchmarking to evaluate compression speed vs ratio.

### What the CLI Should NOT Do
- Bloat the binary with heavy graphic engines or unnecessary runtime dependencies.
- Emit messy, unstructured, or unaligned text dumps to standard output.
- Silently overwrite files without `--force` or explicit user intent.
- Execute path operations without verifying canonical directory boundaries.

---

## 2. Product Decision Filter

When evaluating new features or pull requests, apply this filter:
1. **Is it terminal-native?** Does it adhere to short labels, predictable short-flags (`-o`, `-l`, `-e`, `-f`), and dense summaries?
2. **Is it cross-platform?** Does it run identically on macOS, Linux, and Windows path separators (`filepath.ToSlash` / `filepath.FromSlash`)?
3. **Is it safe by default?** Does it prevent path escapes, unsafe symlink targets, and memory bloat?
4. **Is it low-overhead?** Does it stream I/O rather than buffering entire archives in RAM?

---

## 3. Repository Map & Hotspot Ownership

```
.
├── cmd/               # CLI layer (Cobra command hierarchy, flags, UI invocations)
│   ├── root.go        # Global flags, root execution, and custom help templates
│   ├── pack.go        # Compression command (aliases: zip, compress, c, p)
│   ├── unpack.go      # Decompression command (aliases: unzip, extract, x, u)
│   ├── list.go        # Archive metadata inspector (aliases: ls, inspect, l)
│   ├── bench.go       # Multi-level compression performance benchmark
│   └── version.go     # Semantic version & environment reporting
├── internal/
│   ├── compress/      # Core streaming zip creation engine
│   ├── extract/       # Safe zip extraction engine
│   ├── safety/        # Path traversal & Zip-Slip security boundary checks (CRITICAL HOTSPOT)
│   ├── stats/         # Byte, ratio, duration, and speed calculations
│   └── ui/            # Dense terminal rendering, table formatting, colors, badges
├── pkg/types/         # Domain models, options, and summary structs
├── main.go            # Minimal binary entry point
├── Makefile           # Build, test, and cross-compilation targets
├── LICENSE            # MIT License
├── README.md          # User documentation
├── RULES.md           # Operational rules for agents & maintainers
├── SECURITY.md        # Security policies & vulnerability reporting
├── SECURITY_AUDIT.md  # Threat matrix & security audit verification
└── TRADEMARK.md       # Trademark & naming policy
```

### Hotspot Ownership & Critical Boundaries
- **`internal/safety/`**: The security gateway. Any modification to `ValidateDestinationPath` or `ValidateSymlinkTarget` must pass all tests in `internal/safety/safety_test.go` and `internal/extract/extract_test.go`. Never loosen path sanitization.
- **`internal/compress/` & `internal/extract/`**: Core I/O streaming paths. Must always clean temporary files on failure and never buffer unlimited bytes into memory.

---

## 4. Commands & Critical Safety Rules

1. **Path Boundary Validation**: Never create a file or directory without checking `safety.ValidateDestinationPath(destDir, entryName)`.
2. **Symlink Boundary Checks**: Never resolve a symlink to a target outside the extraction directory.
3. **Safe Overwrites**: Fail cleanly if a target file exists unless the `--force` (`-f`) flag is explicitly set.
4. **Exclusions**: Always honor glob patterns (e.g., `.git*`, `.DS_Store`, `node_modules`).
5. **No Color Mode**: Automatically respect `NO_COLOR` environment variable and `--no-color` flag.

---

## 5. Working Rules & Code Practices

- **SOLID Principles**: Maintain clean separation between CLI commands (`cmd`), core business logic (`internal/compress`, `internal/extract`), security checks (`internal/safety`), and presentation (`internal/ui`).
- **Standard Library First**: Utilize Go standard library (`archive/zip`, `compress/flate`, `path/filepath`, `io`, `os`) for core algorithms to minimize supply-chain surface.
- **Comprehensive Unit Tests**: Every feature and fix must include automated test coverage with `go test -v ./...`.
- **Zero Unchecked Errors**: Always handle or return errors explicitly.

---

## 6. GitHub Operations

### Triage & PR Verification
- **Re-read Live Context**: Re-read the live issue or PR title, body, comments, state, labels, and author language before any public reply or closeout.
- **Check Open PRs First**: Run `gh pr list --state open --search '<issue number or keyword>'` at the start of triage, alongside reading the code, not after a patch already exists. When a PR addresses the issue, the default path is review, maintainer-edit if needed, then squash merge; do not land an equivalent fix on main and close the contributor's PR as superseded. Self-fixing is for a stale, misdirected, or absent PR, and refusing a PR requires naming the mergeable alternative. When an issue's closeout is in scope, done means the PR is merged or properly declined, not just that main is pushed.
- **Small Mechanical Fixes on Contributor PRs**: Small mechanical fixes on a contributor PR belong on the contributor's branch, not in a review comment. Once the maintainer has authorized the merge, confirm `gh pr view <num> --json maintainerCanModify`, run `gh pr checkout <num>`, make the change, and commit with the contributor as author (`git -c user.name=... -c user.email=... commit`, taking the address from `gh pr view <num> --json author` or the branch's `git log --format=%ae`) so the squash carries no maintainer `Co-authored-by`. Push with a bare `git push`: `gh pr checkout` sets `branch.<name>.pushRemote` to the fork, while `git push origin HEAD` lands an unrelated branch on this repo and the PR never sees the change. When `maintainerCanModify` is false, or the change is a design choice, a refactor, or work only the author can test, it stays a review comment.
- **CI Traps on Contributor PRs**: `gh pr ready` starts nothing, because every pull_request trigger here uses the default `[opened, synchronize, reopened]` types and excludes `ready_for_review`, so close and reopen the PR instead; and a first-time contributor's runs sit at `action_required` until approved through `repos/<owner>/<repo>/actions/runs/<id>/approve`.

### Mole Mac App Separation Rules
- **Keep Repositories Separate**: Keep CLI issues and Mole Mac app issues separate. A fix in `mole-mac` does not imply a close in this CLI repo, and a CLI fix does not prove a Mac app issue is fixed unless the Mac app release path is verified.
- **Closing Bugs/Features**: When closing a fixed bug or shipped feature, use project wording from the issue context and include the expected release path only when confirmed.
- **Mole Mac Invitation Etiquette**: Leave it off by default. A resolved defect reply is complete without it, and a paid product appended to a bug answer reads as a pitch the reporter did not ask for. Add it only when the thread itself supplies the reason: the reporter said the CLI was hard to use, asked for something that is Mole Mac's job rather than the CLI's, or is plainly not a terminal user. Never add it when the reply corrects the reporter's own misreading, when the reporter contributed the fix, when the thread already concerns Mole Mac, or on PR thank-you notes, feature requests, and questions. When it does belong, it is one final sentence kept separate from the resolution facts:
  - Chinese: `也欢迎试试我的 Mole Mac：https://mole.fit/，更易用，也更精致。`
  - English: `You’re also welcome to try my Mole Mac app at https://mole.fit/ for a more polished, easier-to-use experience.`
  - When in doubt, leave it out and close the reply on the reporter's next step.
- **Discussion Content Cleanup**: When the maintainer classifies a Discussion as cleanup-only, such as spam, an empty or accidental post, duplicate promotion, or obsolete housekeeping with no technical answer needed, close it directly without replying. Do not apply this shortcut to substantive bug reports, Q&A, feature requests, or not-planned product decisions; those still need a concise disposition before closure.
- **Remote Diagnostics for Unreproducible Reports**: For Mole Mac reports, give the reporter exactly one command to paste into Terminal: `curl -fsSL 'https://mole.fit/downloads/Mole-Diagnose.command' | bash`. Never tell them to download, open, or double-click the file. When it finishes they email the `Mole-Diagnose-*.zip` it leaves on the Desktop to `hi@mole.fit`; never ask them to attach the archive publicly because it contains local paths and logs. For CLI-only issues, prefer the relevant command output or JSON status.

### Issue Closeout Pipeline
- **Default Pipeline**: Once a fix is confirmed, commit lands on `main` (that alone makes it installable via nightly), verify the fix is actually on `main`, then reply in the reporter's language, opening with `@reporter`, in short paragraphs rather than one block, with the concrete update command: update command now, the next stable release only when that path is confirmed.
- **Maintainer Authorization**: Closing needs the maintainer's word, but that word covers the whole pipeline: "该回复回复，该关闭关闭" or an equivalent authorizes commit, reply, and close in one turn, so run them to the end instead of returning for a separate confirmation at each step. The closing comment should invite reopening if the problem persists.
- **Announcements**: Announcements are a separate artifact from the changelog: one tweet above the fold with no line break, leading with what the tool does for the user and ending with the GitHub link; public copy about the CLI never positions it against the Mac app (the CLI is free and open source, Mole Mac is the polished paid path, both appear together); WeChat is opt-in for release announcements, never included by default.

---

## 7. Release Operations & Tag Integrity

### Critical Tag-Driven Rules
- **NEVER Rewrite History Reachable by Published V* Tags**: Never rewrite history that a published `V*` tag can reach, in any editor or agent. Every descendant commit gets a new SHA, tags follow the rebuilt commits, and GitHub writes the tag's commit SHA into `archive/refs/tags/<TAG>.tar.gz` as a pax global header, so the tarball's checksum changes while every file stays byte for byte identical. Anything that pinned a hash derived from that tarball is then silently wrong, starting with package formula managers (e.g. Homebrew).
- **Checksum Invalidation**: Before any `filter-branch`, `filter-repo`, or history-rewriting rebase, list the published tags the rewritten range reaches and treat every downstream checksum derived from them as invalidated; if that list is not empty, do not rewrite.
- **AI Trailers**: Strip AI co-author trailers when merging a pull request, never retroactively.
- **Tag Pushes**: Use tag-driven flow via `release.yml` on capital-V tag pushes (`vX.Y.Z`). Restate which distribution channels a release-flavored run will touch and confirm with the maintainer before acting; channel scope is specified by the maintainer, never inferred.

---

## 8. Issue Version Preflight

- **Version Verification**: Before asking a reporter to submit a new issue or repeat a reproduction, have them update to the latest stable release, rerun the CLI, and verify the actual running version (`compressor version`).
- **Feature Requests**: For feature requests, first check whether the latest stable release already provides it.
- **Evidence Verification**: Compare reported versions with the live release; a checked "latest" box or an old preview test is not version evidence.
- **Update Failures**: Reports about an update failure must remain possible: collect the installed version and the update error instead. Apply this to every issue intake template and support reply; never require Preview for ordinary reports or dismiss a privacy/security report because the app is old.
