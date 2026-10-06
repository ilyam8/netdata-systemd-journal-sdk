//! Synthetic writer/reader for tests/interoperability/run_live_growth.py.
use journal_core::file::{
    EntryField, JournalFile, JournalFileOptions, JournalWriter, Mmap, MmapMut,
};
use sha2::{Digest, Sha256};
use std::io::{BufRead, Write};
use std::path::Path;
use uuid::Uuid;

fn checkpoint(stage: &str, entries: usize) {
    println!(
        "{}",
        serde_json::json!({"stage": stage, "entries": entries})
    );
    std::io::stdout().flush().unwrap();
    let mut ack = String::new();
    assert!(std::io::stdin().lock().read_line(&mut ack).unwrap() > 0);
}

fn read(path: &Path) -> Result<(), Box<dyn std::error::Error>> {
    let reader = JournalFile::<Mmap>::open_path(path, 64 * 1024)?;
    let mut entries = Vec::new();
    reader.entry_offsets(&mut entries)?;
    let mut hashes = Vec::new();
    for entry in entries {
        let mut fields = Vec::new();
        reader.entry_data_object_offsets(entry, &mut fields)?;
        let mut hash = None;
        for field in fields {
            reader.visit_data_payload_at(field, &mut Vec::new(), |payload| {
                if payload.starts_with(b"MESSAGE=") {
                    hash = Some(hex::encode(Sha256::digest(payload)));
                }
                Ok(())
            })?;
        }
        hashes.push(hash.ok_or("missing MESSAGE")?);
    }
    println!("{}", serde_json::to_string(&hashes)?);
    Ok(())
}

fn run() -> Result<(), Box<dyn std::error::Error>> {
    // These are caller-selected inputs; argv[0] is never used as trusted identity.
    // nosemgrep: rust.lang.security.args-os.args-os
    let args: Vec<_> = std::env::args_os().collect();
    let operation = args.get(1).ok_or("missing read|write operation")?;
    let path = Path::new(args.get(2).ok_or("missing journal path")?);
    if operation == "read" {
        return read(path);
    }
    if operation != "write" {
        return Err("unknown operation".into());
    }
    let compact = args.get(3).is_some_and(|value| value == "compact");
    let repo = journal_core::repository::File::from_raw_path(path).ok_or("invalid journal path")?;
    let id = Uuid::from_u128;
    let mut file = JournalFileOptions::new(id(1), id(2), id(3))
        .with_file_id(id(4))
        .with_compact(compact)
        .with_data_hash_table_buckets(64)
        .with_field_hash_table_buckets(16)
        .create::<MmapMut>(&repo)?;
    let mut writer = JournalWriter::new(&mut file, 1, id(2))?;
    writer.add_entry(&mut file, &[b"MESSAGE=seed"], 1_700_000_000_000_001, 1)?;
    checkpoint("seed", 1);
    let mut payload = b"MESSAGE=".to_vec();
    payload.resize(9 * 1024 * 1024, b'x');
    let fields = std::iter::once(EntryField::raw(&payload)).chain(std::iter::from_fn(|| {
        // The previous field is prepared, but its ENTRY has not been published.
        checkpoint("prepared", 1);
        None
    }));
    writer.add_entry_fields(&mut file, fields, 1_700_000_000_000_002, 2)?;
    checkpoint("committed", 2);
    file.sync()?;
    Ok(())
}

fn main() {
    if let Err(err) = run() {
        eprintln!("{err}");
        std::process::exit(1);
    }
}
