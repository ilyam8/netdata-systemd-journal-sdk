# SOW-0139 - Go Live Journalctl Boot-Selector Compatibility

## Status

Status: open

Sub-state: pending activation. This SOW tracks a pre-existing test-harness
compatibility failure reproduced on systemd 255; it authorizes no
implementation.

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

- Whether the portable fix is a version-aware selector, a synthetic boot
  fixture adjustment, or a different stock-journalctl invocation with
  equivalent semantics.
- Whether the same invocation occurs in other language-specific tests or
  query/live harnesses and needs one shared repair.

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

Status: pending activation, cross-version investigation, and user design
decision

Problem / root-cause model:

- Stock journalctl 255 interprets the current synthetic-file follow invocation
  differently from the systemd 260.1 environment where the harness originally
  passed.

Evidence reviewed:

- Identical final-tree and base-commit failures plus the passing shared live
  matrix described above.

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

- Pending cross-version command analysis and user-selected design.

Validation plan:

- Run the three focused tests and the complete Go journal suite on systemd 255.
- Run the same cases on systemd 260.1.
- Run regular, compact, compressed, sealed, and interrupted live matrices with
  stock journalctl, stock libsystemd, Rust, and Go readers.

Artifact impact plan:

- AGENTS.md, runtime skills, specs, and production docs should remain
  unchanged for a test-only repair.
- Update SOW ledgers on activation and closure.

Open-source reference evidence:

- Activation should compare the relevant journalctl boot/follow behavior in
  `systemd/systemd` tags matching 255 and 260.1.

Open decisions:

- Exact cross-version stock-journalctl invocation.
- Whether one shared helper can cover every equivalent harness.

## Implications And Decisions

Pending activation and user design selection.

## Plan

1. Inventory all equivalent stock live-reader invocations.
2. Reproduce and explain the systemd 255/260.1 behavior difference.
3. Present test-only design alternatives to the user.
4. Implement only the selected compatibility repair.
5. Rerun focused, full-suite, and cross-version live validation.

## Implementation And Review Plan

Implementation:

- No implementation until activation and user design selection.

Reviewers:

- Recommend whole-SOW external review after local validation and before GitHub
  submission. Use the system-wide `external-reviewers` skill after user
  authorization.

Failure handling:

- Stop if a proposed repair weakens stock live-follow coverage or changes
  production SDK behavior.

## Execution Log

### 2026-07-26

- Created as SOW-0138's explicit mapping for three pre-existing systemd 255
  stock-journalctl test failures.

## Validation

Pending activation.

## Outcome

Pending.

## Lessons Extracted

Pending activation.

## Followup

Pending activation.
