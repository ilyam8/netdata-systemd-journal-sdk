# SOW-0147 - Indexed Snapshots For Domain History

## Status

Status: in-progress
Sub-state: user approved option 1 and prerequisite SDK fixes on 2026-10-06. Implementing Go/Rust writer failure safety, efficient strict verification and bounded indexed snapshots. SDK-first delivery; no publication authorized.

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
- Separate worktree/branch feat/indexed-snapshot starts at local master 8ad648a7b4d36bf2d75ee6ee98b42446ff64a276. Experiment branch remains intact.
- Go/Rust source, public docs, verified examples and current-contract spec are implemented. Validation evidence is below; independent implementation review remains required. Scratch remains under /tmp or .local/.

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
- APIs and approved lifecycle implemented in Go/Rust. Independent implementation review remains pending; no release or DEM integration claim.

Tests or equivalent validation:
- Original modeled interruption evidence remains valid as root-cause evidence. New writer failure and strict-verifier regressions reproduce failing-before cases, including missing final reverse links, zero-reference DATA and failed first-append cleanup.
- TestDesignVerifyRetainedFixture interrupted after 152 seconds; no pass or full runtime estimate claimed. Log /private/tmp/dem-sdk-verification-cost.log.

Real-use evidence:
- Completed spike uses actual SDK writes/reads at 949 MB; production DEM integration remains future step12 validation, not established here.

Reviewer findings:
- Callback/lifetime surface and DEM selected populations accepted. Publication-boundary and crash-reopen integrity finding accepted and reproduced; recovery policy is now approved and its regression belongs in implementation acceptance.

Same-failure scan:
- Traced Go and Rust append publication, Go append-open, Log error propagation and archive ordering; both writer families require their affected paths in implementation verification.

Sensitive data gate:
- Public source and synthetic journal data only. No private telemetry or credentials in tracked artifacts.

Artifact maintenance gate:
- AGENTS.md and runtime skills unchanged. Public snapshot/recovery documentation, executable Go/Rust examples, product-scope spec and validation report carry the delivered contract.
- No output/reference skills exist. SOW is in-progress/current after user approval and status summary records it.

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

Additional implementation evidence:
- Full Go module tests and vet pass. Full journal race suite passed (25.7 s), then focused snapshot/strict verifier race tests passed after bounded capture refinements (5.2 s). All supported DATA compression/layout/access modes, historical unkeyed whole-file fixture, binary names, callback errors/cancellation and concurrent first-array growth are covered.
- Rust core 81, directory writer 9, public SDK 140 pass; two preexisting macOS wheel/root expectations are excluded after baseline reproduction. Shared conformance adapter 15/15 passes. Final logs are /tmp/dem-sdk-rust-final-tests.log and /tmp/dem-sdk-rust-conformance.json.
- Wiki structure validates; 78 harness tests and all 33 marked Go/Rust examples pass. Initial offline cache miss was resolved by copying existing task-local registry cache; no source dependency/version changed.
- Stock systemd 257.13 in a disposable container verifies all seven generated Go/Rust documentation journals. Read-only synthetic mounts only; no host journal or service accessed. This is closed-file verification, not a new Linux live-reader parity claim.
- Go 100k-row collision-heavy snapshot benchmark (six runs): capture median28.1us/12,344B, exact100-row payload traversal includingopen205us/12,456B, FIELD392.5us, all100kpayloads13.31ms/12,608B; eageropen2.195ms/5,155,297B. No cold-cache claim.
- Mixed250kfixture strict verification: Go1.04s/681.6MB cumulative allocations/~494MB process-treepeakRSS; Rust1.64s/~537MB childpeakRSS after cursor optimization. These are offline graph-sized costs, not query-open costs. Rust capture97us, exact2496postings265us, FIELD2.645ms, lazyglobal10.949ms; metadata-only visits differ from Go100kpayload workload.
- Six alternating30krow Go writer trials: baseline143.3ms, changed146.2ms median (~2.0% difference; overlapping ranges). Same mixed32-field structured append, compact layout, defaultlivepublication. No broad speedup/regression claim from this noisy small difference.

## Outcome

Implementation in progress under explicit option-1 approval; no readiness claim.

## Lessons Extracted

Cheap capture consistency checks do not certify arbitrary interrupted graphs. Avoid using the current full verifier as a startup mechanism without measuring and correcting its repeated-posting work.

## Followup

Execute the approved recovery policy and SDK prerequisite before consumer adoption. No extra untracked implementation is deferred.

## Regression Log

No completed claim was reopened. The experiment explicitly excluded production recovery certification; this finding comes from the subsequent production design.

Approval recorded 2026-10-06: user replied "approve" to option 1 and the prerequisite SDK fixes. Gate ready before first source change.
