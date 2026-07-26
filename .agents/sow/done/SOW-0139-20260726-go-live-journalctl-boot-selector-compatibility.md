# SOW-0139 - Go Live Journalctl Boot-Selector Compatibility

## Status

Status: completed

Sub-state: the cross-version test-only repair, same-pattern expansion,
documentation synchronization, validation, and review gates are complete.

## Requirements

### Purpose

Restore the three low-level Go live-writer tests that invoke stock
`journalctl --file --follow --no-tail --boot=all` on systemd 255 without
weakening their live-visibility, interruption, sequence, or stock-reader
coverage.

### User Request

SOW-0138 validation found three full-suite failures outside its changed path.
The same commands fail identically at the unchanged SOW-0138 base commit, so
the Go writer optimization must not absorb an unrelated harness repair.

### Assistant Understanding

Facts:

- `TestGoWriterLiveStockReaders`,
  `TestGoWriterLiveStockReadersStress`, and
  `TestGoWriterLiveInterruptionReopenAndVerify` fail on systemd
  `255.4-1ubuntu8.16`.
- Stock `journalctl` reports
  `No journal boot entry found from the specified boot offset (+0).`
- The Go writer exits with the expected status, and the stock libsystemd
  reader can observe the expected rows where the test reaches that assertion.
- The same three tests and error reproduce from unchanged commit
  `438ab0350488a41611de2d53b6eacd90b4f726be`.
- The shared direct-file live interoperability matrix passes for Go regular,
  compact, compressed-compact, and sealed writers on the same host.

Inferences:

- The failure is tied to the test harness's stock `journalctl` boot selector,
  not to SOW-0138 DATA-link publication.
- A repair must preserve the stronger live-follow oracle rather than merely
  dropping stock-reader coverage.

Unknowns:

- Whether the local systemd 260.1 source build has every dependency needed for
  executable cross-version validation. Missing dependencies will not be
  installed without a separate user decision.

### Acceptance Criteria

- The user selects the test-only compatibility design before implementation.
- All three affected Go live tests pass on systemd 255 and the project's
  systemd 260.1 compatibility target.
- The tests retain stock `journalctl` follow coverage, stock libsystemd
  coverage, ordered-prefix checks, final row checks, interruption/reopen
  behavior, and stock verification.
- Production writer APIs, boot identity, file bytes, and reader semantics
  remain unchanged.
- A same-pattern search covers every committed `--file --follow` stock-reader
  invocation that combines boot selection with synthetic journals.

## Analysis

Sources checked:

- `go/journal/live_concurrency_test.go`
- `tests/conformance/live/run_live_concurrency.py`
- `.agents/sow/current/SOW-0138-20260726-go-writer-entry-link-state-reuse.md`
- Final and base-commit failure logs under
  `.local/validation/sow0138/`

Current state:

- The applicable Go suite passes when the three independently reproduced
  host-journalctl cases are excluded.
- The shared Go live matrix passes all SOW-0138 production-path variants.

Risks:

- Removing boot selection without proving equivalent follow behavior can hide
  real current-boot filtering bugs.
- Version-specific branching can make the harness harder to reason about.
- Changing synthetic boot identity can invalidate cross-version evidence or
  weaken interruption coverage.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:

- systemd 255 parses explicit `--boot=all` as disabling the boot selector, but
  its later follow-mode defaulting sees the resulting false boolean and
  re-enables current-boot filtering. The synthetic low-level writer fixture
  intentionally has no `_BOOT_ID` DATA field, so the re-enabled lookup fails
  with boot offset `+0`.
- systemd 260.1 uses a tri-state boot-selection variable. Explicit
  `--boot=all` resolves it to false and later follow-mode defaulting no longer
  re-enables the filter.
- Adding `--merge` to the explicit-file follow invocation suppresses the
  implicit follow-mode current-boot default on systemd 255. With one explicit
  `--file`, it does not broaden the input source, and systemd 260.1 also keeps
  the boot selector disabled.

Evidence reviewed:

- Identical final-tree and base-commit failures plus the passing shared live
  matrix described above.
- `systemd/systemd @ db11bab38ccf` (`v255`):
  `src/journal/journalctl.c:257-264`, `src/journal/journalctl.c:1078-1082`,
  and `src/journal/journalctl.c:1268-1305`.
- `systemd/systemd @ c0a5a2516d28` (`v260.1`):
  `src/journal/journalctl.c:140-152` and
  `src/journal/journalctl.c:1082-1090`.
- A scratch-only `/tmp` experiment added `--merge` beside `--boot=all` and
  passed 60/60 ordered entries through both stock journalctl poll and follow
  readers on systemd `255.4-1ubuntu8.16`, followed by stock verification.

Affected contracts and surfaces:

- Go live-concurrency tests and their shared stock-journalctl helper.
- Test compatibility across systemd 255 and 260.1.
- No production SDK behavior.

Existing patterns to reuse:

- Shared live matrix polling and libsystemd readers.
- Deterministic synthetic machine and boot IDs.
- Historical systemd-version compatibility validation.

Risk and blast radius:

- Low when confined to test tooling, but high evidentiary risk if the repair
  weakens the stock live-reader oracle.

Sensitive data handling plan:

- Use only deterministic synthetic journal identities and repository-local
  fixtures.

Implementation plan:

- Keep `--boot=all` as the explicit no-boot-filter request.
- Add unconditional `--merge` to the shared stock journalctl follow command.
- Do not add version parsing, change synthetic journal fields, or modify any
  production writer/reader code.
- Search every committed explicit-file stock follow invocation for the same
  pattern.

Validation plan:

- Run the three focused tests and the complete Go journal suite on systemd 255.
- Run the same cases on systemd 260.1.
- Run regular, compact, compressed, sealed, and interrupted live matrices with
  stock journalctl, stock libsystemd, Rust, and Go readers.
- If executable systemd 260.1 validation needs unavailable software, stop
  before installation and report the exact missing dependency.

Artifact impact plan:

- Keep AGENTS.md, runtime skills, product specs, and SDK CLI examples
  unchanged for a test-only repair.
- Synchronize the live-harness README, adapter contract, and language README
  validation inventories with the repaired stock-reader invocation.
- Update SOW ledgers on activation and closure.

Open-source reference evidence:

- `systemd/systemd @ db11bab38ccf` (`v255`).
- `systemd/systemd @ c0a5a2516d28` (`v260.1`).
- The user explicitly authorized a normal upstream clone at
  the requested local checkout for cross-version validation. This is the sole
  repository-boundary write exception for this SOW; build copies and generated
  evidence remain under this repository's ignored `.local/` tree or `/tmp`.

Open decisions:

- Implementation has no open design decision.
- Validation has no open tooling decision. The user authorized Ubuntu's
  packaged `gperf` build dependency after Meson identified it as missing.
- The user selected the long-term repair for the post-review same-pattern
  finding: add `--merge` to all five query-matrix follow cases that explicitly
  request `--boot=all`. Specific-boot and implicit-boot cases remain unchanged.

## Implications And Decisions

1. Cross-version follow invocation
   - Selected: retain `--boot=all` and add unconditional `--merge` for the
     explicit-file stock journalctl follow reader.
   - Rejected: systemd-version parsing, because it adds brittle branching.
   - Rejected: adding `_BOOT_ID` to the low-level writer fixture, because that
     changes the fixture rather than repairing the stock-reader invocation.
   - Rejected: removing stock follow coverage, because it would weaken the
     live-publication oracle.
2. External source checkout
   - Selected: clone the official `systemd/systemd` repository as a normal
     local working checkout.
   - Constraint: do not modify the upstream checkout; any patched build copy
     belongs under the SDK repository's ignored `.local/` tree.
3. Equivalent query-matrix follow cases
   - Selected: add `--merge` to all five follow cases that explicitly request
     `--boot=all`, including cases whose initial entries currently mask the
     systemd 255 defaulting behavior.
   - Rejected: repairing only the three currently failing empty-initial cases,
     because equivalent no-boot-filter requests should use one invocation.
   - Rejected: documenting the query matrix as systemd-260.1-only, because the
     committed stock-reader matrix can remain compatible with systemd 255
     without production behavior changes.

## Plan

1. Clone and verify the approved upstream systemd reference checkout.
2. Inventory all equivalent stock live-reader invocations.
3. Implement the selected shared-harness compatibility repair.
4. Rerun focused, full-suite, and cross-version live validation.
5. Review the complete SOW and close it before release preparation starts.

## Implementation And Review Plan

Implementation:

- Local implementation in this session, limited to the shared live test
  harness and its direct tests if required.

Reviewers:

- Run the already authorized external reviewer set on the completed SOW after
  local validation and before GitHub submission, using the system-wide
  `external-reviewers` skill.

Failure handling:

- Stop if a proposed repair weakens stock live-follow coverage or changes
  production SDK behavior.

## Execution Log

### 2026-07-26

- Created as SOW-0138's explicit mapping for three pre-existing systemd 255
  stock-journalctl test failures.
- Activated after the user selected the long-term release-gate repair.
- Recorded the systemd 255 boolean-defaulting root cause, systemd 260.1
  tri-state behavior, the successful scratch `--merge` experiment, and the
  selected shared-harness design.
- Recorded explicit authorization for a normal local `systemd/systemd`
  checkout for cross-version validation.
- Cloned and verified the official upstream checkout. It is clean on
  `main`, uses the official `systemd/systemd` remote, and resolves `v255` and
  `v260.1` to the commits recorded above.
- The initial same-pattern inventory found the shared stock follow command and
  classified the separate query-matrix cases as indexed-boot fixtures. Later
  executable systemd 255 validation corrected that classification for three
  cases whose reader starts before the first indexed boot entry is appended.
- Added `--merge` beside `--boot=all` in the shared stock follow command and
  updated the live-harness README invocation.
- Focused systemd 255 tests passed:
  `TestGoWriterLiveStockReaders`,
  `TestGoWriterLiveStockReadersStress`, and
  `TestGoWriterLiveInterruptionReopenAndVerify`.
- The complete Go module suite passed on systemd 255.
- A four-feature live matrix passed every Go-writer case with Go, Rust, stock
  journalctl, and stock libsystemd readers. Three Rust-writer cases hit the
  already tracked SOW-0136 explicit-identity failure; the Rust sealed case
  passed. No SOW-0139 changed path caused those failures.
- The isolated systemd 260.1 matrix copied the approved upstream source under
  ignored `.local/` state and stopped during Meson setup because `gperf` is
  missing. No package was installed and the upstream checkout was not
  modified.
- After explicit approval, installed Ubuntu's `gperf` package
  (`3.1-1build1`). No packages were upgraded and no services or containers
  required restart.
- Built the tagged systemd 260.1 `journalctl` and matrix ingester from
  `c0a5a2516d28601fb3afc1a77d7b42fcfe38fced` with low-priority I/O,
  low-priority CPU scheduling, and Ninja concurrency capped at four jobs.
  The approved upstream checkout remained clean and unchanged.
- The three affected Go live tests passed with the built systemd 260.1
  `journalctl` first in `PATH`.
- External review completed with six `PRODUCTION GRADE` verdicts and one
  `NEEDS CHANGES` verdict. The verified finding was documentation drift in the
  authoritative adapter contract and the Go/Rust validation inventories; the
  exact stock-reader invocation still omitted `--merge`.
- Synchronized the adapter contract and Go/Rust validation inventories. SDK
  journalctl examples remain unchanged because they invoke the repository
  rewrite rather than stock systemd 255.
- The second review round returned six `PRODUCTION GRADE` verdicts. MiMo's
  first response was empty, and its single permitted retry returned analysis
  without an accepted verdict, leaving a MiMo verdict coverage gap.
- Independently checked MiMo's partial same-pattern concern against stock
  systemd 255 in a fresh ignored validation directory. Three query-matrix
  cases that start from an empty journal fail with the same boot-offset `+0`
  error before their appended `_BOOT_ID` entries become visible:
  `follow-live-append-no-tail`, `follow-cursor-file-no-tail`, and
  `follow-directory-no-tail`. The earlier blanket indexed-boot-metadata
  exclusion was therefore incomplete.
- The user selected the all-case repair. Added `--merge` to all five
  query-matrix follow cases that explicitly request `--boot=all`; retained the
  specific `--boot=0` and implicit-boot cases unchanged, and synchronized the
  interoperability README.
- The isolated query-matrix follow slice passed 21/21 stock, Go, and Rust
  reader combinations on both systemd `255.4-1ubuntu8.16` and the tagged
  systemd 260.1 build.
- The complete Go module suite passed again after the all-case query-matrix
  repair.

## Validation

Completed:

- Verified: focused three-test Go live suite on systemd `255.4-1ubuntu8.16`.
- Verified: complete `go test ./... -count=1`.
- Verified: Go writer live matrix for regular, compact, compact-zstd, and sealed
  files with Go, Rust, stock journalctl, and stock libsystemd readers.
- Verified: stock verification and structural checks for all four Go-writer
  matrix cases.
- Verified: repository-wide same-pattern inventory.
- Verified: tagged systemd 260.1 build from
  `c0a5a2516d28601fb3afc1a77d7b42fcfe38fced`; the build report contains no
  discrepancy or observation codes.
- Verified: focused three-test Go live suite with the built systemd 260.1
  `journalctl`.
- Verified: second external-review round from Claude, Codex, GLM, MiniMax,
  DeepSeek, and Qwen; all returned `PRODUCTION GRADE`.
- COVERAGE GAP: MiMo returned no accepted verdict after its one permitted
  retry.
- Verified: the repaired query-matrix follow slice passed 21/21 stock, Go, and
  Rust reader combinations on systemd `255.4-1ubuntu8.16`.
- Verified: the same repaired query-matrix follow slice passed 21/21 combinations
  with the tagged systemd 260.1 `journalctl`.
- Verified: complete post-repair `go test ./... -count=1`.
- Verified: final external review returned valid `PRODUCTION GRADE` verdicts
  from Claude, Codex, GLM, MiniMax, DeepSeek, and MiMo. Qwen returned positive
  technical analysis but no accepted verdict, so its final-round verdict is a
  documented coverage gap.
- Reviewer adjudication:
  - The first-round adapter-contract and language-validation documentation
    drift was verified and repaired.
  - The second-round equivalent query-matrix concern was reproduced on
    systemd 255, accepted by the user, repaired across all five explicit
    `--boot=all` cases, and validated on both target versions.
  - Final-round P3 documentation findings were verified and repaired in the
    adapter contract and interoperability coverage inventory.
  - The proposed pager-end/dmesg same-class extension was rejected for this
    SOW: systemd 255 source shows the boolean re-default at
    `src/journal/journalctl.c:1078-1082` is follow-only, and a scratch static
    run stopped on the pre-existing v260-specific
    `file-pager-end-merge-suppresses-implicit-boot` case because systemd 255
    rejects that combination.
- Same-failure search: all committed stock follow invocations combining
  synthetic inputs with explicit `--boot=all` are repaired. Specific-boot and
  implicit-boot oracles remain unchanged.
- Sensitive-data gate: reviewed every durable changed artifact; no secrets,
  personal names, customer data, private endpoints, or raw external-review
  output are present. The first audit run's three credential-assignment hits
  were false positives caused by `PASS:` labels; the labels were changed to
  `Verified:`.
- SOW audit: `.agents/sow/audit.sh` exited zero after the lifecycle move and
  reported a clean initialization, status/directory consistency, durable
  open-source citations, and no sensitive-data patterns.
- Artifact maintenance gate:
  - `AGENTS.md`: unchanged; no workflow or project-wide guardrail changed.
  - Runtime project skills: unchanged; the journal compatibility workflow did
    not change.
  - Specs: unchanged; no product API, file format, writer behavior, or
    journalctl rewrite behavior changed, and the existing v260 parity spec
    already allows `--boot=all --merge`.
  - End-user/operator docs: Go and Rust validation inventories were updated;
    SDK CLI examples remain unchanged because they invoke repository rewrites,
    not stock systemd 255.
  - Conformance docs: the adapter contract, live-harness README, and
    interoperability README now describe the exact repaired invocations.
  - End-user/operator skills: none exist and no output/reference skill changed.
  - SOW lifecycle and ledgers: status, directory, canonical ledger, and root
    convenience ledger are synchronized.

## Outcome

The shared low-level live harness and every equivalent query-matrix follow case
now suppress systemd 255's erroneous implicit current-boot re-default when
`--boot=all` was explicitly requested. The change is confined to test tooling
and documentation; production SDK behavior, journal bytes, indexes, APIs, and
reader/writer semantics are unchanged.

The three release-gating Go tests, the complete Go module suite, and 21/21
stock/Go/Rust follow combinations pass on systemd 255. The same focused tests
and 21/21 follow combinations pass with tagged systemd 260.1.

## Lessons Extracted

- Explicit selector parsing and later default normalization must be reviewed
  together; checking only option parsing missed systemd 255's follow-mode
  re-default.
- Indexed metadata eventually appended to a live file does not protect an
  empty-initial reader: boot selection can fail before the first entry becomes
  visible.
- Same-pattern searches must inspect startup state and timing, not only final
  fixture contents.

## Followup

- SOW-0136 already tracks the unrelated Rust directory-harness explicit
  identity failures observed during the broader live matrix.
- No new product or compatibility follow-up is required for this repair.
  Intermediate systemd 256-259 executables were not run, but the two relevant
  implementation classes were verified from source and executable endpoint
  tests cover systemd 255 and 260.1.
- Full static query-matrix compatibility with systemd 255 is explicitly not a
  project contract: the matrix contains v260-specific parser behavior that
  stock systemd 255 rejects. That is not deferred work for this SOW.
