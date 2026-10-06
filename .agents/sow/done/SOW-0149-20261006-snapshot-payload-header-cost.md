# SOW-0149 - Snapshot Payload Header Cost

## Status

Status: completed
Sub-state: validated implementation committed as 0d365e2; independent main-agent review found no blockers.

## Requirements

### Purpose

Remove unnecessary payload-traversal work exposed by approved DEM indexed-history integration, preserving snapshot safety and semantics.

### User Request

Investigate full DATA parsing in Go SnapshotEntry.VisitPayloads, measure fair SDK and DEM before/after comparisons, use minimal object metadata when supported by evidence, validate and commit locally before main independent review. No cache, pool, speculative Rust change, subagent or remote action.

### Assistant Understanding

Facts:
- DEM broad 10k-row RUM queries regressed versus the old saved scan; the supplied CPU profile attributes 84 percent cumulatively to readDataHeaderAt.
- SnapshotEntry.VisitPayloads uses s.readData, parsing the entire 64-byte DATA header. Its payload helper uses only the 16-byte object header.
- Capture, FIELD traversal and postings still need complete DATA headers.
- Rust payload traversal already uses its existing borrowed DATA view; no equivalent struct-copy optimization is established.

Inferences:
- Repeated decoding and copying six unused uint64 fields per payload may explain measurable cost. The profile alone is insufficient proof.

Unknowns:
- No unresolved implementation question. Controlled measurements establish the bounded effect below.

### Acceptance Criteria

- Payload traversal reads only needed object metadata if controlled SDK and DEM benchmarks support the optimization.
- Preserve type, minimum size, committed offset/extent bounds, decompression, cancellation and callback lifetime checks; capture/posting readers keep full metadata.
- Compression/remap/corruption and race tests pass, along with relevant reader tests.
- Record measured benefit and limits, update durable evidence, audit and commit locally before main independent review.

## Analysis

Sources checked:
- SDK AGENTS.md, completed SOW-0147/0148, docs/Indexed-Snapshots.md and product-scope snapshot contract.
- go/journal/{indexed_snapshot.go,indexed_snapshot_visit.go,indexed_snapshot_capture.go,reader_entry.go,reader_unique.go,format.go}.
- Parent-provided DEM benchmark/profile evidence; current benchmark source is inspected read-only.

Current state:
- Clean SDK branch feat/indexed-snapshot at 529675f before work. No current SDK SOW overlaps.
- The payload helper validates type/size, reads payload and decompresses using only header.object.

Risks:
- Accidentally weakening committed-bound checks or parsing less than the payload layout needs; dedicated negative traversal tests will cover this.
- CPU-profile attribution and changing consumer sources can mislead performance claims; build both variants against identical source via an SDK-only overlay and alternate process runs.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:
- Full DATA header parsing and by-value copies are unnecessary for payload access; only type, size and compression flags are needed. Existing normal reader payload paths already parse the object header alone.

Evidence reviewed:
- Local source and supplied profile; existing 100k-row indexed snapshot benchmark and DEM 10k-row query benchmark; runtime journal compatibility and orchestration skills.

Affected contracts and surfaces:
- Internal Go snapshot payload read and shared payload helper signature/callers, focused tests and performance evidence. No public behavior, on-disk format, recovery policy or Rust change.

Existing patterns to reuse:
- parseObjectHeader, reader readSlice, snapshot committed-bound helpers, current decompressor and payload callback mechanism.

Risk and blast radius:
- Keep the existing full DATA parser for every caller requiring hash, FIELD or posting metadata. Shared helper callers pass its object metadata directly.

Sensitive data handling plan:
- Synthetic benchmark journals and public code only. Scratch/binaries/caches under temporary task storage; no live journal or host identity probes.

Implementation plan:
1. Preserve baseline source/binaries; use a minimal object-header payload path and retain full DATA reads for other operations.
2. Add negative payload traversal tests, run focused/full relevant Go suites and race tests, and compare six alternating SDK and DEM benchmark pairs.
3. Record evidence, audit and locally commit coherent validated implementation; request main independent review of that commit and record its outcome separately.

Validation plan:
- Test active payload callback behavior on regular/compact layouts, supported compression, tiny windows, cancellation and malformed type/size/offset/extent.
- Run SDK 100k all-payloads and DEM 10k broad/narrow query benchmarks on separately compiled identical-source variants. Compare elapsed time and allocations; do not infer effect solely from profile samples.

Artifact impact plan:
- AGENTS.md and runtime skills: no workflow change.
- Specs and public docs: contract unchanged; update only if investigation reveals a necessary semantic change, which would require routing the decision.
- Performance evidence: documentation/indexed-snapshot-validation.md records controlled follow-up results.
- End-user/operator skills: none exist.
- SOW lifecycle/status: track this bounded performance follow-up; keep completed prior SOWs intact.

Open-source reference evidence:
- netdata/netdata @ 9a2e0cbd5bc68fb559b39964238b5de1f1004236 plus uncommitted DEM integration changes; src/go/plugin/dem/rum/history benchmark source inspected read-only. Both consumer binaries were built from the same checked source hashes. Local SDK parser and reader contracts establish the implementation behavior.

Open decisions:
- No product fork. Retain optimization only with positive controlled measurements and unchanged contracts.

## Implications And Decisions

1. Approved consumer integration authorizes this coupled internal performance correction. Do not change capture cost, recovery policy or public behavior.
2. Benchmark identical consumer source; preserve baseline SDK through an overlay. No cache/pool or speculative sibling-language optimization.
3. Commit validated implementation before requesting main independent review, as explicitly assigned. This requires a separate review/SOW closure commit after the review rather than claiming review before it happens.

## Plan

Execute the gate plan as one bounded follow-up.

## Implementation And Review Plan

Implementation:
- SDK-only owner; consumer source stays read-only. Temporary benchmark runners may be created under the task scratch directory.

Reviewers:
- Main independent review after validated local commit. No subagents or external harness.

Failure handling:
- Reject unsupported optimization claims, preserve test evidence and fix contract violations before committing. Keep SOW active until review and audit are complete.

## Execution Log

### 2026-10-06

- Confirmed clean source and no current SDK work. Read profile and payload parser owners; preserved baseline source before implementation.
- Replaced only payload traversal full DATA reads with bounded object-header parsing; narrowed shared payload helper to object metadata. Full capture/FIELD/posting readers remain unchanged.
- Six alternating pairs establish the SDK and consumer timing effects recorded in documentation/indexed-snapshot-validation.md. Consumer source manifest digest: e049e9365897213b76531c80854aa1ff4667425d4d0bfb7e428740d6fa83e99a.
- Full Go tests, vet and selected race suite pass. Malformed payload checks pass on baseline and optimized paths, establishing unchanged safety checks.
- Committed validated implementation as 0d365e2 before review as explicitly assigned. Main independent review found no blockers. Completed and moved this SOW in the required separate review-closure commit.

## Validation

Acceptance criteria evidence:
- Minimal object metadata is used only for payload access. Controlled timing improves SDK broad traversal by 23.0 percent and DEM broad queries by 18.2-19.6 percent, with unchanged allocation behavior.
- New public-operation tests preserve all offset/type/size/extent and decompression rejection checks. Existing compression/remap/callback/cancellation/concurrency tests pass.
- Main independently reviewed committed 0d365e2 and found no blockers; bounded SDK performance follow-up is complete.

Tests or equivalent validation:
- go test -overlay /private/tmp/dem-sdk-payload-before/overlay.json ./journal -run '^TestIndexedSnapshotRejectsMalformedPayloadObjects$' -count=1 passed: all 36 new cases retain baseline behavior.
- go test ./... passed, including full journal package (17.692s).
- go test -race ./journal -run 'TestIndexedSnapshot|TestReader.*Unique|Test.*Unique' -count=1 passed (5.361s).
- go vet ./..., git diff --check and bash .agents/sow/audit.sh passed.
- Six alternating process pairs at 300ms per selected benchmark: SDK exact/all-payloads and DEM sessions15m/sessions_all/errors_overview_all/error_selected_all. Before/after source overlays avoid consumer drift. SDK broad 13.910 to 10.713ms; DEM broad sessions 8.965 to 7.332ms. All broad reductions p=0.002; exact change is not significant (p=0.394).
- Benchmarks, source hash manifest and benchstat summaries live under /private/tmp/dem-payload-bench. No host cache flush, stock live-systemd or additional platform validation claimed; the operation changes no disk bytes or format semantics.

Real-use evidence:
- Actual DEM RUM query reducers processed 10,000-row synthetic journals in separately compiled current-source variants. The optimization does not erase the entire earlier regression against saved-scan queries; remaining consumer comparison stays with the parent task.

Reviewer findings:
- Main independently reviewed 0d365e2: captured offset/extent checks remain before payload access; DATA type/minimum-size and decompression checks remain in the helper; full headers stay in index/capture paths; cancellation and callback lifetime are unchanged. The malformed-object matrix covers both layouts/access modes and baseline parity. No blocker or fix was identified.

Same-failure scan:
- readData callers needing hash/posting/FIELD metadata retain full headers; only entry payload traversal needs the minimal header.

Sensitive data gate:
- Public source and synthetic fixture data only.

Artifact maintenance gate:
- AGENTS.md and runtime skills unchanged: no workflow or compatibility-policy change.
- Specs and public docs unchanged: frozen bounds, payload behavior and lifecycle contracts remain identical.
- documentation/indexed-snapshot-validation.md records source cause, controlled measurement, safety validation and claim limits.
- No end-user/operator skills exist. This SOW is completed in done and the status summary records closure; the explicit commit-before-review instruction required a separate closure commit.

Specs update:
- Existing frozen-bounds and payload-lifetime requirements remain unchanged.

Project skills update:
- Existing native-index and avoid-repeated-parsing guidance already covers this work.

End-user/operator docs update:
- Existing public contract remains unchanged; internal performance evidence records new measurements.

End-user/operator skills update:
- None exist.

Lessons:
- Profile attribution must be corroborated by controlled elapsed-time benchmarks.

Follow-up mapping:
- Consumer integration remains the parent Netdata history-time task; no separate SDK work deferred.

## Outcome

The measured internal optimization is complete with passing local validation and independent review of 0d365e2. SDK broad payload traversal is 23.0 percent faster in the controlled fixture, and current DEM broad queries improve 18.2-19.6 percent with unchanged allocation behavior. No public semantic, format or Rust change was introduced; remaining consumer comparison work stays with the parent integration task.

## Lessons Extracted

Read only the metadata required by a traversal operation while preserving its committed bounds.

## Followup

Main independent review completed without findings. Consumer integration, including the remaining comparison against its earlier saved-scan path, remains separately owned by the approved Netdata history-time task.

## Regression Log

Bounded integration performance follow-up; no previously claimed compatibility result is invalidated.
