# SOW-0145 - Rust Close Without Retention And v0.8.2

## Status

Status: completed

Sub-state: Rust API, package preparation, documentation, validation, review,
and post-merge publication completed. All eight Rust `0.8.2` crates are indexed
and non-yanked; the exact registry consumer passes on Rust `1.91.0`.

## Requirements

### Purpose

Provide Rust's equivalent of Go `Log.CloseWithoutRetention()` and prepare
publication of all eight Rust SDK crates at `0.8.2` through a new PR.

### User Request

The user requested another PR implementing the same call for Rust and Rust
`0.8.2` publication. Standing delivery instructions require completed SOWs
in the PR, with actual publication after merge.
The prior release instruction, "So the PR must have it done", explicitly
overrode pending publication tracking in SOW-0144. Apply that same narrow
close-out timing here: complete this implementation/preparation record in the
PR and retain authorized post-merge execution, without another pending SOW.

### Assistant Understanding

Facts at the analysis baseline:

- Go `v0.8.2` and root `v0.8.2` are published at
  `51bf47f1f90562f7b7717ca41aaaac74f75f90f0`.
- Rust source and all eight registry crates were at `0.8.1`.
- Rust `Log::close(self)` archived, synced according to configuration, then
  applied retention. No close-without-retention method existed.
- The public SDK re-exports the writer's `Log`, so an inherent method is
  directly available without another facade wrapper.
- Rust closes consume the writer; repeating close or retrying on the same
  object is not a supported Rust operation.

Inferences:

- The approved parity request means an additive consuming
  `Log::close_without_retention(self) -> Result<()>`, preserving the existing
  close and Drop contracts and skipping only close-time retention.
- Staggered Rust `0.8.2` publication uses this PR's later merged commit.
  Existing Git tags remain immutable; Rust-only registry publication does
  not require a new Git tag, as recorded by SOW-0140.

Unknowns:

- The merged Rust PR commit and crates.io propagation time become known
  after merge/publication; neither is assumed as pre-merge evidence.

### Acceptance Criteria

- Both close methods share the archive, sync, empty-file, resource-release,
  and error path; existing `close()` retains its policy behavior.
- Special close skips retention callbacks/deletion. The original one-second
  policy, append, 1.5-second wait, special close, and reopen without retention
  preserves readable entries in both naming modes.
- Older expired rotated archives distinguish special close from normal close.
  Empty/lazy cases, archive errors, and archive-sync policy are covered.
- Rust workspace/package requirements and active install examples use `0.8.2`;
  third-party dependency versions and Rust/Go compiler floors do not change.
- Local validation, independent review, scanner disposition, and artifact
  updates are recorded; completed SOW and implementation ship in one PR commit.
- The concrete authorized post-merge procedure requires all eight Rust crates
  published in dependency order and an exact registry consumer invoking the
  API before publication is reported successful. Existing Go/root tags stay
  fixed. Completed PR preparation does not claim publication has occurred.

## Analysis

Sources checked:

- Project instructions, pending/current SOWs, product scope, SOW-0142 through
  SOW-0144, SOW-0140, and release/compatibility/docs/orchestration skills.
- Rust writer close/Drop/hooks, facade re-exports, rotation/retention tests,
  manifests/lockfile, consumer guides, and repository-local Cargo caches.

Initial state at `51bf47f1f90562f7b7717ca41aaaac74f75f90f0`:

- `rust/src/crates/journal-log-writer/src/log/mod.rs:590` consumes the writer;
  line 627 applies retention after archiving. The existing helper is reused.
- No pending SOW owns this additive API; final v1 publication is separate.

Risks:

- Duplicated close bodies could diverge in durability/error behavior.
- Retention at writer open remains intentional; a new close API cannot undo
  correctly enforced earlier opening policies.
- Rust consumes the writer on errors, unlike Go's reusable pointer receiver;
  documentation must not promise Rust retries or repeated closure.
- Rust registry dependencies need matching `0.8.2` requirements; existing
  published Git tags must not be overwritten to match the later Rust commit.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:

- Rust close applies an old policy before callers can reopen with a new one.
  Go now exposes explicit close-time retention ownership; Rust needs parity.

Evidence reviewed:

- Existing Rust close, empty strict/lazy branches, Drop, retention hooks,
  public re-exports, Go API/tests, and the corrected Go final-review scenarios.
- All eight Rust versions and current Git tags were checked; Rust `1.91.0`
  is installed, and third-party dependencies are already cached locally.

Affected contracts and surfaces:

- Rust directory-writer API, tests, package versions/internal requirements,
  Rust lockfile, consumer Rust/shared docs, product scope, and release tracking.
- No Go implementation, journal format, identity, append/query hot path,
  cache schema, or host-service contract changes.

Existing patterns to reuse:

- Consuming Rust close API, private boolean close helper, synthetic identities,
  existing journal verification/retention hooks, and repository-local caches.

Risk and blast radius:

- One cold explicit-close branch; all existing archive operations stay shared.
  Runtime parsing/writing formats and hot paths are unchanged, so no throughput
  claim or unrelated benchmark/refactor is required.

Sensitive data handling plan:

- Use synthetic fixture identities/entries only. Keep raw scanner output,
  package logs, and credentials out of durable records; caches/logs stay under
  ignored `.local/`. Never inspect the host journal or modify services.

Implementation plan:

1. Add the consuming method using a shared close helper and preserve Drop.
2. Add meaningful policy-transition, normal-close control, callback, empty/lazy,
   error, and archive-sync tests; update Rust/shared docs and product scope.
3. Bump Rust workspace/internal package requirements to `0.8.2`, regenerate
   only local package lock entries, and validate on existing Rust `1.91.0`.
4. Finish review/scanner/artifact gates; complete this SOW in the PR, commit
   explicit paths, push the branch, and create the PR.
5. Publish Rust crates after merge and verify exact registry consumption.

Validation plan:

- Focused writer tests, complete Rust workspace tests, fmt/clippy as applicable,
  exact public SDK consumer, Rust wiki examples, wiki validation, SOW audit,
  whitespace, and package manifest/content checks.
- Reuse existing interoperability/live tests and verify synthetic output with
  installed stock journalctl; record actual version rather than claim v260
  certification from a different tool version.
- Check code-scanning/Codacy findings and record concrete dispositions.
- Dependency-ordered publish dry-runs may require preceding `0.8.2` crates to
  be indexed; run each final dry-run immediately before its authorized publish.

Artifact impact plan:

- AGENTS.md: existing workflow and guards already cover this additive API.
- Runtime project skills: current release/compatibility/docs workflows apply.
- Specs: record Rust close ownership and staggered `0.8.2` publication.
- End-user/operator docs: Rust install pins and Rust/shared writer guidance.
- End-user/operator skills: none are declared in the repository.
- SOW lifecycle: one active SOW; complete in this PR under the user's delivery
  instruction, with post-merge publication procedure in the completed record.
- SOW-status.md: synchronize canonical and root activation/completion entries.

Open-source reference evidence:

- `netdata/systemd-journal-sdk @ 51bf47f1f90562f7b7717ca41aaaac74f75f90f0`
  `rust/src/crates/journal-log-writer/src/log/mod.rs:590`,
  `rust/src/journal/src/lib.rs:66`, `rust/Cargo.toml:39`.
- External format redesign is unnecessary: existing archival code is retained.

Open decisions:

- The user approved Rust parity and version `0.8.2` through another PR. The
  consuming receiver follows Rust's existing public API; existing published
  tags stay fixed. No new product/design fork is needed for this surgical change.

## Implications And Decisions

1. Surgical parity: introduce the consuming Rust method and share close logic;
   retain normal retention, archive failure handling, and Drop behavior.
2. Publish all eight Rust workspace crates at `0.8.2` after this PR merges;
   preserve third-party dependency versions, compiler minimums, and Git tags.
   The patch-version decision is additive: one inherent method is added,
   existing public signatures and struct fields remain compatible.
3. Close this SOW inside the PR as instructed; retain post-merge publication
   procedure without another pending SOW or lifecycle-only commit.

## Plan

Implement API/tests, prepare package/docs metadata, validate and review the
complete change, close/commit/push the PR, then publish after merge.

## Implementation And Review Plan

Implementation:

- Primary implementation is sequential on `feat/rust-close-without-retention`.
  Read-only release investigation is delegated; no agent writes shared files.

Reviewers:

- Independent final review covers the complete locally validated API/release
  change. The user explicitly authorized the recommended external-review batch
  on 2026-10-03 after local validation. That authorization covers meaningful
  whole-SOW review rounds and material corrections in this conversation.

Repository boundary block for every external-reviewer prompt:

```text
CRITICAL REPOSITORY BOUNDARY:
- Do not make changes outside this repository for any reason.
- Repository path: current repository root.
- You may inspect external references read-only when the task requires it.
- Write, edit, delete, move, reset, checkout, install, generate, cache, or format nothing outside this repository.
- The only write exception outside the repository is /tmp.
- Prefer .local/ inside this repository for scratch work, generated temporary files, cloned references, logs, and working notes.
```

Failure handling:

- Investigate failed checks without expanding runtime scope or rewriting tags.
  Do not claim validation, review, publication, or a registry consumer passed
  before actual evidence exists.

## Execution Log

### 2026-10-03

- Confirmed clean master and merged Go release, existing immutable tags, Rust
  close/Drop semantics, package graph, cached minimum toolchain, and no overlap.
- Created the Rust parity branch and recorded the concrete gate before code.
- Implemented the shared close path, six focused tests, version/docs changes,
  and updated exactly 23 local lockfile versions; all third-party lock entries
  remain unchanged.
- The six focused tests pass on Rust `1.91.0`. An initial missing-directory
  error fixture did not trigger rename because the existing helper explicitly
  skips missing source paths. Replaced it with a directory occupying the actual
  archive destination; both close methods return the real rename error.
- The first in-progress SOW audit required a separately named sensitive-data
  validation gate. Added that gate; the corrected in-progress audit passed.
- Independent complete-change final review returned PASS. All six authorized
  external reviewers returned PRODUCTION GRADE with no blocking findings,
  no failed runs, and thirteen non-blocking observations across five classes.
- Restored normal `close()` in the canonical production example, retained
  separate special-close guidance, and reran that compiled/executed example:
  one selected, one passed. Wiki validation again passed all 15 pages.
- Completed this record and moved it with both status ledgers for the same
  implementation commit, as explicitly required for the PR.
- The completed-state audit passed with changed-artifact sensitive scanning;
  directory/status consistency and both ledgers agree, with no sensitive-data
  patterns in the 33 scanned durable artifacts.

## Validation

Acceptance and executable evidence:

- Rust `1.91.0` focused writer checks: six tests pass (one unit and five
  integration tests). The original one-second policy/1.5-second wait/reopen
  sequence passes in strict and chain naming. Five-second older-archive
  controls prove special close preserves archives/deletion callbacks while
  normal close deletes an expired older archive and reports deletion.
- Archive-sync checks cover both naming modes with sync enabled/disabled.
  Empty eager/lazy cases, failing retention-accounting hooks, and real strict
  archive rename errors are exercised; both methods preserve consuming-error
  ownership. No Rust repeat-close or retry claim is made.
- `cargo test --offline --locked --workspace --manifest-path rust/Cargo.toml
  --jobs 2 -- --test-threads=2`: 497 pass, zero fail. Five existing doctest
  snippets are explicitly marked `ignore`: one compatibility migration
  snippet, three guarded-cell illustrations, and one query-builder fragment.
- `cargo test --offline --locked --manifest-path rust/Cargo.toml -p
  systemd-journal-sdk-log-writer --features serde-api --lib --jobs 2 --
  --test-threads=2`: all 15 pass, including archive-sync and serde fixtures.
- The public SDK path consumer imports `journal::Log`, invokes the new method,
  and repeats the original retention-policy handoff in both naming modes.
  Original archive bytes are unchanged after eager reopening; native SDK
  verification and `FileReader` return the exact original message/one row.
  This is source-consumer evidence before merge, not registry-install evidence.
- Installed stock systemd `255.4-1ubuntu8.17` verifies every new nonempty
  close fixture with `journalctl --verify --file`; the policy-handoff fixture
  also returns its exact entry through stock `--directory --output=json`.
  This does not claim fresh systemd v260.1 certification or a new full live
  matrix: append/layout/compression/reader/live-publication paths are unchanged.
- `cargo fmt --manifest-path rust/Cargo.toml --all --check`, whitespace checks,
  wiki validation (15 pages), and all 31 executed wiki examples (14 Rust/17 Go)
  pass. After restoring normal `close()` in the general directory-writer
  example, that exact example was compiled/executed again and passed. The other
  30 examples are unchanged; the dedicated public SDK consumer calls the new API.
- All eight publishable manifests report `0.8.2`, Rust minimum `1.91`, and
  internal `^0.8.2` requirements. Package listings include the writer's
  `src/log/mod.rs`. Only 23 local lock entries change; third-party entries
  remain unchanged even after package/consumer validation.
- The common crate's locked publish dry-run packages and verifies successfully
  on Rust `1.91.0`. The initial offline attempt was rejected because Cargo
  requires its HTTP registry metadata even for this dry-run; the networked
  dry-run passed using repository-local caches. No upload took place.
- The registry crate's final dry-run currently stops because crates.io has no
  common `^0.8.2` version (registry candidates end at `0.8.1`). This is the
  release skill's documented dependency-order condition, not a falsely claimed
  passing check. Run each remaining final dry-run after its predecessors are
  published and indexed, using the post-merge procedure below.
- An initial strict Clippy run propagated lint enforcement into unchanged
  `journal-common/src/time.rs:106,160` and failed on two pre-existing
  `manual_is_multiple_of` warnings. Their modulo operations remain guarded
  against zero. A writer-only `--no-deps -D warnings` run found four other
  existing warnings: `ptr_arg` at `log/chain.rs:27,35` and `too_many_arguments`
  at `log/startup.rs:84,133`. These files are unchanged. The release does not
  impose a new zero-warning policy or add unrelated refactors; the final
  writer-only Clippy run exits successfully under its ordinary warning policy,
  with exactly these four baseline warnings. No warning is attributed to the
  new method or tests.
- The corrected in-progress audit passes with changed-artifact sensitive
  scanning enabled. The completed-state audit also passes, including status/
  directory consistency and sensitive scanning. Release tag targets were
  rechecked remotely and remain unchanged.

Evidence directory:

- Ignored `.local/rust-v0.8.2.zYsQbBT5/logs/` contains focused/workspace/serde
  tests, normal and strict Clippy output, formatter/wiki checks, common and
  dependency-blocked registry dry-runs, package listings, source consumer, and
  sanitized scanner inventories. Executed wiki-example evidence is the same
  run's `docs-examples/manifest.json` (31 selected/31 passed); the final example
  correction is recorded separately in `docs-normal-close/manifest.json`
  (one selected/one passed). Private logs also retain all six structured external
  verdicts, wiki-validator output, stock-systemd version, and remote tag targets.

Same-failure search:

- Searched all Rust close wrappers/re-exports, Drop, archive-sync, policy hooks,
  and neighboring retention/rotation tests. The public SDK directly re-exports
  the writer type; no second close implementation or facade wrapper needs
  modification. Normal close remains in existing rotation/naming tests.
- Scanned all active docs/spec install examples, not only the previous version:
  Rust examples identify `0.8.2`, and Go examples retain the published `0.8.2`.

Review gate:

- Independent whole-change final review returned PASS. A proposed pending
  publication SOW was withdrawn after checking the actual user override in
  completed SOW-0144; acceptance distinguishes completed PR preparation from
  unperformed registry publication. The exact closed artifact state receives
  an independent final check before the implementation commit is submitted.
- The user authorized external review after local validation. All six requested
  whole-SOW reviews returned PRODUCTION GRADE; no P0/P1/P2 findings, transport
  failures, retries, or coverage gaps occurred. Their thirteen P3 observations
  are adjudicated below. No runtime change followed the review batch; the only
  consumer-example correction was separately compiled and executed.

External finding dispositions:

1. Canonical example special-close placement (five observations): accepted and
   fixed. `docs/Rust-API.md` again demonstrates normal `close()` under its
   rotation/retention heading; separate guidance documents the new method and
   consuming ownership. The corrected example and wiki validator both pass.
2. Single-file handoff is not differential (three observations): accepted as
   a coverage limit, not a defect. The user's exact scenario is preserved.
   Normal close protects its just-archived file in `log/chain.rs:649-651`;
   `tests/log_writer/close_without_retention.rs:47-102` independently compares
   deletion/preservation of an expired older rotated archive, and lines 116-137
   compare failing retention accounting. Adding the same control to the literal
   scenario is rejected as redundant; existing controls prove the difference.
3. Retained wiki/systemd evidence (two observations): accepted and corrected.
   Private logs now include wiki-validator output, `journalctl --version`, and
   remote immutable-tag targets alongside executed examples and journal checks.
4. In-progress SOW lifecycle (two observations): accepted as the correct
   pre-review state and completed here. This file is under `done/`, its status is
   completed, and both ledgers are synchronized before the PR commit.
5. Real-sleep test cost (one observation): consciously accepted. The focused
   integration binary took about 21 seconds; duration retention uses real wall
   time, so these controls exercise the actual policy. An injected-clock
   refactor is additional scope and rejected for this surgical API change.

Unrelated observations:

- Existing Clippy/scanner/parser-helper debt retains the concrete dispositions
  below; it does not enter the new close path or block Rust package preparation.
- Unused-variable/import warnings in unchanged documentation examples remain
  non-blocking baseline warnings; all executed examples pass.
- Host and engine are published workspace members, but not entries in
  `[workspace.dependencies]` or dependencies of the public SDK. This unchanged
  graph supports direct package use; their shared internal requirements still
  resolve to `^0.8.2`, as verified by metadata. Adding unused central
  aliases is rejected as unrelated hygiene rather than a release requirement.
- Rust consumes the writer, unlike Go's pointer-receiver idempotence; that is
  the existing Rust ownership contract and is explicitly documented.
- Dependency-ordered remaining dry-runs, registry publication, merge-commit
  inspection, and exact registry consumption remain the authorized post-merge
  procedure, not pre-merge passing claims.

Artifact maintenance gate:

- AGENTS.md: no update; ownership, purity, compiler floor, PR delivery, and
  release safeguards remain unchanged.
- Runtime project skills: no update; existing release dependency ordering,
  verification grammar, close/retention compatibility, and review rules apply.
- Specs: product scope records the consuming Rust API, empty/lazy behavior,
  release version, and staggered Rust source commit without tag rewriting.
- End-user/operator docs: Rust install pins, package README, and writer API
  guides are updated; the canonical example retains ordinary retention closure
  and its final form was compiled/executed.
- End-user/operator skills: none are declared or shipped by this repository;
  no external copied skill is affected.
- SOW lifecycle: SOW-0145 owns the whole change and is completed under `done/`
  with this implementation in the PR under the user's delivery instruction.
- SOW-status.md: canonical and root ledgers both record completion. No unrelated
  SOW lifecycle is changed.
- Lessons/follow-up: maintain consuming Rust ownership and immutable prior
  tags; post-merge publication has the concrete procedure below. Valid quality
  debt retains its existing disposition or tracked SOW; no untracked deferral.

Sensitive data gate:

- Reviewed the new API/tests/docs: identities and messages are synthetic;
  no credentials, host journal entries, customer data, or private endpoints
  enter durable artifacts. Raw package/scanner output stays under `.local/`.

### Scanner Dispositions - 2026-10-03

- Fresh GitHub code-scanning API inventory contains four open CodeQL
  `rust/cleartext-logging` alerts on master commit
  `51bf47f1f90562f7b7717ca41aaaac74f75f90f0`. Alerts #3678/#3679 at
  `rust/src/crates/journal-registry/src/repository/collection.rs:32,42` concern
  in-memory `VecDeque` insertion/removal, which do not log data. Alerts
  #3680/#3681 at `rust/src/crates/journal-log-writer/tests/log_writer.rs:268,293`
  concern failure diagnostics from synthetic journal fixtures; test identities
  are fixed synthetic literals at lines 34/38. The source UUID parser at
  `rust/src/crates/journal-log-writer/src/log/chain.rs:90,107,110` reads journal
  sequence and boot metadata, not authentication secrets. These four findings
  are false positives for this Rust publication; no alert is silently claimed
  fixed or dismissed. This is fresh Rust evidence, not the Go-only rationale
  recorded in SOW-0143.
- Fresh public Codacy Cloud API inventory returns all 50 issues: 32 Go, 13
  Python, five Rust. The rule counts are 21 Go compiler-floor advisories
  (16 High/five Medium), 23 complexity findings, three Rust unsafe-audit
  reminders, one Python subprocess reminder, one undefined helper symbol, and
  one Python whitespace Info. No finding touches the changed close runtime.
- The 21 Go advisories are outside the Rust Cargo dependency graph; this SOW
  neither republishes Go nor changes its compiler floor. SOW-0141's earlier
  compiler disposition covers 13 advisories; it is not claimed to clear eight
  newer advisories. Go remediation is rejected within this Rust-only scope,
  preserving the Netdata floor contract.
- Rust unsafe reminders at `rust/src/cmd/journalctl/output.rs:367,372,377`
  already have explicit null/initialization checks and safety comments at
  lines 363-380. The CLI is unpublished (`Cargo.toml` has `publish = false`).
  These unchanged audit reminders are dispositioned without unrelated changes.
- The 23 complexity findings are genuine existing threshold debt outside the
  new close path. Rust instances are `rust/src/journal/src/parse.rs:34` and
  `rust/src/journal/src/tests/facade.rs:429`; remaining instances are Go and test
  harnesses. Existing SOW-0141 dispositions and user-parked SOW-0097/0098 debt
  remain the authority; no unapproved complexity refactor is added here.
- Python subprocess use at `tests/parser-parity/run_parser_parity.py:94-105`
  passes an explicit argument vector without a shell in an existing harness.
  The undefined `MANIFEST` at lines 77-83 is a real latent defect, but its
  `_short_alias()` helper has no invocation or assignment elsewhere. Remediation
  is rejected within this Rust API/publication scope; it does not enter any
  published crate or executed release path. It is not called a false positive.
- Python whitespace findings at `tests/interoperability/go_fixture_writer.py:67,260`
  refer to tabs inside generated Go string literals, not Python indentation;
  this is a style false positive. Only sanitized aggregate inventories and
  file/line evidence are retained here; scanner response details stay ignored.

## Outcome

Rust's additive consuming `Log::close_without_retention()` API and all eight
`0.8.2` package preparations are complete, locally validated, and reviewed for
the new PR. Normal close behavior is preserved. This SOW is completed inside
the PR under the user's explicit close-out instruction.

At PR preparation, registry publication awaited merge and was not claimed as
completed. PR #3 subsequently merged at
`8ad648a7b4d36bf2d75ee6ee98b42446ff64a276`; all eight Rust `0.8.2` crates are
now published, indexed, and non-yanked. Their package checksums and clean VCS
metadata match that merged source, and the exact registry consumer passes on
Rust `1.91.0` in both naming modes. The publication receipt below records the
post-merge evidence. Existing Go/root tags retain their published Go commit.

## Lessons Extracted

The Go-only publication did not provide Rust API parity; parity needs an
explicit Rust implementation and registry release.
The literal one-file policy handoff does not alone distinguish the close calls;
expired older-archive and callback controls provide the required proof.

## Followup

Post-merge Rust publication and exact registry consumption belong to this
approved SOW's delivery procedure, with completion in the PR as instructed.

### Post-Merge Publication Procedure

1. Confirm this Rust PR is merged, fetch its exact master merge commit, and
   verify a clean source tree at that commit. Existing root `v0.8.2` and
   `go/v0.8.2` must still peel to `51bf47f1f90562f7b7717ca41aaaac74f75f90f0`.
2. Use the installed Rust `1.91.0` toolchain and repository-local Cargo cache,
   target, and temporary directories; no toolchain install or floor change.
3. Dry-run and publish each `0.8.2` crate with the locked dependency graph,
   waiting for its registry index entry before advancing: common, registry,
   core, host, log-writer, index, engine, then public SDK. The release skill
   permits dependent dry-runs after predecessor publication because new path
   versions are validated against the registry.
4. Check all eight exact `0.8.2` registry versions are present and non-yanked.
   Inspect packaged `.cargo_vcs_info.json` against this PR's merged commit.
5. Run a clean exact-version registry consumer on Rust `1.91.0`, without path
   patches, which imports `journal::Log`, calls `close_without_retention()`,
   and verifies the one-second policy handoff and readable original entry in
   both naming modes. Report registry publication only after these checks.
6. Keep the SOW completed inside the PR as requested. This authorized delivery
   procedure does not require another pending SOW or rewriting published tags.

## Regression Log

This is additive Rust parity, not a regression reopening.

## Publication Verification - 2026-10-03

Completed the authorized post-merge procedure after PR #3 merged. No SOW
status or directory transition is needed: this record was already completed
inside that PR under the user's explicit delivery instruction.

Release source and gates:

- Clean source commit: `8ad648a7b4d36bf2d75ee6ee98b42446ff64a276`.
  Its complete Git tree matches reviewed commit
  `ca535ddff30b322be71725f1a2925230b8f6c2be`; independent final review carried
  PASS to the exact merged revision.
- All six merged-commit workflows passed: Coverage, CodeQL, Codacy SARIF,
  Verify Doc Examples, Publish Wiki, and Code Quality.
- Fresh master scanning still has the same four CodeQL findings #3678-#3681
  already dispositioned above; no new CodeQL finding entered this release.
- Existing `v0.8.2` and `go/v0.8.2` still peel to
  `51bf47f1f90562f7b7717ca41aaaac74f75f90f0`. No release tag was changed.

Published packages, in dependency order:

| Package | Version | Verification |
| --- | --- | --- |
| `systemd-journal-sdk-common` | `0.8.2` | indexed, non-yanked, checksum/VCS verified |
| `systemd-journal-sdk-registry` | `0.8.2` | indexed, non-yanked, checksum/VCS verified |
| `systemd-journal-sdk-core` | `0.8.2` | indexed, non-yanked, checksum/VCS verified |
| `systemd-journal-sdk-host` | `0.8.2` | indexed, non-yanked, checksum/VCS verified |
| `systemd-journal-sdk-log-writer` | `0.8.2` | indexed, non-yanked, checksum/VCS verified |
| `systemd-journal-sdk-index` | `0.8.2` | indexed, non-yanked, checksum/VCS verified |
| `systemd-journal-sdk-engine` | `0.8.2` | indexed, non-yanked, checksum/VCS verified |
| `systemd-journal-sdk` | `0.8.2` | indexed, non-yanked, checksum/VCS verified |

Publication and consumer validation:

- Each final `cargo publish --locked --dry-run` passed immediately before its
  upload on Rust/Cargo `1.91.0`. Each preceding package was published and indexed
  before proceeding to its dependents; all eight real uploads succeeded.
- The first upload attempt stopped before any upload because the selected Cargo
  expected the default registry's credential environment variable. Corrected
  credential forwarding in the private release wrapper, then successfully
  published. Credentials remained in memory; no token was recorded.
- Downloaded each exact crate from the registry, matched its SHA-256 checksum
  with registry/index metadata, and checked `.cargo_vcs_info.json`: every package
  records the clean merged source commit above and Rust minimum `1.91`.
  The published writer source includes `close_without_retention()`.
- A fresh consumer pins both public SDK and writer to `=0.8.2`, with no path
  dependency or patch. On Rust `1.91.0`, it calls `journal::Log`'s new method
  after the original one-second policy/1.5-second wait, eagerly reopens without
  retention, confirms unchanged archive bytes, verifies the journal natively,
  and reads the exact original message/one row through `FileReader`. Both chain
  and strict naming pass. Its five SDK packages in `Cargo.lock` are registry
  sources with checksums matching the verified publication receipts.
- An optional offline metadata command requested an uncached Android-specific
  package even though this consumer runs natively. Registry provenance was
  instead verified directly from the consumer's locked source/checksum records;
  no cross-target consumer claim is made.
- Ignored `.local/rust-v0.8.2.zYsQbBT5/logs/` retains all final dry-run/upload
  outputs, eight sanitized package receipts, consumer/toolchain output,
  locked registry provenance, merged CI, scanner inventory, and remote tags.
  Downloaded verification archives remain under the same run's
  `registry-artifacts/`; no raw credential or scanner output enters this record.

Artifact maintenance and follow-up mapping:

- Existing SOW-0145 and both status ledgers now record actual publication.
  Status remains completed under `done/`; no new or reopened SOW is needed.
- Product scope and active install examples already specify `0.8.2` and its
  staggered Rust publication without moving Go tags; these remain accurate,
  so no additional spec or consumer-doc edit is needed.
- AGENTS.md and runtime skills need no update: publication used the existing
  authorized release order, purity boundaries, compiler floor, and credential
  safeguards. No end-user/operator skills are declared or shipped.
- All publication steps and exact registry consumption are implemented.
  Existing unrelated quality debt retains the dispositions above; no new
  deferred publication work remains. Receipt maintenance is submitted through
  the PR workflow, preserving the user's release-delivery instruction.
- Publication-receipt validation: `git diff --check` and the SOW audit with
  changed-artifact sensitive-data scanning both pass. No source changed after
  publication; the receipt updates only this completed SOW and both ledgers.
