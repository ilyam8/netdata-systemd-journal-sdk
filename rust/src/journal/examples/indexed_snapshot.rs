//! Run against a caller-provided immutable file, or while the caller excludes
//! its writer: cargo run -p systemd-journal-sdk --example indexed_snapshot --
//! /path/to/file.journal FIELD value
use journal::{IndexedSnapshot, IndexedSnapshotOptions, SnapshotControl, verify_index};
use std::time::Instant;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // These are caller-selected inputs; argv[0] is never used as trusted identity.
    // nosemgrep: rust.lang.security.args-os.args-os
    let args: Vec<_> = std::env::args_os().collect();
    if args.len() != 4 {
        return Err("expected journal-path field value".into());
    }
    let path = std::path::Path::new(&args[1]);
    let field = args[2].to_str().ok_or("field must be UTF-8")?.as_bytes();
    let value = args[3].to_str().ok_or("value must be UTF-8")?.as_bytes();
    let control = SnapshotControl::default();
    let started = Instant::now();
    verify_index(path, &control)?;
    println!("verify_index_us={}", started.elapsed().as_micros());
    let started = Instant::now();
    let mut snapshot = IndexedSnapshot::open(
        path,
        IndexedSnapshotOptions {
            capture_fields: vec![field.to_vec()],
            capture_values: vec![(field.to_vec(), value.to_vec())],
            ..Default::default()
        },
        &control,
    )?;
    println!(
        "open_us={} entries={} captured={:?}",
        started.elapsed().as_micros(),
        snapshot.entry_count(),
        snapshot.captured_value(field, value)?
    );
    let started = Instant::now();
    let mut matches = 0;
    snapshot.visit_match(field, value, &control, |entry| {
        let _metadata = entry.metadata();
        matches += 1;
        Ok(())
    })?;
    println!(
        "match_us={} matches={matches}",
        started.elapsed().as_micros()
    );
    let started = Instant::now();
    let mut field_matches = 0;
    snapshot.visit_field(
        field,
        &control,
        |candidate| Ok(candidate == value),
        |_| {
            field_matches += 1;
            Ok(())
        },
    )?;
    println!(
        "field_us={} matches={field_matches}",
        started.elapsed().as_micros()
    );
    let started = Instant::now();
    let mut entries = 0;
    snapshot.visit_entries(&control, |_| {
        entries += 1;
        Ok(())
    })?;
    println!(
        "entries_us={} entries={entries}",
        started.elapsed().as_micros()
    );
    Ok(())
}
