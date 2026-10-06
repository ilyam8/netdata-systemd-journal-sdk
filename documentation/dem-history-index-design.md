# DEM history-time and SDK indexed snapshots

Status: approved for implementation, 2026-10-06. User explicitly approved option 1 and prerequisite SDK fixes. The user approved the measured SDK-extension direction, its disclosed costs and SDK-first / DEM-second delivery. This document makes that direction concrete; it does not approve remote publication. DEM has no backward-compatibility obligation. Cloud remains the final initiative step.

## Goal and boundaries

A narrow incident range MUST select retained RUM observations by Agent receipt time and synthetic attempts by actual start time, independently of append delay. Selective queries MUST use native journal indexes without opening-time entry-offset materialization. Broad queries MUST remain exact; no population cap or saved-time shortcut is allowed. Retention remains whole-file saved-age/byte retention.

The SDK owns file-format parsing, stable snapshot bounds, index traversal and payload lifetime. DEM owns schema, timestamp meaning, bucket encoding, query path selection and domain reduction. No database, side index, persistent cache, generic numeric query language, implicit lock helper or host discovery is introduced. Existing SDK readers, facades and Explorer contracts remain unchanged.

## SDK prerequisite: Go and Rust

Provide the same small idiomatic API in both required product languages. Illustrative Go shape (names are routine implementation choices):

```go
type IndexedSnapshotOptions struct {
    Reader        ReaderOptions
    CaptureFields [][]byte
    CaptureValues []Field
}
type CapturedValue struct { Present bool; EntryCount uint64 }

func OpenIndexedSnapshot(ctx context.Context, path string, opts IndexedSnapshotOptions) (*IndexedSnapshot, error)
func (s *IndexedSnapshot) EntryCount() uint64
func (s *IndexedSnapshot) CapturedValue(name, value []byte) (CapturedValue, error)
func (s *IndexedSnapshot) VisitMatch(ctx context.Context, name, value []byte, visit func(*SnapshotEntry) error) error
func (s *IndexedSnapshot) VisitField(ctx context.Context, name []byte, accept func([]byte) (bool, error), visit func(*SnapshotEntry) error) error
func (s *IndexedSnapshot) VisitEntries(ctx context.Context, visit func(*SnapshotEntry) error) error
func (s *IndexedSnapshot) Close() error
func (e *SnapshotEntry) VisitPayloads(visit func([]byte) error) error
```

SnapshotEntry exposes copied sequence/realtime/monotonic/boot metadata when needed, not raw mutable Reader or arbitrary file offsets. Rust mirrors the operations with scoped `&mut SnapshotEntry<'_>` callbacks, Drop for cleanup and lightweight explicit cancellation control. No async runtime dependency is needed. Names and values are byte-safe. Options copy their capture declarations; callers cannot mutate the captured contract afterward.

### Capture and lifetime

- Opening MUST occur while the caller excludes appends/rotation/truncation for that file. Immutable files need no active writer exclusion. The SDK does not acquire implicit locks. DEM already owns the required admission gate.
- Capture a complete owned header, committed object and entry bounds, declared FIELD heads and declared exact-value counts. No offset vector, complete index copy or candidate set is constructed. Metadata from mmap is not frozen merely because file size is frozen.
- Requested value counts are captured under exclusion and stay immutable. Validate direct/terminal postings and array metadata against the committed boundary before exposing a count; a raw mutable DATA count can include an unpublished entry. During capture, requested chains beyond published bounds are inconsistencies, not absence. An absent value returns Present=false, count=0. Asking for an undeclared captured field/value is an explicit error; it does not inspect mutable live metadata.
- Capture all query file handles before DEM releases admission. Once captured, supported append-only writers may continue while traversal reads the original committed population. No guarantee covers truncation, in-place modification or uncoordinated capture during an append. Existing retention handle ownership is preserved; platform-specific unlink behavior is not broadened.
- Establish the committed tail from frozen global entry count and native ENTRY-array metadata for every format; validate newer tail hints against that boundary rather than trusting them independently. Historical headers without the newer hint remain supported. Empty files require consistent empty metadata and remain empty after later appends. Reject inconsistent publication state; never mistake missing historical metadata for empty history.
- Exact hash lookups stop before following post-capture object offsets. FIELD enumeration starts from its captured head. Posting/global-array pointers and entries are checked against captured bounds before dereference. Validate chain progress/cycles and object types, not only a loop counter.
- Copy bounded posting chunks before invoking entry callbacks. In VisitField, finish the value predicate and release its payload guard before traversing postings. Entry/payload views are valid only for their callbacks. No nested snapshot operations during callbacks; payload callbacks cannot retain borrowed bytes or reenter the view. Rust enforces borrowing; Go detects expired views/reentry where practicable without changing allocation complexity.
- A snapshot is single-consumer and non-concurrent. Independent queries use independent snapshots. Close releases every owned handle/window even after error.

### Query semantics and errors

- VisitMatch walks one exact DATA value's postings. Valid absence is empty; unavailable/corrupt native indexes are errors. No full-row fallback hides an indexed operation's changed cost.
- VisitField enumerates existing distinct FIELD values, applies a byte-value predicate, then streams postings of accepted values. One entry may be visited for several accepted values of a multivalued field. This is explicitly a posting traversal, not a deduplicated union or sorted query. DEM's single-bucket invariant makes each candidate unique for its query.
- VisitEntries lazily streams the captured global entry population for parity, broad fallback consumers and debugging; no eager O(entries) allocation. It has documented O(entries) work.
- Return callback, I/O, structural and decompression errors. Cancellation is checked in capture loops, hash/FIELD traversal, posting/global-array chunks and entry payload loops; a syscall or individual decompression remains cooperative, not a hard latency bound. Callers discard partial aggregates on failure.
- Preserve regular/compact layouts, keyed/unkeyed hashing, historical supported header sizes and supported DATA compression. Whole-file `.journal.zst` follows the existing staging/decompression path and explicitly has full-file opening cost; DEM uses ordinary `.journal` files. This API is additive, not a promise that pre-existing facade snapshots now have the new capture contract.

## DEM adoption

### Encoding and completeness

- Keep one process-owned journal; the recommended recovery choice uses the existing SDK strict active/archive naming mode and updates file discovery together. Every row carries exactly one current DEM schema marker.
- RUM uses an explicitly named observation timestamp, replacing ambiguous TSUnixUS / DEM_TS_US native names. Synthetics add a scalar actual-start timestamp that MUST agree with StartedUS in the self-contained JSON. Both phases use that same immutable start timestamp.
- Use separate RUM-observation and synthetic-start bucket fields, canonical decimal `floor(timestamp_us / 60_000_000)`. Minute width is fixed internally, not another operator option.
- The shared append envelope derives schema/bucket fields from the validated kind/timestamp; domain encoders own their other payload fields. Reject caller-supplied reserved envelope fields and duplicate required scalars before an append is attempted. Avoid a parallel copy of domain state in the journal layer.
- Each file snapshot captures the current schema marker's posting count together with total entry count. Nonempty files qualify only if the marker covers all committed entries. Missing/unsupported/partial schema produces an explicit incompatible-history error; empty files qualify. No automatic deletion, migration, silent file omission or old-layout adapter.
- Coverage is a producer provenance invariant under valid native indexes, not a full corruption verifier. Selected rows are validated for scalar multiplicity, canonical timestamp/bucket agreement, domain identity and synthetic scalar/JSON agreement. Structural SDK failures fail the query. Unselected arbitrary semantic corruption is not claimed to be detected.

### Query paths

- Open every retained owned file under admission; no saved-time file pruning. Capture the relevant domain bucket FIELD and schema count, then release admission before traversal.
- Narrow lists probe each intersecting minute's exact bucket value; broad/sparse ranges enumerate existing bucket values and accept those in range. The crossover only changes work, never result completeness; measure it using the real DEM reducer before selecting its internal value.
- Always check exact timestamps after bucket selection. Inclusive Unix-second bounds include their whole second. Avoid converting unconstrained seconds into overflowing microseconds. Literal zero is epoch; omitted and relative defaults follow the reviewed domain contract.
- RUM sessions/errors summarize selected retained observations. Apply user-ID session filtering after aggregation, so other selected observations in a matching session still count. Sort selected sessions at microsecond precision before limiting; Function default sort and exposed precision agree.
- Session timeline uses exact session-ID postings over all retained time, followed by site/kind verification, and preserves the pending/retained occurrence-aware union. Activity-only records remain summary-only. Picker bounds do not clip identity context.
- Synthetic lists use start buckets and reduce both phases before outcome filtering/order/limit. Completion wins over a retained start regardless of wall-clock rollback; complete-only evidence is self-contained, unmatched start remains unknown. Positive completion earlier than start is valid; duration remains independently measured. Normal producer emits one immutable completion, not an update stream.
- Synthetic detail uses run-ID postings plus job/kind verification, preserving active-first diagnosis. No picker bounds or saved-time restriction.
- Preserve redaction, sampling/missing-evidence truth, retained disabled-job history and conditional error-fingerprint details. Errors/cancellation fail the Function rather than return a partial exact-looking aggregate.

## Verified interruption finding and approved recovery boundary

Adversarial review found that exclusion alone does not establish native-index integrity after an interrupted append. Go publishes posting links before committing global n_entries, and writes tail_entry_offset before that final count. Append-open currently checks header/format rather than graph consistency; a subsequent append can preserve orphan postings below a newer valid tail. In the baseline implementation, writer errors did not prevent later append or clean-state publication; the corrected lifecycle described below poisons writers after uncertain mutation.

A deterministic Go interruption model used the real writer and OpenWithOptions, changing only synthetic temporary journal metadata at the documented publication boundary. The postings-before-header window produced five global and five prototype-indexed rows; the tail-before-count window produced five global and six indexed rows. After reopen and another real append, global scan returned six and indexed traversal seven for both windows. VerifyFile rejected both damaged graphs. This is modeled process interruption, not a claim that an actual ENOSPC/power-loss fault was injected. The committed `go/journal/interrupted_publication_test.go` reproduces both header-publication windows with synthetic files and verifies strict rejection before and after a real reopen/append. Run `cd go && go test ./journal -run TestVerifyIndexRejectsInterruptedPublicationAfterReopen -count=1 -v`. This regression establishes the recovery invariant without depending on the historical prototype or temporary artifacts.

The existing verifier is not suitable for unconditional startup use without work: dataReferencesEntry rebuilds and linearly searches a DATA posting list for each referencing ENTRY. A retained-fixture trial was stopped after 152 seconds without finishing the first 250k-row file. This is an aborted cost probe, not a passing test or full-runtime estimate. verify_graph.go also tolerates a missing reverse link for the last entry, so current VerifyFile alone does not prove all required index coverage.

The narrow corrective direction is an SDK prerequisite that prevents a mutating writer from continuing or publishing a clean state after an uncertain failure, while still releasing resources. Pre-mutation validation errors do not poison the writer. Add efficient, strict index-consistency verification before uncertain files are reused or archived; do not repair, delete, omit or replay their records automatically. The indexed reader still has a valid-native-graph precondition and bounded local capture checks; it is not a full recovery engine.

The user approved option 1; option 2 records the rejected alternative:

1. Recommended: switch new DEM journals to the SDK existing StrictSystemdNaming mode: dem.journal is active, dem@...journal files are archives. Current IdentityMode: LogIdentityStrict does not select that mode; current active/archive paths can coincide. Include both filename classes in discovery/retention ownership. Verify every preexisting dem.journal across all retained machine directories regardless of header state before NewLog or any other startup mutation, plus other files without proven eligibility. Trust new-schema archives only when produced by the fixed SDK lifecycle with successful archive file sync before rename and poisoning after uncertain mutation. The schema marker alone is insufficient, and prior WIP files are rejected rather than adopted. External modification and arbitrary storage corruption are outside that archive-provenance guarantee. This avoids a full retained-history scan at every restart.
2. Verify every retained file before startup use. This has the simplest integrity boundary and detects more historical corruption, but adds work proportional to all retained data on every restart. The current quadratic verifier must be corrected first; the actual resulting cost must be measured before delivery.

For both choices, poisoned writer cleanup MUST bypass normal close/archive and empty-file deletion, including a failed first append with cached n_entries=0. DEM MUST keep archive sync enabled. Archives with unexpected state, missing/partial schema coverage or quarantined status fail explicitly. An empty unmarked archive establishes no encoder provenance; a cheap strict empty-graph check may prove there is no retained evidence to omit, otherwise reject it.

Both choices reject damaged files intact and expose unavailable history. Existing DEM startup currently exits if its history store cannot open, so validation failure would prevent plugin startup unless a separately approved degraded-history mode is introduced. Do not silently add that mode. In-process uncertain history writes fail closed; exact scope of measurement availability must be recorded before DEM adoption. User approval resolves this fork; dependent implementation may proceed.

Read-only Astra reviews accepted the callback API and DEM population contract. The interruption finding was accepted and reproduced; its corrective scope and option 1 are now explicitly approved. SDK implementation starts first. User approval of the measured two-step direction remains valid and is not being requested again.

## Delivery and evidence

1. SDK prerequisite: separate branch off local SDK master; implement approved option 1 and include the necessary writer/verification correction before the snapshot API can be adopted; Go/Rust public API, tests, docs and current-contract spec. Reuse existing parsing/mapping primitives, not a copied journal parser. Compare indexed and scan results against independently generated fixtures, including multivalue duplicates and binary fields. Exercise normal/historical headers, compact/regular/unkeyed files, compressed payloads, capture failure/close, cancellation and malformed chains. Stress supported append-only writers during snapshot reads and preserve baseline reader behavior. Record unsupported local platform/tooling validation honestly. Benchmark narrow/broad/open/allocations against spike and sibling implementation. No minimum-version increase or new dependency.
2. DEM adoption: after SDK prerequisite is reviewed/available, migrate producer/read/query/Function/docs/tests together in the reusable Netdata worktree. Use a local module override for pre-release validation only; final dependency must reference the published SDK version before merge readiness. Validate producer-to-journal-to-Function semantics with real SDK and independently asserted expected rows. Compare six alternating benchmark runs, with allocation accounting, at representative retention scale. Update final Cloud mapping; Cloud source stays postponed.

User approval of the two-step direction persists. Public release/tag/push, deleting incompatible local history, and any scope expansion beyond this design still need explicit authorization. Actual production compatibility is established by implementation validation, not experiment measurements or reviewer agreement.
