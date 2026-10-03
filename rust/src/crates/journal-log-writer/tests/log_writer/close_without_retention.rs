use super::*;
use std::sync::atomic::AtomicBool;
use std::time::Duration;

#[test]
fn close_without_retention_preserves_one_second_policy_reload() {
    for strict in [false, true] {
        let dir = TempDir::new().unwrap();
        let config = test_config()
            .with_strict_systemd_naming(strict)
            .with_retention_policy(
                RetentionPolicy::default().with_duration_of_journal_files(Duration::from_secs(1)),
            );
        let mut first = Log::new(dir.path(), config).unwrap();
        write_test_entry(&mut first, &[b"MESSAGE=policy reload"]).unwrap();
        std::thread::sleep(Duration::from_millis(1500));
        first.close_without_retention().unwrap();

        let files = journal_file_paths(&dir);
        assert_eq!(files.len(), 1);
        let archive = &files[0];
        let bytes = fs::read(archive).unwrap();
        let second = Log::new(
            dir.path(),
            test_config()
                .with_strict_systemd_naming(strict)
                .with_open_mode(LogOpenMode::Eager),
        )
        .unwrap();
        assert_eq!(fs::read(archive).unwrap(), bytes);
        let file = JournalFile::<Mmap>::open_path(archive, 4096).unwrap();
        assert_eq!(
            file.journal_header_ref().state,
            JournalState::Archived as u8
        );
        assert_eq!(file.journal_header_ref().n_entries, 1);
        verify_journalctl_file(archive);
        if let Some(rows) = read_journal_directory_json(second.journal_directory(), &[]) {
            assert_eq!(rows.len(), 1);
            assert_eq!(rows[0]["MESSAGE"], "policy reload");
        }
        second.close_without_retention().unwrap();
    }
}

#[test]
fn close_without_retention_preserves_expired_rotated_archives() {
    for strict in [false, true] {
        for skip_retention in [false, true] {
            let dir = TempDir::new().unwrap();
            let config = test_config()
                .with_strict_systemd_naming(strict)
                .with_rotation_policy(RotationPolicy::default().with_number_of_entries(1))
                .with_retention_policy(
                    RetentionPolicy::default()
                        .with_duration_of_journal_files(Duration::from_secs(5)),
                );
            let observer = Arc::new(RecordingObserver::default());
            let mut log =
                Log::new_with_lifecycle_observer(dir.path(), config, observer.clone()).unwrap();
            write_test_entry(&mut log, &[b"MESSAGE=older"]).unwrap();
            write_test_entry(&mut log, &[b"MESSAGE=current"]).unwrap();
            let active = log.active_path().unwrap().to_path_buf();
            let files = journal_file_paths(&dir);
            assert_eq!(
                files.len(),
                2,
                "rotation must retain the pre-expiry archive"
            );
            let older = files.iter().find(|path| **path != active).unwrap();
            let older_bytes = fs::read(older).unwrap();
            std::thread::sleep(Duration::from_millis(5100));

            if skip_retention {
                log.close_without_retention().unwrap();
                assert_eq!(fs::read(older).unwrap(), older_bytes);
                assert_eq!(count_journal_files(&dir), 2);
            } else {
                log.close().unwrap();
                assert!(
                    !older.exists(),
                    "normal close must delete the expired archive"
                );
                assert_eq!(count_journal_files(&dir), 1);
            }
            let events = observer.events.lock().unwrap();
            let deleted = events
                .iter()
                .filter(|event| matches!(event, LogLifecycleEvent::RetainedDeleted { .. }))
                .count();
            assert_eq!(deleted, usize::from(!skip_retention));
            for archive in journal_file_paths(&dir) {
                let file = JournalFile::<Mmap>::open_path(&archive, 4096).unwrap();
                assert_eq!(
                    file.journal_header_ref().state,
                    JournalState::Archived as u8
                );
                verify_journalctl_file(&archive);
            }
        }
    }
}

struct FailingArtifactSizer(AtomicBool);

impl LogArtifactSizer for FailingArtifactSizer {
    fn journal_artifact_size(&self, _: &Path) -> journal_log_writer::Result<u64> {
        if self.0.load(Ordering::Relaxed) {
            return Err(std::io::Error::other("retention accounting failure").into());
        }
        Ok(0)
    }
}

#[test]
fn close_without_retention_skips_retention_accounting() {
    for strict in [false, true] {
        for skip_retention in [false, true] {
            let dir = TempDir::new().unwrap();
            let sizer = Arc::new(FailingArtifactSizer(AtomicBool::new(false)));
            let config = test_config()
                .with_strict_systemd_naming(strict)
                .with_retention_policy(RetentionPolicy::default().with_size_of_journal_files(1));
            let mut log =
                Log::new_with_hooks(dir.path(), config, None, Some(sizer.clone())).unwrap();
            write_test_entry(&mut log, &[b"MESSAGE=retention callback"]).unwrap();
            sizer.0.store(true, Ordering::Relaxed);
            if skip_retention {
                log.close_without_retention().unwrap();
            } else {
                assert!(matches!(log.close(), Err(WriterError::Io(_))));
            }
            assert_eq!(count_journal_files(&dir), 1);
            verify_journalctl_file(&journal_file_paths(&dir)[0]);
        }
    }
}

#[test]
fn close_without_retention_preserves_empty_and_lazy_behavior() {
    for strict in [false, true] {
        for eager in [false, true] {
            let dir = TempDir::new().unwrap();
            let config = test_config()
                .with_strict_systemd_naming(strict)
                .with_open_mode(if eager {
                    LogOpenMode::Eager
                } else {
                    LogOpenMode::Lazy
                });
            let log = Log::new(dir.path(), config).unwrap();
            log.close_without_retention().unwrap();
            let expected = usize::from(eager && !strict);
            assert_eq!(journal_file_paths(&dir).len(), expected);
            for archive in journal_file_paths(&dir) {
                let file = JournalFile::<Mmap>::open_path(&archive, 4096).unwrap();
                assert_eq!(
                    file.journal_header_ref().state,
                    JournalState::Archived as u8
                );
                assert_eq!(file.journal_header_ref().n_entries, 0);
            }
        }
    }
}

#[test]
fn close_without_retention_returns_archive_errors() {
    for skip_retention in [false, true] {
        let dir = TempDir::new().unwrap();
        let mut log = Log::new(dir.path(), test_config().with_strict_systemd_naming(true)).unwrap();
        write_test_entry(&mut log, &[b"MESSAGE=archive failure"]).unwrap();
        let active = log.active_path().unwrap().to_path_buf();
        let file = JournalFile::<Mmap>::open_path(&active, 4096).unwrap();
        let header = file.journal_header_ref();
        let archive = log.journal_directory().join(format!(
            "system@{}-{:016x}-{:016x}.journal",
            uuid::Uuid::from_bytes(header.seqnum_id).simple(),
            header.head_entry_seqnum,
            header.head_entry_realtime,
        ));
        drop(file);
        // A directory at the archive path forces a real rename failure.
        fs::create_dir(&archive).unwrap();
        let result = if skip_retention {
            log.close_without_retention()
        } else {
            log.close()
        };
        assert!(matches!(result, Err(WriterError::Io(_))));
        assert!(active.exists());
        verify_journalctl_file(&active);
    }
}
