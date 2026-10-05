use super::io::*;
use super::*;

struct PostingCursor {
    remaining: u64,
    array: u64,
    index: usize,
    last_entry: u64,
}

impl GraphVerifier<'_> {
    fn strict_array_chain(
        &self,
        arrays: &mut HashSet<u64>,
        start: u64,
        mut count: u64,
        label: &str,
        hint: Option<(u64, u64)>,
    ) -> Result<Vec<u64>, String> {
        let entries = self.walk_entry_array_chain(start, count, label)?;
        let mut current = start;
        if count == 0 && hint.is_some_and(|hint| hint != (0, 0)) {
            return Err("nonzero empty tail array hint".into());
        }
        while count > 0 {
            self.source.check()?;
            if !arrays.insert(current) {
                return Err("shared ENTRY_ARRAY".into());
            }
            let array = &self.entry_arrays[&current];
            let used = count.min(array.items.len() as u64);
            count -= used;
            if count == 0 {
                if hint.is_some_and(|hint| hint != (current, used)) {
                    return Err("tail array hint mismatch".into());
                }
                if array.next != 0
                    || array.items[used as usize..]
                        .iter()
                        .any(|offset| *offset != 0)
                {
                    return Err("uncommitted array slots or successor".into());
                }
            }
            current = array.next;
        }
        Ok(entries)
    }

    // Each DATA posting cursor advances once per ENTRY reference. Repeated
    // values cost O(total references), without a second incidence graph.
    pub(super) fn validate_strict_indexes(&self) -> Result<(), String> {
        // Parsing each walked table validates its header pointer and extent.
        // A header pointer alone can refer to an unpublished object after tail.
        for kind in [OBJECT_TYPE_DATA_HASH_TABLE, OBJECT_TYPE_FIELD_HASH_TABLE] {
            if self.counts[kind as usize] != 1 {
                return Err("missing or duplicate committed hash table object".into());
            }
        }
        let mut arrays = HashSet::new();
        let entries = self.strict_array_chain(
            &mut arrays,
            self.header.entry_array_offset,
            self.header.n_entries,
            "global entry array",
            if self.header.header_size >= 264 {
                Some((
                    u32_at(self.source, 256)? as u64,
                    u32_at(self.source, 260)? as u64,
                ))
            } else {
                None
            },
        )?;
        let mut cursors = HashMap::new();
        let mut indexed_data = HashSet::new();
        let buckets = self.header.data_hash_table_size / HASH_ITEM_SIZE;
        if buckets == 0 {
            return Err("DATA hash table unavailable".into());
        }
        for bucket in 0..buckets {
            let item = self.header.data_hash_table_offset + bucket * HASH_ITEM_SIZE;
            let mut current = u64_at_u64(self.source, item)?;
            let tail = u64_at_u64(self.source, item + 8)?;
            let mut previous = 0;
            while current != 0 {
                self.source.check()?;
                if current <= previous || !indexed_data.insert(current) {
                    return Err("invalid DATA hash chain".into());
                }
                let data = self
                    .data_objects
                    .get(&current)
                    .ok_or("missing indexed DATA")?;
                if data.hash % buckets != bucket {
                    return Err("DATA hash bucket mismatch".into());
                }
                if data.n_entries == 0 || data.entry_offset == 0 {
                    return Err("DATA missing postings".into());
                }
                if !self.entry_objects.contains_key(&data.entry_offset) {
                    return Err("missing inline ENTRY".into());
                }
                cursors.insert(
                    current,
                    PostingCursor {
                        remaining: data.n_entries,
                        array: data.entry_array_offset,
                        index: 0,
                        last_entry: 0,
                    },
                );
                let mut last_entry = data.entry_offset;
                for entry in self.strict_array_chain(
                    &mut arrays,
                    data.entry_array_offset,
                    data.n_entries - 1,
                    "DATA postings",
                    if self.compact {
                        Some((
                            u32_at_u64(self.source, current + 64)? as u64,
                            u32_at_u64(self.source, current + 68)? as u64,
                        ))
                    } else {
                        None
                    },
                )? {
                    if entry <= last_entry {
                        return Err("orphan or duplicate DATA posting".into());
                    }
                    last_entry = entry;
                }
                previous = current;
                current = data.next_hash_offset;
            }
            if previous != tail {
                return Err("DATA bucket tail mismatch".into());
            }
        }
        if indexed_data.len() != self.data_objects.len() {
            return Err("missing DATA index or reverse ENTRY link".into());
        }
        let mut last = 0;
        for offset in entries {
            self.source.check()?;
            if offset <= last {
                return Err("global entries do not progress".into());
            }
            last = offset;
            for data in &self.entry_objects[&offset].items {
                self.source.check()?;
                let cursor = cursors
                    .get_mut(data)
                    .ok_or("ENTRY references missing indexed DATA")?;
                // The trusted-unique writer API can preserve duplicate items,
                // while the native reverse list contains the entry once.
                if cursor.last_entry == offset {
                    continue;
                }
                if cursor.remaining == 0 {
                    return Err("missing DATA reverse link".into());
                }
                let expected = if cursor.last_entry == 0 {
                    self.data_objects[data].entry_offset
                } else {
                    let array = self
                        .entry_arrays
                        .get(&cursor.array)
                        .ok_or("missing posting array")?;
                    let entry = *array
                        .items
                        .get(cursor.index)
                        .ok_or("missing posting slot")?;
                    cursor.index += 1;
                    if cursor.index == array.items.len() {
                        cursor.array = array.next;
                        cursor.index = 0;
                    }
                    entry
                };
                if expected != offset {
                    return Err("DATA/ENTRY reverse link mismatch".into());
                }
                cursor.last_entry = offset;
                cursor.remaining -= 1;
            }
        }
        if cursors.values().any(|cursor| cursor.remaining != 0) {
            return Err("orphan DATA posting".into());
        }
        drop(cursors);
        if arrays.len() != self.entry_arrays.len() {
            return Err("orphan ENTRY_ARRAY".into());
        }
        let buckets = self.header.field_hash_table_size / HASH_ITEM_SIZE;
        if buckets == 0 {
            return Err("FIELD hash table unavailable".into());
        }
        let mut indexed_fields = HashSet::new();
        let mut field_data = HashSet::new();
        let mut names = HashSet::new();
        for bucket in 0..buckets {
            let item = self.header.field_hash_table_offset + bucket * HASH_ITEM_SIZE;
            let mut current = u64_at_u64(self.source, item)?;
            let tail = u64_at_u64(self.source, item + 8)?;
            let mut previous = 0;
            while current != 0 {
                self.source.check()?;
                if current <= previous || !indexed_fields.insert(current) {
                    return Err("invalid FIELD hash chain".into());
                }
                let (hash, next, head, name) =
                    self.fields.get(&current).ok_or("missing indexed FIELD")?;
                if hash % buckets != bucket || !names.insert(name.as_slice()) {
                    return Err("FIELD hash bucket or name mismatch".into());
                }
                if *head == 0 {
                    return Err("FIELD has no DATA chain".into());
                }
                let mut data_offset = *head;
                let mut last_data = u64::MAX;
                while data_offset != 0 {
                    self.source.check()?;
                    if data_offset >= last_data || !field_data.insert(data_offset) {
                        return Err("invalid FIELD DATA chain".into());
                    }
                    if self.data_names.get(&data_offset) != Some(name) {
                        return Err("FIELD DATA name mismatch".into());
                    }
                    let data = self
                        .data_objects
                        .get(&data_offset)
                        .ok_or("missing FIELD DATA")?;
                    last_data = data_offset;
                    data_offset = data.next_field_offset;
                }
                previous = current;
                current = *next;
            }
            if previous != tail {
                return Err("FIELD bucket tail mismatch".into());
            }
        }
        if indexed_fields.len() != self.fields.len() || field_data.len() != self.data_objects.len()
        {
            return Err("orphan FIELD or missing FIELD DATA link".into());
        }
        Ok(())
    }
}
