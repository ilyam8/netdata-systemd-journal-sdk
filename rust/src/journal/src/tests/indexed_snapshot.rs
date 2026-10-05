use super::*;
use journal_core::file::{EntryField, EntryWriteOptions};

fn fixture(
    compact: bool,
    compression: Compression,
) -> (
    tempfile::TempDir,
    PathBuf,
    JournalFile<MmapMut>,
    JournalWriter,
) {
    let dir = tempfile::tempdir().unwrap();
    let parent = dir.path().join("01010101010101010101010101010101");
    std::fs::create_dir(&parent).unwrap();
    let path = parent.join("system.journal");
    let file = RepoFile::from_path(&path).unwrap();
    let mut journal = JournalFile::<MmapMut>::create(
        &file,
        JournalFileOptions::new(test_uuid(1), test_uuid(2), test_uuid(3))
            .with_compact(compact)
            .with_compression(compression)
            .with_compress_threshold(8),
    )
    .unwrap();
    let writer = JournalWriter::new(&mut journal, 1, test_uuid(2)).unwrap();
    (dir, path, journal, writer)
}

fn append(journal: &mut JournalFile<MmapMut>, writer: &mut JournalWriter, index: u64) {
    let text = format!("MESSAGE={} {}", index, "repeated text ".repeat(100));
    writer
        .add_entry_fields_with_options(
            journal,
            [
                EntryField::raw(b"SCHEMA=1"),
                EntryField::raw(b"BUCKET=a"),
                EntryField::raw(b"BUCKET=b"),
                EntryField::raw(b"\xff=\x00\xfe"),
                EntryField::raw(text.as_bytes()),
            ],
            1_000_000 + index,
            index,
            EntryWriteOptions::default().field_name_policy(FieldNamePolicy::Raw),
        )
        .unwrap();
}

#[test]
fn indexed_snapshot_bounds_postings_values_and_payload_lifetimes() {
    for compact in [false, true] {
        for compression in [
            Compression::None,
            Compression::Zstd,
            Compression::Lz4,
            Compression::Xz,
        ] {
            let (_dir, path, mut journal, mut writer) = fixture(compact, compression);
            for index in 1..=17 {
                append(&mut journal, &mut writer, index);
            }
            let control = SnapshotControl::default();
            verify_index(&path, &control).unwrap();
            let mut snapshot = IndexedSnapshot::open(
                &path,
                IndexedSnapshotOptions {
                    reader: ReaderOptions::snapshot().with_window_size(4096),
                    capture_fields: vec![b"BUCKET".to_vec(), b"ABSENT".to_vec()],
                    capture_values: vec![
                        (b"SCHEMA".to_vec(), b"1".to_vec()),
                        (b"SCHEMA".to_vec(), b"absent".to_vec()),
                    ],
                },
                &control,
            )
            .unwrap();
            for index in 18..=400 {
                append(&mut journal, &mut writer, index);
            }
            assert_eq!(snapshot.entry_count(), 17);
            assert_eq!(
                snapshot.captured_value(b"SCHEMA", b"1").unwrap(),
                CapturedValue {
                    present: true,
                    entry_count: 17
                }
            );
            assert!(
                !snapshot
                    .captured_value(b"SCHEMA", b"absent")
                    .unwrap()
                    .present
            );
            assert!(snapshot.captured_value(b"SCHEMA", b"undeclared").is_err());
            let mut seen = Vec::new();
            snapshot
                .visit_match(b"SCHEMA", b"1", &control, |entry| {
                    seen.push(entry.metadata().seqnum);
                    let mut binary = false;
                    entry.visit_payloads(|payload| {
                        binary |= payload == b"\xff=\x00\xfe";
                        Ok(())
                    })?;
                    assert!(binary);
                    Ok(())
                })
                .unwrap();
            assert_eq!(seen, (1..=17).collect::<Vec<_>>());
            seen.clear();
            snapshot
                .visit_entries(&control, |entry| {
                    seen.push(entry.metadata().seqnum);
                    Ok(())
                })
                .unwrap();
            assert_eq!(seen, (1..=17).collect::<Vec<_>>());
            let mut count = 0;
            snapshot
                .visit_field(
                    b"BUCKET",
                    &control,
                    |_| Ok(true),
                    |_| {
                        count += 1;
                        Ok(())
                    },
                )
                .unwrap();
            assert_eq!(count, 34, "multivalued rows deliberately repeat");
            snapshot
                .visit_field(b"ABSENT", &control, |_| panic!(), |_| panic!())
                .unwrap();
            assert!(
                snapshot
                    .visit_field(b"UNDECLARED", &control, |_| Ok(true), |_| Ok(()))
                    .is_err()
            );
            let cancelled = || true;
            assert!(matches!(
                snapshot.visit_entries(
                    &SnapshotControl {
                        cancelled: Some(&cancelled)
                    },
                    |_| Ok(())
                ),
                Err(SdkError::Cancelled)
            ));
            let mut count = 0;
            assert!(
                snapshot
                    .visit_match(b"BUCKET", b"a", &control, |_| {
                        count += 1;
                        Err(SdkError::Unsupported("callback"))
                    })
                    .is_err()
            );
            assert_eq!(count, 1);
        }
    }
}

#[test]
fn indexed_empty_snapshot_stays_empty_after_append() {
    let (_dir, path, mut journal, mut writer) = fixture(false, Compression::None);
    let control = SnapshotControl::default();
    let mut snapshot =
        IndexedSnapshot::open(&path, IndexedSnapshotOptions::default(), &control).unwrap();
    append(&mut journal, &mut writer, 1);
    snapshot
        .visit_entries(&control, |_| panic!("empty snapshot grew"))
        .unwrap();
    snapshot
        .visit_match(b"SCHEMA", b"1", &control, |_| panic!("empty snapshot grew"))
        .unwrap();
}

#[test]
fn strict_index_rejects_last_entry_missing_reverse_link() {
    for compact in [false, true] {
        let (_dir, path, mut journal, mut writer) = fixture(compact, Compression::None);
        for index in 1..=3 {
            append(&mut journal, &mut writer, index);
        }
        let offset = journal
            .find_data_offset(journal.hash(b"SCHEMA=1"), b"SCHEMA=1")
            .unwrap()
            .unwrap();
        {
            let mut data = journal.data_mut(offset, None).unwrap();
            data.header.n_entries = NonZeroU64::new(2);
        }
        journal.sync().unwrap();
        assert!(verify_index(&path, &SnapshotControl::default()).is_err());
    }
}

#[test]
fn writer_failure_poison_preserves_bytes_and_prevalidation_reuses() {
    for compact in [false, true] {
        let (_dir, path, mut journal, mut writer) = fixture(compact, Compression::None);
        assert!(writer.add_entry(&mut journal, &[b"invalid"], 1, 1).is_err());
        assert!(!writer.is_poisoned());
        append(&mut journal, &mut writer, 1);
        assert!(
            writer
                .add_entry(&mut journal, &[b"NEW=data", b"invalid"], 2, 2)
                .is_err()
        );
        assert!(writer.is_poisoned());
        let before = std::fs::read(&path).unwrap();
        assert!(matches!(
            writer.add_entry(&mut journal, &[b"MESSAGE=retry"], 3, 3),
            Err(JournalError::WriterPoisoned)
        ));
        drop(writer);
        drop(journal);
        assert_eq!(std::fs::read(path).unwrap(), before);
    }
}

#[test]
fn indexed_historical_and_whole_file_compressed_fixtures() {
    let control = SnapshotControl::default();
    let dir = repo_root().join("fixtures/systemd/test-data/no-rtc");
    for item in std::fs::read_dir(dir).unwrap() {
        let path = item.unwrap().path();
        if path.file_name().unwrap() != "system.journal.zst" {
            continue;
        }
        let mut snapshot =
            IndexedSnapshot::open(&path, IndexedSnapshotOptions::default(), &control)
                .unwrap_or_else(|err| panic!("{}: {err}", path.display()));
        let mut count = 0;
        snapshot
            .visit_entries(&control, |entry| {
                entry.visit_payloads(|_| Ok(()))?;
                count += 1;
                Ok(())
            })
            .unwrap();
        assert_eq!(count, snapshot.entry_count());
        verify_index(&path, &control)
            .unwrap_or_else(|err| panic!("strict {}: {err}", path.display()));
    }
}

#[test]
fn indexed_capture_rejects_unpublished_postings_and_tail_hints() {
    for compact in [false, true] {
        let (_dir, path, mut journal, mut writer) = fixture(compact, Compression::None);
        for index in 1..=5 {
            append(&mut journal, &mut writer, index);
        }
        let before = *journal.journal_header_ref();
        append(&mut journal, &mut writer, 6);
        // Deterministic interruption at the last header publication step:
        // DATA links and tail hints are published, but global count is still old.
        journal.journal_header_mut().n_entries = before.n_entries;
        let options = IndexedSnapshotOptions {
            capture_values: vec![(b"SCHEMA".to_vec(), b"1".to_vec())],
            ..Default::default()
        };
        assert!(
            IndexedSnapshot::open(&path, options.clone(), &SnapshotControl::default()).is_err()
        );
        let header = journal.journal_header_mut();
        header.tail_entry_offset = before.tail_entry_offset;
        header.tail_entry_seqnum = before.tail_entry_seqnum;
        header.tail_entry_realtime = before.tail_entry_realtime;
        header.tail_entry_monotonic = before.tail_entry_monotonic;
        header.tail_entry_array_offset = before.tail_entry_array_offset;
        header.tail_entry_array_n_entries = before.tail_entry_array_n_entries;
        assert!(IndexedSnapshot::open(&path, options, &SnapshotControl::default()).is_err());
        assert!(verify_index(&path, &SnapshotControl::default()).is_err());
    }
}

#[test]
fn indexed_snapshot_survives_concurrent_append_and_array_growth() {
    for compact in [false, true] {
        for initial in [1, 2, 512] {
            let (ready_tx, ready_rx) = std::sync::mpsc::channel();
            let (resume_tx, resume_rx) = std::sync::mpsc::channel();
            let (done_tx, done_rx) = std::sync::mpsc::channel();
            let thread = std::thread::spawn(move || {
                let (_dir, path, mut journal, mut writer) = fixture(compact, Compression::Zstd);
                for index in 1..=initial {
                    append(&mut journal, &mut writer, index);
                }
                ready_tx.send(path).unwrap();
                resume_rx.recv().unwrap();
                for index in initial + 1..=4096 {
                    append(&mut journal, &mut writer, index);
                }
                done_rx.recv().unwrap();
            });
            let path = ready_rx.recv().unwrap();
            let control = SnapshotControl::default();
            let mut snapshot = IndexedSnapshot::open(
                &path,
                IndexedSnapshotOptions {
                    capture_fields: vec![b"MESSAGE".to_vec()],
                    capture_values: vec![(b"SCHEMA".to_vec(), b"1".to_vec())],
                    ..Default::default()
                },
                &control,
            )
            .unwrap();
            resume_tx.send(()).unwrap();
            for _ in 0..8 {
                let mut seqnums = Vec::new();
                snapshot
                    .visit_match(b"SCHEMA", b"1", &control, |entry| {
                        entry.visit_payloads(|_| Ok(()))?;
                        seqnums.push(entry.metadata().seqnum);
                        Ok(())
                    })
                    .unwrap();
                assert_eq!(seqnums, (1..=initial).collect::<Vec<_>>());
                let mut count = 0;
                snapshot
                    .visit_field(
                        b"MESSAGE",
                        &control,
                        |_| Ok(true),
                        |_| {
                            count += 1;
                            Ok(())
                        },
                    )
                    .unwrap();
                assert_eq!(count, initial);
            }
            done_tx.send(()).unwrap();
            thread.join().unwrap();
        }
    }
}

#[test]
fn strict_index_rejects_empty_entry_metadata_and_field_orphans() {
    let (_dir, path, mut journal, mut writer) = fixture(false, Compression::None);
    journal.journal_header_mut().tail_entry_realtime = 1;
    assert!(verify_index(&path, &SnapshotControl::default()).is_err());
    journal.journal_header_mut().tail_entry_realtime = 0;
    append(&mut journal, &mut writer, 1);
    let field = journal
        .find_field_offset(journal.hash(b"SCHEMA"), b"SCHEMA")
        .unwrap()
        .unwrap();
    journal
        .field_mut(field, None)
        .unwrap()
        .header
        .head_data_offset = None;
    assert!(verify_index(&path, &SnapshotControl::default()).is_err());
}
