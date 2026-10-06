# SOW-0148 - Indexed Snapshot Archive State

## Status

Status: completed
Sub-state: bounded accessor implemented, independently inspected and validated; local commit includes lifecycle closure.

## Requirements

### Purpose

Let recovery inspect captured archive state without reopening a journal or scanning entries.

### User Request

Approved DEM option-1 recovery needs to reject archive names whose captured header is not archived. Expose equivalent Go IsArchived and Rust is_archived accessors, preserve frozen metadata, validate, and commit locally. No publication or new recovery policy.

### Assistant Understanding

Facts:
- SOW-0147 is completed. Both snapshot implementations already copy the complete header.
- Neither snapshot exposes its header or archive state publicly.
- Pending SDK work covers other integration, release, clock, filter and performance scopes; none overlaps this accessor.

Inferences:
- Comparing the existing owned state byte is sufficient; no extra reader, parsing, storage or compatibility adapter is needed.

Unknowns:
- No unresolved implementation or product decision.

### Acceptance Criteria

- Go IsArchived() bool and Rust is_archived() -> bool return true only for captured archived state.
- Captures of online and offline files return false; archiving after capture does not change the earlier result.
- Public docs and product spec describe frozen state and distinguish it from integrity certification.
- Focused Go/Rust tests, docs checks, review and local SOW audit pass before completion.

## Analysis

Sources checked:
- AGENTS.md; completed SOW-0147; documentation/dem-history-index-design.md option 1.
- go/journal/indexed_snapshot.go and tests; rust/src/journal/src/indexed_snapshot.rs and tests/indexed_snapshot.rs.
- docs/Indexed-Snapshots.md and .agents/sow/specs/product-scope.md.

Current state:
- Clean feat/indexed-snapshot at 280056d before work. No public cheap frozen-state alternative exists.

Risks:
- Reading the live mapping instead of the owned header would change eligibility after capture; tests must distinguish both paths.
- Archived state alone does not prove clean lifecycle provenance or graph integrity; documentation must preserve that distinction.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:
- Consumer recovery cannot inspect private snapshot state and would otherwise need a second read outside the captured contract.

Evidence reviewed:
- Source owners and completed design above; pending/current SOW inventory; journal compatibility, orchestration and documentation skills.

Affected contracts and surfaces:
- Two additive constant-time snapshot accessors, tests, docs and current product spec. No file format or reader/writer lifecycle change.

Existing patterns to reuse:
- EntryCount/entry_count reads owned header metadata. Existing snapshot fixtures use actual SDK writers.

Risk and blast radius:
- One comparison per accessor; existing capture and traversal stay unchanged. Frozen header behavior is covered in both languages.

Sensitive data handling plan:
- Synthetic journals and public source only. No credentials, telemetry or personal data in artifacts. Tool caches remain in temporary task storage.

Implementation plan:
1. Add matching accessors and real writer lifecycle tests for active, offline, archived and capture-before-archive states.
2. Document semantics, validate both languages and docs, obtain main-agent independent inspection, audit and commit the completed follow-up.

Validation plan:
- Focused snapshot tests in both languages; wiki structure and affected verified examples; diff checks and SOW audit.
- Tests call public accessors against actual writer state transitions. No performance benchmark is required for an owned-byte comparison without I/O or allocation.

Artifact impact plan:
- AGENTS.md: unchanged workflow.
- Runtime project skills: unchanged workflow and compatibility guidance.
- Specs: add frozen archived-state metadata contract.
- End-user/operator docs: update Indexed-Snapshots API and recovery guidance.
- End-user/operator skills: none exist.
- SOW lifecycle: narrow follow-up; completed SOW-0147 remains unchanged.
- SOW-status.md: record follow-up start and completion.

Open-source reference evidence:
- No external source inspection is needed; format state constants and copied headers are already implemented locally.

Open decisions:
- None. The approved recovery boundary requires this metadata, and the assignment authorizes the accessor and local commit.

## Implications And Decisions

1. Preserve approved option-1 recovery. A true accessor result is one provenance condition, not standalone integrity proof.
2. Use existing frozen header storage in both implementations. No new I/O, dependency or public header exposure.
3. Main agent independently inspects the bounded change; no subagents or remote actions are authorized for this assignment.

## Plan

Execute the two gate steps above as one coherent SDK follow-up.

## Implementation And Review Plan

Implementation:
- One implementation owner for SDK files only. Main owns consumer integration separately.

Reviewers:
- Main-agent independent inspection requested for this small locally understandable accessor. No external harness or additional agent will be launched.

Repository boundary block for every external-reviewer prompt:

```text
CRITICAL REPOSITORY BOUNDARY:
- Do not make changes outside this repository for any reason.
- Repository path: current repository root.
- You may inspect external references read-only when the task requires it.
- Write, edit, delete, move, reset, checkout, install, generate, cache, or format nothing outside this repository.
- The only write exception outside the repository is /tmp.
- Prefer .local/ inside this repository for scratch work, generated temporary files, cloned references, logs, and external-agent working notes.
```

Failure handling:
- Fix local failures within this bounded scope; preserve unrelated changes. Report concrete tool limitations without claiming validation.

## Execution Log

### 2026-10-06

- Confirmed clean worktree, completed SOW-0147 and absence of overlapping current work. Recorded ready gate before implementation.
- Added matching accessors reading only copied header bytes; no new I/O, allocation or graph traversal.
- Actual Go archive and Rust directory-writer close transitions prove frozen state; Go checks both mmap and ReadAt. Offline fixtures return false.
- Focused language suites, both updated published examples, wiki structure and initial SOW audit passed.
- Main-agent independent inspection found no blocker. Completed this follow-up and ran the lifecycle audit before the local commit.

## Validation

Acceptance criteria evidence:
- Go IsArchived and Rust is_archived use the existing owned captured header. Online/offline false and archived true are covered; capture before a real archive remains false afterward.
- Public docs and product-scope spec distinguish the frozen state from graph integrity and lifecycle provenance.

Tests or equivalent validation:
- Go: go test ./journal -run TestIndexedSnapshot -count=1 passed (1.884s), Go 1.27.1 darwin/arm64.
- Rust: cargo test -p systemd-journal-sdk indexed_snapshot --offline passed all 12 selected tests, Rust 1.91.
- python3 tests/docs/check_wiki_docs.py validated all 16 wiki pages.
- python3 tests/docs/verify_examples.py --only go-indexed-snapshot and --only rust-indexed-snapshot each compiled and ran the updated public example successfully.
- git diff --check and bash .agents/sow/audit.sh passed. Caches used task-local temporary storage or repository .local tooling paths.
- No new stock-systemd or platform matrix is claimed: no journal bytes, writer lifecycle, mapping or traversal behavior changed.

Real-use evidence:
- Go Writer.ArchiveTo and Rust Log.close produced the archived files used by the tests. Snapshots opened before and after those public operations return distinct expected metadata. Go Writer.CloseOffline and a Rust low-level offline header fixture cover the other valid non-archived state.

Reviewer findings:
- Main agent independently inspected Go/Rust accessors, real archive-lifecycle tests and docs/spec. Confirmed both read already-owned state with no reread/allocation or integrity claim; no concrete findings. This bounded review covers only the accessor addition, retaining SOW-0147 review for other contracts.

Same-failure scan:
- Public-method searches found only private frozen state in both snapshot implementations; existing reader metadata is not the same snapshot object.

Sensitive data gate:
- Public source and synthetic journal fixtures only.

Artifact maintenance gate:
- AGENTS.md and runtime skills need no change: no workflow or format-handling change.
- Specs and end-user docs record the additive frozen-state accessor.
- No end-user skills exist.
- SOW lifecycle: completed follow-up moved to done and committed with implementation; SOW-0147 remains unchanged.
- SOW-status.md records completion; no other active SDK work was advanced.

Specs update:
- .agents/sow/specs/product-scope.md describes owned archive-state capture and its integrity limits.

Project skills update:
- No new reusable workflow; existing native-metadata guidance applies.

End-user/operator docs update:
- docs/Indexed-Snapshots.md describes accessor semantics and integrity limits; both examples exercise the public methods.

End-user/operator skills update:
- None exist in this project.

Lessons:
- A frozen metadata accessor avoids reopening the same journal for recovery eligibility.

Follow-up mapping:
- Consumer adoption remains Netdata history-time step12; no additional SDK work deferred.

## Outcome

Implemented and validated equivalent frozen archive-state accessors in both SDKs, with independent main-agent inspection and no findings. Consumer recovery can use the captured state without reopening the file. No remote action or publication performed.

## Lessons Extracted

Recovery eligibility must use the same captured header as the indexed population.

## Followup

Consumer integration remains the approved Netdata step12. No remote SDK publication is authorized.

## Regression Log

No completed SOW claim is reopened; this is a narrow additive consumer prerequisite.
