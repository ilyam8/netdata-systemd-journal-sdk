# SOW-0137 - Rust Writer Residual Array-Open Performance

## Status

Status: open

Sub-state: pending activation. This SOW tracks non-blocking equivalent
offset-array open patterns found during SOW-0135 review; it authorizes no
implementation.

## Requirements

### Purpose

Measure and, if justified, remove remaining redundant offset-array object opens
in the canonical Rust writer without changing journal bytes, publication order,
or index correctness.

### User Request

SOW-0135 delivered one measured compact DATA-link optimization. Its
same-pattern search and external review found related ref-then-mut and
re-read-before-mut patterns in the global ENTRY array and regular/fallback DATA
array paths. These are possible additional gains, not correctness blockers for
SOW-0135.

### Assistant Understanding

Facts:

- Global ENTRY-array append reads the tail capacity and then opens the same
  array mutably to write.
- Regular DATA-array and compact fallback traversal can compute the tail
  capacity during the chain walk, then reopen the same tail read-only and
  mutably.
- New array allocation opens the new array to size it and callers reopen it to
  write the first slot.
- SOW-0135's measured workload was compact with valid cached tails, so it did
  not isolate the regular/fallback paths.

Inferences:

- Some object validation/mmap work may still be avoidable, especially for
  regular output and the per-entry global ENTRY array.
- The paths have different header ownership and publication lifecycles, so the
  SOW-0135 compact DATA-tail merge must not be copied mechanically.

Unknowns:

- The measurable cost of each pattern on compact and regular workloads.
- Whether a single mutable guard is compatible with each path's allocation,
  growth, header publication, and failure ordering.

### Acceptance Criteria

- Profile and benchmark each candidate with identical compact and regular
  parameters before editing.
- Present design alternatives and publication/failure-order implications to the
  user before implementation.
- Accept an optimization only when it materially reduces the targeted path.
- Preserve exact bytes, all ENTRY and DATA array chains, live visibility,
  compact/regular correctness, public APIs, and bounded memory.

## Analysis

Sources checked:

- `.agents/sow/current/SOW-0135-20260726-rust-writer-entry-link-state-reuse.md`
- `rust/src/crates/journal-core/src/file/writer_entry_arrays.rs`
- SOW-0135 external review outputs under
  `.local/reviews/sow0135-round1/`

Current state:

- SOW-0135 removes the current-entry DATA reopen and non-full compact-tail
  double open on its measured path.
- Other array-open patterns remain correct but may be more expensive than
  necessary.

Risks:

- Combining read and mutation too early can change validation-before-mutation
  behavior on malformed files.
- Reusing a guard across allocation or growth can violate borrow rules or alter
  array-before-header publication ordering.
- Optimizing unmeasured paths can add complexity without throughput value.

## Pre-Implementation Gate

Status: pending activation, measurement, and user design decision

Problem / root-cause model:

- Candidate writer paths reopen an offset-array object after already reading
  the metadata needed to select or size the mutation.

Evidence reviewed:

- SOW-0135 source, benchmark/profile results, same-pattern search, and external
  review.

Affected contracts and surfaces:

- Canonical Rust global ENTRY-array publication.
- Regular DATA entry arrays.
- Compact DATA authoritative fallback.
- Array allocation/growth if measurement includes it.

Existing patterns to reuse:

- SOW-0135 single compact-tail guard where publication and failure ordering are
  equivalent.
- Structural oracle and live compatibility matrices.

Risk and blast radius:

- Medium. A defect can damage global entry traversal or DATA inverted indexes.

Sensitive data handling plan:

- Use deterministic synthetic repository fixtures only.

Implementation plan:

- Pending measurement and user-selected scope/design.

Validation plan:

- Focused compact and regular unit tests for append, growth, malformed state,
  reopen, and publication order.
- Full affected Rust tests and writer benchmark target check.
- Structural oracle for all ENTRY/DATA array chains.
- Stock verification, indexed queries, live matrix, byte identity, and matching
  before/after profiles.

Artifact impact plan:

- AGENTS.md, specs, skills, and end-user docs are expected to remain unchanged
  for an internal compatible optimization.
- Update SOW ledgers on activation and closure.

Open-source reference evidence:

- Activation should recheck systemd v260.1 entry-array append ordering.

Open decisions:

- Which candidate paths have enough measured cost to justify implementation.
- Whether they can share a guard-reuse primitive without obscuring publication
  order.

## Implications And Decisions

Pending activation.

## Plan

1. Benchmark/profile compact and regular candidate paths.
2. Present evidence and design options to the user.
3. Implement only the selected causal improvement.
4. Validate bytes, indexes, live visibility, and profiles.

## Implementation And Review Plan

Implementation:

- No implementation until activation and user design selection.

Reviewers:

- Recommend whole-SOW external review after local validation and before GitHub
  submission. Use the system-wide `external-reviewers` skill after user
  authorization.

Failure handling:

- Stop on any byte/index/publication-order change or non-material measured gain.

## Execution Log

### 2026-07-26

- Created from SOW-0135 same-pattern review as a non-blocking measured
  performance follow-up.

## Validation

Pending activation.

## Outcome

Pending.

## Lessons Extracted

Pending activation.

## Followup

Pending activation.
