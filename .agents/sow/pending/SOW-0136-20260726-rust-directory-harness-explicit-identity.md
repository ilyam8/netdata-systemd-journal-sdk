# SOW-0136 - Rust Directory Harness Explicit Identity

## Status

Status: open

Sub-state: pending activation. This SOW tracks validation-harness drift found
while closing SOW-0135; it authorizes no implementation by itself.

## Requirements

### Purpose

Restore the committed Rust directory-mode compact and live interoperability
matrices after the strict writer contract began requiring explicit
caller-provided identity.

### User Request

SOW-0135 validation found that the committed Rust directory-mode test writer
fails before journal creation because its synthetic test configuration supplies
no machine ID. The production writer contract must remain strict; the harness
must adapt rather than reintroducing implicit host discovery or generated
identity in production paths.

### Assistant Understanding

Facts:

- `tests/interoperability/run_compact_matrix.py` selects the Rust test writer in
  directory mode.
- `tests/interoperability/run_live_matrix.py` also defaults the Rust test writer
  to directory mode for unsealed feature cases.
- `rust/src/internal/testcmd/livewriter/src/main.rs` constructs its directory
  `Origin` with `machine_id: None`.
- The current high-level directory writer correctly rejects the missing
  identity before journal creation.
- SOW-0135 direct-file compact and live matrices passed, so this is
  validation-tooling drift rather than evidence of a low-level writer defect.

Inferences:

- The test tool needs an explicit deterministic synthetic identity supplied by
  its command surface or harness invocation.
- Hard-coding host identity discovery or weakening production validation would
  violate the runtime-purity and strict writer contracts.

Unknowns:

- Whether the narrowest maintainable interface is a test-tool machine-ID
  argument, a complete synthetic `Origin` fixture option, or explicit harness
  fixture wiring.
- Whether other committed directory-mode test runners use the same incomplete
  configuration and should be repaired in the same SOW.

### Acceptance Criteria

- The user selects the synthetic-identity interface before implementation.
- Rust directory-mode compact and live interoperability cases create journals
  without host identity discovery.
- The full committed compact and live feature matrices pass for the Rust
  directory writer with deterministic synthetic IDs.
- Production writer identity requirements, runtime purity, public SDK defaults,
  and host-observation rules remain unchanged.
- A same-pattern search covers every internal Rust directory-writer test tool
  and interoperability runner.

## Analysis

Sources checked:

- `AGENTS.md`
- `.agents/skills/project-journal-compatibility/SKILL.md`
- `.agents/sow/current/SOW-0135-20260726-rust-writer-entry-link-state-reuse.md`
- `tests/interoperability/run_compact_matrix.py`
- `tests/interoperability/run_live_matrix.py`
- `rust/src/internal/testcmd/livewriter/src/main.rs`

Current state:

- Direct-file validation exercises the canonical low-level writer successfully.
- Directory-mode harness coverage is blocked before the changed writer path is
  exercised.

Risks:

- A test-only repair could accidentally leak implicit identity generation into
  production code.
- Different harnesses could choose inconsistent synthetic identities and lose
  deterministic output.
- A narrow one-runner fix could leave equivalent committed matrix failures.

## Pre-Implementation Gate

Status: pending activation and user design decision

Problem / root-cause model:

- The test harness still constructs a directory writer without the explicit
  machine identity now required by the production contract.

Evidence reviewed:

- The sources and SOW-0135 failure evidence listed above.

Affected contracts and surfaces:

- Internal Rust `livewriter` test command.
- Compact and live interoperability harnesses.
- Synthetic test identity only; no production SDK identity behavior.

Existing patterns to reuse:

- Repository deterministic UUID fixtures.
- Explicit caller-provided identity in production writer constructors.

Risk and blast radius:

- Low to medium if confined to internal test tools; unacceptable if it weakens
  production identity validation or adds host discovery.

Sensitive data handling plan:

- Use fixed synthetic UUIDs only. Do not read or record host identity.

Implementation plan:

- Pending user selection of the test-tool identity interface after activation.

Validation plan:

- Run the Rust directory writer through all compact compression variants and
  the complete live feature matrix.
- Require stock journalctl, stock libsystemd, Go/Rust readers, structural
  validation, final verification, and deterministic synthetic identity.
- Run affected Rust test-tool checks, formatting, and project audit.

Artifact impact plan:

- AGENTS.md: no change expected.
- Runtime project skills: no change expected unless the repair exposes a
  reusable validation rule not already covered.
- Specs: no production contract change expected.
- End-user/operator docs: no change expected.
- SOW status ledgers: update on activation and closure.

Open-source reference evidence:

- None required yet.

Open decisions:

- Test-tool argument versus harness-owned fixture wiring.
- Scope of equivalent internal directory-mode test runners.

## Implications And Decisions

Pending activation.

## Plan

1. Inventory every internal directory-mode writer harness.
2. Present the concrete identity-interface options and evidence to the user.
3. Implement only the selected test-only design.
4. Rerun the full directory-mode matrices and audit.

## Implementation And Review Plan

Implementation:

- No implementation until this SOW is activated and the user selects the
  design.

Reviewers:

- Recommend whole-SOW external review after local validation and before GitHub
  submission. Use the system-wide `external-reviewers` skill after user
  authorization.

Failure handling:

- Stop if the repair requires a production API/default change, implicit host
  discovery, or nondeterministic identity.

## Execution Log

### 2026-07-26

- Created as the explicit follow-up mapping for the pre-existing directory-mode
  validation-harness drift recorded by SOW-0135 and its external review.

## Validation

Pending activation.

## Outcome

Pending.

## Lessons Extracted

Pending activation.

## Followup

Pending activation.
