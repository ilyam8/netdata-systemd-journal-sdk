# SOW-0135 - Rust Writer Entry-Link State Reuse

## Status

Status: completed

Sub-state: Implementation, local validation, whole-SOW external review, final
audit, follow-up mapping, and lifecycle gates completed. Local commit is created
with the SOW close.

## Requirements

### Purpose

Reduce repeated DATA and compact offset-array validation/mmap work in the Rust
journal writer without changing journal bytes, indexes, public APIs, durability,
live visibility, or default semantics.

### User Request

The user selected the long-term-best production design after SOW-0134 isolated
the current writer path: reuse DATA link metadata already resolved for the
current entry, open a non-full compact tail once for validation and mutation,
and preserve the existing final DATA-header publication ordering.

The change remains local. Do not publish, tag, push, or release an SDK version,
and do not modify the protected external Netdata benchmark checkout.

### Assistant Understanding

Facts:

- `JournalFile::find_data_offset_parts()` already reads the matched DATA header
  and payload while resolving each field:
  `rust/src/crates/journal-core/src/file/file_payload.rs:415-439` and
  `rust/src/crates/journal-core/src/file/file_payload.rs:444-529`.
- The lookup result retains only the DATA offset. It discards `n_entries`,
  `entry_array_offset`, and compact tail fields that are present in the bytes
  already read.
- `JournalWriter::link_data_to_entry()` reopens the same DATA object mutably to
  recover those fields:
  `rust/src/crates/journal-core/src/file/writer_entry_arrays.rs:435-470`.
- Steady compact append opens the tail offset-array read-only to get capacity,
  then opens the same object mutably to write:
  `rust/src/crates/journal-core/src/file/writer_entry_arrays.rs:346-405`.
- After the array item is written, `append_data_entry_array_link()` reopens the
  DATA object mutably to publish the new tail/count:
  `rust/src/crates/journal-core/src/file/writer_entry_arrays.rs:503-527`.
- Every read-only object open repeats header slice lookup, typed validation,
  bounds checks, full-object lookup, and typed parsing:
  `rust/src/crates/journal-core/src/file/file.rs:769-811`.
- Every existing mutable object open repeats header slice lookup, typed
  validation, type/size/bounds checks, full-object mutable lookup, and typed
  parsing:
  `rust/src/crates/journal-core/src/file/file_mut.rs:342-420`.
- SOW-0134's exact 100,000-row repeating workload measured a windowed median of
  120,601.200 rows/s. Its focused profile attributed leading self-time to link
  publication, field addition, both hashes, mmap lookup, typed validation, and
  mutable-object sizing. Whole-file mmap did not materially separate
  throughput.
- `JournalWriter::create_successor()` constructs a new writer for the new file:
  `rust/src/crates/journal-core/src/file/writer.rs:407-415`.
- The trusted-unique fast path promises that one entry contains no duplicate
  complete DATA payloads. The default path still sorts and removes duplicates:
  `rust/src/crates/journal-core/src/file/writer.rs:597-607`.

Inferences:

- The first DATA reopen and the separate compact-tail capacity open are
  avoidable on the normal unique-field steady path.
- Entry-local state is safer than a writer-lifetime payload cache: it has no
  cross-file identity, collision, eviction, invalidation, reopen, or successor
  lifecycle.
- DATA-header publication must remain after the offset-array write. Publishing
  the count/tail first could expose an index entry whose array item is not yet
  initialized.
- Duplicate offsets under misuse of the trusted-unique option need an
  authoritative disk fallback. Two pre-publication snapshots for the same DATA
  object would otherwise make the second link use stale state.

Unknowns:

- The throughput improvement and exact symbol redistribution are measurable
  outcomes, not assumptions.
- The profile cannot establish a universal SDK or NetFlow capacity result.

### Acceptance Criteria

- Existing DATA lookup returns an internal link-state snapshot without changing
  the public `find_data_offset_parts()` API.
- `EntryItem` carries the snapshot only for the current entry. It does not
  create writer-lifetime cardinality growth.
- New DATA objects receive the correct empty link state.
- Unique existing DATA links use the snapshot instead of the first mutable DATA
  reopen.
- A non-full compact tail is validated and mutated through one mutable object
  guard.
- Offset-array data is written before the final DATA count/tail publication,
  preserving current visibility and failure ordering.
- Default duplicate elimination is unchanged. Repeated offsets under
  trusted-unique misuse fall back to authoritative disk state rather than
  consuming a stale snapshot.
- Compact and regular journal bytes, DATA inverted indexes, query results,
  crash/reopen behavior, successor behavior, public APIs, live publication, and
  default semantics remain unchanged.
- The implementation adds no persistent cache, public option, dependency, file
  format, configuration, or producer-specific behavior.
- The exact SOW-0134 workload demonstrates a material reduction in the targeted
  path. If benchmark/profile evidence does not separate from noise, the
  production optimization is not accepted merely because tests pass.

## Analysis

Sources checked:

- `AGENTS.md`
- `.agents/skills/project-agent-orchestration/SKILL.md`
- `.agents/skills/project-journal-compatibility/SKILL.md`
- `.agents/sow/specs/product-scope.md`
- `.agents/sow/current/SOW-0134-20260726-rust-writer-data-link-performance.md`
- `rust/src/crates/journal-core/src/file/file_payload.rs`
- `rust/src/crates/journal-core/src/file/file.rs`
- `rust/src/crates/journal-core/src/file/file_mut.rs`
- `rust/src/crates/journal-core/src/file/writer.rs`
- `rust/src/crates/journal-core/src/file/writer_entry_arrays.rs`
- `rust/src/crates/journal-core/src/file/writer_tests.rs`
- `rust/src/internal/testcmd/writer_core_bench/src/main.rs`
- `tests/benchmarks/run_writer_core_benchmarks.py`

Current state:

- A steady compact link opens DATA twice and the tail offset-array twice at the
  high-level object-access layer.
- The matched DATA header was already read during `add_data()`.
- The existing publication path correctly writes the offset-array link before
  publishing the DATA count/tail.

Risks:

- A stale snapshot can lose or overwrite an inverted-index link.
- Changing publication order can expose incomplete indexes to live readers or
  leave a crash-inconsistent header.
- Trusted-unique duplicate misuse creates intra-entry state changes not visible
  in the pre-publication snapshot.
- Compact and regular layouts store different tail metadata.
- Optimizing object validation must not weaken validation on input, fallback,
  reopen, or malformed-file paths.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:

- DATA lookup reads and validates the bytes needed to identify the object, but
  discards link metadata from the same header. Link publication rereads that
  DATA object to reconstruct the discarded state.
- Compact steady publication then reads the tail offset-array for capacity and
  reopens it mutably for the actual write.
- These repeated object opens invoke the profiled mmap lookup, object sizing,
  bounds, and typed-validation path for every linked field.

Evidence reviewed:

- The source and SOW evidence listed under `## Analysis`.
- The SOW-0134 benchmark, hashes, journal verification, whole-file diagnostic,
  and focused profile.
- `systemd/systemd @ c0a5a2516d28601fb3afc1a77d7b42fcfe38fced`
  (`v260.1`):
  - `src/libsystemd/sd-journal/journal-file.c:2120-2177` writes/grows the
    entry-array before updating the owning count/tail.
  - `src/libsystemd/sd-journal/journal-file.c:2180-2234` links one DATA object
    through `entry_offset`, `entry_array_offset`, `n_entries`, and compact tail
    fields.
  - `src/libsystemd/sd-journal/journal-file.c:2236-2290` links the ENTRY and
    each DATA inverted index.

Affected contracts and surfaces:

- Internal DATA lookup result in `file_payload.rs`.
- Internal `EntryItem` state in `writer.rs`.
- DATA inverted-index publication in `writer_entry_arrays.rs`.
- Focused writer tests and SOW-0134 benchmark/profile evidence.
- No public API, file format, configuration, dependency, reader, Go
  implementation, compression behavior, FSS behavior, or publication cadence
  is intentionally changed.

Existing patterns to reuse:

- The manual lightweight DATA lookup header parsing in `file_payload.rs`.
- `EntryItem` as current-entry, sorted per-field writer state.
- Existing offset-array growth and malformed compact-tail fallback.
- Existing raw/structured byte-identity, compact, regular, stock verification,
  reopen, successor, live publication, and trusted-unique tests.

Risk and blast radius:

- Medium risk within the Rust low-level writer's DATA inverted-index path.
- A defect can write a journal that verifies structurally but has missing or
  incorrect DATA-to-ENTRY query indexes.
- Blast radius includes compact and regular Rust-written files and all callers
  of `JournalWriter`; public source compatibility remains unchanged.
- Memory grows only by a fixed metadata snapshot for each already-retained
  current `EntryItem`. The vector is cleared per entry and reused. There is no
  workload-cardinality or writer-lifetime cache.
- Because the state is entry-local, it cannot cross reopen, successor, or
  journal-file identity boundaries and has no collisions or evictions.

Sensitive data handling plan:

- Tests and benchmarks use only synthetic IDs, documentation-range addresses,
  and deterministic payloads.
- Durable artifacts contain no credentials, customer data, private endpoints,
  personal data, or proprietary incident content.

Implementation plan:

1. Add a private copyable resolved-DATA link-state type containing
   `n_entries`, `entry_array_offset`, and optional compact tail metadata.
2. Add an internal DATA lookup variant that returns offset plus link state while
   leaving the public offset-only API unchanged.
3. Capture link state from the matched DATA bytes already obtained for payload
   comparison; synthesize empty state for newly created DATA.
4. Store the state in `EntryItem` and use it for first-link, promotion, and
   steady append selection.
5. Detect adjacent duplicate DATA offsets after sorting. Preserve the
   trusted-unique misuse behavior with an authoritative on-disk fallback for
   every repeated offset.
6. Merge compact tail capacity validation and non-full mutation into one
   mutable offset-array guard. Preserve the existing fallback and growth path.
7. Keep the final DATA mutation after the offset-array write and update no
   entry-local state after partial failure.
8. Add focused invariants and run compatibility, benchmark, and profile gates.

Validation plan:

- Focused tests for regular/compact link-state decoding, new-DATA empty state,
  first link, promotion, steady append, tail growth, malformed-tail fallback,
  duplicate default behavior, trusted-unique duplicate fallback, reopen, and
  successor behavior.
- `cargo test -p systemd-journal-sdk-core`.
- `cargo test -p systemd-journal-sdk-log-writer`.
- `cargo check -p writer_core_bench`.
- `cargo fmt --all --check`.
- Existing raw/structured byte-identity tests and deterministic byte-identity
  matrix for all final states.
- Compact interoperability matrix and stock `journalctl --verify --file`.
- Live one-writer/multiple-reader matrix because DATA link publication is
  visible to active readers.
- Verify/query checks proving every DATA inverted index and returned row set.
- Interleaved before/after measurements with identical SOW-0134 parameters,
  preserved binary hashes, enough repetitions to distinguish the delta from
  host noise, and no universal-capacity wording.
- Before/after profiles with identical parameters, zero lost samples, and
  enough duration to assess the targeted cluster rather than rank a
  750-sample tie.
- Complete diff and equivalent-pattern search before reporting.
- Whole-SOW external review after local validation under the standing
  authorization.

Artifact impact plan:

- AGENTS.md: no change expected; project workflow policy is unchanged.
- Runtime project skills: no change expected; compatibility and review rules
  already cover this work.
- Specs: no behavior change is intended; record byte/index/API compatibility
  evidence rather than changing product contracts.
- End-user/operator docs: no change expected because this is an internal
  compatible optimization.
- End-user/operator skills: none exist for this surface.
- SOW lifecycle: activate only after SOW-0134 closes; close with implementation,
  evidence, review, and lifecycle state together.
- SOW-status.md: add pending now and update on activation/closure.

Open-source reference evidence:

- `systemd/systemd @ c0a5a2516d28601fb3afc1a77d7b42fcfe38fced`
  - `src/libsystemd/sd-journal/journal-file.c:2120-2290`

Open decisions:

- The user selected the entry-local resolved-link-state and single-tail-guard
  design. No product/design decision remains open.
- The one-active-SOW lifecycle dependency was resolved when SOW-0134 completed
  and committed at `bbf28f877c37a584254ea58b6b3026cf46fb9ba6`.

## Implications And Decisions

1. Production design:
   - Selected: entry-local resolved DATA link state plus one mutable compact
     tail guard.
   - Classification: long-term-best.
   - Reason: removes repeated work at its source without persistent cache
     lifecycle, collision, eviction, or cross-file correctness risks.
2. Payload cache:
   - Rejected for this SOW.
   - Reason: the previous cache has mixed workload evidence and would combine a
     separate hashing/lookup design with link-publication state reuse.
3. Whole-file mmap:
   - Rejected as the primary change.
   - Reason: SOW-0134 observed only a 0.911% median difference with overlapping
     ranges.
4. Public configuration:
   - Rejected.
   - Reason: the optimization is internal and must preserve default semantics.

## Plan

1. Close SOW-0134 after review remediation and re-review.
2. Activate this SOW and preserve the exact pre-change benchmark binary.
3. Implement the approved internal state flow and compact tail single-open path.
4. Run focused and full correctness/interoperability validation.
5. Run identical before/after benchmark and profile comparisons.
6. Review, remediate verified findings, and report the local diff/commit.

## Implementation And Review Plan

Implementation:

- The project manager implements the approved design directly after activation.
- Stop if the design requires a public API, file-format, durability,
  compatibility, persistent-cache, or publication-order decision not recorded
  here.

Reviewers:

- External review has standing authorization for this conversation.
- Follow the system-wide `external-reviewers` skill after complete local
  validation; do not copy volatile reviewer or harness instructions here.

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

- Stop on any unexpected public/runtime contract expansion, index mismatch,
  byte delta, live-reader regression, ambiguous performance evidence, or audit
  failure.
- Record the failure and return to the user if the approved design cannot be
  applied without deviation.

## Execution Log

### 2026-07-26

- Created as the user-selected production follow-up from SOW-0134.
- No production source was changed under this pending SOW.
- SOW-0134 completed its diagnostic, validation, external-review, audit, and
  lifecycle gates and committed at
  `bbf28f877c37a584254ea58b6b3026cf46fb9ba6`.
- Activated SOW-0135 with the user-selected entry-local resolved-link-state and
  single compact-tail guard design unchanged.
- Implemented the approved design in the canonical `journal-core` writer:
  - DATA lookup now exposes a private matched-object state snapshot containing
    `n_entries`, `entry_array_offset`, and compact tail metadata without
    changing the public offset-only lookup API.
  - Each already-retained current-entry item carries that snapshot. Newly
    created DATA starts with an explicit empty snapshot.
  - Trusted-unique duplicate misuse clears the snapshot after the first
    adjacent duplicate offset so every later publication rereads authoritative
    on-disk state.
  - Compact non-full tail append now validates and writes through one mutable
    offset-array guard.
  - Offset-array writes still precede final DATA count/tail publication.
- Added six focused tests covering new-DATA empty state, default duplicate
  elimination, trusted-unique duplicate fallback marking and real publication,
  regular/compact state growth, compact tail growth, and malformed compact-tail
  fallback.
- Preserved and identified the pre-change diagnostic release binary, built the
  changed release binary, and ran an alternating release benchmark with two
  warmups and twelve measurements per binary.
- Ran matching before/after profiles and validated the full DATA inverted-index
  chains in representative 100,000-row journals.
- The committed directory-mode interoperability runners currently fail before
  journal creation because their Rust test writer does not provide the explicit
  synthetic machine ID now required by the strict directory-writer contract.
  This is pre-existing validation-harness drift, not a failure in the changed
  low-level writer. The same compact and live feature sets were therefore run
  through the test writer's direct-file mode without changing the harness.
- The all-language deterministic byte-identity runner generated and verified
  the Rust and Go online/offline/archived files, then stopped because this host
  has no `meson` executable for its pinned systemd v260.1 reference build.
  No package was installed. Before/changed deterministic Rust bytes and the
  Rust/Go final-state pairs provide compatibility evidence, but the systemd
  v260.1 three-way byte-identity proof remains unavailable in this environment.
- Ran the authorized whole-SOW external review with the exact requested
  reviewers. Prompt SHA-256:
  `2ca1d75b3fbda3d6118ee96d31697e58bdca81a8f7043fd68d424dfc86d4859d`.
  All seven invocations completed and returned `PRODUCTION GRADE`.
- Review follow-up tracking:
  - SOW-0136 tracks the pre-existing explicit-identity directory-harness drift.
  - SOW-0137 tracks measured analysis of remaining canonical writer array
    reopens outside the selected compact DATA-tail change.
- Final `git diff --check` passed.
- Final `bash .agents/sow/audit.sh` passed with one current SOW before closure,
  eleven pending SOWs including SOW-0136/SOW-0137, clean status/directory
  consistency, valid pre-implementation gates, sanitized open-source
  references, no sensitive-data patterns, and a clean final verdict.
- The post-lifecycle audit also passed after moving SOW-0135 to `done/` and
  updating both ledgers; it reported no current SOW and clean final status.

## Validation

Acceptance criteria evidence:

- Internal state flow:
  - `rust/src/crates/journal-core/src/file/file_payload.rs:32-62` defines the
    private generic lookup result and bounded resolved-link snapshot.
  - `file_payload.rs:433-593` keeps the public offset-only lookup intact while
    the writer-only variant extracts link state from the already-read matched
    DATA bytes.
  - `rust/src/crates/journal-core/src/file/writer.rs:26-31` stores one snapshot
    in each current `EntryItem`; the vector is cleared and reused per entry.
  - `writer.rs:754-793` captures existing state and supplies explicit empty
    state for new DATA.
- Correct fallback:
  - `writer.rs:599-615` preserves default deduplication and marks every
    trusted-unique repeated offset after the first for an authoritative disk
    reread.
  - `rust/src/crates/journal-core/src/file/writer_entry_arrays.rs:409-491`
    selects snapshot or disk state without changing the public writer API.
  - Focused compact/regular duplicate tests observe four DATA inverted-index
    links from two rows containing two intentionally repeated payloads each.
- Compact tail and publication order:
  - `writer_entry_arrays.rs:346-380` opens the cached compact tail once for
    validation and non-full mutation, with the existing chain-walk fallback.
  - `writer_entry_arrays.rs:523-547` writes/grows the offset-array link before
    reopening DATA for final count/tail publication, preserving the previous
    visibility and failure order.
- Bounded memory and file identity:
  - The state adds three optional scalar values to each already-retained
    current-entry item. It has no writer-lifetime collection, payload copy,
    hash collision, eviction, journal-file key, reopen state, or successor
    state.
  - Reopen and successor writers reconstruct current-entry state from their own
    `JournalFile` lookup. Existing high-level reopen/rotation tests passed.
- Compatibility:
  - No public API, option, dependency, format field, object order, publication
    cadence, compression mode, FSS mode, or reader code changed.
  - All twelve 100,000-row before/changed measurement pairs are byte-for-byte
    identical.
  - Raw and structured writers remain byte-identical in compact and regular
    formats, with hashes identical to SOW-0134.

Tests or equivalent validation:

- `cargo fmt --manifest-path rust/Cargo.toml --all --check`: passed.
- `cargo test --manifest-path rust/Cargo.toml -p
  systemd-journal-sdk-core`: passed, 80 unit tests; 3 documentation tests
  ignored as before.
- `cargo test --manifest-path rust/Cargo.toml -p
  systemd-journal-sdk-log-writer`: passed, 6 unit tests, 51 integration tests,
  and 1 documentation test.
- `cargo check --manifest-path rust/Cargo.toml -p writer_core_bench`: passed.
- `git diff --check`: passed.
- Focused new tests:
  - `new_data_entry_item_starts_with_empty_link_state`
  - `default_duplicate_elimination_keeps_one_resolved_state`
  - `trusted_unique_duplicate_offsets_use_authoritative_fallback_state`
  - `resolved_data_link_state_tracks_regular_and_compact_tail_growth`
  - `trusted_unique_duplicate_fallback_reads_updated_data_state`
  - `compact_invalid_cached_tail_falls_back_to_authoritative_array_chain`
- Raw/structured byte identity, 512 NetFlow-shaped rows, trusted unique,
  uncompressed, publication disabled:
  - compact raw and structured:
    `0d8c003eb289788efc94e66c06ade3230f715b69c2f01a991f51c6bc4116d32f`
  - regular raw and structured:
    `1b08a3f28dfe4d190428e74f5a1e56c489e4e10ed13dcb7fb3af1852433edaec`
  - all four files passed stock `journalctl --verify --file`.
- Compact direct-file interoperability on stock systemd
  `255 (255.4-1ubuntu8.16)`:
  - none, zstd, xz, and lz4; 256 entries each;
  - structural oracle, stock verify/JSON/export/indexed binary match,
    libsystemd binary reads, and Go/Rust JSON/export reads;
  - 40 passed, 0 failed;
  - evidence:
    `.local/benchmarks/sow0135/compact-file-mode/20260726T112032Z/summary.json`.
- Live direct-file interoperability on the same stock systemd:
  - regular, zstd, xz, lz4, compact, compact-zstd, compact-xz, compact-lz4,
    and sealed;
  - 30 entries per feature, six concurrent polling readers (two each for
    stock, Go, and Rust), one live libsystemd reader, final reads, structural
    oracle, and stock verification;
  - 9 feature cases passed, 0 failed;
  - evidence:
    `.local/benchmarks/sow0135/live-file-mode/20260726T112252Z/summary.json`.
- DATA inverted-index validation:
  - the structural oracle walked and validated every DATA entry-array chain in
    representative before/changed 100,000-row journals;
  - each file contained 386 DATA objects and 2,529 ENTRY_ARRAY objects, with
    all 2,529 arrays referenced and both results passing;
  - stock indexed before/changed query counts matched: 100,000 for
    `FLOW_VERSION=v5`, 1,172 for one repeated source, 782 for another, and 391
    for a five-field identity conjunction;
  - evidence:
    `.local/benchmarks/sow0135/index-validation-summary.json`.
- Deterministic final-state byte identity:
  - Rust and Go were byte-identical for online, offline, and archived states;
  - hashes were
    `400eae423394117fe5cbd1775df1241b86be57b22a1afbd6f4e68a0f76865571`,
    `25800a4a660cd8de84b40a9f813dc43262ace7f662f3602d59b5ad73776bd920`,
    and
    `b8d7518141ea06dd6908fa393627beb2388a3a8e413f3d9f9e0afc57c3919a5a`;
  - all retained the expected DATA hash-chain depth of 3;
  - full systemd v260.1 comparison was blocked by missing `meson`, as recorded
    above.

Real-use evidence:

- Exact focused workload: 100,000 rows, NetFlow-v5 repeating-256, compact,
  uncompressed, structured fields, trusted unique payloads, live publication
  disabled, windowed mmap.
- Preserved pre-change binary:
  `.local/benchmarks/sow0134/binaries/writer_core_bench-sow0134-diagnostic`,
  SHA-256
  `ebc7892b9b65fd1963dec45150ce37ac98c939d377ba21080e74440dc0caae20`.
- Changed binary:
  `.local/benchmarks/sow0135/binaries/writer_core_bench-changed`, SHA-256
  `358070e0173e47249cb71eaa254fb34ff44c4ea3a688d61c614d8449d3ff276e`.
- Alternating release measurements, two warmups and twelve measurements per
  binary:
  - before: min 110,608.256, median 119,711.172, max 121,884.200 rows/s;
  - changed: min 139,145.935, median 145,635.679, max 148,539.203 rows/s;
  - median improvement: 21.656%;
  - all twelve paired output journals were byte-identical;
  - evidence:
    `.local/benchmarks/sow0135/interleaved-release/summary.json`.
- Matching 300,000-row cycle profiles:
  - before: 115,676.156 rows/s, 2,281 samples, zero lost;
  - changed: 130,742.046 rows/s, 2,134 samples, zero lost;
  - instrumented throughput improvement: 13.024%;
  - `JournalWriter::publish_entry_links` self-time: 6.91% to 1.55%;
  - `WindowManager::get_window`: 6.84% to 4.37%;
  - `WindowManager::get_slice`: 5.87% to 4.10%;
  - the former `compact_data_tail_capacity` and
    `append_to_existing_data_tail` symbols disappeared;
  - relative self-time is not additive and some remaining symbols rose in
    percentage as total work decreased; no claim is made that every symbol
    decreased;
  - before `perf.data` SHA-256:
    `ffc6d40cbc2ab169fe91364c21a4a773108a608170df5101e2edf9ff0730d519`;
  - changed `perf.data` SHA-256:
    `14b7b88cda45ab768343ac9ed8a395f0848662390e41749ff5f4c284b1309a00`;
  - profile journals were byte-identical with SHA-256
    `d8cfec09c1b57481639b2b59c92d54fd5d1c7e9957d4613b2fbe9c333b578219`.
- These are workload- and host-specific SDK results. They do not establish
  universal SDK capacity or prove the collector's authoritative
  100,000-datagram/s end-to-end result.

Reviewer findings:

- Round 1 outputs:
  `.local/reviews/sow0135-round1/{claude,codex,glm,minimax,deepseek,mimo,qwen}.txt`.
- Verdicts:
  - Claude: `PRODUCTION GRADE`
  - Codex: `PRODUCTION GRADE`
  - GLM: `PRODUCTION GRADE`
  - MiniMax: `PRODUCTION GRADE`
  - DeepSeek: `PRODUCTION GRADE`
  - Mimo: `PRODUCTION GRADE`
  - Qwen: `PRODUCTION GRADE`
- No reviewer reported an in-blast-radius P0, P1, or P2 defect. Codex reported
  the pre-existing directory-harness identity drift as unrelated P2; it is now
  tracked by pending SOW-0136.
- Verified non-blocking P3 observations:
  - The public offset-only lookup now decodes two additional `u64` fields from
    the already-fetched 64-byte header even though its `()` result discards
    them. This is a tiny per-DATA-probe reader/filter cost, not a semantic,
    mmap, validation, or allocation change. Avoiding it would complicate this
    shared internal path without evidence of a reader regression, so it is
    accepted for this writer SOW.
  - `optional_nonzero_u64_field<T>` expresses the fixed struct-layout bound
    through a debug assertion against `size_of::<T>()`, while the actual slice
    length is guaranteed by its only caller. This is safe for the two
    compile-time `offset_of!(DataObjectHeader, ...)` call sites but mildly
    indirect. No source change is justified for this private fixed-use helper.
  - `DataLookupResult<T>` lost unused `Debug`, `Clone`, and `Copy` derives.
    Nothing requires those traits; restoring them would be cosmetic only.
  - Snapshot promotion can allocate an ENTRY_ARRAY before the later mutable
    DATA guard is constructed, whereas the old path opened DATA first. The
    lookup already enforces aligned offset, object type, arena bounds, compact
    prefix size, full object availability, and matching payload. Under the
    supported one-writer contract, the subsequent zerocopy split cannot acquire
    a new failure condition. The theoretical externally-mutated-file scenario
    is outside the writer contract and does not justify restoring the removed
    hot-path validation.
  - Equivalent redundant array opens remain in global ENTRY-array and
    regular/fallback DATA paths. They are correct, were not isolated by the
    selected compact workload, and are tracked for measurement/design in
    SOW-0137 rather than being copied blindly into this SOW.
  - Existing dead-looking branches in the authoritative duplicate fallback,
    minor helper duplication, mixed offset-access style, and missing explanatory
    comments are maintainability notes without functional effect.
- Discarded or corrected findings:
  - One reviewer described compact DATA objects with declared size 64 through
    71 as a pre-existing panic risk. `validate_data_payload_info()` already
    requires the compact `payload_prefix_size` of 72 before the slice is
    returned, so this scenario is rejected with `InvalidObjectSize`.
  - One reviewer said no extra parsing occurs on the public offset-only path.
    The path does decode the two new header fields; the correct disposition is
    the first verified P3 above.
  - A reviewer inferred removed documentation from the two merged compact-tail
    helpers. Those helpers had no API documentation; the observation is not a
    defect.
- Unrelated/coverage observations:
  - Directory-mode harness identity drift: tracked by SOW-0136.
  - Legacy `jf` writer old path: product-scope decision remains unchanged and
    pending SOW-0098 already tracks broad legacy/core duplication debt.
  - Missing local systemd v260.1 reference build: environmental evidence gap;
    no repo implementation is required.
  - Downstream NetFlow end-to-end retest: owned by the integrating worker.
- No source remediation or repeat review was required because there was no
  verified blocking finding and no material post-review source change.

Same-failure scan:

- Searched the Rust workspace for every `EntryItem`, DATA lookup, DATA link
  publication, compact tail, and offset-array mutation pattern.
- The only canonical current writer path is the changed `journal-core` path.
- `rust/src/crates/jf/journal_file/src/writer.rs` contains a separate legacy
  writer with equivalent old link behavior. Product scope explicitly designates
  `journal-core` plus `journal-log-writer` as canonical and recommends new
  integrations use those crates. Changing the legacy public compatibility
  writer would expand this measured production SOW without NetFlow evidence and
  is not required for the requested target.
- Equivalent global ENTRY-array and regular/fallback DATA-array reopens remain.
  Their state/publication lifecycles differ from the selected compact-tail path,
  and the focused benchmark did not isolate them. Pending SOW-0137 requires
  measurement and a user design decision before any implementation.

Sensitive data gate:

- The plan uses only synthetic/public evidence and contains no raw sensitive or
  personal data.

Artifact maintenance gate:

- AGENTS.md: no update required; no project workflow or public contract changed.
- Runtime project skills: no update required; existing writer compatibility and
  external-review rules covered the work.
- Specs: no update required; public/file-format semantics are unchanged.
- End-user/operator docs: no update required; the optimization is internal.
- End-user/operator skills: none apply to this internal path.
- SOW lifecycle: completed and moved to `done/` with implementation, evidence,
  follow-up SOWs, ledgers, and source in one commit.
- SOW-status.md: completed state is recorded in both ledgers.

Specs update:

- None. No product contract changed.

Project skills update:

- None. No reusable workflow gap was introduced by the source change. The
  pre-existing directory-mode validation-harness drift is recorded as a
  remaining tooling issue rather than silently changing identity behavior.

End-user/operator docs update:

- None. No public API or operator behavior changed.

End-user/operator skills update:

- None.

Lessons:

- Reusing state already validated during current-entry DATA lookup removes more
  work than a writer-lifetime payload cache while avoiding cache lifecycle,
  identity, collision, and eviction risks.
- Compact tail metadata is useful only as a hint. A correct optimization keeps
  an authoritative chain-walk fallback for malformed or stale tail metadata.
- Exact byte identity plus a structural walk of every DATA entry-array chain is
  materially stronger compatibility evidence than closed-file verification
  alone.
- Validation harnesses must evolve with strict explicit-identity contracts;
  otherwise directory-mode tests can fail before exercising the writer path.

Follow-up mapping:

- The NetFlow worker owns the authoritative identical 100,000-datagram/s
  collector retest after integrating the local SDK change. This is outside this
  repository and is explicitly rejected as a new local SOW.
- Pending SOW-0136 tracks the Rust directory-mode compact/live validation
  harness repair with explicit deterministic synthetic identity and an
  unchanged strict production writer contract.
- Pending SOW-0137 tracks measurement and user design decisions for equivalent
  canonical writer array reopens.
- The full pinned systemd v260.1 three-way byte-identity matrix should be rerun
  where its existing build dependency is available. This is an environmental
  validation rerun, not a repository implementation, and is explicitly
  rejected as a separate code SOW.

## Outcome

Completed. The canonical Rust writer now reuses entry-local resolved DATA-link
state and one compact-tail mutable guard. The focused release benchmark improved
by 21.656% at the median with byte-identical output, the targeted profile path
decreased, correctness/interoperability gates passed, and all seven authorized
reviewers returned `PRODUCTION GRADE`.

## Lessons Extracted

The production result supports entry-local state reuse rather than a
writer-lifetime DATA cache: it achieved a 21.656% median improvement on the
focused workload with byte-identical output and no cross-file state.

## Followup

The authoritative NetFlow end-to-end retest remains with the integrating
worker. SOW-0136 tracks validation-harness identity drift, and SOW-0137 tracks
remaining measured writer array-open work. The unavailable local systemd v260.1
reference build is recorded as an environmental validation rerun.

## Regression Log

None yet.
