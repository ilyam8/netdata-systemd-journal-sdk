# SOW-0141 - v0.8.0 Dependency Refresh

## Status

Status: in-progress

Sub-state: local implementation, validation, and external-review disposition
complete. Awaiting the final release checkpoint; current GitHub code-scanning
and Codacy results remain a mandatory post-push, pre-tag/publication gate.

## Requirements

### Purpose

Refresh the Rust and Go SDK dependency graphs comprehensively, align declared
minimum toolchains with the current Netdata consumer contract, remove dead
dependency declarations, and prepare a validated `0.8.0` release candidate.

### User Request

The user selected:

- match the Go module minimum to Netdata and validate both the minimum and
  Netdata packaging toolchains;
- align the Rust minimum to Netdata and validate both the minimum and a current
  toolchain;
- update actively used direct dependencies to the latest non-yanked stable
  releases compatible with those toolchains;
- refresh transitive resolution, remove unused dependency declarations, and
  retain an intentional pre-release only where no newer compatible API line
  exists;
- supersede the untagged Go `0.7.9` plan and release `0.8.0` only if all
  correctness, interoperability, integration, and performance gates pass.

### Assistant Understanding

Facts:

- The repository is clean at
  `d0dce35cfa7020b03c5cb3971e23a8df4481cd18`.
- Rust `0.7.9` crates are already published, but neither `v0.7.9` nor
  `go/v0.7.9` exists locally or remotely.
- `go/go.mod` declares `go 1.26`.
- `netdata/netdata @ 8005e84c0c14cbee336d5e50ad731067ce326990`
  declares `go 1.26.2` in `src/go/go.mod`, downloads Go `1.26.5` in
  `packaging/check-for-go-toolchain.sh`, and selects `^1.26` in
  `.github/workflows/build.yml`.
- `rust/Cargo.toml` declares `rust-version = "1.85"`, while the locked
  `roaring 0.11.4` dependency declares Rust `1.90.0`; the current SDK MSRV
  metadata is already false.
- Netdata declares `rust-version = "1.91"` in `src/crates/Cargo.toml`.
- A compatible Cargo lock refresh currently proposes 88 package changes,
  including `spin 0.9.8` to non-yanked `0.9.9`.
- Direct Rust releases beyond current requirement lines include cache,
  compression, hash, cryptography, random, and build dependencies. The largest
  expected migration risks are `foyer 0.20.1` to `0.22.3`, `lru 0.16.4` to
  `0.18.1`, compression major/minor-line changes, `rand 0.9` to `0.10`, and
  RustCrypto digest-line changes.
- All four Go direct dependencies have newer releases. Netdata's resolved graph
  already selects `github.com/klauspost/compress v1.19.1` and
  `golang.org/x/sys v0.47.0`.
- Fourteen external workspace dependency declarations are not referenced by
  any Rust member manifest. `md5` is referenced by
  `journal-core/Cargo.toml` but has no Rust source use.
- The intentional direct `zerocopy 0.9.0-alpha.0` requirement is the newest
  live `0.9` API-line release; stable `0.8.x` is an older API line rather than
  a newer replacement.
- GitHub reports zero open Dependabot alerts. A Go vulnerability scan under
  Go `1.26.5` reports no reachable vulnerabilities in the current Go module
  graph.

Inferences:

- A comprehensive refresh is appropriate for a pre-1.0 minor release, but not
  for the already-published Rust patch release.
- Passing unit tests alone cannot establish readiness because cache,
  compression, hashing, and transitive scheduler/runtime changes can alter
  journal compatibility or performance without changing APIs.
- The correct "latest" policy is the latest non-yanked stable release that
  compiles with Go `1.26.2` or Rust `1.91`, not an unconstrained newest release
  requiring a newer consumer toolchain.

Unknowns:

- Which direct Rust dependency upgrades require source migration rather than a
  manifest-only change.
- Whether the refreshed dependency graph changes journal bytes, index results,
  cache persistence compatibility, or measured writer/reader throughput.
- Whether every transitive dependency selected under the Rust `1.91` resolver
  contract actually compiles on Rust `1.91`; declared dependency metadata and
  an exact-toolchain build must both pass.

### Acceptance Criteria

- `go/go.mod` declares `go 1.26.2` and no SDK-specific `toolchain` directive.
- The Go module resolves the latest stable compatible versions of all four
  direct dependencies and is tidy.
- `rust/Cargo.toml` declares Rust `1.91`.
- Every actively used direct Rust dependency is on the latest non-yanked stable
  compatible release, except the recorded intentional `zerocopy` pre-release
  API line.
- Unused workspace and package dependency declarations are removed with
  same-pattern evidence.
- Cargo resolution contains no yanked package and is generated with the Rust
  `1.91` compatibility contract.
- Public Rust and Go APIs, default behavior, file format, journal indexes,
  durability semantics, and runtime-purity contracts remain unchanged unless
  an unexpected migration decision is presented to and approved by the user.
- Rust builds and tests pass on exact Rust `1.91` and the available current
  Rust `1.97.1`.
- Go builds and tests pass on exact Go `1.26.2` and Netdata's packaging
  toolchain Go `1.26.5`.
- Compact/regular, compression, byte-identity, sealing, live, verifier, and
  Rust/Go interoperability gates pass.
- A scratch-copy Netdata integration build/tests resolve the local `0.8.0`
  candidate without modifying any Netdata checkout.
- Focused before/after cache, reader, and repeated-structured-writer benchmarks
  show no unexplained material regression. Any repeatable regression stops
  release preparation for a user decision.
- Active consumer docs and specs identify `0.8.0`; release package dry-runs
  pass.
- No tag, registry publication, or release is performed without a final
  explicit user checkpoint after all evidence is reported.

## Analysis

Sources checked:

- `AGENTS.md`
- `.agents/skills/project-release-tagging/SKILL.md`
- `.agents/skills/project-docs-authoring/SKILL.md`
- `.agents/skills/project-journal-compatibility/SKILL.md`
- `.agents/sow/done/SOW-0072-20260530-dependency-and-package-hygiene.md`
- `.agents/sow/done/SOW-0140-20260726-v0-7-9-release-readiness.md`
- `.agents/sow/pending/SOW-0066-20260530-v1-release-and-registry-publication.md`
- `.agents/sow/specs/product-scope.md`
- `rust/Cargo.toml`
- `rust/Cargo.lock`
- Rust workspace member manifests and dependency call sites
- `go/go.mod`
- `go/go.sum`
- Go compression and platform dependency call sites
- active consumer docs under `README.md`, `rust/README.md`, and `docs/`
- current crates.io sparse-index metadata for direct Rust dependencies
- current Go module version queries
- current GitHub dependency alerts and dependency graph

Current state:

- The SDK source and previous SOW closeout are committed and pushed.
- The `0.7.9` Rust publication is complete, but the coordinated Git/Go release
  was deliberately not created.
- Go dependency updates are small in count, while Rust direct and transitive
  updates include multiple source-compatible and source-breaking version-line
  changes.
- The repository already requires a newer Rust compiler than its declared
  `1.85` through `roaring 0.11.4`.

Risks:

- Cache-library migration can change serialized cache behavior, recovery, or
  resource usage.
- Compression-library migration can change encoded bytes, corruption handling,
  or interoperability.
- Hash and cryptography migrations can change digest construction if adapted
  incorrectly.
- Random-library migration can alter deterministic fixtures if test RNG paths
  are not preserved.
- A broad lock refresh can change runtime scheduling, memory use, or benchmark
  results despite compiling cleanly.
- Raising the declared minimum compilers affects SDK consumers; alignment with
  current Netdata bounds is therefore mandatory.
- Creating release tags or publishing packages before final validation would
  make an incomplete result externally immutable.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:

- Dependency requirements and the lockfile were accumulated across many
  feature releases. They now contain stale direct versions, a yanked transitive
  lock, unreferenced workspace declarations, and an inaccurate Rust minimum
  version. The user selected a coordinated `0.8.0` maintenance release rather
  than layering more patch-release exceptions.

Evidence reviewed:

- The source, dependency, consumer, documentation, prior-SOW, and registry
  evidence listed above.
- `cargo update --dry-run --verbose` reports 88 latest-compatible lock changes.
- A focused dry-run reports only `spin 0.9.8` to `0.9.9` for the yanked-package
  repair.
- `go list -m -u -json all` reports four direct Go updates.
- Exact manifest-reference scans identify unreferenced Rust workspace
  dependencies and the unused `journal-core` `md5` dependency.
- Current Netdata source establishes the consumer toolchain floors and actual
  packaging compiler.

Affected contracts and surfaces:

- Rust and Go minimum compiler metadata.
- Rust workspace/member manifests and lockfile.
- Go module requirements and sums.
- Cache construction and error types in `journal-engine`.
- Compression encoders/decoders in Rust and Go.
- Hashing, FSS, UUID/random, mmap, async/runtime, and platform helpers.
- Published Rust package metadata and Go module resolution.
- Consumer installation docs, product scope, SOW ledgers, and release skill
  checks.
- Netdata Rust and Go dependency integration.

Existing patterns to reuse:

- Sequential dependency-order Rust package validation from SOW-0140.
- Repository-local Cargo/Go caches and target directories under `.local/`.
- Shared interoperability matrices under `tests/interoperability/`.
- Focused repeated-structured writer benchmark and profiling harnesses from
  SOW-0134/SOW-0135/SOW-0138.
- GitHub wiki validation and verified-example harnesses.
- Scratch-copy consumer validation so external checkouts remain unchanged.

Risk and blast radius:

- High within build graphs and medium within product behavior. Public APIs and
  file-format behavior are intended to remain stable, but cache, compression,
  hashing, and runtime dependencies sit on correctness or performance paths.
- Netdata checkout modification is prohibited. Integration evidence will use a
  repository-local ignored scratch copy.
- The protected historical benchmark checkout is not used or modified.

Sensitive data handling plan:

- Dependency names, versions, public commits, test counts, and synthetic
  benchmark data are non-sensitive.
- Raw logs and generated artifacts stay under ignored `.local/`.
- No credentials, registry tokens, private endpoints, customer data, personal
  data, production logs, or proprietary incident material will be written to
  durable artifacts.

Implementation plan:

1. Establish exact Go `1.26.2` and Rust `1.91` validation toolchains in
   repository-local scratch state without changing system defaults.
2. Update Go minimum metadata and all four direct dependencies; tidy and repair
   any source incompatibility.
3. Remove proven-unused Rust dependency declarations, update actively used
   direct requirements to latest compatible stable releases, and refresh the
   lockfile under the Rust `1.91` compatibility contract.
4. Repair compile/API migrations without changing SDK public or journal-format
   contracts. Stop for a user decision if a dependency requires such a change.
5. Run focused tests after each migration cluster, then full Rust/Go,
   compatibility, interoperability, security, docs, and package gates.
6. Compare current-master and changed dependency graphs with identical cache,
   reader, and repeated-structured-writer benchmarks; profile unexpected
   regressions.
7. Validate local SDK consumption from a scratch copy of current Netdata
   without modifying Netdata source.
8. Bump all release metadata and active consumer docs to `0.8.0`, run package
   dry-runs, review the complete diff, and report the final publication
   checkpoint.

Validation plan:

- Rust exact-toolchain checks on `1.91` and `1.97.1`.
- `cargo fmt --all --check`, workspace tests, focused package tests, all-target
  checks, Clippy where supported, docs, and sequential package dry-runs.
- Go exact-toolchain checks on `1.26.2` and `1.26.5`, `go test ./...`,
  `go vet ./...`, `go mod tidy` cleanliness, and `govulncheck`.
- Raw/structured byte identity across final states.
- Compact, compression, mixed-directory, live-concurrency, FSS, directory,
  journalctl-query, and verifier/interoperability matrices relevant to changed
  dependencies.
- Stock systemd `v260.1` file-backed verification against synthetic fixtures.
- Before/after focused benchmarks with identical parameters and repeated runs.
- Scratch-copy Netdata Go and Rust consumer build/tests.
- Active-doc version sweep, wiki validation, verified examples,
  `git diff --check`, and `.agents/sow/audit.sh`.
- Same-pattern scans for yanked packages, stale direct versions, unused
  manifest dependencies, stale `0.7.9` active docs, and unexpected public API
  changes.

Artifact impact plan:

- AGENTS.md: no update expected; product and workflow contracts are unchanged.
- Runtime project skills: release-tagging skill will be used and updated only
  if validation exposes a durable workflow gap.
- Specs: update product scope for `0.8.0` and aligned compiler minimums.
- End-user/operator docs: update active Rust and Go installation examples and
  any compiler-support statements.
- End-user/operator skills: none expected.
- SOW lifecycle: SOW-0141 remains current until dependency, compatibility,
  integration, benchmark, review-disposition, and package-readiness evidence
  is complete. Publication remains a separate final checkpoint.
- SOW-status.md: both canonical and convenience ledgers track SOW-0141.

Open-source reference evidence:

- `netdata/netdata @ 8005e84c0c14cbee336d5e50ad731067ce326990`
  - `src/go/go.mod`
  - `packaging/check-for-go-toolchain.sh`
  - `.github/workflows/build.yml`
  - `src/crates/Cargo.toml`
- `systemd/systemd @ c0a5a2516d28601fb3afc1a77d7b42fcfe38fced`
  - compatibility tag `v260.1`
  - existing project conformance and verifier baselines apply.

Open decisions:

- The user selected options 1A, 2A, and 3A for the planned implementation.
- The user subsequently selected option 1A for the Foyer migration: remove the obsolete public
  `EngineError::FoyerIo(foyer::IoError)` variant and consolidate Foyer failures
  under `EngineError::Foyer(foyer::Error)`.
- None for implementation. The user approved boxing Foyer `0.22.3` errors so
  the public error variant holds `Box<foyer::Error>` and the existing 64-byte
  `EngineError` compile-time bound remains enforced.
- Immutable tags and registry publication remain deliberately unapproved until
  final evidence is reported.

## Implications And Decisions

1. **Toolchain baselines: option A accepted (long-term-best).**
   - Go module minimum becomes `1.26.2`, with validation on `1.26.2` and
     Netdata packaging Go `1.26.5`; no SDK `toolchain` directive.
   - Rust minimum becomes `1.91`, with validation on exact `1.91` and the
     available current `1.97.1`.
2. **Dependency policy: option A accepted (long-term-best).**
   - Update every actively used direct dependency to the latest non-yanked
     stable release compatible with the selected compiler floor.
   - Refresh transitive resolution and remove proven-unused declarations.
   - Keep the intentional `zerocopy 0.9.0-alpha.0` API line unless evidence
     reveals a newer compatible `0.9` release.
3. **Release sequence: option A accepted (long-term-best).**
   - Do not create `v0.7.9` or `go/v0.7.9`.
   - Prepare `0.8.0` and stop before immutable publication/tag actions for a
     final user checkpoint.
4. **Behavioral design constraint.**
   - Dependency migrations may adapt internal calls but must not change public
     APIs, file-format bytes, indexes, durability, runtime purity, or default
     behavior without a new user decision.
5. **Performance constraint.**
   - A repeatable unexplained regression stops release preparation; it is not
     waived merely because tests pass.
6. **Foyer error API: option A accepted (long-term-best).**
   - Foyer `0.22.3` removed its separate public `IoError` type and now returns
     `foyer::Error` throughout.
   - Remove `EngineError::FoyerIo` and preserve the existing
     `EngineError::Foyer` conversion as the single cache-error path.
   - This is an explicitly accepted public pre-1.0 API change for `0.8.0`.
7. **Foyer error representation: boxed option accepted (long-term-best).**
   - Store the 72-byte Foyer error allocation behind a pointer in
     `EngineError::Foyer`.
   - Preserve automatic `From<foyer::Error>` conversion with a manual
     implementation.
   - Preserve the existing `size_of::<EngineError>() <= 64` invariant; the
     additional allocation occurs only when constructing a Foyer error.
8. **Foyer cache API coupling: existing dependency decision applied.**
   - `FileIndexCache` is a public alias for `foyer::HybridCache`; updating the
     approved Foyer dependency from `0.20` to `0.22` therefore also updates
     that alias's available methods.
   - Treat this as part of the accepted pre-1.0 Foyer migration for `0.8.0`,
     document the exact `0.22` coupling for direct engine consumers, and do
     not add a new Foyer re-export or wrapper API during release preparation.

## Plan

1. Establish compiler-floor validation and dependency inventory.
2. Upgrade Go dependencies and validate Go behavior.
3. Upgrade Rust dependencies in bounded clusters and validate each cluster.
4. Run full correctness, compatibility, security, and documentation gates.
5. Run before/after benchmarks and investigate regressions.
6. Validate scratch-copy Netdata integration.
7. Prepare `0.8.0` metadata/package dry-runs and report the publication
   checkpoint.

## Implementation And Review Plan

Implementation:

- Implementation remains local to this session. Dependency clusters are
  migrated sequentially so failures and performance changes have attributable
  causes.

Reviewers:

- The previous release's review waiver covered only that completed change set.
  Recommend a fresh whole-SOW external review after local validation because
  this SOW changes build graphs and migration code. Do not run external
  reviewers without current user authorization.

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

- Stop immediately if an upgrade requires an unapproved public API,
  file-format, durability, cache-schema, or compatibility change.
- Record and investigate every test, interoperability, security, integration,
  and benchmark failure before continuing.
- Do not publish, tag, force, reset, checkout, or alter external repositories.

## Execution Log

### 2026-07-27

- Recorded the user-selected toolchain, comprehensive dependency, and `0.8.0`
  release-sequence decisions.
- Completed the pre-implementation gate from current SDK, registry, and Netdata
  evidence.
- Installed exact Rust `1.91.0`, Go `1.26.2`, and Go `1.26.5` toolchains under
  ignored repository-local `.local/` state without changing system defaults.
- Updated the Go module floor to `1.26.2`, refreshed all four direct Go
  dependencies, and ran:
  - Go `1.26.2`: `go test -p 4 ./... -count=1` and `go vet ./...`;
  - Go `1.26.5`: `go test -p 4 ./... -count=1`.
  All completed successfully.
- Updated the Rust compiler floor to `1.91`, removed fifteen unreferenced
  external workspace declarations plus the unused `journal-core` `md5`
  inheritance, and refreshed the lock
  graph to the latest Rust-`1.91`-compatible resolution. Cargo locked 97
  packages and removed the yanked `spin 0.9.8` path.
- Repaired source-compatible Rand `0.10` and RustCrypto HMAC `0.13` trait
  imports.
- Exact Rust `1.91` all-target checking reached the Foyer migration and stopped:
  Foyer `0.22.3` replaced builder objects with configuration objects and
  consolidated `IoError` into `foyer::Error`. The builder migration is
  internal, but removing or changing public
  `EngineError::FoyerIo(foyer::IoError)` requires the user decision mandated
  by the behavioral design constraint.
- The user selected option 1A: remove `EngineError::FoyerIo` and use the
  existing `EngineError::Foyer` variant for all Foyer `0.22.3` failures.
- Applied the approved Foyer builder/error migration. The next exact Rust
  `1.91` check reached the existing
  `size_of::<EngineError>() <= 64` assertion and stopped because
  `foyer::Error` is now 72 bytes. A repository-local probe confirmed a
  representative inline error enum is 72 bytes and its boxed equivalent is 24
  bytes; implementation paused for the required representation decision.
- The user approved the boxed representation after clarification that the
  72-byte Foyer value remains intact on the heap while `EngineError` stores a
  pointer and continues enforcing its at-most-64-byte invariant.
- Exact Rust `1.91` and Go `1.26.2` closed-file matrix validation exposed the
  pre-existing strict-identity regression already tracked by SOW-0136:
  `rust/src/internal/testcmd/livewriter/src/main.rs` still supplies neither a
  machine ID nor boot ID to the strict high-level directory writer. A rebuild
  in the default `rust/target` directory confirmed this is not a stale-binary
  or dependency-migration failure.
- The user selected surgical option 1A: pause this SOW, reopen the originating
  SOW-0115 regression, repair every equivalent internal test-tool call site
  with fixed synthetic identity, validate it, and then resume this SOW.
- SOW-0115 regression repair completed. The shared Rust directory livewriter
  and conformance adapter now supply fixed synthetic machine/boot IDs.
  Closed-file 32/32, binary 18/18, compression 72/72, compact 80/80, and full
  live feature 18/18 gates passed, along with the adapter case, runtime purity,
  Rust `1.91`/`1.97.1` focused checks, formatting, whitespace, and SOW audit.
  SOW-0136 closed as superseded and this dependency SOW resumed.
- The mixed-directory matrix then exposed a second pre-existing harness-only
  regression: Go and Rust correctly print the stock-compatible non-quiet
  `--list-boots` table header, while the old mixed harness counts that header
  as a boot. The user selected option 1B: reopen originating SOW-0121, normalize
  table headers, and compare exact boot rows against stock systemd rather than
  retain the weak count-only assertion.
- SOW-0121 regression repair completed. The mixed-directory harness now strips
  only the exact table header, compares complete normalized boot rows against a
  complete stock oracle, and rejects incomplete stock oracles. The
  full-feature systemd 255 matrix passed 42/42; tagged v260.1 directory and
  query matrices passed; focused Go `1.26.2` and Rust `1.91` journalctl tests
  passed. The local tagged v260.1 build reports `-XZ -GCRYPT`, so its expected
  inability to serve as the full mixed-format oracle is recorded rather than
  misclassified as an SDK failure. This dependency SOW resumed.
- Completed identical baseline-versus-refreshed dependency benchmarks from
  commit `d0dce35cfa7020b03c5cb3971e23a8df4481cd18`:
  - Rust repeated-structured compact writer median changed from
    139,536.957 to 140,452.120 rows/s, a 0.656% improvement, with every paired
    output byte-identical.
  - The initial unpinned Go writer result was noisy and was rejected as
    evidence. A CPU-80-pinned 200,000-row rerun changed from 133,033.478 to
    134,190.943 rows/s, a 0.870% improvement, with every paired output
    byte-identical.
  - Reader medians changed by -0.484%, -2.038%, and -2.726% for the three Rust
    paths, and +0.593% and +0.791% for the two Go paths. Every paired checksum
    matched against the same 128 MiB fixture.
  - Foyer cache first-build and miss medians changed by +0.975% and +3.155%.
    Reopen and hit percentages were +9.694% and +3.695%, but the absolute
    median increases were only about 52 and 33 microseconds. Both versions
    produced the expected miss result and successfully reopened the cached
    hit. A separate scratch probe also reopened a Foyer `0.20.1` cache with
    `0.22.3`.
- Completed dependency freshness checks on the selected compiler floors:
  root and nested Rust `cargo update --dry-run --verbose` each lock zero
  changes; Go reports no update for any of its four direct dependencies.
- Prepared the `0.8.0` candidate metadata. All eight Rust packages and their
  internal dependency requirements now identify `0.8.0`; active Rust and Go
  installation examples identify `0.8.0`; the product spec and getting-started
  guide record Rust `1.91` and Go `1.26.2` minimums.
- Validated the current Netdata consumer at
  `netdata/netdata @ 8005e84c0c14cbee336d5e50ad731067ce326990`
  from an ignored repository-local archive. The original checkout remained
  clean and read-only. Go `1.26.5` targeted SNMP trap packages passed. Rust
  `1.91` selected all seven local `0.8.0` packages and passed the NetFlow
  package: 573 unit tests and one integration test passed, 26 manual tests were
  ignored, and none failed.
- The first scratch Rust attempt demonstrated why a version-matching consumer
  requirement is part of integration evidence: a `0.7.8` lock ignored local
  `0.7.9` patches. That non-candidate build was stopped, the scratch-only
  requirements were set to the candidate version, and the successful run
  explicitly compiled all seven local `0.8.0` packages.
- Final exact-toolchain validation passed:
  - Rust `1.91.0` and `1.97.1`: root formatting plus 484 root tests and 17
    nested-JF tests on each toolchain;
  - Go `1.26.2` and `1.26.5`: complete `go test ./...` and `go vet ./...` on
    each toolchain;
  - wiki validation: 15 pages; verified examples: 31/31; docs harness tests:
    77/77.
- Package readiness passed for all eight `cargo package --list` gates. The
  common crate completed `cargo publish --dry-run --allow-dirty --locked`.
  The registry crate then stopped only because crates.io does not yet contain
  common `0.8.0`, confirming the expected sequential-publication constraint.
  No registry upload, tag, push, or release was performed.
- Static review found two remaining unused internal workspace dependency
  declarations. The same-pattern scan then found four more root workspace
  declarations without member inheritance and 27 additional package
  declarations without a source or feature reference. The final cleanup
  removes 21 root workspace declarations and 28 package declarations in total;
  the subsequent scan reports zero remaining candidates. Exact Rust `1.91`
  and `1.97.1` all-target checks and 484-test root workspace runs pass; the
  nested-JF workspace passes all-target checks and 17 tests on both toolchains.
- A final freshness rerun selected newly available transitive build dependency
  `toml_parser 1.1.3+spec-1.1.0` through `cbindgen`; the lock was refreshed and
  the next Cargo dry-run selected zero changes.
- Post-change security validation passed:
  - Go `1.26.5` `govulncheck ./...`: no vulnerabilities found;
  - OSV queried all 246 locked Rust registry package/version pairs. It found no
    vulnerability advisory and three informational unmaintained-crate
    advisories: `bincode 1.3.3` and `paste 1.0.15` through Foyer, and
    `fxhash 0.2.1` through `hashers`. These are transitive requirements of the
    selected current direct dependencies.
- Reanalyzed the reader results as paired baseline/changed measurements rather
  than comparing only aggregate medians. Every case's 95% bootstrap interval
  includes zero, and exact two-sided sign-test p-values range from `0.3438` to
  `1.0000`; the raw changes therefore do not establish a repeatable reader
  regression.
- Corrected review-confirmed release records: SOW-0121's completed regression
  status, stale Go directive text in both ledgers, the two unused internal
  declarations, and direct-engine documentation for the public Foyer `0.22`
  coupling.
- Post-review package/docs gates pass: 8/8 `cargo package --list`, the common
  crate's full publish dry-run, 15 wiki pages, and 31/31 executable examples.
  `cargo fmt --all --check`, `git diff --check`, and the SOW audit also pass.

## Validation

Acceptance criteria evidence:

- `go/go.mod` declares Go `1.26.2` without a `toolchain` directive and uses the
  latest direct compatible compression/platform releases.
- `rust/Cargo.toml` and the nested JF workspace declare Rust `1.91`. The active
  Rust direct dependencies are current for that compiler contract; the sole
  pre-release requirement remains the intentional newest `zerocopy 0.9` API
  line.
- Twenty-one unused root workspace declarations and 28 unused package
  dependency declarations were removed. Cargo root/nested dry-runs now select
  zero changes; the lock contains no yanked or unavailable registry packages.
- Public/file-format/default behavior is unchanged except for the explicitly
  accepted pre-1.0 removal of `EngineError::FoyerIo` and the boxed payload type
  of `EngineError::Foyer`, plus the public `FileIndexCache` alias following
  Foyer `0.22` methods instead of `0.20`. `EngineError` remains at most 64
  bytes.
- Full compiler-floor/current-toolchain, compatibility, Netdata integration,
  package, docs, and performance evidence below passes.
- The candidate identifies `0.8.0` consistently. No immutable action was
  performed.

Tests or equivalent validation:

- Rust exact `1.91.0` and `1.97.1`, each:
  - `cargo fmt --all --check`: PASS;
  - root `cargo check --workspace --all-targets`: PASS;
  - root `cargo test --workspace --all-targets`: 484 passed, zero failed;
  - nested JF all-target check/tests: 17 passed, zero failed.
- Go exact `1.26.2` and `1.26.5`, each:
  - `go test ./...`: PASS;
  - `go vet ./...`: PASS;
  - exact-floor `go mod tidy`: clean.
- Go `1.26.5` `govulncheck ./...`: PASS, no vulnerabilities found.
- Rust OSV lock scan: 246 registry package/version pairs queried; no
  vulnerability advisory. The three returned records are informational
  unmaintained-crate advisories for transitive `bincode`, `fxhash`, and
  `paste`.
- Rust/Go/systemd closed-file identity: 32/32 PASS.
- Binary matrix: 18/18 PASS; compression: 72/72 PASS; compact: 80/80 PASS;
  live: 18/18 PASS; mixed directory: 42/42 PASS; lock: 4/4 PASS.
- Tagged systemd v260.1 directory and query matrices: PASS with no failures.
- Full-feature systemd 255 verifier matrix: 63/63 PASS across 9 positive
  formats, 12 negative corruptions, and stock/Go/Rust verification.
- Dataset byte identity against tagged systemd v260.1: online, offline, and
  archived outputs are each Rust/Go/systemd byte-identical; top-level
  `all_equal` is true and all DATA hash chains retain depth three.
- Documentation: 15 wiki pages valid, 31/31 executable examples pass, and
  77/77 documentation harness tests pass.
- Packaging: 8/8 package listings pass; common `0.8.0` full publish dry-run
  passes; the next crate reports only the expected unpublished-common
  dependency-order constraint.
- `git diff --check`: PASS.

Real-use evidence:

- Current Netdata scratch-copy integration:
  - Go `1.26.5` SNMP trap collector packages using the local Go module: PASS;
  - Rust `1.91` NetFlow package using all seven local `0.8.0` crates: 574 tests
    passed, 26 ignored, zero failed.
- The source checkout remained clean at
  `8005e84c0c14cbee336d5e50ad731067ce326990`; only ignored SDK-local scratch
  state was written.
- Performance evidence is workload- and host-specific. It establishes no
  universal capacity claim: the Rust and pinned Go writer medians improved,
  reader checksum identity held, and cache timing changes were small in
  absolute terms. The reader aggregate-median range was -2.726% to +0.791%,
  but paired bootstrap intervals all include zero and sign tests do not
  establish a repeatable change.

Reviewer findings:

- Authorized static reviewers were run as one whole-SOW batch before the
  nonbehavioral finding-fix cleanup:
  - GLM, DeepSeek, and Qwen returned `PRODUCTION GRADE` with P3-only notes.
  - Claude returned `NEEDS CHANGES` with two P2 and five P3 findings.
  - Codex could not inspect the current worktree because its local sandbox
    setup failed and its repository mirror predates this SOW. Its reported P2
    is a review-coverage failure, not a source finding; no verdict is inferred.
  - MiniMax produced no output and timed out on both the original invocation
    and the single permitted retry. No verdict is inferred.
  - MiMo was not run, as explicitly directed by the user.
- Claude P2 `unrecorded_public_api_break_fileindexcache_foyer`: verified.
  `FileIndexCache` is a public alias to Foyer's `HybridCache`; the accepted
  `0.20` to `0.22` migration changes directly callable cache methods. The SOW
  now records that pre-1.0 coupling and the lower-level Rust package guide
  documents the required Foyer `0.22` method/type line. No new public re-export
  or wrapper was added.
- Claude P2 `missing_post_change_vulnerability_scan_evidence`: verified.
  Post-change `govulncheck` reports no Go vulnerability. The Rust OSV scan
  covers every locked registry package/version pair and reports no security
  vulnerability; its three results are informational unmaintained-crate
  advisories. Current GitHub code scanning and Codacy cannot analyze an
  unpushed commit, so the branch must be pushed and those results must be
  captured and dispositioned before any tag or registry publication.
- Claude P3 findings:
  - repaired the completed SOW-0121 regression status;
  - removed the two reported unused internal workspace declarations, then
    extended the same-pattern scan to all manifests and removed every remaining
    source/feature-unreferenced declaration;
  - corrected stale Go directive wording in both ledgers;
  - documented the Foyer coupling with the P2 fix;
  - rejected the claimed repeatable reader regression: the review cited raw
    aggregate medians, while paired analysis shows every 95% bootstrap interval
    includes zero and every sign test is nonsignificant.
- DeepSeek P3 `base64-023-deprecated-engine-trait`: rejected as factually
  incorrect. Base64 `0.23.0` does not deprecate `Engine`; the source continues
  to define the trait and its documentation imports it for method resolution.
  An exact Rust `1.91` check produces no deprecation warning.
- All remaining reviewer P3s are nonblocking and require no code change:
  pre-existing legacy-JF unused import/Clippy/lockfile observations, normal
  transitive version duplication and pre-release selection, the accepted
  Foyer error display/API change, and expected pre-tag `0.8.0` documentation.
- No verified P0/P1/P2 source or release-record finding remains. Codex and
  MiniMax are explicitly recorded coverage gaps rather than invented verdicts.

Same-failure scan:

- Active docs/specs/manifests contain no stale `0.7.8` or `0.7.9` consumer
  release reference.
- Root and nested Cargo update dry-runs select zero changes; the direct Go
  module report shows no available update.
- Registry metadata scan found no yanked/missing locked package. Manifest
  reference scanning found no member inheritance or source use for any removed
  dependency declaration.
- The final all-manifest same-pattern scan reports zero package dependencies
  without a source/feature reference and zero root workspace dependencies
  without member inheritance.
- Both SIGBUS handler copies use the Rust-`1.97`-accepted function-pointer cast.
- Both HMAC call sites import the required `KeyInit` trait.
- The only Clippy failure remains the pre-existing nested legacy JF
  `clippy::mut_from_ref` deny in `journal_file/src/file.rs`; this SOW neither
  introduced nor concealed it.

Sensitive data gate:

- Planning evidence contains only public repository paths, versions, commits,
  and synthetic-test policy. No sensitive data is present.

Artifact maintenance gate:

- AGENTS.md: no update; workflow and runtime contracts did not change.
- Runtime project skills: no update; release-tagging and docs-authoring rules
  remain accurate.
- Specs: product scope identifies `0.8.0`, Rust `1.91`, and Go `1.26.2`.
- End-user/operator docs: active Rust and Go installation examples identify
  `0.8.0`; Getting Started records both compiler minimums; the lower-level
  Rust package guide records the Foyer `0.22` coupling exposed through the
  engine crate.
- End-user/operator skills: none expected.
- SOW lifecycle: current/in-progress until the final user release checkpoint
  and post-push remote scanning gate.
- SOW-status.md: both ledgers continue to identify SOW-0141 as current.

Specs update:

- `.agents/sow/specs/product-scope.md` now records `0.8.0` and the selected
  minimum compilers.

Project skills update:

- None required.

End-user/operator docs update:

- `README.md`, `rust/README.md`, `docs/Getting-Started.md`, `docs/Go-API.md`,
  `docs/Rust-API.md`, and `docs/Rust-Crates-And-Packages.md` identify the
  coordinated `0.8.0` candidate.

End-user/operator skills update:

- None expected.

Lessons:

- Cargo patches do not override a lock to a semver-incompatible package
  version. Consumer integration must prove that the candidate package version
  was actually selected, not merely that a patch table exists.
- For noisy sub-two-second writer runs, CPU-pinned interleaving and more pairs
  are required before classifying a dependency regression.
- Percentage changes on sub-millisecond cache reopen/hit operations need the
  absolute time delta beside them to avoid overstating practical impact.
- A stock systemd binary's feature list is part of oracle validity. The local
  v260.1 build remains valid for byte/directory/query evidence but cannot
  replace the full-feature systemd 255 XZ/FSS oracle.

Follow-up mapping:

- Sequential Rust publication, root/Go tags, branch push, and registry
  verification remain final-checkpoint actions in this SOW, not follow-up code
  work.

## Outcome

The local `0.8.0` dependency-refresh candidate is implemented and validated.
All correctness, compatibility, integration, documentation, packaging-list,
and performance gates pass. The only blocked dry-run is the expected crates.io
dependency order after the common crate. External-review findings are
dispositioned with the documented Codex/MiniMax coverage gaps. The user's
explicit immutable-release checkpoint and post-push remote scanning remain
outstanding.

## Lessons Extracted

- Exact consumer toolchain metadata must be validated against the consumer
  build, not inferred from nominal language compatibility.
- Dependency refresh benchmarks need stable pinning before a noisy initial
  result can be treated as causal.
- Candidate integration must verify selected package versions in build output;
  a configured local override alone is insufficient evidence.

## Followup

- None for code. Publication proceeds sequentially only after explicit user
  approval and clean/dispositioned post-push GitHub code-scanning and Codacy
  results.

## Regression Log

None yet.
