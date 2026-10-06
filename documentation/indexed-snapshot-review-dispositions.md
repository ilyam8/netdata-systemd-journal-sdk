# Indexed snapshot review dispositions

This sanitized report records the supplied-comment decisions for the indexed
snapshot change. Raw review bodies and local execution logs are not required to
interpret the decisions. Source comments were assessed against code and tests;
no GitHub replies or thread-resolution actions were requested or performed.

## Initial supplied review

The initial file contained 23 comments, including repeated reports. There were
22 accepts and one rejection. Existing tests and the validation report carry
reproduction and execution evidence. Later findings below correct the earlier
incomplete claim that both implementations checked all empty-file metadata.

| ID | Finding | Decision and evidence |
|---|---|---|
| C01 | SHA-256 digest collision buckets | Reject: hypothetical distinct-payload SHA-256 collision, with no established production trigger. The verifier uses all 256 bits and compares payload bytes. Additional buckets do not address a demonstrated defect. |
| C02 | Fixed-memory claim | Accept: document Rust scratch proportional to widest ENTRY and decompression memory. |
| C03 | Expired Go metadata detection | Accept: only VisitPayloads rejects an expired view; exported metadata remains readable. |
| C04 | Go preflight poisoning | Accept: mark actual possible storage/publication-state mutations. Reused DATA plus oversized ENTRY remains retryable; prior DATA/sealing mutation poisons. |
| C05 | 32-bit initializer overflow | Accept root cause, reject skip-only remedy: checked uint64 aggregate sizing before arithmetic/conversion, negative counts rejected. Production linux/386 cross-build passes. |
| C06 | Rust failure wording | Accept: distinguish errors before versus after possible mutation. |
| C07 | Missing Linux live evidence | Accept: the rerun passed all 18 stock/libsystemd/Go/Rust live feature cases and final keyed verification; see indexed-snapshot-validation.md. |
| C08 | Rust example argv panic | Accept: args_os preserves path argument; field/value UTF-8 errors are returned. Existing SDK path limitations remain controlled errors. |
| C09 | Empty active tail boot metadata | Accept: initial correction covered Rust strict/snapshot checks. The second review also found the missing Go checks; see H04 below. |
| C10 | FIELD allocation before captured bound | Accept: validate fixed header, type, minimum size and captured extent before reading variable payload. |
| C11 | Unsynchronized concurrent-growth test | Accept: coordinate appends inside visitor execution, covering array growth deterministically. |
| C12 | Rust preflight poisoning | Accept: same mutation-boundary invariant as C04, including sealing/publication state. |
| C13 | Invalid SOW state wording | Accept: use defined lifecycle state completed after review/validation; reopen while corrections are in flight. |
| C14 | Temporary-only interruption evidence | Accept: committed self-contained TestVerifyIndexRejectsInterruptedPublicationAfterReopen reproduces both publication cut points, including reopen and append. Corrected historical global/indexed count explanation. |
| C15 | Go failure wording | Accept: paired with C06 and mutation-boundary fixes. |
| C16 | Present DATA with zero count means absence | Accept: only missing lookup offset denotes absence; present zero-posting DATA is corruption. |
| C17 | Missing arrays/links/slots silently truncate | Accept: one required-posting rule; refresh ambiguous cached zero scalars because counts may be newer than cached links. A fresh zero is corrupt; positive future offsets clip. Applied to Go and Rust. |
| C18 | Empty journals without indexes | Accept: absent table permitted only with zero corresponding population and zero header pointer/size; required/duplicate tables still rejected. |
| C19 | Missing profiles/comparable performance evidence | Accept evidence correction: record reproducible current hot-path profiles, before/after trials and limits of systemd/sibling comparison. |
| C20 | Multiple postings with no array | Accept, duplicate of C17. |
| C21 | Zero posting slot | Accept, duplicate of C17. |
| C22 | Empty tail boot ID | Accept, duplicate of C09. |
| C23 | Repeated DATA within one ENTRY | Accept: reverse indexes represent distinct DATA×ENTRY membership. Go strict consumes once, matching Rust and stock journalctl 257.13 verification of the six-entry duplicate-reference fixture. |

## Second supplied review

The two new bot findings were read from the updated local file. The four human
comments were fetched individually with curl from the GitHub REST review-comment
endpoint. They originally reviewed 280056d; applicability was reassessed after
rebasing this branch onto local master 5cb48c8. The rebase changed no Go/Rust source.

| ID | Finding | Decision and current handling |
|---|---|---|
| B01 | Dispositions exist only in ignored local notes | Accept: this tracked report records all 23 original decisions plus this round; raw comments remain local. |
| B02 | Go strict verification skips empty metadata | Accept: the early return omits current-file head/time/boot checks. Address through the shared empty-population invariant, preserving inherited sequence state. Overlaps H03/H04. |
| H01 | Declared arena excludes committed objects | Accept: physical bytes alone do not certify declared file bounds. Verify object and table extents against the declared arena, and reject unsafe append-open before mapping, truncation or header publication. Covers both layouts and languages. After validation, Go preserves existing preallocation as its writable arena so reopen and later publication cannot shrink valid padding. |
| H02 | Empty rotated files inherit a tail sequence | Accept: tail_entry_seqnum is chain state and may be nonzero without current-file entries. Snapshots and strict verification must accept it while preserving frozen zero-entry population after a later append. |
| H03 | Empty Go head metadata passes strict verification | Accept: nonzero current-file head metadata can survive reopen and corrupt a later first append. Same invariant fix as B02; compatibility verification remains distinct. |
| H04 | Modern empty boot metadata is unchecked | Partly already fixed: the previous correction covered Rust. Accept the remaining Go gap and align historical header/flag semantics across strict and snapshot paths. |

Human source links:

- [H01: declared arena](https://github.com/netdata/systemd-journal-sdk/pull/5#discussion_r4192562428)
- [H02: inherited tail sequence](https://github.com/netdata/systemd-journal-sdk/pull/5#discussion_r4192562436)
- [H03: empty head metadata](https://github.com/netdata/systemd-journal-sdk/pull/5#discussion_r4192562442)
- [H04: empty boot metadata](https://github.com/netdata/systemd-journal-sdk/pull/5#discussion_r4192562451)

Pinned format evidence is systemd/systemd at
c0a5a2516d28601fb3afc1a77d7b42fcfe38fced (v260.1),
src/libsystemd/sd-journal/journal-file.c:440 and :596-675: rotation inherits the
tail sequence counter; header loading checks the arena, table descriptors and
active tail metadata with historical field-presence semantics.

## Scope and validation interpretation

The fix targets the shared file-header invariants, not individual error strings
or fixture-specific exceptions. Strict verification certifies the complete
committed graph; append-open uses bounded header/tail checks and does not silently
perform a full graph scan. Applications still own strict recovery checks for
uncertain files.

The shared regular/compact byte fixtures distinguish stock compatibility from
safe SDK reuse. Stock 257.13 accepts an isolated empty head sequence and a declared
arena ending inside the final object's payload, although the SDK recovery paths
must reject them: a later append preserves the wrong head sequence, or mapping
the declared size can truncate committed bytes. This does not weaken the strict
recovery invariant. The combined head-sequence/time case from the human review
is rejected by stock.

Final test, interoperability and independent-review results are recorded in
[indexed-snapshot-validation.md](indexed-snapshot-validation.md).
