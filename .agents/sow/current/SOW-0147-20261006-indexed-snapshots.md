# SOW-0147 - Indexed Snapshots For Domain History

## Status

Status: in-progress
Sub-state: reopened for the confirmed live-reader arena-growth regression from human comment 4193955216. The user approved separating live read bounds from stable-file integrity validation and authorized native subagent verification. No push or publication.

## Requirements

### Purpose

Expose efficient exact native-index traversal over a committed snapshot, so incident-history queries do not scan or allocate offsets for every retained entry.

### User Request

After the Go experiment, user agreed to bounded SDK snapshots/index traversal, consumer-owned time buckets and SDK prerequisite followed by DEM adoption. Clean end state matters more than churn. DEM is unused WIP without compatibility promises. No push, release or publication was authorized.

### Assistant Understanding

Facts:
- The completed opt-in Go experiment demonstrates narrow-query value; ordinary Reader opens eagerly materialize all offsets.
- Current writer publication is multi-step. A modeled interruption followed by OpenWithOptions and append yields global/indexed membership disagreement.
- SDK product languages are Go and Rust. Existing readers/facades and disk format remain valid; the new surface is additive.

Inferences:
- A small callback API and capture declarations are sufficient; no time-language/query-planner SDK is necessary.
- Writer failure handling and a truthful restart integrity boundary are prerequisites for exact indexed adoption.

Unknowns:
- No unresolved product decision. Implementation/validation may reveal defects that must be resolved within the approved contract or explicitly raised if they require a material new choice.

### Acceptance Criteria

- Go and Rust expose equivalent bounded index snapshots with byte-safe exact matching, FIELD predicates and selected payload callbacks.
- Capture and post-capture append traversal share one proven committed population, preserve historical/layout/compression support, and propagate errors/cancellation.
- Writer/interruption integrity preconditions and recovery policy are implemented and tested before DEM adoption; no false completeness promise.
- Relevant conformance, cross-language/live tests, benchmarks, docs/examples, independent review and project audit pass before readiness.

## Analysis

Concrete design and reviewed findings: documentation/dem-history-index-design.md.

Sources checked:
- go/journal/{reader.go,reader_unique.go,reader_entry.go,explorer.go,writer.go,writer_init.go,writer_arrays.go,writer_objects.go,log.go,verify.go,verify_graph.go}.
- rust/src/crates/journal-core/src/file/{file.rs,file_payload.rs,file_iterators.rs,writer.rs}; rust/src/journal/src/lib.rs.
- .agents/sow/specs/product-scope.md; parked SOW-0123 and SOW-0125 remain separate and unexecuted.
- netdata/netdata @ 6af30bddb83a5fd47c2f64e16d40a7f59ea944ad: src/go/plugin/dem/{journal,rum/history,synthetic/history}, src/go/cmd/demplugin/main.go.
- Experiment branch spike/dem-history-time @ fa5a832c74dae456225e536eb8f95b19d4c5d54d: experiments/go-history-time/README.md and tagged Go prototype.

Current state:
- Separate worktree/branch feat/indexed-snapshot is rebased onto local master 5cb48c8. Experiment branch remains intact.
- Go/Rust source, public docs, verified examples and current-contract spec are implemented. Validation evidence is below; independent implementation review is complete. Scratch remains under /tmp or .local/.

Risks:
- Callback mapping lifetime, mutable FIELD/posting structures, older header metadata and interrupted publication interact. Frozen file size alone is insufficient.
- Verification can block startup; existing verifier has repeated-posting quadratic work and a last-entry reverse-link exception.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:
- Existing eager entry enumeration hides indexed query speed. Independent index/header publication can leave a graph that a reopen accepts but indexed and global traversal interpret differently.

Evidence reviewed:
- SDK AGENTS.md; project-journal-compatibility and project-agent-orchestration skills; project-docs-authoring for eventual public examples; source/spec/experiment owners above; reviewed joint design and deterministic interruption reproduction.

Affected contracts and surfaces:
- New idiomatic Go/Rust snapshot API; approved writer uncertain-error and explicit pre-reuse validation behavior; native graph errors, callback lifetime and cancellation; tests, docs and current product spec.
- DEM owns schema/timestamps/buckets and strict eligibility. No SDK producer-specific names, new dependency, compiler minimum change, host discovery, implicit lock, database or cache.

Existing patterns to reuse:
- Header-only Go opening, native FIELD/DATA/hash/posting traversal, Rust guarded payload callbacks, current layouts/hash/compression and file ownership; existing verifier graph representations where appropriate without retaining quadratic work.

Risk and blast radius:
- Reader API is additive. Writer failure/reopen policy reaches existing SDK consumers and follows the explicit operational choice recorded below. No automatic recovery/deletion or compatibility migration is authorized.

Sensitive data handling plan:
- Public code and synthetic identities/fixtures only; no credentials or customer data. External source references use repository identity/commit and relative paths. Scratch outputs remain under /tmp.

Implementation plan:
1. Approved: selective startup verification of uncertain active files, strict active/archive naming and synchronized clean-archive provenance; preserve and report damage, no automatic repair/deletion. Implement the SDK primitives and safe writer lifecycle needed by this policy.
2. Implement the prerequisite writer/verification consistency contract and callback snapshots in Go/Rust using existing format owners; include tests/docs/specs and benchmark evidence. Sequence coherent local commits, no remote operations.
3. Independently review actual implementation and resolve verified defects; audit and close only when the approved target is complete. DEM adoption is a separate Netdata step12 deliverable after SDK availability.

Validation plan:
- Reproduce publication windows and test reopen/append consistency, poison/no-clean-state behavior and selected restart policy.
- Test all capture counts/bounds, empty/historical files, keyed/unkeyed/compact/regular/compressed DATA, binary/multivalue semantics, callback failures and cancellation, concurrent append growth and tiny mmap windows.
- Independently expected IDs for exact/predicate/global paths; compare Go/Rust selected results and costs. Run existing SDK compatibility tests appropriate to changed code, relevant live/file-backed systemd matrices when available, public example validation and full local SOW audit. Record unavailable tooling instead of claiming parity.

Artifact impact plan:
- AGENTS.md: no policy change.
- Runtime project skills: capture reusable snapshot/integrity rules if implementation establishes them.
- Specs: update product-scope with shipped contracts only after implementation.
- End-user/operator docs: document new API choice, exclusion/lifetimes/errors/recovery and compiled examples.
- End-user/operator skills: none exist in this repository.
- SOW lifecycle: in-progress/current after explicit recovery-policy approval. Parked operator/clock work is not resumed.
- SOW-status.md: records active implementation and will record completion after review.

Open-source reference evidence:
- netdata/systemd-journal-sdk @ 8ad648a7b4d36bf2d75ee6ee98b42446ff64a276 and experiment fa5a832c74dae456225e536eb8f95b19d4c5d54d, owners above.
- netdata/netdata @ 6af30bddb83a5fd47c2f64e16d40a7f59ea944ad, source owners above.

Open decisions:
- Resolved: user replied "approve" to option 1 and prerequisite SDK fixes. Use selective verification with proven clean archives, distinct active/archive names and failure preservation. Do not re-request this approval; no automatic recovery or remote publication is authorized.

## Implications And Decisions

1. User approved bounded Go experiment; its result is evidence, not release certification.
2. User then approved the measured SDK direction and SDK-first/DEM-second delivery. This authorizes concrete design and ordinary preparation; remote publication remains unapproved.
3. User explicitly approved option 1 and prerequisite SDK fixes after the reproduced interruption finding: verify uncertain active files before startup mutation, trust only proven clean archives, preserve damaged files and fail closed. The existing startup failure implication and archive sync/strict naming requirements were disclosed. Naming and callback API shape remain routine details.

## Plan

Follow the approved ordered gate plan. Complete SDK prerequisite before Netdata adoption; no remote actions.

## Implementation And Review Plan

Implementation:
- Main owns Go snapshot/writer integration, shared docs/SOW and final validation. Bounded delegated Go verifier and Rust implementations may proceed under disjoint ownership; no overlapping writers or delegated commits.

Reviewers:
- User authorized native read-only subagents in this initiative. Two Astra reviewers covered SDK feasibility and DEM population/completeness; focused SDK challenge found the publication issue. No external reviewer harness or GitHub publication was used. Follow standing user review directions for eventual implementation review.

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
- Record concrete reproductions, preserve user/source changes, correct verified defects within approved scope, and route material public behavior forks to the user. Do not claim readiness from reviewer agreement or an aborted benchmark.

## Execution Log

### 2026-10-06

- Created isolated SDK implementation worktree from master. Drafted/reviewed joint API/DEM contract.
- Modeled publication interruptions with an overlay-only Go test against experiment checkout: scan five versus indexed six before final count; six versus seven after reopen/append. Verifier rejects graph. Test passes by asserting the disagreement; it does not establish a fixed implementation.
- Stopped full-verifier retained-fixture cost probe after 152 seconds without first-file completion. Only verified task-owned PID was signaled; no unrelated processes affected.
- Initial design audit passed before the explicit option 1 approval. Implementation then proceeded in disjoint Go snapshot/writer/verifier and Rust scopes; source checks below precede the coherent local review commit.

## Validation

Acceptance criteria evidence:
- APIs and approved lifecycle implemented in Go/Rust. Independent implementation review and focused correction checks are complete; no release or DEM integration claim.

Tests or equivalent validation:
- Original modeled interruption evidence remains valid as root-cause evidence. New writer failure and strict-verifier regressions reproduce failing-before cases, including missing final reverse links, zero-reference DATA and failed first-append cleanup.
- TestDesignVerifyRetainedFixture interrupted after 152 seconds; no pass or full runtime estimate claimed. Log /private/tmp/dem-sdk-verification-cost.log.

Real-use evidence:
- Completed spike uses actual SDK writes/reads at 949 MB; production DEM integration remains future step12 validation, not established here.

Reviewer findings:
- Two fresh Astra reviewers independently assessed Go snapshot/index/lifecycle and Rust/parity after commit 501fd32. Accepted and reproduced: strict Go FIELD chain order mismatch, historical header boot/monotonic mismatch in both snapshots, and Rust hash tables outside the committed object graph. Main also corrected a matching Go FIELD with missing DATA head being reported as absence. New regression tests failed before source fixes and pass after cd0dfaf.
- Focused independent confirmation on cd0dfaf found no remaining blockers in either scope. Snapshot clipping proof holds under valid graph, capture exclusion and append-only guarantees. Existing broader coverage remains valid. No optional style/coverage suggestion extended the cycle.
- One focused-review tool attempt failed an automated content classification; the same ordinary file-format review completed after a concise restatement, without changing scope or performing a restricted action.

Same-failure scan:
- Traced Go and Rust append publication, Go append-open, Log error propagation and archive ordering; affected paths are covered by failure preservation tests. Rechecked historical header feature flags and walked-table provenance in both implementations.

Sensitive data gate:
- Public source and synthetic journal data only. No private telemetry or credentials in tracked artifacts.

Artifact maintenance gate:
- AGENTS.md and runtime skills unchanged. Public snapshot/recovery documentation, executable Go/Rust examples, product-scope spec and validation report carry the delivered contract.
- No output/reference skills exist. SOW is completed/done after review and validation; status summary records it.

Specs update:
- product-scope records additive snapshot and strict recovery invariants.

Project skills update:
- Existing journal-compatibility skill already owns native-index and source-purity requirements; new API contracts live in public docs/spec/source.

End-user/operator docs update:
- docs/Indexed-Snapshots.md, reader/writer/API navigation and Go/Rust READMEs; all 33 marked examples compile/run. Context import support has a harness regression test.

End-user/operator skills update:
- None exist in this repository.

Lessons:
- Exclusion freezes writer activity, not consistency after prior failure; index-count coverage depends on valid graph provenance.

Follow-up mapping:
- SDK prerequisite remains this SOW. Consumer adoption remains Netdata SOW-20261005-dem-framework-12-history-time; Cloud remains final step04. Existing operator/clock pending SOWs are separate, unchanged.

Reference reconciliation:
- `rg -n 'RotationRetriesAfterArchiveCleanupFailure|retry.*archive|retry.*rotation|restoreErr' go rust docs` returned no remaining references to the retired archive retry/rollback behavior.
- `rg -n 'archive_existing_active_file|ErrWriterFailed|is_poisoned|archiveTo\(' go/journal rust/src/crates/journal-log-writer/src/log rust/src/crates/journal-core/src/file/writer.rs` returned 31 owner/test references. All are intentionally retained lifecycle entry points, error guards or regression expectations; no compatibility retry path remains.
- SDK reader/facade/Explorer surfaces remain intentionally retained with their original contracts. New APIs are additive. DEM schema/naming/query migration and Cloud remain the explicitly staged consumer work, not missing SDK cleanup.

Additional implementation evidence:
- After review fixes, Go snapshot/all-verifier race checks pass (9.38 s), final Go vet passes, all 10 Rust snapshot tests independently pass, and final Rust public suite passes 142 tests with the two baseline platform failures filtered.
- Both public Go/Rust APIs strictly verify and traverse exact/FIELD/global indexes in all seven Go/Rust documentation writer fixtures. Log .local/indexed-snapshot-cross-language.log. This adds cross-language snapshot evidence to the stock closed-file checks.
- Same mixed250kfixture Go metadata-only probe: capture116.5us, exact2496postings209.7us, FIELD2.717ms, lazyglobal11.464ms. Results agree with Rust counts; timings are warm one-shot measurements, not a ranking.
- Full Go module tests and vet pass. Full journal race suite passed (25.7 s), then focused snapshot/strict verifier race tests passed after bounded capture refinements (5.2 s). All supported DATA compression/layout/access modes, historical unkeyed whole-file fixture, binary names, callback errors/cancellation and concurrent first-array growth are covered.
- Rust core 81, directory writer 9, public SDK 142 pass; two preexisting macOS wheel/root expectations are excluded after baseline reproduction. Shared conformance adapter 15/15 passes. Initial core/writer/conformance logs are /tmp/dem-sdk-rust-final-tests.log and /tmp/dem-sdk-rust-conformance.json; the final public-suite log is .local/indexed-snapshot-rust-final.log (142 pass, 2 baseline failures filtered).
- Wiki structure validates; 78 harness tests and all 33 marked Go/Rust examples pass. Initial offline cache miss was resolved by copying existing task-local registry cache; no source dependency/version changed.
- Stock systemd 257.13 in a disposable container verifies all seven generated Go/Rust documentation journals. Read-only synthetic mounts only; no host journal or service accessed. This is closed-file verification, not a new Linux live-reader parity claim.
- Go 100k-row collision-heavy snapshot benchmark (six runs): capture median28.1us/12,344B, exact100-row payload traversal includingopen205us/12,456B, FIELD392.5us, all 100kpayloads13.31ms/12,608B; eageropen2.195ms/5,155,297B. No cold-cache claim.
- Mixed250kfixture strict verification: Go1.04s/681.6MB cumulative allocations/~494MB process-treepeakRSS; Rust1.64s/~537MB childpeakRSS after cursor optimization. These are offline graph-sized costs, not query-open costs. Rust capture97us, exact2496postings265us, FIELD2.645ms, lazyglobal10.949ms; metadata-only visits differ from Go100kpayload workload.
- Six alternating30krow Go writer trials: baseline143.3ms, changed146.2ms median (~2.0% difference; overlapping ranges). Same mixed 32-field structured append, compact layout, defaultlivepublication. No broad speedup/regression claim from this noisy small difference.

## Outcome

SDK prerequisite is ready for review/release. All in-scope code, tests and documentation are committed locally, and verified review blockers are resolved. No source change or final dependency pin is claimed for DEM yet.

## Lessons Extracted

Cheap capture consistency checks do not certify arbitrary interrupted graphs. Avoid using the current full verifier as a startup mechanism without measuring and correcting its repeated-posting work.

## Followup

Publish the reviewed SDK version through the user-owned release workflow, then execute Netdata step12 adoption. This is the approved staged boundary; no SDK implementation is deferred and no remote action is authorized.

## Regression Log

The initial experiment excluded production recovery certification; the production design added it. The completed implementation was subsequently reopened for the supplied-comment corrections documented below.

Approval recorded 2026-10-06: user replied "approve" to option 1 and the prerequisite SDK fixes. Gate ready before first source change.


## Regression - 2026-10-06

Authorization: user requested checking supplied .local/gh-review-bot-comments.md for feat/indexed-snapshot, ignoring GitHub state and fixing only real production defects, misleading claims and flaky tests. This fixes the approved contract; no new architecture or behavior fork is selected. The saved file contains 23 comments including repeated posting/empty-tail findings. Each receives an explicit accept/reject disposition in the final evidence ledger.

Pre-Implementation Gate for corrections: ready.

Problem / root-cause model:
- Go FIELD capture currently loads the variable-length payload before checking captured object bounds. Local index walks also conflate absent values or future publication with malformed zero/missing links. Verify exact live-publication invariants before tightening checks.
- Failure flags are set at broad writer entry points; investigate whether compact-capacity errors can occur before mutation and unnecessarily prevent a valid smaller retry. The Go 32-bit failure fixture may overflow integer size arithmetic.
- Strict Go/Rust verification differs on empty indexes, empty tail boot metadata and repeated ENTRY DATA references. Establish production-reachable format cases before selecting corrections. A theoretical SHA-256 collision alone does not establish a production defect.
- Documentation overstates fixed scratch/expired-view detection and retains temporary-only evidence; previous readiness lacks a recorded fresh Linux live matrix despite the active-SOW gate.

Evidence reviewed: supplied comments only; current sources at ac38ddb; completed SOW0147/0148/0149; product-scope indexed snapshot/strict verification contract; project-journal-compatibility native indexes, live/verify matrix and synthetic-only rules; project-docs-authoring compiled example contract; prior benchmark/profile and conformance evidence. No live GitHub source is in scope.

Affected surfaces: Go indexed snapshots, strict verifier and writer failure tests; Rust strict verifier/writer/snapshot tests and example; associated README/wiki/spec/validation claims. Existing APIs, file format, compiler minimums and reader/facade ownership remain unchanged.

Existing patterns: reuse native header checks, committed extent helpers, snapshot publication bounds, strict posting cursors, mutation tracking and current deterministic synthetic fixtures. Avoid per-query graph validation or new persistent state.

Risk and blast radius: rejecting valid concurrent append snapshots, silently returning partial indexed results, poisoning reusable writers, breaking historical empty-file compatibility and weakening failure preservation. All fixes must preserve valid append clipping and corruption errors. No host journal/service actions; no automatic recovery or data deletion.

Sensitive data handling: synthetic files and public source only. Saved comments and raw logs stay ignored under .local. Tracked evidence uses repo-relative paths and sanitized aggregates.

Implementation plan:
1. Verify every supplied finding and same-cause related paths. Main owns docs/tracking/matrix validation; disjoint Go snapshot, Go writer/verifier and Rust source/test assignments investigate concrete cases. No overlapping writers or delegated commits.
2. Reproduce accepted defects before changing source when feasible. Correct only established bugs/misleading contract claims/flaky synchronization; record rejected speculative or optional items with evidence.
3. Run relevant Go/Rust suites, race/vet, snapshot/corruption regressions, portability and public examples. Run file-backed Linux live/verification matrices in a disposable task-owned environment where practical; if unavailable, qualify readiness rather than claim completion. Benchmark/profile changed hot paths with comparable workloads; document limits of stock/sibling comparison.
4. Commit coherent validated implementation locally before independent cross-scope review under the standing native-subagent authorization; resolve verified findings. Audit, disposition all 23 comments, and close the reopened SOW only when acceptance criteria are met. No remote mutation.

Validation plan: failing-before/passing-after regressions; exact result multisets and valid concurrent growth; Go 32-bit fixture validation; empty historical formats and Go/Rust agreement; scope-focused full suites; production-path benchmarks; documentation/example checks; Linux live stock journalctl/libsystemd plus Go/Rust readers, all feature modes and final verify including FSS, with counts/durations/version recorded.

Artifact impact: correct wiki/READMEs and current spec for precise lifetime/memory/failure semantics; retain durable reproduction pointers and current performance/live evidence in documentation/indexed-snapshot-validation.md. SOW-status tracks reopened0147; no new SOW for a regression. No runtime skill/policy change is required unless a repeatable gap is established.

Open decisions: none for restoring the approved contract. Scope is fixed by user instruction. A genuinely new public-contract fork must be raised with evidence before implementation.

Correction dispositions: 22 accepted supplied comments (including duplicate reports), one rejected theoretical SHA-256 collision report without a production trigger. The tracked report documentation/indexed-snapshot-review-dispositions.md enumerates all 23 original findings and the second round; committed regressions and the validation report carry the evidence. Raw review bodies remain ignored.

User steering: fixes must eliminate the cause/class, rather than accumulate symptom patches. The correction uses three shared invariants: (1) validate captured extents before variable reads and refresh only ambiguous cached zero posting scalars after an observed live count; (2) poison only after possible storage or publishable in-memory mutation, including sealing/hash-depth bookkeeping; (3) verify graph populations and unique DATA×ENTRY membership consistently across layouts/languages. Valid concurrent append controls cover each repaired pointer/slot path.

The required Linux arm64 build exposed an existing ABI assumption in Rust's two user/group-name lookup buffers (i8 versus libc::c_char). Both buffers now use the platform ABI type; searching rust/src found no further identical assumptions. This bounded portability correction is necessary to validate actual SDK binaries, with no public API or identity-discovery policy change. Baseline ac38ddb has the same two offending buffers; the Linux compiler reports E0308 before the correction.

Correction validation checkpoint: full Go module tests and vet pass; full journal race 29.117 s; Linux arm64 unprivileged Rust public/core/directory suites 150/83/9 pass. All 33 wiki examples, 16-page structure and 78 harness tests pass. Stock 257.13 live matrix 18/18 and verifier matrix 63/63 pass, with 100 entries/writer, 10 ms pacing, two polling readers each stock/Go/Rust and one libsystemd reader. Sealed first-run environment lacked libgcrypt20; corrected task image and complete rerun pass. No v260.1 or long-duration stress claim. Linux/386 production and Windows/amd64 cross-builds pass; 386 test binary requires an overlay for pre-existing unrelated test timestamp overflow, not a runtime pass. Before/after snapshots and isolated 12-pair writers show no significant regression; current hot-path profile commands and attribution are in the validation report. Independent correction review of dbfc60a is complete: separate read-only reviewers covered Rust, Go writer/verifier and Go snapshots/docs, without reviewing their own implementations. No code blocker remained; one stale baseline-tense sentence was corrected directly. The Go reviewer reran focused regressions (1.178 s).

Current sibling writer evidence: six alternating-order trials share 30,000 mixed 32-field rows, compact/uncompressed/unsealed format, equal bucket counts, every-entry publication and equivalent append-only timers. Go 172.6 ms, Rust windowed 227.8 ms, Rust whole-file 208.0 ms medians; trusted-unique bypass disabled. Documentation records exact workload/profile commands and why stock/snapshot comparison is not equivalent. No general throughput or language ranking claim.

Final review disposition: root causes are addressed at shared validation/mutation/graph boundaries, not by zero-link exemptions, fixture skips, blanket failure poisoning or speculative hash-collision buckets. The extra libc::c_char correction restores the required Linux arm64 build. All implementation is in local commit dbfc60a; no remote state was queried or changed. Final documentation closure is complete. DEM journal, RUM history and synthetic history tests pass with the local SDK override; the sandbox-only RUM identity mismatch was reproduced on ac38ddb and disappears when the explicit macOS identity-helper read is permitted. Netdata source remains unchanged.

Completion assessment: approved clean end state and scope rechecked; shared-boundary fixes, coupled docs/spec/tests, direct validation and three independent cross-scope reviews are complete. No verified blocker remains. Existing external publication/consumer release staging remains unchanged; no new in-scope work is deferred. Theoretical SHA-256 collision hardening is rejected, not deferred. SOW audit passes; completed file returns to done.

Correction reference reconciliation: `rg -n 'arena\.writeAt' go/journal` returns no references to the removed duplicate arena write wrappers. `rg -n 'appendMutated|required_posting|refreshPostingOffset|read_fresh_bytes_at' go/journal rust/src` returns 32 references, all retained mutation tracking, required-offset helpers/callers or regression checks. The canonical primitives replace duplicate wrappers; no compatibility path or deferred coupled reference remains.

## Second review corrections - 2026-10-06

Authorization: user requested rebase onto local master, reassess the two current saved bot comments, fetch four linked human comments using curl/GitHub API, and fix only real production issues, misleading statements or flaky tests. Prior direction to eliminate shared causes remains in force. Request fixes the goal; no new user-owned architecture decision is selected.

Pre-Implementation Gate for second corrections: ready.

Target: preserve all approved snapshot/lifecycle behavior and local master's publication receipt; reject unsafe/inconsistent declared file extents and empty-file metadata while permitting lawful inherited sequence counters; make review dispositions inspectable in tracked sanitized evidence. Existing compatibility contracts and historical header semantics must be checked explicitly. No automatic repair, data deletion, publication or GitHub replies/resolutions.

Evidence: current branch7657e62; local master5cb48c8 adds publication receipt/status docs only; current saved bot file; REST comments4192562428/2436/2442/2451 from netdata/systemd-journal-sdk PR5, originally reviewing280056d. Comments are leads until reproduced. Relevant sources: strict header/object graph validators, snapshot capture, append-open sizing and mapped arenas, current empty/historical tests, published SDK spec. Skills: project-agent-orchestration, project-journal-compatibility and selected-comment PR review workflow.

Root-cause model: physical-file bounds are not a substitute for the format's declared arena bounds. Empty-file validation must distinguish current-file ENTRY metadata from an inherited chain sequence counter and historical boot semantics. Reuse one invariant owner per language across affected strict/snapshot paths where their contracts match; avoid one-off header-field exceptions or changing compatibility verification merely to align strict checks.

Plan: (1) record this gate, rebase onto local master, reconcile status docs without dropping either receipt; (2) reproduce findings and inspect same-cause paths, then implement bounded corrections with regular/compact and historical positive/negative cases; (3) add tracked per-comment disposition, validate affected Go/Rust paths and stock interoperability/byte preservation, commit coherent fixes before independent review; (4) complete needed review, audit and close SOW. No unrelated work or dependency additions.

Validation: failing-before/passing-after tests; exact unchanged bytes on rejected guarded reuse; inherited sequence positive controls and frozen empty snapshots after later appends; Go/Rust strict parity and stock checks where applicable; full affected suites, race/vet as relevant. Preserve earlier live/performance evidence where no assumption changes; rerun relevant matrices for affected extent/reopen paths. Rebase source tree must equal the pre-rebase source tree because master changes docs only.

Risk: false rejection of valid rotated or historical files, accidental shrink on reopen, and false recovery certification. Core runtime remains pure file-format code. All fixtures use synthetic identities and task-local files; no host journal/service action. Raw API responses remain ignored in .local; tracked summary excludes reviewer identities/personal data. Native independent review remains authorized. Open decisions: none unless investigation reveals a materially different public contract.

Rebase completed onto local master5cb48c8. Conflict was solely .agents/sow/SOW-status.md: retained both master's published Rust receipt and branch records. `git diff c534c46 HEAD -- go rust` was empty immediately after rebase (HEAD0ae1ac4), proving source preservation. BotB01 accepted by tracked per-comment report; BotB02 overlaps humanH03/H04. All four fetched human leads remain applicable at least partly: H04 Rust was fixed earlier, Go still omitted the check. Pinned systemd source confirms inherited tail sequence and declared arena/current-file metadata distinction. Go reproduction fails both layouts for unsafe arena reuse, rotated-empty snapshots and empty metadata certification; Rust is being corrected under the same class-level contract.

Second-review implementation checkpoint: Go/Rust now share header-level arena and empty-population invariants across strict verification, snapshots and append reuse. Original declared extents must validate before any mutation; empty rotated tail sequence42 remains legal, while current-file head/time/offset/active-boot claims do not. The boot check requires both field presence (header>=272) and its compatibility flag. Strict graph walks enforce raw object extents; physical alignment padding remains separately bounded. Rust generic readers preserve the historical damaged-tail tolerance, and capture the header before measuring the growing file to avoid stale-length/live-header comparisons. Go investigation also reproduced open/close truncation of legal physical padding; the same allocation-class correction includes preserving that space through reuse.

Second-review evidence: same-byte Go/Rust matrix26/26 per language (8 positive,18 negative); stock257.13 accepts all8positive and rejects14negative, with the four intentional stricter recovery cases explained in the tracked disposition report. Final Linux Rust public161/core83/directory9 pass. Linux stock/libsystemd/Go/Rust live18/18 and verify63/63 pass; entries100, cadence10ms, two polling readers per language plus one libsystemdreader, final ordered reads and keyed verify. Wiki structure16pages and compiled/run examples33/33 pass. Go full module and focused race/vet pass before the final padding correction; final rerun follows that correction. Before/after capture benchmark medians: Go31.19/31.28us with unchanged allocation counts; Rust46.76/47.22us; six alternating pairs, overlapping ranges, different workloads not a sibling ranking. Detailed interpretation and commands are in documentation/indexed-snapshot-validation.md. No remote state mutation.

Reference search: `rg -n 'appendArenaFileSize|validateDeclaredArena|validateArenaObject|validateEmptyEntryMetadata|validated_arena_end|validate_committed_arena|validate_empty_entry_metadata' go/journal rust/src` finds no removed appendArenaFileSize helper and lists the shared helper definitions and required caller sites. All retained matches are active validators/callers; none represent a deferred compatibility path. Current scope still excludes automatic repair/deletion and a hidden full graph pass on append-open. No new public API or dependency fork.

Artifact assessment: product-scope spec and Indexed-Snapshots wiki document declared arena, lawful inherited sequence and bounded append-open semantics. Tracked disposition report replaces the ignored-only review ledger. Validation report distinguishes current versus prior runs, stock compatibility versus stronger recovery and historical reader tolerance. AGENTS/runtime skills need no policy update: their current bounds/native-index/live-validation rules already apply. There are no end-user output skills. Raw API bodies and diagnostic logs stay ignored; sanitized aggregate evidence contains no reviewer identities or credentials. SOW audit and diff checks pass at this checkpoint. Independent review follows a coherent validated commit; no prior-round review is claimed as coverage of these changes.

Final local-validation checkpoint: Go module tests pass (journal22.896s), focused race12.828s and modulevet pass after the allocation correction. Production linux/386 and windows/amd64 cross-builds pass. The new fail-before/pass-after padding regression confirms open/close and later append preserve physical preallocation and strict validity. Original declared bounds are validated before adopting physical allocation into the writable header, so this does not repair or accept the corrupt arena fixtures. Capture benchmark evidence remains applicable because only append-open changed afterward. Shared fixtures/live/verify evidence remains valid for the header/reader changes; final reopen tests cover the additional Go allocation step. Sensitive-data audit and diff check pass. Implementation is ready for the required independent read-only review, not yet declared complete.

Independent review checkpoint for4fcb79d: reviewer reproduced a shipping blocker in the same allocation class on Rust. A valid compact three-entry journal with its declared arena ending at the raw final ENTRY passes strict verification; open_for_append followed by sync shrinks physical8MiB to3745356 and removes four required alignment bytes, after which strict verification fails. Concrete chain: JournalFile::sync passes declared end into WindowManager::sync, which set_len truncates. Earlier positive padding tests covered Rust read/verify and Go full reuse, but missed Rust sync on the same input. This is coupled repair under the existing safe-reuse target, not a deferred new feature.

Bounded repair plan: preserve the existing Rust allocation ownership in WindowManager, distinguishing retained initial/published file capacity from temporary mapping growth. Sync and live publication must retain initial/published allocation while continuing to trim temporary oversized mapping windows. Add fail-before/pass-after public reopen/sync/realappend cases for both layouts and mapping strategies; retain existing temporary-window tests. Verify tail append alignment if the same concrete valid fixture exposes it. No change to generic reader tolerance, original declared-extent validation, public API or full-graph cost. Revalidate affected core/public/writer suites and live behavior, commit the coherent correction, and return the focused diff/evidence to the same independent reviewer. Gate remains ready for this fixed-target bug repair.

Independent review also reproduced rejected-open mutation on a small file: original physical2088bytes, final object header at2048 fits but its48byte payload exceeds the arena; Rust opens a grow-capable writer window before its final bound check, returns an error, yet leaves4096bytes. Short files can similarly grow through the initial mutable header mapping. The same pre-mutation target therefore requires fixed-size header/tail preflight through unmutating reads before writable mappings. Existing fulltail constructor checks must also read without growing an already-mutable file. Tests must use tight physical sizes and compare exact bytes, not only8MiBfixtures that mask mapping granularity. A separate actual-append check reproduced MisalignedOffset after a valid unaligned compact tail; constructor now follows the ordinary append path's aligned object size. These are coupled manifestations of the same boundary invariant; no broader rewrite or contract fork is introduced.

Independent-review correction validation: Rust shared original-FD fixed-header/tail preflight now precedes writable maps; existing constructors reuse the uncached positional read path for tail validation. WindowManager retains initial/published allocation separately from temporary mapping growth, and resumed append offsets follow aligned object end with checked addition. Six public regressions fail before/pass after across regular/compact, both strategies, reopen/existing mutablefile, sync/publication/later append and exact rejected small-file preservation; new core control preserves intentional trimming. Final indexed33/33 and Linux public167/core84/directory9 pass. Linux live matrix rerun18/18 passes with the same synthetic workload/settings. Scoped formatting/diff checks pass. Same-failure search for logical_size/set_len/arena_size confirms sync and post_change share the retained allocation invariant, while original bounded validation no longer invokes grow-capable tail mapping. No new per-entry read/syscall is added. Coupled spec and tracked report updated; focused independent rereview follows the correction commit.

Final allocation hot-path measurement: six alternating Rust30k-row structured writer pairs against0ae1ac4, windowed/compact/everyentrypublication, median205.7ms before and204.3ms after with overlapping ranges. No measurable regression; no new per-entry syscall/read/allocation. Audit including changed-artifact sensitive scan and diff checks pass. Concrete retained-allocation/preflight references were searched with `rg -n 'logical_size|set_len|retained_size|validate_committed_arena_header|read_file_exact_at' rust/src/crates/journal-core/src/file/{file.rs,file_mut.rs,mmap.rs,writer.rs}`. Remaining set_len uses are initial creation, permitted temporary mapping growth, bounded retained-allocation publication/sync, and tests; none is an unvalidated append-open path. No coupled item is deferred.

Second-review closure: independent read-only reviewer completed full correction assessment of4fcb79d and focused rereview ofb3d27d9, retaining the earlier coverage of unchanged code. Both independently reproduced blockers now pass: valid compact reopen/sync preserves8388608bytes and strict validity; malformed tight-tail open rejects while retaining2088bytes exactly. Reviewer traced fixed header/tail positional validation before writable mapping, retained allocation versus temporary growth, and aligned resumed append. No remaining verified blocker or material review gap. Additional optional expansion was not made a completion requirement.

Final target/scope assessment: all six current supplied findings receive tracked dispositions; H04 was already fixed in Rust at the earlier head, while the remaining Go/historical-semantics gap is now fixed. No current finding is rejected in full. Previous23-comment dispositions remain22accepted/1rejected with their rationale. Root-cause correction covers declared graph bounds, empty-file metadata ownership and safe writable allocation/reuse across languages; no compatibility adapters, hidden graph scan, auto repair, dependency change or unrelated source edit. DEM journal/RUM-history/synthetic-history suites pass against localSDK; Netdata source unchanged. Original approved release/adoption staging remains externally owned and tracked; no new in-scope work is deferred. Public docs/spec and tracked validation/disposition report hold the durable contracts. Raw comments/repro logs stay ignored. Audit and sensitive-data/diff checks pass; completed SOW returns to done. Local commits4fcb79d andb3d27d9 carry the validated implementations; no push.

## Regression - 2026-10-06: Live Reader Arena Growth

Approval and scope: the user accepted the reproduced finding and proposed validation-boundary correction, then authorized implementation and native subagent verification where useful. Preserve the public live-reader contract, strict recovery/snapshot safety and current writer behavior; no new dependency, remote action or consumer change. Main owns implementation; one native Astra high read-only reviewer will challenge the interacting validation contracts and final correction. Existing implementation review remains relevant outside this affected surface; no external CLI review is requested.

Gate: ready. The approved clean end state distinguishes safe reads of committed entries during normal writer growth from validation of a stable, writer-excluded arena. Full declared-arena consistency remains mandatory for strict verification, append-open and indexed snapshots. Ordinary readers validate physical mapping/access extents without requiring the rounded future allocation to exist already. Remove the mistaken universal validation at ordinary open, retain useful shared bounds mechanisms, and complete coupled tests/docs/spec/review evidence.

Reproduction: at base f35889c, the public Rust lazy field iterator provides a deterministic observation point after preparing a9MiB MESSAGE DATA object but before publishing the next ENTRY, with default publication cadence1, no compression,64KiB mappings,64DATA/16FIELD buckets and synthetic IDs. Both regular/compact layouts declare16,777,216bytes while physical size is9,502,720bytes and n_entries remains1. Rust open returns ObjectExceedsFileBounds; Go opens and reads the committed seed. After append, both traverse2entries. No SDK source or journal-byte mutation is needed. Source and raw output: .local/human-arena-probe/. Human source: https://github.com/netdata/systemd-journal-sdk/pull/5#discussion_r4193955216.

Root cause: stable declared-arena consistency was applied in the shared Rust ordinary-reader constructor. Rust object_added publishes rounded arena_size before end-of-entry post_change extends physical allocation; capturing header before stat does not resolve that publication order. Prior growth tests inspect completed/synced writes and miss this interleaving. The correction must target validation ownership rather than a numeric special case or weakening strict recovery.

Plan:
1. Add deterministic regular/compact regressions exercising existing/new readers during growth, demonstrate failure first, and extend shared interoperability coverage across Rust/Go writers and readers at an8MiB boundary.
2. Separate ordinary reader mapping/access safety from stable arena validation through the existing header/file owners; preserve strict verification, append-open and indexed snapshot negative cases. Inspect callers and record reference-search dispositions.
3. Run affected Rust/Go suites, live/verification matrices with file-backed stock tooling where available, and bounded before/after reader-open cost evidence. Update contracts, review dispositions and validation evidence; commit validated implementation before independent final review.
4. Resolve verified review findings, run project audit, close the original SOW with current evidence, and make any validated review/documentation follow-up commit. No push or release.

Risks/validation: an overly broad relaxation could map invalid tables or accept a malformed stable snapshot; an overly strict check could reject ordinary publication states. Negative tests must prove safe bounds and strict rejection still hold, while live tests must read committed data at the actual writer transition instead of synthesizing headers. No extra hot-path syscalls/scans should be introduced. Public/synthetic evidence only. Tracked docs/specs and this SOW will explain the validation boundary; unchanged archive-state and payload-optimization SOWs remain closed.


Implementation checkpoint: ordinary Rust opening now uses validate_reader_mappings, retaining checked declared-end arithmetic, on-disk header validity, paired/aligned table descriptors and bounds against min(declared,physical). Stable validated_arena_end composes those checks with full extent and tail checks; its indexed-capture and append/recovery callers remain. No writer behavior or Go runtime changes. Added deterministic SDK growth matrix and both-language unit coverage, plus32ordinary-open malformed cases and beyond-file/overflow cases in every stable validation test family. Updated published explanation, product scope, journal compatibility guidance and tracked review/validation reports.

Validation before final review: regression failed on base f35889c and passes after correction. Linux Rust core86/public167 pass; directory writer9pass unprivileged after the root-only permission-fixture failure. Go all-module tests/journal race/module vet pass. Deterministic growth matrix20/20observations passes on macOS and Linux. Full stock/live18/18cases and verifier63/63results pass with systemd257.13;30entries/writer,2polling readers/language,1libsystemd follower, all9feature modes. Wiki check and18Go/15Rust examples pass; offline Rust rerun resolved only a DNS dependency-fetch failure. Open benchmark six alternating10k-open pairs:11.934→11.973microseconds median(+0.32%, overlapping ranges), same file/source except correction, Rust1.91macOSarm64, no cold-cache guarantee. Raw evidence .local/human-arena-probe/; tracked report documentation/indexed-snapshot-validation.md.

Reference search: `rg -n 'validated_arena_end|validate_reader_mappings|validate_committed_arena' rust/src -g '*.rs'` reports11hits: ordinary open migrated; shared mapping/full helpers retained; indexed capture and both append/reuse validation paths retained. Strict verifier independently checks source length in verify_graph/header.rs. Go `validateDeclaredArena` remains confined to indexed capture, strict graph verification and append-open. No missed ordinary-open caller, adapter or deferred coupled path. Main self-review confirms unchanged stable rejection invariants and no new hot-path syscall or allocation. Pending only independent final review and audit/closure for this correction; no new product decision or release action.
