# SOW-0140 - v0.7.9 Release Readiness

## Status

Status: in-progress

Sub-state: implementation, local validation, and authorized external review are
complete; source commit/push, authoritative GitHub Coverage/Codacy/CodeQL
closeout, and authorized Rust crates.io publication remain.

## Requirements

### Purpose

Prepare the validated Rust and Go writer optimizations for a patch release,
publish the Rust `0.7.9` crates after green CI, and leave immutable Git tags and
the Go module release for a separate user decision.

### User Request

The user accepted keeping both writer correctness/performance changes, asked
for the next release, accepted the recommended release-readiness plan, selected
the comprehensive test/static-analysis repair options, and directed the work to
proceed. Ordinary repository edits and validation retries under that approved
plan do not require production-migration approval.

### Assistant Understanding

Facts:

- `master` is clean at `cbb71fb04eeb054acdfad099dd66b4b10d9695a0`.
- Rust and Go resolved-DATA link-state reuse are committed at `438ab03` and
  `688d67b`; the systemd 255 follow compatibility repair is committed at
  `cbb71fb`.
- GitHub Coverage run `30224014038` fails only because 11 Rust `journalctl`
  unit tests try to launch a hard-coded workspace debug binary that does not
  exist in the coverage target directory.
- GitHub Codacy SARIF run `30224014022` reports findings in completed SOW
  history, retired experiments, and nine active files/lines.
- Current CodeQL reports three findings from the current Rust/Go analysis and
  two stale Python findings last analyzed at `a2361ab`; the current workflow no
  longer analyzes Python after the Python/Node product retirement.
- The released Rust workspace version and documented Rust dependency example
  are `0.7.8`; no local or remote `v0.7.9` or `go/v0.7.9` tag exists.

Inferences:

- This is a patch release because the writer optimizations and gate repairs do
  not change the public API, file format, or default behavior.
- A release must not proceed while Coverage and Codacy are red.
- The clean Rust CLI-test repair is to make binary-spawning tests integration
  tests, where Cargo guarantees `CARGO_BIN_EXE_journalctl`, instead of allowing
  tests to use a stale or unrelated binary.

Unknowns:

- Whether every dependent Rust crate can complete `cargo publish --dry-run`
  before its `0.7.9` internal dependencies exist on crates.io. The release
  workflow explicitly allows recording this registry-order constraint while
  still validating local packages.

### Acceptance Criteria

- Rust coverage no longer depends on a pre-existing
  `.local/cargo-target/debug/journalctl`.
- The 11 end-to-end Rust CLI assertions execute the Cargo-built binary and pass
  from a fresh target directory.
- Codacy excludes only completed SOW history, retired experiments, and the
  existing lockfile exclusion; all nine active Codacy findings are fixed.
- The three current CodeQL findings are fixed without weakening behavior; the
  two stale Python findings are dispositioned with commit evidence.
- Rust and Go tests, formatting, docs examples, interoperability, writer
  correctness, SOW audit, and relevant release packaging gates pass.
- Rust release metadata and consumer docs consistently identify `0.7.9`.
- Authorized whole-SOW external review has no unresolved blocking findings.
- Release-prep source is committed and pushed with green required CI.
- All eight Rust `0.7.9` crates are dry-run and published in dependency order
  only after the release-prep commit is pushed and required CI is green.
- No release tag or Go module tag is created or pushed without separate user
  approval.

## Analysis

Sources checked:

- `rust/src/cmd/journalctl/main.rs:4708-5059`
- `rust/src/cmd/journalctl/Cargo.toml`
- `tests/coverage/run_rust_coverage.sh`
- GitHub Coverage run `30224014038`
- `.codacy.yml`
- `.github/workflows/codacy-sarif.yml`
- GitHub Codacy SARIF run `30224014022`
- GitHub open code-scanning alerts exported under ignored
  `.local/release-readiness/`
- `go/cmd/journalctl/output.go:254-263`
- `rust/src/crates/journal-host/src/platform/linux.rs:82-95,135-145`
- `.github/workflows/codeql.yml`
- `.agents/skills/project-release-tagging/SKILL.md`
- `rust/Cargo.toml`, `rust/Cargo.lock`, `README.md`, and `rust/README.md`

Current state:

- `run_cli()` in the Rust binary's unit-test module first checks an environment
  variable that Cargo guarantees for integration tests, not unit tests, then
  falls back to `.local/cargo-target/debug/journalctl`. Coverage builds under
  `rust/target/llvm-cov-target`, so all 11 subprocess tests fail with
  `No such file or directory`; the other 21 binary unit tests pass.
- The nine active Codacy findings are two insecure-temp-literal test paths, one
  subprocess-import annotation, two truly unused imports, and four missing
  blank-line findings in the two SOW status ledgers.
- The Go timestamp path parses `_SOURCE_REALTIME_TIMESTAMP` as `uint64` and
  converts the entire value to `int64` before `time.UnixMicro`. At
  `math.MaxUint64`, this wraps to `-1` microsecond. Splitting into seconds and
  microsecond remainder, as the Rust renderer already does, preserves the full
  journal `uint64` timestamp.
- The two Rust CodeQL unused-variable alerts are false positives caused by
  implicit captured-format arguments. Equivalent machine-id and boot-id error
  paths should use explicit format arguments so the variables remain
  unambiguous to both readers and the analyzer.
- The two stale Python CodeQL alerts were fixed by `e17f694`; their alert
  instances remain open only because Python was removed from the CodeQL
  workflow when it became a tooling language rather than a product SDK.

Risks:

- Moving tests incorrectly could reduce parser coverage or make integration
  tests depend on the host journal. The moved tests must retain every assertion
  and use only synthetic paths/options.
- Timestamp repair could accidentally truncate the sub-second remainder. A
  maximum-`uint64` regression test must prove the split representation.
- Broad static-analysis exclusions could hide active code. Exclusions are
  limited to `.agents/sow/done/**`, `experiments/**`, and `rust/Cargo.lock`.
- Release version changes affect all eight publishable Rust packages through
  workspace inheritance and lockfile resolution. Full workspace tests and
  package inspection are required.
- Rust crates.io publication is authorized only after green required CI;
  immutable tag creation and the Go module release remain outside this SOW's
  authority.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:

- The source changes intended for `0.7.9` are locally validated but the release
  branch has two red CI gates. Coverage is non-hermetic because unit tests
  launch an unrelated hard-coded binary path. Codacy analyzes historical and
  retired artifacts while also reporting nine small active issues. CodeQL
  exposes one real Go integer-wrap defect, two Rust analyzer-clarity findings,
  and two already-fixed stale Python instances. These must be repaired or
  explicitly dispositioned before release metadata changes are accepted.

Evidence reviewed:

- The source, workflow, alert, and CI evidence listed in `## Analysis`.
- GitHub Coverage log: 21 Rust `journalctl` unit tests passed and exactly 11
  `run_cli()` tests failed at `main.rs:4725` while spawning the missing fallback
  binary.
- A focused Go range probe showed the existing maximum `uint64` conversion
  renders `1969-12-31T23:59:59.999999Z`, while a seconds/remainder split renders
  the positive full-range timestamp with `551615` microseconds preserved.
- SOW-0116 records `experiments/` as retired non-product Python/Node code.
- SOW-0084 records the local convention of narrow inline Bandit subprocess
  annotations and Codacy `exclude_paths` support.

Affected contracts and surfaces:

- Test architecture: Rust `journalctl` unit versus integration tests.
- Go portable `journalctl` timestamp rendering for extreme untrusted journal
  values.
- Rust diagnostic wording implementation, not emitted wording.
- Static-analysis scope and status-ledger Markdown.
- Rust workspace package versions, lockfile, and Rust consumer examples.
- GitHub code-scanning alert state for two obsolete Python instances.

Existing patterns to reuse:

- Cargo integration tests use `env!("CARGO_BIN_EXE_<name>")`.
- Rust `journalctl` already splits `uint64` microseconds into seconds and
  nanoseconds before timestamp construction.
- Active Python harnesses use narrow `# nosec B404` annotations.
- Repository scratch paths use `.local/`.
- Release package and tag rules come from
  `.agents/skills/project-release-tagging/SKILL.md`.

Risk and blast radius:

- Product behavior changes only for previously wrapped Go timestamps above
  `math.MaxInt64` microseconds; they will retain their real positive value.
- Rust changes are test placement and analyzer-visible formatting only.
- Static-analysis scope stops scanning history and explicitly retired SDK
  experiments, but continues scanning active Rust, Go, tests, docs, SOW
  ledgers, and tooling.
- Version bump applies to all workspace-inherited Rust packages. No Go source
  version constant exists; Go release identity comes from the future
  `go/v0.7.9` tag.

Sensitive data handling plan:

- CI metadata, code-scanning exports, reviewer output, and package logs stay in
  ignored `.local/` scratch storage. Durable SOW evidence contains only public
  repository paths, commit hashes, alert numbers, aggregate results, and
  sanitized summaries. No tokens, credentials, private endpoints, customer
  data, or proprietary incident payloads will be recorded.

Implementation plan:

1. Move all 11 Rust binary-spawning tests unchanged into
   `rust/src/cmd/journalctl/tests/cli_subprocess.rs`, use Cargo's compile-time
   binary path, and remove the non-hermetic fallback helper.
2. Add the three agreed Codacy exclusions and fix the nine active Codacy
   findings using narrow source changes.
3. Preserve full-range Go CLI timestamps by passing split
   seconds/nanoseconds to `time.Unix` in every live renderer, add focused
   regression coverage, and make equivalent Rust captured-format uses
   explicit. Preserve the existing exported `BootInfo` field types in this
   patch release and record their legacy signed-range limitation.
4. Disposition the two already-fixed stale Python CodeQL alerts as fixed with
   commit evidence after current source validation.
5. Bump the Rust workspace and internal dependency versions to `0.7.9`, refresh
   the lockfile, update every active Rust and Go consumer version example, and
   add the missing consumer-doc sweep to the release-tagging skill.
6. Run focused, full, interoperability, coverage, packaging, docs, static
   analysis, audit, and whole-SOW review gates.
7. Commit and push the validated release-prep source, wait for CI, then stop at
   the publication checkpoint.

Validation plan:

- Fresh-target `cargo test -p journalctl`.
- Focused Go `journalctl` tests including maximum-`uint64` timestamp rendering.
- `cargo test --workspace`, `go test ./... -count=1`, and
  `cargo fmt --all --check`.
- Rust/Go compact and regular writer tests, byte identity, inverted-index
  correctness, verification/interoperability, and relevant live/file-backed
  matrices.
- Local focused Bandit, Pylint, markdownlint, and Codacy configuration checks;
  pushed Codacy SARIF and CodeQL runs are authoritative closeout.
- Rust and Go coverage workflows, docs checks, verified examples,
  `git diff --check`, and `.agents/sow/audit.sh`.
- Rust package listing and publish dry-runs in dependency order, recording any
  expected registry-order constraint without publishing.
- Same-pattern searches for binary-path fallbacks, unsafe integer conversions,
  captured-format false positives, and active files accidentally covered by
  exclusions.
- Authorized read-only external review of the complete SOW and diff.

Artifact impact plan:

- AGENTS.md: no update expected; repository workflow and product guardrails do
  not change.
- Runtime project skills: update the release-tagging skill if the active
  consumer-doc version sweep is missing; do not copy volatile external-reviewer
  instructions.
- Specs: no update expected; public API and journal-format behavior do not
  change.
- End-user/operator docs: update all active Rust and Go dependency examples to
  `0.7.9`, including published `docs/**` pages that may still identify an older
  release.
- End-user/operator skills: no copied/output skill is affected.
- SOW lifecycle: one release-readiness SOW; actual publication remains a
  separate final checkpoint after this SOW.
- SOW-status.md: both canonical and convenience ledgers are updated at
  activation and completion.

Open-source reference evidence:

- No new external source reference is required for these repository-local gate
  repairs. The systemd interoperability baselines already recorded by SOW-0139
  remain the relevant upstream evidence.

Open decisions:

- None. The user selected the comprehensive gate/static-analysis repair and
  release-preparation options, accepted the recommendations, authorized the
  review session, and explicitly directed implementation to proceed.

## Implications And Decisions

1. **Decision 1 - repair release gates before release: option A accepted
   (long-term-best).** Fix the non-hermetic test architecture rather than
   bypassing coverage or prebuilding an unrelated fallback binary.
2. **Decision 2 - restore a meaningful Codacy green baseline: option A accepted
   (long-term-best).** Exclude completed SOW history and retired experiments,
   retain active surfaces, and fix every active finding.
3. **Decision 3 - prepare `0.7.9` and initially stop before publication: option
   A accepted (surgical).** Complete source, version, validation, review,
   commit, push, and CI readiness before any publication. Decision 7 later
   supplies the required Rust crates.io approval but does not authorize tags.
4. **Decision 4 - external review authorization.** Standing authorization in
   the current conversation covers the user-selected reviewer set. The
   system-wide external-reviewers skill remains the source of truth for the
   current harnesses and execution details.
5. **Decision 5 - ordinary source-work approval scope.** Repository edits and
   validation retries that follow this approved plan are not production-system
   migrations and do not require repeated approval. Unexpected design,
   compatibility, publication, destructive, service, or system-package changes
   still stop for a new decision.
6. **Decision 6 - no further external-review round.** After round 2 found only
   documentation/process omissions, the user explicitly waived another review
   rerun. Mimo's truncated round-2 output is ignored and is not retried.
7. **Decision 7 - publish Rust `0.7.9` to crates.io.** The user requires the
   released crates for downstream Netdata integration and explicitly approved
   crates.io availability. After the final commit is pushed and required CI is
   green, dry-run and publish all eight Rust crates in dependency order. This
   does not authorize creating or pushing Git tags or a Go module release.

## Plan

1. Repair hermetic CLI-test execution and active analyzer findings.
2. Validate behavior before changing release metadata.
3. Bump and validate `0.7.9` release metadata and packages.
4. Run whole-SOW external review and address verified findings.
5. Commit, push, verify CI, close this release-readiness SOW, and request the
   final publication decision.

## Implementation And Review Plan

Implementation:

- The project manager performs all source changes and validation sequentially.
- Each change retains focused evidence and is reviewed as one complete SOW
  after local validation.

Reviewers:

- External review is authorized for the current conversation. Use the
  system-wide `external-reviewers` skill at the whole-SOW boundary; do not copy
  volatile harness or model instructions into this SOW.

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

- Stop on unexpected product/design, packaging, compatibility, or external-state
  behavior. Record the evidence, revise the plan, and obtain a user decision
  before deviating. Ordinary test failures are investigated and repaired within
  the approved scope.

## Execution Log

### 2026-07-26

- Re-established a clean `master` after the host reboot and disk migration.
- Captured current GitHub Coverage, Codacy, and CodeQL evidence.
- Completed root-cause and same-pattern analysis.
- Activated SOW-0140 with the approved decisions and a ready
  pre-implementation gate.
- Moved all 11 binary-spawning Rust `journalctl` tests from the binary unit-test
  module to `tests/cli_subprocess.rs` and replaced the environment/fallback
  lookup with Cargo's compile-time `CARGO_BIN_EXE_journalctl`.
- Added the approved Codacy exclusions and repaired all nine active Codacy
  findings: repository-local synthetic output paths, one narrow Bandit
  annotation, two unused imports, and status-ledger Markdown spacing.
- Replaced the Go full-width unsigned-to-signed microsecond conversion with a
  seconds/nanoseconds split and added maximum-`uint64`
  `_SOURCE_REALTIME_TIMESTAMP` regression coverage.
- Rewrote equivalent Rust machine-id, boot-id, and journal-directory captured
  format arguments explicitly without changing emitted errors.
- Dismissed stale CodeQL alerts 3385 and 3397 with evidence that `e17f694`
  already fixed them and that Python is no longer a product-language CodeQL
  target.
- Bumped the Rust workspace/internal dependency version and active Rust
  consumer examples from `0.7.8` to `0.7.9`; Cargo refreshed all workspace
  package versions in `rust/Cargo.lock`.
- Confirmed no externally public Rust or Go declaration changed since
  `v0.7.8`; the semver classification remains patch.
- External-review round 1 completed with all seven authorized reviewers.
  Two reviewers returned `NEEDS CHANGES`, four returned `PRODUCTION GRADE`,
  and one returned `PRODUCTION GRADE` with an explicit local-worktree coverage
  gap caused by its deterministic sandbox failure.
- Independently verified both blocking findings: published `docs/**` consumer
  examples still identified `0.7.7`, and the live Go
  `formatHeaderTimestamp(uint64)` path still narrowed and multiplied the full
  timestamp in `int64`.
- Resolution recorded before edits: update every active consumer example and
  the release-tagging checklist; use the already-approved seconds/remainder
  timestamp design in the live Go header/list-boots formatter with focused
  maximum-`uint64` coverage. Do not change the exported signed `BootInfo`
  fields in a patch release; document that pre-existing public-API limitation
  separately.
- Updated all active Rust and Go install examples under the root README,
  `rust/README.md`, and published `docs/**` pages to `0.7.9`.
- Added the missing active-consumer-version sweep to
  `project-release-tagging`, explicitly preventing scans that check only the
  immediately previous version.
- Repaired the live Go header/list-boots formatter with the same
  seconds/remainder design and added maximum-`uint64` regression coverage.
- Made the moved Rust subprocess suite self-documenting and supplied a
  synthetic nonexistent `--file` to the false synchronize-on-exit test so it
  cannot fall through to host journal discovery.
- Round-2 review found no product-code blocker. One reviewer identified two
  remaining release-artifact omissions: the active product-scope spec still
  used `0.7.4`, and the release checklist used a fixed file list that omitted
  that spec.
- Updated the active product-scope install example to `0.7.9` and replaced the
  fixed release-checklist list with a runnable recursive scan across active
  README, docs, and spec roots.
- The user waived another external-review rerun and instructed that Rust
  `0.7.9` must be available on crates.io for downstream integration.
- Because Git tags and a Go module release were not authorized, retained Go
  install examples at the current `v0.7.8` tag while Rust examples identify
  the crates.io-bound `0.7.9` release.

## Validation

Acceptance criteria evidence:

- Fresh-target Cargo execution built the actual `journalctl` binary and passed
  21 unit tests plus all 11 moved subprocess tests. No hard-coded
  `.local/cargo-target/debug/journalctl` fallback remains.
- `.codacy.yml` excludes only `.agents/sow/done/**`, `experiments/**`, and
  `rust/Cargo.lock`. A path-policy probe confirmed active/pending SOWs, Rust,
  Go, tests, and docs remain included.
- The Go maximum-`uint64` source timestamp renders
  `586524-01-19T08:01:49.551615+00:00`; the previous implementation wrapped
  the same value to one microsecond before the Unix epoch.
- The live Go header/list-boots formatter also preserves maximum-`uint64`
  microseconds instead of overflowing during nanosecond conversion.
- Every active Rust install example identifies crates.io version `0.7.9`;
  every active Go install example identifies the current module tag `v0.7.8`.
  `v0.7.9` and `go/v0.7.9` remain absent locally and remotely.
- The release-tagging skill's runnable recursive scan covers README, published
  docs, and active specs without a fixed page list.
- The full Rust/Go package-source comparison against `v0.7.8` found no external
  public declaration changes.

Tests or equivalent validation:

- `cargo fmt --all --check`: passed.
- Fresh-target `cargo test --offline -p journalctl`: passed 21 unit plus 11
  integration tests.
- `cargo check --offline --workspace`: passed at workspace version `0.7.9`.
- `cargo check --offline -p writer_core_bench`: passed.
- Final `cargo test --offline --workspace`: passed 489 tests across 41 result
  groups, with only the existing ignored documentation tests.
- Focused Go timestamp/output tests: passed.
- `go vet ./...`: passed.
- Final `go test ./... -count=1`: passed every package.
- Post-review focused Go `journalctl` tests, `go vet ./...`, and
  `go test ./... -count=1`: passed.
- Post-review Rust `journalctl` validation passed all 21 unit and 11
  integration tests; the synchronize-on-exit no-op case no longer consults
  the default host journal source.
- Go writer benchmark build: passed.
- Windows AMD64 `go test -c ./journal`: passed; artifact SHA-256
  `789ba6e58e5811a10d7f69bff356d20373d48184530bac4ce89e5e85d2d1fedc`.
- Python benchmark suite: 25 passed.
- Docs harness unit suite: 50 passed.
- Verified examples: 31/31 passed across Rust and Go.
- Go-writer regular closed-file interoperability: 16/16 checks passed across
  stock, Rust, and Go readers on systemd 255.
- Go-writer compact interoperability: 10/10 structural, stock, libsystemd,
  Rust, and Go checks passed on systemd 255.
- Full verifier matrix: all 63 results passed across 9 positive formats and 12
  negative corruptions with stock, Rust, and Go verifiers.
- Rust/Go deterministic online, offline, and archived outputs were
  byte-identical and stock-verifiable. Their hashes remained the established
  `400eae423394117fe5cbd1775df1241b86be57b22a1afbd6f4e68a0f76865571`,
  `25800a4a660cd8de84b40a9f813dc43262ace7f662f3602d59b5ad73776bd920`,
  and `b8d7518141ea06dd6908fa393627beb2388a3a8e413f3d9f9e0afc57c3919a5a`.
- `cargo publish --dry-run --allow-dirty --locked` fully packaged and verified
  `systemd-journal-sdk-common 0.7.9`.
- The next full dry-run correctly stopped because crates.io does not yet have
  `systemd-journal-sdk-common 0.7.9`; this is the expected dependency-order
  constraint documented by the release skill. `cargo package --list` succeeded
  for all eight publishable crates, and full dry-runs for dependent crates must
  occur sequentially after each predecessor is published under final user
  approval.
- YAML parsing, active-path exclusion probes, Markdown heading/list spacing,
  Python syntax checks, `git diff --check`, and the project SOW audit passed.
- Post-review wiki validation passed all 15 pages, and the docs harness passed
  all 77 tests.
- Local Rust LCOV generation remains unavailable because `cargo-llvm-cov` is
  not installed. No installation was performed; pushed Coverage CI is the
  authoritative coverage closeout.

Real-use evidence:

- The deterministic ingesters exercised both optimized writers in all three
  final states, produced exactly matching 8 MiB journals, and passed stock
  verification.
- Stock systemd 255 and libsystemd consumed the changed Go writer's regular and
  compact files; both SDK readers consumed the same files.
- The Rust writer's compact and regular direct writer correctness remains
  covered by the 80-test `journal-core` suite, including stock verification,
  resolved-link-state invariants, raw/structured byte identity, regular and
  compact tails, fallback, deduplication, compression, and sealing.
- The generic Rust directory-mode matrix attempt reproduced the known
  deterministic-identity harness failure tracked by pending SOW-0136:
  `machine ID error: machine id is required`. No production writer failure
  occurred and this release SOW did not weaken the explicit-identity contract.
  Consequently, that generic directory-mode matrix supplies no release
  evidence; direct writer, log-writer, file interoperability, byte-identity,
  and stock-verifier evidence listed above provide the release coverage until
  SOW-0136 repairs the harness.

Reviewer findings:

- Round 1 artifacts are under
  `.local/release-readiness/review-round-1/`.
- Verified blocking P2 findings:
  - stale published consumer install examples under `docs/**`;
  - incomplete same-defect repair in the live Go header/list-boots timestamp
    formatter.
- Verified non-blocking findings include the pre-existing signed `BootInfo`
  public fields, the known SOW-0136 directory-harness gap, and cosmetic
  test-comment/documentation suggestions. The signed public fields require a
  separate compatibility decision and are not changed in a patch release.
- The subprocess-file comments and host-source isolation suggestions were
  accepted and implemented. Other P3 observations were either informational,
  pre-existing, contradicted by project-required behavior, or outside scope:
  sequential registry dry-runs, unavailable local LLVM coverage, dual required
  SOW ledgers, retired-experiment exclusion, existing Rust formatting style,
  timestamp behavior beyond stock systemd's printable range, and the public
  signed `BootInfo` compatibility limitation.
- Codex could not inspect the local worktree because its own deterministic
  bubblewrap setup failed. Its returned verdict is retained but is not counted
  as complete working-tree review evidence.
- Material fixes require a complete round-2 review with the same authorized
  reviewer set unless the user waives that gate.
- Round 2 completed for six reviewers; Mimo returned truncated output without
  the required XML verdict. The user explicitly directed that Mimo be ignored
  and that no further reviewer rerun occur.
- Round-2 adjudication verified the only blocking findings as release-artifact
  omissions: `.agents/sow/specs/product-scope.md` still identified `0.7.4`,
  and the new release-skill sweep used an incomplete fixed file list. Both are
  corrected by the final artifact edits.

Same-failure scan:

- No production or test source retains the removed Rust
  `.local/cargo-target/debug/journalctl` fallback.
- The only `CARGO_BIN_EXE_journalctl` use is the new integration-test
  compile-time path.
- No live Go CLI timestamp renderer converts a complete journal `uint64`
  microsecond value to `int64` before splitting seconds and remainder.
- The legacy exported `journal.BootInfo` API still stores its timestamps as
  `int64`; changing those public field types is intentionally excluded from
  this patch release and requires a separate compatibility decision.
- Equivalent `source` formatting in both Linux machine-id and boot-id candidate
  readers now uses explicit arguments.
- The nine active Codacy source patterns are gone; the remaining `os` imports
  in parser-parity scripts are used.
- Active Rust install examples use `0.7.9`, and active Go install examples use
  the current `v0.7.8` module tag; historical SOW evidence was intentionally
  preserved.

Sensitive data gate:

- Durable artifacts contain no raw secrets, credentials, bearer tokens,
  customer data, personal data, private endpoints, or proprietary payloads.
  Scanner exports, package logs, and generated journals remain ignored under
  `.local/`.

Artifact maintenance gate:

- AGENTS.md: no update; project workflow and guardrails did not change.
- Runtime project skills: `project-release-tagging` now requires an
  all-active-consumer-doc version sweep and warns against checking only the
  immediately previous release.
- Specs: `product-scope.md` now identifies the current `0.7.9` Rust dependency;
  public API, file format, indexes, defaults, and durability contracts did not
  otherwise change.
- End-user/operator docs: root/Rust READMEs, the active product spec, and Rust
  examples under published `docs/**` now show `0.7.9`; Go examples remain at
  the current `v0.7.8` module tag because no Go release is authorized.
- End-user/operator skills: no copied/output skill is affected.
- SOW lifecycle: in-progress in `.agents/sow/current/`.
- SOW-status.md: activation updates included in both ledgers.

Specs update:

- `product-scope.md` required a version-only correction from `0.7.4` to
  `0.7.9`; no public journal contract text changed.

Project skills update:

- `project-release-tagging` now includes the consumer-doc version sweep that
  round-1 review proved was missing.

End-user/operator docs update:

- Root/Rust READMEs, the active product spec, and published Rust install
  examples use `0.7.9`; Go examples retain the current `v0.7.8` tag.

End-user/operator skills update:

- No output/reference skill consumes the changed version examples or internal
  implementation.

Lessons:

- Binary subprocess assertions belong in Cargo integration tests; checking an
  integration-only environment variable from unit tests and falling back to a
  workspace binary makes clean-target coverage nondeterministic.
- A journal `uint64` microsecond timestamp should be split before constructing
  Go `time.Time`; converting the full value to `int64` silently loses half the
  representable on-disk range.
- Version sweeps must inspect all active consumer docs for any older release,
  not only the immediately previous version.
- Rust dependent-crate crates.io dry-runs are necessarily sequential for a new
  coordinated workspace version, because each predecessor must exist in the
  registry before Cargo can verify the next packaged crate.

Follow-up mapping:

- SOW-0136 already tracks the Rust directory harness's explicit synthetic
  identity repair; it is not broadened into this release SOW.
- Rust crates.io publication is covered by Decision 7 after commit/push and
  green CI. Immutable tags and a Go module release remain separate follow-ups.

## Outcome

Pending.

## Lessons Extracted

Pending.

## Followup

Rust crates.io publication is authorized after the release-prep commit is
pushed and required CI is green. Immutable Git tag creation and the Go module
release remain gated on separate explicit user approval.

## Regression Log

None yet.
