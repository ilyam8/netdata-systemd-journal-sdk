# SOW-0138 - Go Writer Entry-Link State Reuse

## Status

Status: completed

Sub-state: implementation, validation, benchmark/profile comparison, external
review, follow-up mapping, and the SOW move are complete. Ready to commit with
the implementation.

## Requirements

### Purpose

Remove redundant DATA link-state reconstruction from the pure-Go journal
writer, add an exact Go repeating-DATA diagnostic, and preserve journal bytes,
indexes, public APIs, publication order, portability, and default semantics.

### User Request

After the equivalent Rust optimization completed, the user requested the same
causal improvement for Go. The user selected the long-term-best internal design:

- extend the exact repeating-256 benchmark to Go;
- carry resolved DATA link metadata only for the current entry;
- reuse compact tail metadata and combine mapped validation/mutation where
  safe;
- preserve Go's existing default duplicate elimination;
- add no public trusted-unique option.

The user further decided that a correct, structurally cleaner implementation
may remain when the measured result is neutral within noise. A consistent
measured regression, index discrepancy, byte delta, or compatibility failure
still rejects the production change. Benchmark support remains useful
independently.

### Assistant Understanding

Facts:

- `Writer.findData()` reads and parses the matched DATA header, including
  `entryArrayOffset` and `nEntries`, then returns only the object offset:
  `go/journal/writer_objects.go:205-235`.
- `Writer.collectEntryItems()` stores only `offset` and `hash` in each
  `entryItem`: `go/journal/writer.go:491-503` and
  `go/journal/writer.go:642-645`.
- `Writer.linkDataToEntry()` rereads the same DATA header before selecting the
  first-link, promotion, or steady append path:
  `go/journal/writer_arrays.go:185-197`.
- Compact steady append separately rereads the DATA compact-tail fields before
  validating the target offset-array:
  `go/journal/writer_arrays.go:251-269`.
- On Unix, Go maps the complete arena and small reads/writes use direct mapped
  slices: `go/journal/mmap_unix.go:68-99`. Go therefore does not have Rust's
  single-active-window switching cost, so the Rust percentage improvement must
  not be projected onto Go.
- The Go writer always sorts and removes duplicate DATA offsets:
  `go/journal/writer.go:502-503` and
  `go/journal/writer_compression.go:67-75`. It has no public trusted-unique
  option and this SOW adds none.
- The exact `netflow-v5-repeating-256` workload currently exists only in the
  Rust benchmark driver. The shared runner rejects it for Go:
  `tests/benchmarks/run_writer_core_benchmarks.py:137-176` and
  `tests/benchmarks/run_writer_core_benchmarks.py:198-204`.
- The preliminary current-Go mixed-cardinality run is recorded at
  `.local/benchmarks/sow0138-preliminary-go-mixed/`. Five measurements ranged
  from 19,879.721 to 44,176.179 rows/s with a 26,042.152 median. This spread is
  too noisy and the workload shape is wrong, so it is not acceptance evidence.

Inferences:

- Carrying the matched header state into the current `entryItem` removes one
  DATA-header parse for every reused unique payload.
- Carrying compact tail state removes a separate DATA compact-tail reread during
  publication.
- A direct mapped tail helper can validate the offset-array header and write a
  non-full target slot through one mapped span. Portable non-mmap builds must
  retain the existing read/write fallback.
- The gain will probably be smaller than Rust's because Go already uses a
  whole-file mmap without typed-object guards or window switching. This is a
  working theory until identical before/after evidence exists.

Unknowns:

- The magnitude of the Go improvement.
- Whether the targeted functions remain visible in a useful before profile
  after the exact workload is added.
- Whether runtime noise requires more repetitions or an interleaved runner.

### Acceptance Criteria

- The Go benchmark driver produces the exact repeating-256 payload corpus:
  29 application fields, 450 application logical bytes, one fixed synthetic
  boot DATA payload, 30 ENTRY items, 491 total logical DATA bytes, and 256
  repeating identities.
- Raw and structured benchmark modes produce byte-identical journals for the
  exact workload.
- The shared runner accepts the focused workload for Go and Rust, continues to
  reject unsupported systemd combinations, and validates driver metadata
  without language-specific error text.
- Existing DATA lookup returns private resolved link state with the matched
  offset; no public writer API changes.
- New DATA objects receive empty link state.
- `entryItem` retains resolved state only until the current entry is published,
  after which the existing scratch slice is cleared and reused.
- Go's existing sorting and default duplicate elimination remain unchanged.
- Link publication uses the snapshot for first-link, promotion, and later-link
  selection without rereading the DATA header.
- Compact tail state is validated before mutation. Invalid or inconsistent
  tail metadata uses the authoritative array-chain fallback.
- On Unix, a valid non-full compact tail is validated and mutated through one
  mapped span where safe. Non-Unix behavior retains bounded `ReadAt`/`WriteAt`
  fallback semantics.
- Offset-array content and links are written before the final DATA count/tail
  publication, preserving existing crash and live-reader ordering.
- Compact and regular bytes, DATA inverted indexes, indexed query results,
  compression behavior, FSS behavior, reopen behavior, public APIs, live
  publication, and default semantics remain unchanged.
- The production change is retained when correctness and compatibility pass
  and performance improves or is neutral within measurement noise.
- A repeatable regression, byte/index mismatch, or compatibility failure
  rejects the production change.
- Results are described as workload- and host-specific, not as universal Go,
  SDK, or NetFlow capacity.

## Analysis

Sources checked:

- `AGENTS.md`
- `.agents/skills/project-agent-orchestration/SKILL.md`
- `.agents/skills/project-journal-compatibility/SKILL.md`
- `.agents/sow/specs/product-scope.md`
- `.agents/sow/done/SOW-0062-20260530-rust-go-writer-absolute-performance.md`
- `.agents/sow/done/SOW-0134-20260726-rust-writer-data-link-performance.md`
- `.agents/sow/done/SOW-0135-20260726-rust-writer-entry-link-state-reuse.md`
- `go/journal/format.go`
- `go/journal/mmap_unix.go`
- `go/journal/mmap_other.go`
- `go/journal/writer.go`
- `go/journal/writer_arrays.go`
- `go/journal/writer_compression.go`
- `go/journal/writer_objects.go`
- `go/internal/testcmd/writer_core_bench/main.go`
- `tests/benchmarks/README.md`
- `tests/benchmarks/run_writer_core_benchmarks.py`
- `tests/benchmarks/test_run_writer_core_benchmarks.py`
- `tests/benchmarks/test_report_benchmarks.py`

Current state:

- Go's lookup/header access is cheaper than Rust's former path, but the writer
  still parses the same DATA header once during lookup and again during
  publication.
- Compact publication also rereads eight DATA-tail bytes already owned by the
  matched object.
- The exact causal workload cannot currently run against Go.

Risks:

- Stale state can append to the wrong array slot or publish an incorrect DATA
  entry count.
- Mutating through a mapped span before completing validation can corrupt an
  invalid journal rather than falling back or returning an error.
- Changing array-before-DATA publication order can expose incomplete inverted
  indexes to live readers.
- Compact and regular formats store different tail metadata.
- Reopen and successor writers must never retain offsets from another file.
- Extending the benchmark must not silently change the historical default
  corpus or imply that Go supports a trusted-unique public option.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:

- DATA lookup obtains the metadata needed to publish the current entry's
  inverted-index link, but discards it.
- Publication reconstructs that state by rereading the DATA header and compact
  tail.
- Go's mmap makes each individual access cheap, but the work still repeats for
  every reused field and is mechanically avoidable.

Evidence reviewed:

- The source and historical SOW evidence listed above.
- The completed Rust entry-local design, including its publication ordering,
  malformed-tail fallback, byte identity, index validation, and profile.
- A preliminary current-Go run that proves the historical mixed workload and
  non-interleaved measurements are inadequate for acceptance.

Affected contracts and surfaces:

- Private Go writer lookup result and `entryItem` state.
- DATA inverted-index publication for compact and regular Go journals.
- Unix direct-mmap tail access and portable fallback behavior.
- Internal Go writer benchmark driver and shared benchmark orchestration.
- No public API, module dependency, file-format, configuration, reader,
  identity helper, lock helper, or Netdata-specific producer behavior.

Existing patterns to reuse:

- Rust's accepted entry-local state lifetime and array-before-DATA publication
  order, translated to Go rather than copied mechanically.
- Go's existing `dataHeader`, direct mmap accessors, portable read/write
  fallbacks, compact-tail chain fallback, entry scratch reuse, and default
  deduplication.
- Existing structural oracle, raw/structured identity, compact, compression,
  live, verify, and cross-language readers.

Risk and blast radius:

- Medium risk inside the low-level Go writer's DATA inverted indexes.
- A defect may leave a structurally readable journal with incomplete indexed
  query results, so stock verification alone is insufficient.
- Memory increases by fixed metadata attached to each already-retained current
  entry item. It is bounded by fields in one entry and cannot grow with writer
  lifetime or workload cardinality.
- State cannot cross files because it is created during one append, stored in
  the per-entry scratch slice, and cleared before the next append returns.

Sensitive data handling plan:

- Use only deterministic synthetic IDs, documentation-range addresses, and
  repository-local fixtures.
- Durable artifacts contain no credentials, customer data, private endpoints,
  personal data, or proprietary incident material.

Implementation plan:

1. Extend the Go benchmark driver and shared runner with the existing exact
   repeating-256 workload while preserving the default corpus.
2. Add benchmark metadata tests and raw/structured byte-identity coverage.
3. Preserve a pre-change Go benchmark binary and collect identical baseline
   measurements plus a focused CPU profile.
4. Add a private resolved-DATA link-state value and return it from matched DATA
   lookup; synthesize empty state for new DATA.
5. Attach the state to `entryItem` and consume it during link publication.
6. Rework compact-tail append so valid non-full mmap tails validate and mutate
   through one safe mapped span, with the existing portable and authoritative
   fallbacks.
7. Preserve offset-array-before-DATA publication and change no public API.
8. Add focused invariant/fallback tests, run full compatibility validation,
   and perform interleaved before/after benchmark/profile comparisons.
9. Review the complete SOW under the standing external-review authorization,
   remediate verified findings, and rerun affected gates.

Validation plan:

- Focused Go unit tests for:
  - new DATA empty state;
  - existing DATA resolved state;
  - regular and compact first-link, promotion, steady append, and tail growth;
  - duplicate elimination retaining one valid state;
  - malformed compact-tail fallback;
  - portable/non-direct fallback where testable;
  - reopen with state reconstructed from the reopened file.
- `gofmt` check on changed Go files.
- `go test ./...` with repository-local Go caches.
- Python benchmark tests and syntax checks.
- Go writer benchmark build.
- Raw/structured byte identity for compact and regular output.
- Interleaved identical before/after exact-workload measurements with preserved
  binary hashes and enough repetitions to characterize noise.
- Before/after Go CPU profiles of the identical exact workload.
- Structural oracle for all DATA and ENTRY array references.
- Indexed query counts for representative repeated payloads.
- Compact and compression interoperability matrices with stock, Rust, and Go
  readers.
- Live one-writer/multiple-reader matrix.
- Deterministic Rust/Go byte-identity matrix where applicable.
- `git diff --check` and SOW audit.

Artifact impact plan:

- Benchmark README: update the workload from Rust-only to Rust/Go and document
  Go's standard duplicate-elimination semantics.
- AGENTS.md, specs, runtime skills, public SDK docs, and end-user skills: no
  change expected because contracts and workflow remain unchanged.
- SOW ledgers: mark SOW-0138 current now and completed only with implementation,
  evidence, review, and commit.

Open-source reference evidence:

- Reuse the checked systemd reference recorded by SOW-0135:
  `systemd/systemd @ c0a5a2516d28601fb3afc1a77d7b42fcfe38fced`
  - `src/libsystemd/sd-journal/journal-file.c:2120-2290`

Open decisions:

- None. The user selected the long-term-best internal scope, no public
  trusted-unique option, and correctness-first retention when performance is
  positive or neutral within noise.

## Implications And Decisions

1. Implementation scope:
   - Selected: exact Go benchmark plus entry-local resolved DATA state and safe
     compact-tail mapped mutation.
   - Classification: long-term-best.
   - Reason: removes redundant work at its source and leaves durable causal
     evidence without introducing a writer-lifetime cache.
2. Public API:
   - Selected: unchanged.
   - Go retains mandatory sorting and duplicate elimination.
   - No trusted-unique option is added.
3. Performance acceptance:
   - Selected: retain a correct improvement when measurements are positive or
     neutral within noise; reject a consistent regression.
   - Reason: the structural state flow is bounded and removes redundant work,
     but plausible low-level changes must not be retained when evidence shows
     harm.
4. Compatibility:
   - Selected: exact bytes, inverted indexes, publication order, reopen,
     compact/regular behavior, compression, FSS, and live visibility remain
     mandatory gates.

## Plan

1. Add and validate exact Go diagnostic support.
2. Preserve the baseline binary; measure and profile current Go.
3. Implement entry-local DATA link state and compact-tail mapped mutation.
4. Run focused and full correctness/interoperability validation.
5. Run interleaved before/after measurements and profiles.
6. Review, remediate, close the SOW, and commit the complete implementation.

## Implementation And Review Plan

Implementation:

- The project manager implements the approved design directly.
- Stop if implementation requires a public API, file-format, durability,
  compatibility, persistent-cache, or publication-order decision not recorded
  here.

Reviewers:

- External review has standing authorization for the current conversation.
- Use the system-wide `external-reviewers` skill after complete local
  validation; do not copy volatile reviewer or harness details into this SOW.

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

- Stop on any unexpected contract expansion, byte/index mismatch, publication
  order change, compatibility regression, or repeatable performance regression.
- Record inconclusive measurements honestly and increase repetitions or use
  interleaving rather than selecting favorable runs.

## Execution Log

### 2026-07-26

- Confirmed local and remote `master` at
  `438ab0350488a41611de2d53b6eacd90b4f726be`.
- Verified repository object integrity and a clean working tree after two host
  recovery events and a storage migration.
- Confirmed the committed Rust source/SOW and local SOW-0135 benchmark/profile
  evidence remained present.
- Reviewed Go writer lookup, entry collection, mmap, DATA-link publication,
  compact-tail, deduplication, benchmark-driver, and shared-runner paths.
- Ran a preliminary five-measurement current-Go mixed-cardinality benchmark.
  The 19,879.721 to 44,176.179 rows/s range and wrong workload shape make it
  diagnostic-only, not acceptance evidence.
- Presented implementation and performance-acceptance alternatives.
- Recorded the user-selected long-term-best implementation and
  correctness-first, non-regression acceptance policy.
- Added the exact Go `netflow-v5-repeating-256` workload, workload metadata,
  raw/structured generation, direct append-loop CPU profiling, shared-runner
  support, and focused benchmark tests without changing the historical default
  workload.
- Preserved the pre-change benchmark binary before production edits and
  captured exact compact structured baseline, identity, and profile evidence.
- Added current-entry `resolvedDataLinkState`, threaded it from DATA lookup
  through deduplicated `entryItem` publication, and removed the redundant DATA
  header read from link publication.
- Added a validated direct-mmap compact-tail append path with the original
  portable read/write and authoritative-chain fallbacks.
- Added focused tests for new DATA state, regular/compact growth, reopen
  reconstruction, deduplication, non-mmap fallback, and invalid compact-tail
  repair.
- Reverified the migrated working tree, Git object database, benchmark
  binaries, profiles, and their hashes after the host reboot and disk
  migration.
- Ran final interleaved benchmark, profile, byte/index, compact, compression,
  sealing, live, portability, and full-suite validation.
- Ran the authorized whole-SOW external review. Six reviewers completed valid
  static reviews with `PRODUCTION GRADE` verdicts and no P0-P2 findings. One
  reviewer could not access the uncommitted tree because its local sandbox
  initialization failed deterministically; its nominal verdict is excluded
  from acceptance evidence and the coverage gap is recorded below.

## Validation

Acceptance criteria evidence:

- Exact workload generation is asserted by
  `go/internal/testcmd/writer_core_bench/main_test.go`:
  29 application fields, 30 ENTRY items, 450 application logical bytes, 491
  total logical DATA bytes, 256 repeating identities, 385 distinct application
  payloads, and source/destination/interface cardinalities 100/3/256.
- The final compact and regular raw/structured outputs match each other and
  the pre-change binary byte-for-byte:
  - compact SHA-256
    `a3bc1da96e2cb5b98405775f9ea7c34b9d366d0cdd570286def8a29a2c9ca8e2`;
  - regular SHA-256
    `f91ae79968719905bd8f8e3d27f81f378c87ea4098f303829a49947314a79d39`;
  - all four files pass stock `journalctl --verify`;
  - evidence:
    `.local/benchmarks/sow0138/final/identity/summary.json`.
- The final exact-workload before/after journals are byte-identical in all 14
  warmup/measurement pairs and pass stock verification. The structural oracle
  reports identical object counts for before/final:
  100,000 ENTRY, 386 DATA, 30 FIELD, and 2,529 ENTRY_ARRAY objects, with all
  2,529 ENTRY_ARRAY objects referenced. Evidence:
  `.local/benchmarks/sow0138/final-index-structure.json`.
- Stock indexed query counts match before/final:
  `FLOW_VERSION=v5` 100,000; source address 1,172; destination address 39,100;
  four-field conjunction 391; protocol 100,000. Evidence:
  `.local/benchmarks/sow0138/final-index-query-counts.json`.
- Current-entry state is value-owned and bounded by fields in one entry.
  `entryItem` grows from 16 to 48 bytes, or 960 additional bytes for the
  30-item measured row. The scratch slice is reused, and new/reopened/successor
  writers construct state from their own DATA objects; no offset or slice is
  retained across files.
- Public Go APIs, file format, hash/index contents, compression, FSS, identity,
  locking, and publication cadence are unchanged.

Tests or equivalent validation:

- Focused final link-state tests: 6 passed for new DATA, regular/compact tail
  growth, reopen, dedupe, non-mmap fallback, and malformed-tail repair.
- `go test ./... -skip
  '^TestGoWriterLive(StockReaders|StockReadersStress|InterruptionReopenAndVerify)$'
  -count=1`: PASS for all applicable packages, including `journal`.
- Unfiltered `go test ./... -count=1`: all packages except `journal` pass;
  exactly three stock-journalctl live cases fail on systemd
  `255.4-1ubuntu8.16` with
  `No journal boot entry found from the specified boot offset (+0).`
  Running those same three tests from unchanged base commit
  `438ab0350488a41611de2d53b6eacd90b4f726be` reproduces the same failures and
  error. Final and base evidence:
  `.local/validation/sow0138/go-test-all-final.txt` and
  `.local/validation/sow0138/go-live-stock-baseline-head.txt`.
- `go vet ./...`: PASS.
- `gofmt -d` over every changed/new Go file: no diff.
- Windows AMD64 non-Unix fallback compile:
  `go test -c ./journal`: PASS; final test binary SHA-256
  `9d38c26a622e24e72e800859e092a41a93dc9e03b5a6efba71d68dd633c86510`.
- Go benchmark target build: PASS.
- Python benchmark syntax check and `unittest` discovery: 25 passed.
- `git diff --check`: PASS.
- `.agents/sow/audit.sh`: PASS with SOW-0138 in `done/`, SOW-0139 in
  `pending/`, and no current SOW.

Real-use evidence:

- Compact Go writer with stock, libsystemd, Rust, and Go readers: 10/10 checks
  passed for 512 entries.
- Go zstd/xz/lz4 writers with stock, libsystemd, Rust, and Go readers: 36/36
  checks passed for 256 entries.
- Live Go regular and compact writers: 2/2 cases passed with active polling,
  stock libsystemd follow, all final readers, structural inspection, and stock
  verification.
- Live Go sealed and compact-zstd/xz/lz4 writers: 4/4 cases passed with the
  same live and final-reader gates.
- Rust/Go deterministic correctness journals are byte-identical in online,
  offline, and archived states, all with expected hash-chain depth 3. Evidence:
  `.local/validation/sow0138/rust-go-byte-identity-all.json`.
- The optional systemd-ingester leg of the all-language identity helper could
  not run because `meson` is not installed after the host migration. Rust and
  Go generation and stock verification succeeded before that independent
  build failure. No package installation was authorized or needed for this
  SOW; the limitation is recorded in
  `.local/validation/sow0138/byte-identity-all.json`.

Performance evidence:

- Pre-change binary SHA-256:
  `010f011ad09bda368b78a4972895053f951c6722ea62902c0307d6b1de012505`.
- Final binary SHA-256:
  `1ecee318a7ded7bcf8c911c4f2fc7183aed5e27412e5b729aa0c917b2ed11079`.
- Identical 100,000-row compact structured workload, live publication
  disabled, two warmup pairs plus twelve alternating measurement pairs:
  - before min/median/max:
    110,151.310 / 124,606.339 / 131,727.752 rows/s;
  - final min/median/max:
    134,071.783 / 137,945.395 / 141,657.073 rows/s;
  - median improvement: 10.705%;
  - final minimum exceeded the before maximum;
  - all 14 pair outputs were byte-identical;
  - evidence:
    `.local/benchmarks/sow0138/final-interleaved/summary.json`.
- Identical 300,000-row direct append-loop profiles:
  - before: 116,580.556 rows/s, 2.57 seconds of samples;
  - final: 136,758.876 rows/s, 2.18 seconds of samples;
  - profile-run throughput change: +17.308%;
  - `readDataHeader` cumulative CPU fell from 0.41 to 0.22 seconds;
  - `linkDataToEntry` cumulative CPU fell from 0.65 to 0.29 seconds;
  - `publishEntryObject` cumulative CPU fell from 0.78 to 0.35 seconds;
  - `findData` cumulative CPU grew from 0.39 to 0.55 seconds because compact
    tail resolution deliberately moved into lookup; overall publication and
    total append time still fell materially;
  - before profile SHA-256:
    `f4b5635ba7b88cf4aaffc636c595e8189ed03d63a5bcdef520de6aa5c97e086b`;
  - final profile SHA-256:
    `7e109aa45abaa3ea0437d429f1c5535e7d25fee170689ae29a78825d4042d3cf`;
  - both profile journals are byte-identical with SHA-256
    `d8cfec09c1b57481639b2b59c92d54fd5d1c7e9957d4613b2fbe9c333b578219`.
- These results are host- and workload-specific diagnostics, not universal Go,
  SDK, or NetFlow capacity claims.

Reviewer findings:

- Valid reviews: six `PRODUCTION GRADE` verdicts, no P0-P2 findings.
- Verified non-blocking P3 themes:
  - `entryItem` grows 16 to 48 bytes, bounded by one entry and offset by the
    measured throughput gain;
  - the `mappedCompactTailUnavailable` branch intentionally reaches the
    portable fallback implicitly after the switch;
  - a full compact tail may request a direct span four bytes past the object,
    but the extra bytes are never read or written; an arena-boundary miss
    safely takes the legacy fallback;
  - the private mapped helper assumes its compact-only caller;
  - direct `bench_command` systemd rejection and historical mixed-workload
    shape have less focused unit coverage than the new Go workload;
  - several naming/assertion observations are readability or defense-in-depth
    issues, while functional byte, reopen, portability, and index evidence
    remains complete.
- Discarded reviewer claims:
  - `pprof.StopCPUProfile` cannot have a dropped error because the Go API
    returns no error;
  - `offsetArrayCapacity` does not duplicate validation; it centralizes the
    same validation for two call sites;
  - compact-tail bounds errors remain safe because both the mapped and
    pre-existing portable paths intentionally fall back to the authoritative
    chain.
- No P3 warranted post-review production churn. They are recorded for
  maintainability and do not block the accepted design or evidence.
- One requested reviewer returned exit 0 but explicitly reported that its
  deterministic bubblewrap sandbox initialization failure prevented access to
  the uncommitted tree. Its verdict is excluded rather than treated as review
  evidence. The failure was not transient and was not retried under the
  external-review failure policy. Raw review records:
  `.local/reviews/sow0138-round1/`.

Same-failure scan:

- Production `readCompactDataTail` now has one caller:
  `resolvedDataLinkState`; link publication does not reread the DATA tail.
- `linkDataToEntry` has one production caller, which supplies the current
  entry's resolved state.
- Remaining `readDataHeader` uses are DATA hash-chain lookup/depth and generic
  object access, not equivalent current-entry link-state reconstruction.
- The shared runner continues to reject systemd and duplicate selections for
  the focused workload through `validate_workload_languages`.

Sensitive data gate:

- Changed code, tests, benchmark documentation, and SOW artifacts contain only
  deterministic synthetic IDs and documentation-range addresses. No raw
  credentials, bearer tokens, customer/community identifiers, personal data,
  private endpoints, or proprietary incident details are present.

Artifact maintenance gate:

- AGENTS.md: no update; public workflow and project guardrails are unchanged.
- Runtime project skills: no update; the existing journal-compatibility and
  SOW workflows already cover this internal optimization.
- Specs: no update; public SDK behavior and journal format are unchanged.
- End-user/operator docs: updated
  `tests/benchmarks/README.md` because the focused diagnostic now supports Go.
  No SDK/operator guide changes are needed for a private compatible path.
- End-user/operator skills: none exist or are affected.
- SOW lifecycle: SOW-0138 completes and moves to `done/` with its code.
  Pre-existing systemd 255 live-test compatibility is tracked separately by
  SOW-0139; Rust residual performance remains SOW-0137.
- SOW-status.md: canonical and root ledgers updated for SOW-0138 completion and
  SOW-0139 pending status.

Specs update:

- No spec update was needed because no public contract, file format, default,
  query behavior, or compatibility guarantee changed.

Project skills update:

- No project skill update was needed because the implementation reused the
  established benchmark, compatibility, SOW, and external-review workflows.

End-user/operator docs update:

- `tests/benchmarks/README.md` documents exact Rust/Go workload support and
  Go's unchanged sorting/deduplication semantics.

End-user/operator skills update:

- No output/reference skills exist in this repository, and the benchmark
  documentation change creates no externally copied skill surface.

Lessons:

- Whole-file mmap removes Rust's window-switching cost but does not make
  repeated typed header reconstruction free; eliminating redundant Go
  publication reads still produced a repeatable 10.705% median gain.
- Entry-local resolved state provides the reuse benefit without a
  writer-lifetime cache, collision policy, or cross-file invalidation risk.
- Stock verification alone cannot prove DATA inverted-index completeness;
  structural reference coverage and indexed query counts remain mandatory.
- Full-suite failures must be reproduced at the unchanged base before being
  attributed to a low-level writer optimization.

Follow-up mapping:

- SOW-0139 tracks the pre-existing systemd 255 stock-journalctl
  `--boot=all` live-test compatibility failure; it authorizes no implementation.
- SOW-0137 continues to track independent Rust residual array-open performance.
- Missing local `meson` for the optional systemd-ingester byte-identity leg is
  recorded as an environment limitation, not a repository defect or required
  installation.

## Outcome

Completed. The Go writer now reuses resolved DATA entry-link state for the
current entry and validates/mutates a non-full compact tail through one mapped
span when available, while retaining portable and authoritative fallbacks.
The exact repeating-256 workload improved by 10.705% median in interleaved
measurements, with byte-identical files, identical inverted indexes, passing
compact/regular/compression/sealing/live/reopen/portability validation, and no
public API or default-semantic change.

## Lessons Extracted

- Current-entry state is the narrowest safe lifetime for reusable file offsets.
- Profiles must distinguish moved work from eliminated work: compact-tail
  resolution moved into lookup, while duplicate header parsing and publication
  cost fell.
- Cross-version stock-tool harness failures require unchanged-base
  reproduction and separate tracking before excluding affected cases.

## Followup

- The Rust residual global ENTRY/regular-array work remains independently
  tracked by SOW-0137.
- The pre-existing Go live stock-journalctl systemd 255 boot-selector failure
  is tracked by SOW-0139.

## Regression Log

- No SOW-0138 regression observed.
