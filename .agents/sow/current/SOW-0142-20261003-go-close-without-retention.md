# SOW-0142 - Go Close Without Retention

## Status

Status: in-progress
Sub-state: user-approved additive Go API prerequisite for DEM policy reload; local implementation/validation/review only, no push or release.

## Requirements

### Purpose

Allow a consumer to close/archive its current journal before reopening with changed retention, without deleting archives under a superseded policy.

### User Request

User explicitly approved a small journal SDK API to archive/close without applying old retention, preserving hourly DEM updates and idle expiry. User requested an isolated worktree and identified the existing SDK clone.

### Assistant Understanding

Facts: Log.Close applies configured retention after archive. No public reconfiguration or no-retention close API exists. DEM config reload can relax age before its next sweep; a real-SDK failing regression retained one rather than two sessions. Inference: an additive close variant fixes the ownership boundary without undocumented pointer aliasing. Publication is not authorized; Netdata adoption requires a later published revision.

### Acceptance Criteria

- Add Log.CloseWithoutRetention using exactly the existing archive/durability/resource release path, skipping retention only.
- Preserve normal Close behavior, idempotence, empty/lazy handling, strict/systemd naming and archive failure semantics.
- No disk-format, append, identity, dependency or Rust behavior changes.
- Actual SDK tests demonstrate preserved expired archives and later enforcement under a changed policy, docs/spec match, required local checks and read-only independent review pass.

## Analysis

Source: netdata/systemd-journal-sdk @ ac43b6c99807b6625bb38cf8619c7473bc323a54, go/journal/log.go and log_retention.go; Go source matches published go/v0.8.1. Pending SOWs cover other integrations/performance/filter work; none owns close-retention transitions. Product scope currently documents normal Close/retention. User approval expands the DEM no-SDK-change boundary for this narrow prerequisite only.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model: Close runs stale retention before consumers can construct a new Log with an increased allowance; real DEM regression confirms archive loss.

Evidence reviewed: AGENTS.md, project-agent-orchestration and project-journal-compatibility skills, docs authoring guidance, current/pending SOW status, product-scope Go writer contract, existing close/retention/rotation tests and pinned consumer regression.

Affected contracts and surfaces: additive Go Log method, shared close helper, Go tests, consumer README/API/wiki prose, product-scope and SOW status. Rust and existing callers keep their current behavior.

Existing patterns to reuse: one existing Log close body, archiveActive and current default sync behavior, synthetic test identities and temporary fixture directories. No new policy storage, callbacks or format machinery.

Risk and blast radius: cold explicit close API; normal Close must continue retention. Callers selecting the variant own subsequent enforcement. Archive errors must not claim successful closure. No append/query hot-path code changes or performance claim.

Sensitive data handling plan:

- Synthetic identities/rows; only relative source pointers in durable evidence; no live journals, host identity, credentials or personal names.

Implementation plan:
1. Extract close helper with retention boolean; expose CloseWithoutRetention and retain Close defaults.
2. Add actual journal coverage for aged archives/changed-policy reopen, both naming modes, idempotence, empty/lazy close and archive failure; align Go docs/wiki/spec.
3. Run Go race/vet/docs checks and audit, prepare coherent local commit before independent review under current user instructions, fix verified blockers and finalize SOW with validated work. Publication remains outside authorization.

Validation plan: targeted failing-before/passing-after old-policy regression, full Go tests/race and vet, wiki validator/examples appropriate to prose-only change, reference search and whitespace/audit. No live journal or host service. Existing archive/format tests exercise unchanged bytes; stock Linux tools are unavailable on this macOS host, so no new stock-systemd certification is claimed.

Artifact impact plan:
- AGENTS.md: current collaboration guidance already owns workflow; no policy change.
- Runtime project skills: API contract belongs in docs/spec/tests; no new workflow.
- Specs: add no-retention close variant to current Go writer slice.
- End-user/operator docs: Go README/API and wiki writer descriptions clarify explicit caller retention ownership.
- End-user/operator skills: repository declares none; no external copy affected.
- SOW lifecycle: current prerequisite, complete after validation/review; release readiness stays separate.
- SOW-status.md: update activation/completion summary.

Open-source reference evidence: netdata/netdata consumer base9f789a14a4 plus current task changes; only read-only consumer inspection outside SDK. SDK reference netdata/systemd-journal-sdk @ ac43b6c99807b6625bb38cf8619c7473bc323a54.

Open decisions: scope/API purpose explicitly approved. No SDK push/tag/release or Netdata dependency publication authorized.

## Implications And Decisions

1. User selected SDK API over requiring restart for retention changes. CloseWithoutRetention makes retention ownership explicit and avoids private pointer identity assumptions.
2. Current user instructions select Astra read-only subagents for independent review; use the available collaboration tool for this bounded SDK review. Legacy external harness is not invoked. User instructions take precedence over skill workflow preferences.

## Plan

Implement the additive method/tests/docs as one coherent prerequisite; do not bundle release or unrelated pending work.

## Implementation And Review Plan

Main implements SDK files. Independent reviewer must be read-only, no child agents, no live host journal, no dependency commands using external caches. Include the canonical repository boundary block. Follow current user review timing/commit instructions; Current user instructions require committing validated implementation before independent review. This intentionally splits the local validation checkpoint from final reviewed lifecycle completion; the higher-priority requested commit timing supersedes this repository’s usual one-commit close preference. No history rewrite or push is authorized.

## Execution Log

### 2026-10-03

- Existing source clone clean on master at ac43b6c; created isolated feat/dem-retention-close worktree. User approval recorded before source edits. Temporary reference clone was not used for implementation.

## Validation

API implemented; final-file-state full Go race suite and vet PASS. Wiki validator (15 pages), all 17 Go wiki examples, diff whitespace and local SOW audit PASS. Initial full race suite caught an incorrect new empty-chain test expectation; test/docs now preserve existing Close behavior. Actual synthetic fixtures prove expired archives survive no-retention close, relaxed reopen preserves them and tighter reopen removes them; strict/chain archive state, lazy/eager/idempotence and archive-failure retry pass. Independent review remains required; no readiness claim before review.
Sensitive data gate:

- Synthetic source scope, no sensitive material recorded.
Same-failure scan: no documented setter or close-without-retention exists; old Close must retain its normal policy behavior.
Artifact maintenance gate: docs/spec updates planned above; no new workflow or external skills. Completion audit/status and review remain required.
Follow-up mapping: SDK publication/version pin is consumer delivery sequencing and is not authorized by this SOW; DEM local SOW Q17 tracks it. No SDK release is bundled.

## Outcome

Implementation in progress; no external publication.

## Lessons Extracted

Consumer reconfiguration must use explicit SDK contracts rather than shallow policy pointer aliasing.

## Followup

Release/pin sequencing is tracked in the consuming DEM initiative Q17; no new SDK release task is inferred.

## Regression Log

This additive API addresses a consumer transition gap; existing Close behavior remains correct for its configured policy.
