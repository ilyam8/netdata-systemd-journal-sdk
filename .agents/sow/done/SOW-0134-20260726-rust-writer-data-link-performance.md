# SOW-0134 - Rust Writer Repeating-DATA Diagnostics

## Status

Status: completed

Sub-state: Diagnostic tooling, evidence, review remediation, final whole-SOW
review, and closure validation completed. Production writer behavior remains
unchanged.

## Requirements

### Purpose

Create a reproducible SDK-local benchmark and profile for the repeated structured
DATA/link-publication workload that dominates a production-shaped NetFlow v5
collector run. Use that evidence to identify the cleanest production writer
optimization without assuming that the removed recent-DATA cache is the answer.

### User Request

The user approved the recommended diagnostic-first progression and the addition
of a backward-compatible workload preset to the existing internal
`writer_core_bench`. Repository rules, journal compatibility, the protected
Netdata benchmark checkout, and the no-publish/no-release boundary remain
mandatory.

### Assistant Understanding

Facts:

- The SDK repository starts clean at
  `94e383b9ec0d5abb753c90332696cdea6c6192a1`, the released v0.7.8 source.
- The retained production-shaped collector result accepted
  67,733.418 records/s at approximately one fully occupied CPU core, versus
  66,006.026 records/s before the retained collector-local change.
- The post-change profile's leading SDK writer self-time includes
  `WindowManager::get_window` at 9.23%,
  `JournalWriter::add_entry_field_item` at 5.76%, typed-object validation at
  5.09%, `JournalFile::hash_parts` at 4.35%,
  `WindowManager::get_slice` at 4.30%, mutable DATA access at 4.15%, Jenkins
  hashing at 3.66%, and `JournalWriter::publish_entry_links` at 3.61%.
- The selected application row is a deterministic 29-field structured row with
  450 logical `KEY=value` bytes. Across its 256 repeating identities, 26
  payloads are fixed, source address has cardinality 100, destination address
  has cardinality 3, and input interface has cardinality 256, for 385 distinct
  complete application DATA payloads.
- The high-level log writer prepends one fixed raw
  `_BOOT_ID=<32-hex-digits>` field. The actual low-level writer call therefore
  has 30 ENTRY items and 491 logical DATA payload bytes. The focused direct
  benchmark must reproduce that raw-plus-structured low-level call shape while
  keeping the historical workload unchanged.
- The SDK benchmark currently generates only the historical mixed-cardinality
  32-field workload and already supports compact/regular format, structured/raw
  API mode, trusted-unique payloads, publication cadence, windowed/whole-file
  mmap diagnostics, journal verification, warmups, repetitions, and standard
  reports.
- Historical commit `3e6ff79b8fa8f4d0f18a1314230fe0cec93a1672`
  removed a 4,096-slot recent-DATA cache after the existing mixed-cardinality
  benchmark showed no material Rust throughput benefit. A later private
  production-shaped experiment improved the selected repeating workload by
  7.34%, but remained lossy and did not establish an acceptable cache design.

Inferences:

- The existing mixed-cardinality corpus is insufficient to decide repeated-DATA
  writer behavior because eight fields are unique per row and eight have
  cardinality 2,048.
- A writer-only benchmark matching row shape and SDK options can isolate the
  causal DATA resolution/link/mmap path without packet decoding, UDP scheduling,
  tiering, or collector field-store work.
- Comparing windowed and whole-file diagnostic modes plus existing mmap counters
  can determine whether a narrower WindowManager investigation is justified,
  but cannot by itself prove a safe production strategy.

Unknowns:

- The relative cost of DATA resolution, compact link publication, object
  validation, and window lookup on the exact SDK-local repeating workload is not
  yet measured.
- The production optimization design is intentionally unresolved and excluded
  from this SOW. The user will select it from evidence-backed options after this
  diagnostic SOW reports.

### Acceptance Criteria

- `writer_core_bench` accepts an explicit repeating-256 workload while keeping
  the current mixed-cardinality 32-field workload as the unchanged default.
- Focused tests prove 29 application fields, 30 low-level ENTRY items, 450
  application logical payload bytes, 491 total logical DATA payload bytes,
  256-row identity repetition, and 385 distinct application DATA payloads plus
  one fixed boot DATA payload.
- The Rust benchmark driver and report record workload identity, application
  fields per row, low-level ENTRY items per row, and both logical-byte counts;
  non-default Rust-only workload results cannot be silently compared against
  unsupported systemd or Go drivers.
- Raw and structured API modes remain byte-identical for the new workload when
  invoked with otherwise identical deterministic options.
- A preserved current-HEAD release binary and changed diagnostic binary have
  recorded hashes and use isolated repository-local or `/tmp` outputs.
- The current writer is measured with two warmups and ten measurements using
  compact format, no DATA compression, structured fields, trusted unique
  payloads, disabled live publication, windowed mmap, and identical parameters.
- A focused current-writer profile records samples/loss and the targeted symbol
  distribution. Whole-file mmap is used only as a diagnostic comparison.
- No production SDK writer, reader, file-format, public API, default benchmark
  dataset, Netdata source, release, tag, remote branch, or package is changed.

## Analysis

Sources checked:

- `AGENTS.md`
- `.agents/skills/project-agent-orchestration/SKILL.md`
- `.agents/skills/project-journal-compatibility/SKILL.md`
- `.agents/sow/specs/product-scope.md`
- `.agents/sow/done/SOW-0062-20260530-rust-go-writer-absolute-performance.md`
- `rust/src/internal/testcmd/writer_core_bench/src/main.rs`
- `tests/benchmarks/run_writer_core_benchmarks.py`
- `tests/benchmarks/report_benchmarks.py`
- `tests/benchmarks/README.md`
- `rust/src/crates/journal-core/src/file/writer.rs`
- `rust/src/crates/journal-core/src/file/writer_entry_arrays.rs`
- `rust/src/crates/journal-core/src/file/file_mut.rs`
- `rust/src/crates/journal-core/src/file/mmap.rs`
- ktsaou/netdata @ `06385acd27891da2b9aacdb94ae0e56c3659c025`
  - `src/crates/netflow-plugin/src/ingest/encode.rs`
  - `src/crates/netflow-plugin/src/ingest/service/init.rs`
  - `src/crates/netflow-plugin/src/ingest_capacity_bench_wire.rs`
  - `src/crates/netflow-plugin/src/decoder/protocol/legacy.rs`
  - `src/crates/netflow-plugin/src/flow/record/journal.rs`
  - `src/crates/netflow-plugin/src/flow/record/journal/*.rs`

Current state:

- `writer_core_bench` hard-codes 32 fields in row generation and result
  metadata.
- Its shared runner hard-codes 32 fields in report parameters and has no
  workload selector.
- Direct Rust runs already expose compact, no-compression, structured,
  trusted-unique, no-live-publication, and mmap controls. The new workload must
  additionally prepend the fixed boot field as one raw `EntryField`, matching
  the high-level writer before calling the same low-level core path.
- `JournalWriter::link_data_to_entry()` opens DATA to read link metadata,
  releases it, updates the offset-array tail, then opens DATA again to publish
  the new count/tail.
- Compact tail append reads an offset-array object to validate capacity and then
  opens it again mutably to append.
- WindowManager checks one active window and otherwise scans existing windows.

Risks:

- Changing the default corpus would invalidate historical benchmark comparisons.
- Allowing a Rust-only workload to run alongside unchanged Go/systemd drivers
  would create a false cross-language comparison.
- Row-generation time or allocations inside the append timer would contaminate
  SDK writer measurements.
- Unpinned CPU frequency, background load, journal verification inside the
  timer, or unequal binaries would make small deltas untrustworthy.
- Naming the workload as a universal capacity model would overstate
  host-specific evidence.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:

- Production-shaped evidence identifies repeated keyed DATA lookup, independent
  Jenkins hashing, compact DATA inverted-index link publication, repeated typed
  object validation, and window lookup as the dominant remaining SDK cluster.
  The current SDK-local benchmark has the wrong cardinality distribution to
  distinguish those causes. The first required change is therefore a
  benchmark-only, exact repeating-row preset; production behavior must remain
  unchanged until its evidence supports a user-selected design.

Evidence reviewed:

- The source files and benchmark/profile evidence listed under `## Analysis`.
- The row-shape derivation is grounded in the synthetic NetFlow v5 wire
  generator, v5 decoder, sparse FlowRecord journal encoder, and measured
  450-byte logical-row report.
- The existing cache-removal diff and SOW-0062 benchmark protocol were reviewed
  rather than treating the private cache experiment as a confirmed regression.

Affected contracts and surfaces:

- Internal Rust benchmark CLI and JSON output.
- Shared writer benchmark runner/report configuration for Rust-only diagnostic
  workload selection.
- Benchmark documentation and focused internal tests.
- No runtime SDK or public API contract is affected.

Existing patterns to reuse:

- Pre-materialized `BenchField` rows and append-only timing boundaries in
  `writer_core_bench`.
- Two warmups, ten measurements, standard JSON reporting, stock
  `journalctl --verify`, and API-mode byte identity from SOW-0062.
- Existing `--api-mode`, `--trusted-unique-payloads`,
  `--live-publish-every-entries`, and `--mmap-strategy` controls.

Risk and blast radius:

- Blast radius is limited to non-published internal benchmark tooling and its
  documentation.
- The historical default must be byte-for-byte equivalent at the generated-row
  level and retain the same CLI default.
- The runner must reject a non-default workload unless only the Rust language
  is selected.
- Generated journals are synthetic and disposable; no live journal or host
  service is accessed.

Sensitive data handling plan:

- Use only deterministic documentation-range addresses, synthetic IDs, and
  fixed numeric values.
- Do not copy host-assigned ephemeral ports, customer data, private endpoints,
  personal data, credentials, or proprietary log rows into durable artifacts.
- Evidence records only repository-relative code paths, public commit identity,
  sanitized metrics, commands, and hashes.

Implementation plan:

1. Add a parsed workload selector and exact repeating-256 row generator to the
   Rust internal writer benchmark while preserving the existing default. For
   this direct structured workload only, pass the fixed boot payload as a raw
   `EntryField` followed by the 29 structured fields, exactly matching the
   high-level writer's low-level call shape.
2. Add focused generator invariants and dynamic application-field,
   ENTRY-item, and logical-byte metadata. Reject the non-default workload on the
   directory surface because that separate harness is outside this focused
   core diagnostic.
3. Extend the shared runner/reporter with workload identity, a Rust-only
   non-default guard, and dynamic row-shape metadata.
4. Document the workload and exact diagnostic invocation.
5. Validate behavior, preserve the current-HEAD binary independently, then run
   the approved baseline measurements and profile.

Validation plan:

- `cargo test -p writer_core_bench`
- `cargo check -p writer_core_bench`
- `cargo fmt --all --check`
- Python syntax checks for changed benchmark scripts.
- Default-workload CLI smoke proving unchanged 32-field identity.
- Repeating-workload CLI smoke proving 29 application fields, 30 ENTRY items,
  450 application bytes, and 491 total logical DATA bytes.
- Rust raw/structured byte-identity comparison for compact and regular output.
- Stock `journalctl --verify` for kept compact and regular journals.
- Two-warmup/ten-measurement compact repeating workload run.
- Same-parameters windowed/whole-file diagnostic run.
- `perf record`/`perf report` with sample/loss and binary hashes recorded.
- `git diff --check`, focused same-pattern search, and `.agents/sow/audit.sh`.

Artifact impact plan:

- AGENTS.md: update implementation ownership, external-review policy, and the
  superseded historical routing note.
- Runtime project skills: update orchestration and docs-authoring workflows to
  remove stale routing assumptions and defer reviewer mechanics to the
  system-wide `external-reviewers` skill.
- SOW framework: update the template, active/pending implementation-and-review
  sections, and both status ledgers for the current policy.
- Specs: no update expected because this SOW adds diagnostic tooling, not a
  product contract.
- End-user/operator docs: `tests/benchmarks/README.md` will document the
  internal benchmark workload and limitations.
- End-user/operator skills: no output/reference skill is affected.
- SOW lifecycle: one new SOW; production implementation will require a separate
  user decision and SOW rather than an untracked deferred edit.
- SOW-status.md: update when this SOW starts and when it completes.

Open-source reference evidence:

- ktsaou/netdata @ `06385acd27891da2b9aacdb94ae0e56c3659c025`
  supplies the public synthetic wire shape and writer-call evidence cited
  above. No proprietary corpus or external package source was needed.

Open decisions:

- None block this diagnostic SOW.
- Production writer design is explicitly outside this SOW and must be presented
  as numbered evidence-backed alternatives after the diagnostics.

## Implications And Decisions

1. Diagnostic progression:
   - Options presented: diagnostic-first, immediate compact-tail surgery, or
     cache-first.
   - User decision: diagnostic-first.
   - Classification: long-term-best.
   - Reason: current evidence does not distinguish window scanning, repeated
     object validation, DATA resolution, and unavoidable on-disk mutation well
     enough to choose a production design safely.
2. Benchmark organization:
   - Options presented: add a backward-compatible workload preset to
     `writer_core_bench`, or create a separate executable.
   - User decision: add the preset to the existing benchmark.
   - Classification: long-term-best.
   - Reason: this reuses established timing, JSON, verification, mmap, and
     byte-identity controls without duplicating infrastructure.
3. Production behavior:
   - Decision: no writer-path implementation in this SOW.
   - Reason: the user must see Phase 1 evidence before selecting the production
     cache/mutation/window design.
4. Implementation and review policy:
   - Decision: the project manager owns implementation work directly. Remove
     mandatory implementation routing from live repository instructions without
     adding a blanket prohibition on future user-directed routing.
   - Decision: live repository instructions do not name reviewer models or
     harnesses. They defer to the system-wide `external-reviewers` skill.
   - Decision: external review is the default user-waivable code-review gate.
     Once the user authorizes external review, that authorization applies for
     the current conversation under the boundaries in the system-wide skill.
   - Decision: preserve completed SOWs and past execution records as historical
     evidence. They do not establish current routing.
   - Classification: long-term-best.
   - Reason: implementation ownership remains clear while volatile reviewer
     selection and execution details have one current source of truth.
5. Production follow-up:
   - Decision: carry already-resolved DATA link state with each current
     `EntryItem`, consume that state during link publication, and open the
     compact tail once for validation/mutation while preserving final DATA
     header publication ordering.
   - Decision owner: user.
   - Classification: long-term-best.
   - Boundary: this SOW remains diagnostic-only. The production change requires
     a separate SOW after SOW-0134 closes.
6. External review:
   - Decision: the user authorized `claude`, `codex`, `glm`, `minimax`,
     `deepseek`, `mimo`, and `qwen` for this conversation.
   - Round 1 ran against the complete SOW-0134 changed surface after local
     validation. Standing authorization covers the material-fix re-review.

## Plan

1. Implement and test the exact benchmark workload without changing defaults.
2. Validate runner/report compatibility and journal correctness.
3. Preserve and hash current-HEAD and changed benchmark binaries separately.
4. Run the approved baseline measurements, diagnostic mmap comparison, and
   profile.
5. Recommend whole-SOW external review after local validation, follow the
   system-wide `external-reviewers` skill if authorized, record evidence, close
   locally, and present the production design decision.

## Implementation And Review Plan

Implementation:

- The project manager implements the benchmark code and focused tests, runs the
  benchmark/profile work, and records validation evidence.

Reviewers:

- External review is authorized for this conversation.
- Review the complete SOW and changed surface after local validation. Material
  review fixes require the same whole-SOW scope on re-review.

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

- Stop on unexpected scope expansion, runtime contract changes, test
  regressions, or ambiguous benchmark evidence.
- If external review is authorized, record and disposition reviewer findings
  against the complete SOW before closure.
- An audit failure blocks closure until repaired and re-run.

## Execution Log

### 2026-07-26

- Verified clean starting commit and clean SOW audit.
- Read the applicable project instructions, skills, product scope, prior writer
  performance SOW, current benchmark code, writer hot paths, external synthetic
  row generator, and immutable benchmark/profile evidence.
- Recorded the approved diagnostic-first and existing-benchmark decisions.
- No production SDK or external repository file was changed.
- The designated `llm-netdata-cloud/minimax-m3-coder` implementation run failed
  before editing any source file with an `Unexpected server error`
  (`err_5184ef51`). No fallback run occurred.
- The user superseded the mandatory implementation-routing policy and approved
  direct project-manager implementation, repository-local reviewer-policy
  cleanup, and system-wide reviewer-skill indirection. No SDK source file was
  changed as part of that policy decision.
- Updated `AGENTS.md`, the orchestration and docs-authoring project skills, the
  SOW template, all current/pending implementation-and-review sections, and both
  status ledgers. Completed SOW evidence was preserved unchanged.
- Policy search confirmed that live instruction files contain no reviewer model
  names or harness commands and that `AGENTS.md` contains one historical
  mandatory-delegation note. Remaining model-specific references are past
  actions in SOW/status evidence.
- `git diff --check` passed and `.agents/sow/audit.sh` completed cleanly after
  the policy update.
- Added the backward-compatible `netflow-v5-repeating-256` direct Rust workload
  to `writer_core_bench`. The historical mixed-cardinality workload remains the
  CLI default and its generated two-row shape is covered by an exact focused
  regression test.
- The new generator reproduces the measured synthetic v5 row:
  29 application fields, 450 application logical bytes, one fixed raw boot
  field, 30 low-level ENTRY items, 491 total logical DATA bytes, 256-row
  repetition, and 385 distinct application payloads.
- Extended the shared runner and reporter with workload/shape metadata, an
  exact Rust-only guard for non-default workloads, driver metadata validation,
  API-identity metadata, and a before/after workload compatibility gate.
- Added focused Rust and Python tests and documented the diagnostic command and
  non-universal interpretation.
- Preserved release binaries under
  `.local/benchmarks/sow0134/binaries/`:
  - starting commit binary SHA-256:
    `ec01b4a14db8d9fe23f072c7c6e46b5f9426f2be46d01afe26e50f9c1342b5cb`;
  - diagnostic binary SHA-256:
    `ebc7892b9b65fd1963dec45150ce37ac98c939d377ba21080e74440dc0caae20`.
  The diagnostic binary and its 100,000-row measurement/profile evidence
  predate the later behavior-preserving review fixes. The final source retained
  identical 512-row journal bytes in both formats after those fixes, but the
  preserved diagnostic executable hash is not claimed as a rebuild of the
  final source.
- The exact compact windowed measurement completed with two warmups and ten
  measurements:
  - min: 118,231.355 rows/s;
  - median: 120,601.200 rows/s;
  - max: 122,152.052 rows/s;
  - report SHA-256:
    `0340c7f08e1c9b7e5fa9596b834f967b97271e9e4776d60264d9de8d8b95511e`.
- The same whole-file diagnostic completed with two warmups and ten
  measurements:
  - min: 118,005.079 rows/s;
  - median: 121,700.145 rows/s;
  - max: 122,831.789 rows/s;
  - report SHA-256:
    `4dd17771433e0e1c55338090358a1690233dcbe77944cfee2f5e74f83355ad9c`.
- Whole-file median append throughput was 0.911% above windowed, while the
  observed ranges overlapped. It reduced median map/remap counts from 17/11 to
  8/6, but did not produce a material throughput separation.
- The focused current-writer profile used the same 100,000-row compact,
  structured, trusted-unique, no-live-publication, windowed workload:
  - 750 cycle samples, zero lost;
  - `publish_entry_links`: 6.91% self;
  - `add_entry_field_item`: 6.87%;
  - Jenkins hashing: 6.58%;
  - `WindowManager::get_slice`: 6.25%;
  - typed-object validation: 6.13%;
  - journal DATA hashing: 6.11%;
  - `WindowManager::get_window`: 3.92%;
  - mutable-object sizing: 3.83%;
  - profile SHA-256:
    `eec882c269f258d0b84f34f440a8efc5698d49eaa21edc8e5549efd1d7562a94`.
- An initial unprivileged perf probe was blocked by
  `perf_event_paranoid=4`. The approved focused profile was recorded with
  non-interactive elevated perf permissions against only the benchmark child;
  no sysctl, watchdog, service, or host configuration was changed. A local
  debuginfod connection stalled the first report render; the exact report PID
  started by this session was terminated and report generation was rerun with
  network debuginfod disabled.
- The protected Netdata checkout was inspected read-only only. Its currently
  observed 9 modified and 1 untracked files remain present.
- The user selected the long-term-best production follow-up described in
  Decision 5 and authorized the seven-reviewer session recorded in Decision 6.
- External review round 1 completed. Five reviewers returned
  `PRODUCTION GRADE`; two returned `NEEDS CHANGES`.
- Independent adjudication verified two blocking findings:
  1. the newly added directory-surface shape metadata described only the
     application fields and omitted the high-level writer's fixed boot field;
  2. the focused test checked declared metadata but did not exercise the helper
     that constructs the 30 low-level ENTRY items.
- Remediation plan:
  1. omit the new direct-workload shape metadata from the unchanged directory
     report;
  2. extract one entry-field construction helper used by the benchmark and
     focused tests;
  3. assert both raw and structured low-level shapes and bind the boot DATA
     payload to the writer boot ID in tests;
  4. correct the SOW's statistical and whole-file interpretation wording;
  5. rerun focused/full validation and the same whole-SOW reviewer set.
- Implemented the two blocking remediations:
  - directory output again emits only its historical `fields_per_row` shape
    field and no direct-only workload metadata;
  - `prepare_direct_entry_fields()` is the single construction path used by
    the benchmark and focused raw/structured shape assertions.
- The focused test now proves the boot DATA payload matches the writer boot ID.
- Added pending
  `SOW-0135-20260726-rust-writer-entry-link-state-reuse.md` with the selected
  production design and complete pre-implementation analysis. No production
  source was changed.
- External review round 2 completed. Six reviewers returned
  `PRODUCTION GRADE`; one returned `NEEDS CHANGES`.
- Independent adjudication rejected the claim that default Rust driver shape
  metadata could be missing without detection: `driver_workload_errors()`
  reads every required field without a fallback and reports a mismatch for a
  missing value. A focused test now proves that behavior directly.
- Independent adjudication accepted the reporter's partial-metadata finding:
  an explicitly workload-tagged report could omit one or more row-shape fields
  and still compare with a complete report. Explicit workload reports now
  require all four shape fields; true legacy reports with no `workload` key
  retain the documented default-workload compatibility path.
- Added focused tests for incomplete explicit report rejection, legacy-default
  report compatibility, missing default Rust driver metadata rejection, and
  direct entry-field API routing for raw, structured, prefixed, and unprefixed
  workloads.
- External review round 3 completed after those material fixes. Six reviewers
  returned `PRODUCTION GRADE`. The Codex reviewer could not access the local
  worktree because its sandbox failed before every read-only command with
  `bwrap: loopback: Failed RTM_NEWADDR: Operation not permitted`; its
  fail-closed response identified no code defect and is treated as an
  unrecovered reviewer coverage gap rather than a product verdict.
- Round-3 findings were P3 maintenance/documentation observations only. Final
  cleanup added a derived total-byte invariant, documented the reporter's
  explicit/legacy metadata rule, removed one dead reporter constant, and
  broadened the historical-evidence wording to match retained SOW/status
  records. These trivial cleanups do not require another external-review round.

## Validation

Acceptance criteria evidence:

- Exact generator invariants are covered by
  `netflow_repeating_workload_matches_low_level_writer_shape`.
- The default workload regression is covered by
  `mixed_cardinality_default_row_shape_is_unchanged`.
- Direct-only behavior is covered by a Rust unit test and a real CLI rejection.
- The runner's exact Rust-only guard, selected workload command propagation,
  driver shape validation, and non-Rust rejection are covered by Python unit
  tests and a real CLI rejection.
- Compact and regular raw/structured output hashes matched exactly at 512 rows:
  - compact:
    `0d8c003eb289788efc94e66c06ade3230f715b69c2f01a991f51c6bc4116d32f`;
  - regular:
    `1b08a3f28dfe4d190428e74f5a1e56c489e4e10ed13dcb7fb3af1852433edaec`.
- Stock systemd 255.4 `journalctl --verify --file` passed for all four kept
  smoke journals and for every non-warmup shared-runner journal.
- The exact 100,000-row windowed run, whole-file diagnostic, and current-writer
  profile are recorded under `.local/benchmarks/sow0134/`.

Tests or equivalent validation:

- `cargo test -p writer_core_bench`: 3 passed.
- `cargo check -p writer_core_bench`: passed.
- `cargo test -p systemd-journal-sdk-core`: 74 unit tests passed; 3 doc tests
  ignored as expected.
- `cargo test -p systemd-journal-sdk-log-writer`: 58 tests passed in total
  (6 library, 51 integration, and 1 compile doc test).
- `python3 -m unittest tests/benchmarks/test_run_writer_core_benchmarks.py
  tests/benchmarks/test_report_benchmarks.py`: 25 passed.
- `cargo fmt --all --check`: passed.
- Python syntax compilation: passed.
- `git diff --check`: passed after the final SOW evidence update.
- Review-fix validation repeated the focused Rust and Python suites, core and
  log-writer suites, cargo check, formatting, and diff checks with the same
  pass counts.
- A rebuilt one-row directory smoke report retained
  `fields_per_row: 32` and omitted `workload`,
  `application_fields_per_row`, `entry_items_per_row`, and both new logical-byte
  fields. Its archived journal passed stock `journalctl --verify --file`.
- Review-fix compact and regular 512-row raw/structured journals remained
  byte-identical with the same hashes recorded above, and all four passed stock
  verification.
- Round-2 remediation validation repeated the focused Rust tests, Rust
  benchmark check, Python suites, formatting, and diff checks. The latest
  counts are 3 Rust benchmark tests and 25 Python tests, all passing.
- Final post-review cleanup validation retained the same 3 Rust benchmark and
  25 Python test passes; `cargo check`, Python syntax compilation,
  `cargo fmt --all --check`, and `git diff --check` passed.
- The final `.agents/sow/audit.sh` run completed cleanly with one completed
  current SOW before the lifecycle move. Durable-artifact personal-name,
  live-instruction reviewer-command/model, and mandatory-delegation policy
  scans passed.

Real-use evidence:

- Shared-runner compact windowed report:
  `.local/benchmarks/sow0134/windowed/compact-none-fss-off-api-structured-field-live-every-0-workload-netflow-v5-repeating-256-trusted-unique-mmap-windowed-20260726T084827936428Z/report.json`.
- Shared-runner whole-file report:
  `.local/benchmarks/sow0134/whole-file/compact-none-fss-off-api-structured-field-live-every-0-workload-netflow-v5-repeating-256-trusted-unique-mmap-whole-file-20260726T084941677985Z/report.json`.
- Standard comparison report:
  `.local/benchmarks/sow0134/windowed-vs-whole-file.md`.
- Profile:
  `.local/benchmarks/sow0134/profiles/current-writer-windowed-20260726T085244Z/`.
- The external NetFlow baseline report supplied with the user request retained
  SHA-256
  `6c645d6c8e2e6acf363cce35618507bdb2ae68e61514c4937320989c0edcea0c`.

Reviewer findings:

- Round 1 outputs are stored under `.local/reviews/sow0134-round1/`.
- Verified blocking findings:
  - `directory_surface_entry_shape_metadata_wrong` (`claude`, P2): accepted.
    The new directory report fields described 32/817 while the high-level
    writer appends a fixed 41-byte boot DATA item, so the actual shape is
    33/858. Remediation removes the direct-only shape metadata from the
    unchanged directory report.
  - `low-level-entry-shape-is-only-metadata-tested` (`minimax`, P2), also
    reported as `entry_items_per_row_never_asserted_against_actual_append`
    (`claude`, P3): accepted. Remediation makes the benchmark and test share the
    exact entry-field construction helper and asserts both API modes.
- Verified non-blocking findings included the duplicated boot-ID literal,
  unresolved ordering within the small profile sample, whole-file outcome
  wording, unused one-time structured buffer allocation, default mixed-row
  byte-count behavior above one million rows, intentionally redundant
  compatibility metadata, and local pre-commit/untracked state. The boot-ID
  invariant and wording issues are included in remediation; the remaining
  items do not invalidate the 100,000-row diagnostic and are recorded as
  maintenance limitations rather than scope expansion.
- Rejected or out-of-scope findings:
  - a canonical boundary block is mandatory in reviewer prompts, not in every
    historical pending SOW; SOW-0123 therefore does not require a speculative
    plan rewrite;
  - model names in historical evidence are explicitly permitted;
  - the untracked focused test is part of the intended uncommitted change, not
    missing content;
  - pre-existing benchmark-CI and writer-directory reporter limitations are
    outside this SOW.
- Process note: `deepseek` reported running the SOW audit despite the
  read-only prompt explicitly prohibiting audit/test execution. Its static
  findings were independently adjudicated, but that reviewer did not follow
  the requested review procedure.
- Round 2 outputs are stored under `.local/reviews/sow0134-round2/`.
- Six round-2 reviewers returned `PRODUCTION GRADE`; `minimax` returned
  `NEEDS CHANGES`.
- Round-2 finding dispositions:
  - `default-workload-shape-metadata-is-not-actually-validated-for-rust-driver`
    (`minimax`, P2): rejected as factually incorrect. The validator uses
    `driver.get(field)` without the summary helper's fallback, so a missing
    default-workload field becomes a mismatch. A focused regression test now
    demonstrates the rejection.
  - `reporter-compatibility-guard-allows-partially-missing-row-shape`
    (`minimax`, P2): accepted. Explicit workload reports now require complete
    row-shape metadata, while a tested legacy report with no explicit workload
    remains compatible with a complete current default report.
  - `entry_field_routing_condition_untested` (`claude`, P3): accepted as useful
    regression coverage. The append branch now uses
    `requires_entry_fields_api()`, whose focused test covers both workload
    presets and both API modes.
  - Remaining P3 observations concern the historical mixed workload above one
    million rows, one-time allocation, metadata duplication, non-Rust inferred
    summaries, pre-existing CI wiring, and historical evidence. They do not
    affect the 100,000-row Rust-only diagnostic or its compatibility boundary
    and are recorded as maintenance limitations rather than scope expansion.
- Round 3 outputs are stored under `.local/reviews/sow0134-round3/`.
- Six round-3 reviewers returned `PRODUCTION GRADE`: `claude`, `glm`,
  `minimax`, `deepseek`, `mimo`, and `qwen`.
- The Codex reviewer was unable to inspect the worktree because every local
  read-only command failed in its sandbox before execution. The failure is
  deterministic for that invocation, was not retried under the reviewer-skill
  failure policy, and leaves one explicit reviewer coverage gap.
- No round-3 P0, P1, or P2 code, compatibility, security, or operational
  finding was verified. Reported P3 observations were independently checked:
  - accepted cleanup: derive the 491-byte total in the Rust invariant test,
    remove the unused `WRITER_WORKLOAD_KEYS`, document explicit-versus-legacy
    report metadata, qualify the stock verifier version, clarify the preserved
    diagnostic-binary timeline, and align historical-evidence policy wording;
  - retained maintenance limitations: the historical mixed workload's static
    byte count above one million rows, non-Rust summary inference, malformed
    summary robustness, metadata duplication, a one-time unused buffer, manual
    benchmark-test CI coverage, and pre-existing directory harness behavior;
  - rejected factual ambiguity: the `6c645d...` hash belongs to the external
    NetFlow baseline report supplied by the user, not a regenerated SDK
    benchmark report. The wording now makes that provenance explicit.
- Final review found no blocking issue. Closure audit passed.

Same-failure scan:

- Searched the benchmark suite for remaining hard-coded writer-core workload
  metadata and command builders. The separate directory and systemd helpers
  retain their historical 32-field corpus intentionally; the shared runner
  prevents the Rust-only diagnostic from reaching them.
- The production writer pattern remains unchanged for the follow-up decision:
  steady compact publication opens DATA mutably, opens the tail array once for
  capacity validation and again for mutation, and then opens DATA mutably a
  second time to publish the new count/tail.

Sensitive data gate:

- Current SOW content uses synthetic/public evidence and contains no raw secret,
  credential, customer, private-endpoint, or personal data.

Artifact maintenance gate:

- AGENTS.md: implementation/review policy updated and validated.
- Runtime project skills: orchestration and docs-authoring policy updated and
  validated.
- Specs: no update needed; diagnostic tooling and repository workflow changed,
  but no product behavior or public contract changed.
- End-user/operator docs: benchmark README updated.
- End-user/operator skills: none exist for this benchmark surface.
- SOW lifecycle: current/in-progress is consistent.
- SOW-status.md: current entry added with SOW activation.

Specs update:

- No update needed because no SDK behavior, public contract, data format, or
  compatibility guarantee changed.

Project skills update:

- The orchestration and docs-authoring runtime skills were updated for the
  approved implementation/reviewer policy. Journal-compatibility rules remain
  correct and required no change.

End-user/operator docs update:

- `tests/benchmarks/README.md` documents the workload, exact row shape,
  Rust-only guard, diagnostic invocation, and interpretation boundary.

End-user/operator skills update:

- No output/reference skill exists for the benchmark documentation, so no skill
  update is needed.

Lessons:

- Whole-file mmap reduces mapping operations but did not materially separate
  throughput, so changing the default mapping strategy is not supported.
- The 750-sample profile supports the combined
  mutation/validation/hashing/link cluster. It does not statistically resolve
  the ordering of similarly sized individual symbols within that cluster.
- The exact writer-only workload sustains a median 120,601 rows/s on this host.
  That is not an end-to-end capacity result: at 100,000 rows/s the writer alone
  consumes most of one core's available time, leaving insufficient budget for
  UDP receive, decode, field handling, and collector control work.
- The profile reproduces the same mutation/validation/hashing/link cluster as
  the production-shaped collector profile, validating this benchmark as a
  causal follow-up tool.

Follow-up mapping:

- The user-selected per-entry resolved-link-state design is tracked by pending
  `SOW-0135-20260726-rust-writer-entry-link-state-reuse.md`.

## Outcome

The diagnostic tooling is implemented and locally validated without changing
the production writer or default benchmark corpus. Whole-file mmap did not
materially separate throughput in this run, so the evidence does not support
changing the default mapping strategy as the primary solution. The next
production design targets repeated DATA resolution and compact
link-publication metadata/object access.

## Lessons Extracted

- Exact workload shape matters: the historical mixed-cardinality corpus does
  not represent a 256-identity repeating flow workload.
- Mapping operation counts alone are not a throughput proxy.
- Raw/structured byte identity can cover a mixed raw boot field plus structured
  application fields without changing the public writer API.

## Followup

Production writer changes are deliberately excluded. The selected production
design is tracked by pending SOW-0135. All verified review findings were
remediated, final whole-SOW review found no blocking issue, and only the closure
audit/lifecycle commit remains for SOW-0134.

## Regression Log

None yet.
