---
name: project-release-tagging
description: "Use when creating, checking, or pushing release tags for this repository, especially Go module release tags."
---
# Project Release Tagging

## Purpose

Ensure published releases are consumable by all SDK users, including Go users
whose module lives under the `go/` subdirectory.

For the complete release process, use
[project-release](../project-release/SKILL.md). This skill supplies focused
package and tag rules; `RELEASING.md` maintains the CI and maintainer commands.

## Scope

Use this skill when:

- creating a release tag;
- checking whether a release tag exists locally or remotely;
- pushing a release tag;
- preparing a release that should be consumed by `go get`.

## Mandatory Knowledge

- This repository has a root release tag, for example `v0.2.0`.
- The Go module path is declared in `go/go.mod` as
  `github.com/netdata/systemd-journal-sdk/go`.
- Because the Go module is in the repository subdirectory `go/`, each Go
  module release must also have a tag prefixed with that subdirectory, for
  example `go/v0.2.0`.
- The root release tag and the Go submodule release tag must point to the same
  commit unless a SOW records a deliberate split release decision.
- Rust crates.io publication uses project-prefixed package names. The public
  SDK package is `systemd-journal-sdk`; lower-level packages are
  `systemd-journal-sdk-common`, `systemd-journal-sdk-registry`,
  `systemd-journal-sdk-core`, `systemd-journal-sdk-host`,
  `systemd-journal-sdk-log-writer`, `systemd-journal-sdk-index`, and
  `systemd-journal-sdk-engine`.
- Rust crates must be publish-dry-run and published in dependency order:
  common, registry, core, host, log-writer, index, engine, public SDK. If a
  SOW changes dependencies, update the order in that SOW and this skill.
- Go's module reference defines this rule: if a module is defined in a
  repository subdirectory, the semantic version tag name is prefixed with the
  module subdirectory followed by `/`.
- Once pushed, release tags must not be moved or deleted without explicit user
  approval. Go module proxies and checksum databases may cache the old tag
  target.

## CI Release Workflow

- Prefer the canonical repository's manual `Release` workflow for Rust
  publication after preparation has merged. Read `RELEASING.md` for setup and
  execution. The workflow file is `.github/workflows/release.yml`; helper and
  recovery tests live under `tests/release/`.
- The user explicitly starts a run on master with version `X.Y.Z` and the full
  merged release source SHA. Workflow code and release source are separate
  checkouts; the source may precede the workflow's merge commit.
- Configure the GitHub `release` environment for exactly the master branch,
  and configure all eight crates for Trusted Publishing with owner `netdata`,
  repository `systemd-journal-sdk`, workflow `release.yml`, environment `release`.
- CI checks both language suites at source-declared compiler minimums before
  obtaining publishing permission. Each new crate is dry-run, then receives
  fresh temporary permission and is uploaded in the established order.
- API/index availability, non-yanked status, archive checksum and clean source
  SHA must agree before a published version can be skipped. Preserve the
  original version/source inputs when resuming a partial release.
- All eight Rust crates and an exact-version registry consumer pass before
  a maintainer pushes paired annotated tags using existing GitHub access.
  Copy the exact source from the successful CI summary, create missing tags
  there, and push the new pair atomically to the canonical repository.
  Correct existing siblings are preserved; conflicting tags stop the release.
- Any maintainer with repository write access can start the CI workflow.
  CI uses crates.io Trusted Publishing; maintainers with tag-push permission
  own GitHub tag creation. No additional GitHub App or account token is
  configured for this workflow.
- After the tag push, run the guide's `verify-go` helper command to verify
  both annotated targets, matching Rust archive sources and a fresh
  exact-version Go consumer. It reads the
  selected source's Go minimum, so verification supports historical source
  while the working branch has advanced. The helper does not change tags.
- Actual version publication belongs to its execution SOW. Completing the
  automation implementation SOW does not mean a version has been published.
- A new crate that does not already exist needs an explicitly authorized
  initial publication before Trusted Publishing can be configured for it.

## Manual Fallback Workflow

1. Confirm the intended Rust crate and Go module versions. They are normally
   the same, for example `0.2.0`, but may differ when a SOW records a
   Rust-only or Go-only release.
2. Update and verify every active consumer install example for the intended
   version. Scan the active documentation and specification roots
   dynamically; do not limit the scan to a fixed file list or only the
   immediately previous version:

   ```bash
   grep -RInE \
     'systemd-journal-sdk.*version = "[0-9]+\.[0-9]+\.[0-9]+"|systemd-journal-sdk/go@v[0-9]+\.[0-9]+\.[0-9]+' \
     --include='*.md' --include='*.toml' \
     README.md rust/README.md docs .agents/sow/specs
   ```

   Verify every returned Rust example identifies the intended Rust crate
   version and every Go example identifies the intended Go module version.
3. Verify the worktree is clean:

   ```bash
   git status --short --branch
   ```

4. Verify the Go module path:

   ```bash
   sed -n '1,20p' go/go.mod
   ```

5. Check local and remote tags before creating anything:

   ```bash
   git tag -l 'v0.2.0' 'go/v0.2.0'
   git ls-remote --tags origin refs/tags/v0.2.0 refs/tags/v0.2.0^{} refs/tags/go/v0.2.0 refs/tags/go/v0.2.0^{}
   ```

6. If either tag exists at a different commit, stop and ask the user. Do not
   move, delete, force-push, or recreate tags without explicit approval.
7. Create annotated tags on the intended commit:

   ```bash
   git tag -a v0.2.0 <commit> -m 'v0.2.0'
   git tag -a go/v0.2.0 <commit> -m 'go/v0.2.0'
   ```

8. Push the branch first, then push both tags:

   ```bash
   git push origin <branch>
   git push origin v0.2.0 go/v0.2.0
   ```

9. Verify remote tag targets:

   ```bash
   git ls-remote --tags origin refs/tags/v0.2.0 refs/tags/v0.2.0^{} refs/tags/go/v0.2.0 refs/tags/go/v0.2.0^{}
   ```

10. Report the peeled tag commit hashes to the user.

## Rust crates.io Workflow

Before publishing Rust crates:

1. Verify `rust/Cargo.toml` workspace package version matches the intended
   release.
2. Check public API compatibility since the previous release. Public Rust
   struct field additions are source-breaking for downstream exhaustive struct
   literals unless the struct is already `#[non_exhaustive]`; public method
   additions are normally additive. Record the semver decision in the release
   SOW before tagging.
3. Verify publishable internal dependencies include both `package = ...` and
   `version = ...` so Cargo can replace path dependencies with registry
   dependencies.
4. Run `cargo publish --dry-run` for each publishable package in dependency
   order:

   ```bash
   cargo publish --manifest-path rust/src/crates/journal-common/Cargo.toml --dry-run
   cargo publish --manifest-path rust/src/crates/journal-registry/Cargo.toml --dry-run
   cargo publish --manifest-path rust/src/crates/journal-core/Cargo.toml --dry-run
   cargo publish --manifest-path rust/src/crates/journal-host/Cargo.toml --dry-run
   cargo publish --manifest-path rust/src/crates/journal-log-writer/Cargo.toml --dry-run
   cargo publish --manifest-path rust/src/crates/journal-index/Cargo.toml --dry-run
   cargo publish --manifest-path rust/src/crates/journal-engine/Cargo.toml --dry-run
   cargo publish --manifest-path rust/src/journal/Cargo.toml --dry-run
   ```

5. Cargo verifies publishable path dependencies against crates.io. For a new
   release version, dependent crate dry-runs may fail until the previous crate
   in dependency order has been published and indexed. In that case, dry-run
   and publish one crate at a time in dependency order: dry-run `common`,
   publish `common`, then dry-run `registry`, publish `registry`, and continue.
6. Publish in the same dependency order only after the package's dry-run passes
   and the SOW review gate is satisfied.
7. Never record crates.io tokens or credential details in durable artifacts.

## Validation Checklist

- Root tag exists locally and remotely when the release includes a root tag.
- Go submodule tag exists locally and remotely when the release includes a Go
  module tag.
- Both peeled tag targets are the same commit when both tags are included.
- The branch containing that commit is pushed.
- Every active consumer install example identifies the intended Rust crate or
  Go module version.
- `git status --short --branch` is clean after release work.
- Rust crates.io dry-runs pass for all publishable Rust packages when Rust
  packages are part of the release.
- Rust package publication is recorded in the SOW with package names and
  versions, without credential details.

## Evidence

- `go/go.mod`: Go module path.
- Go Modules Reference, "Mapping versions to commits":
  `https://go.dev/ref/mod#vcs-version`.
