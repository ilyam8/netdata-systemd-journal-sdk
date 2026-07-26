# SOW-0135 - Rust Writer Entry-Link State Reuse

## Status

Status: open

Sub-state: Design selected and pre-implementation analysis complete. Activation
is blocked only on reviewed closure of SOW-0134.

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

Status: blocked

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
- Gate status is blocked only by the one-active-SOW lifecycle dependency on
  SOW-0134.

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

## Validation

Acceptance criteria evidence:

- Pending implementation.

Tests or equivalent validation:

- Pending implementation.

Real-use evidence:

- Pending implementation.

Reviewer findings:

- Pending implementation.

Same-failure scan:

- Pending implementation.

Sensitive data gate:

- The plan uses only synthetic/public evidence and contains no raw sensitive or
  personal data.

Artifact maintenance gate:

- AGENTS.md: pending closure assessment.
- Runtime project skills: pending closure assessment.
- Specs: pending closure assessment.
- End-user/operator docs: pending closure assessment.
- End-user/operator skills: pending closure assessment.
- SOW lifecycle: open/pending is consistent.
- SOW-status.md: pending entry added with SOW creation.

Specs update:

- Pending closure assessment.

Project skills update:

- Pending closure assessment.

End-user/operator docs update:

- Pending closure assessment.

End-user/operator skills update:

- Pending closure assessment.

Lessons:

- Pending implementation.

Follow-up mapping:

- Pending implementation.

## Outcome

Pending.

## Lessons Extracted

Pending.

## Followup

None yet.

## Regression Log

None yet.
