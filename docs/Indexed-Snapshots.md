# Indexed Snapshots

Use `IndexedSnapshot` when an application needs exact field/value postings or
FIELD-value predicates over one fixed population, without allocating all entry
offsets at open. The application owns time semantics, aggregation and any union
across files. Existing file/directory readers, facades and Explorer keep their
existing contracts; this API does not change their snapshot behavior.

## Choose The Operation

| Operation | Work and result |
|---|---|
| Open/capture | Header, global array headers and declared FIELD heads/value counts; no entry-offset vector |
| Exact match | Native DATA hash lookup, then that value's original postings |
| FIELD predicate | Captured FIELD's distinct values, then accepted values' original postings |
| All entries | Lazy global ENTRY-array traversal; O(entries) work with bounded scratch |
| Payload visitor | Decode only the visited entry's payloads; bytes live only during the callback |
| Strict verification | Offline graph certification with time and memory proportional to graph size |

FIELD traversal is a posting traversal: a row with two accepted values is
visited twice. There is no implicit deduplication, sorting or multi-file merge.
An application requiring a union must implement that contract explicitly.
Missing values return no rows. Missing/corrupt native indexes return errors;
there is no hidden full-row fallback. Discard partial aggregates after an error.

## Capture And Ownership

The caller MUST exclude append, rotation and truncation while opening each
snapshot. The SDK acquires no implicit lock. After capture, supported append-only
writers can resume; the snapshot continues to visit its original population.
Truncation and in-place modification are unsupported. For a consistent multi-file
query, open every file while holding the application's admission gate.

Declare FIELDs used by predicate traversal and exact values whose counts will be
needed. Declarations are copied. A captured value distinguishes absent from
present, and its count does not change after append. Asking for an undeclared
count or FIELD is an error. Exact-match traversal itself needs no declaration.

Entry views and payload slices MUST NOT be retained. Copy what is needed before
the callback returns. A snapshot has one consumer; nested snapshot operations
are unsupported. Rust scopes the borrows. Go detects nested operations and use
of an entry view outside its callback, but retaining a reused view is still a
caller error. Independent queries use separate snapshots.

Context/control cancellation is checked between objects, chunks and payloads.
One syscall or decompression is cooperative, not interruptible. Regular and
compact layouts, historical unkeyed files and supported DATA compression work.
Whole-file `.journal.zst` inputs use existing staging and incur full-file
decompression at open; ordinary `.journal` files use bounded reader windows.

## Go

This immutable fixture needs no writer exclusion. For an active file, hold the
application's writer gate through `OpenIndexedSnapshot`.

<!-- verify-example: lang=go id=go-indexed-snapshot -->
```go
ctx := context.Background()
snapshot, err := journal.OpenIndexedSnapshot(ctx,
    "/var/log/journal/example/system.journal",
    journal.IndexedSnapshotOptions{
        CaptureFields: [][]byte{[]byte("PRIORITY")},
        CaptureValues: []journal.Field{journal.StringField("PRIORITY", "6")},
    })
if err != nil { return err }
defer snapshot.Close()
count, err := snapshot.CapturedValue([]byte("PRIORITY"), []byte("6"))
if err != nil { return err }
fmt.Println("committed", snapshot.EntryCount(), "matching", count.EntryCount)
return snapshot.VisitMatch(ctx, []byte("PRIORITY"), []byte("6"),
    func(entry *journal.SnapshotEntry) error {
        return entry.VisitPayloads(func(payload []byte) error {
            if bytes.HasPrefix(payload, []byte("MESSAGE=")) {
                fmt.Println(string(payload))
            }
            return nil
        })
    })
```

Zero `IndexedSnapshotOptions.Reader` uses the default production reader options.
Explicit nonzero reader options retain their access mode; bounds are always
snapshot bounds. `VisitField` accepts a value predicate;
`VisitEntries` streams all captured rows without allocating an offset vector.
`Close` is idempotent and releases resources even after traversal errors.

## Rust

<!-- verify-example: lang=rust id=rust-indexed-snapshot -->
```rust
use journal::{IndexedSnapshot, IndexedSnapshotOptions, SnapshotControl};
let control = SnapshotControl::default();
let mut snapshot = IndexedSnapshot::open(
    "/var/log/journal/example/system.journal",
    IndexedSnapshotOptions {
        capture_fields: vec![b"PRIORITY".to_vec()],
        capture_values: vec![(b"PRIORITY".to_vec(), b"6".to_vec())],
        ..Default::default()
    },
    &control,
)?;
let count = snapshot.captured_value(b"PRIORITY", b"6")?;
println!("committed {} matching {}", snapshot.entry_count(), count.entry_count);
snapshot.visit_match(b"PRIORITY", b"6", &control, |entry| {
    entry.visit_payloads(|payload| {
        if payload.starts_with(b"MESSAGE=") {
            println!("{}", String::from_utf8_lossy(payload));
        }
        Ok(())
    })
})?;
# Ok::<(), Box<dyn std::error::Error>>(())
```

Use `SnapshotControl { cancelled: Some(&predicate) }` for cooperative
cancellation. `visit_field` and `visit_entries` mirror the Go operations.
Resources are released on drop.

## Integrity Before Recovery

Snapshot capture checks local bounds and requested metadata; it does not
certify every index edge. Its precondition is a valid native index graph.
An interrupted append can leave postings ahead of the committed global count.
A later append must not turn those orphan links into apparently valid history.

For uncertain files, call Go `VerifyIndex(ctx, path)` or Rust
`verify_index(path, &control)` while all writers are excluded, **before** opening
a writer or directory `Log`. Strict verification requires complete bidirectional
ENTRY/DATA membership, FIELD/hash coverage, consistent arrays and payload hashes.
It rejects incomplete publication, including a missing final reverse link.
It does not authenticate sealed TAGs; use the existing keyed verifier for that.
`VerifyFile`/`verify_file` retain their compatibility-oriented tolerance and do
not establish this stronger recovery guarantee.

Strict verification is an offline operation, not query-open work. It retains a
graph-sized working set; it is not constant-memory. Applications may avoid
re-verifying known finalized archives only with a documented provenance policy:
fixed writer lifecycle, successful archive file sync, distinct active/archive
names, and no external mutation. A schema marker alone is not proof of index
integrity. Always verify an uncertain active file, regardless of its header state.

After an uncertain mutating failure, writers reject further mutation and close
only resources, without publishing clean metadata or deleting an empty-looking
file. Go exposes `ErrWriterFailed`, preserving the original error; Rust exposes
its failed-writer error. Pre-mutation validation failures do not poison a writer.
No automatic repair, deletion or replay is part of strict verification. See
[[Writer-APIs|Writer APIs]] for lifecycle and durability choices.
