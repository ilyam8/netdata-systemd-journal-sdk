package journal

import (
	"bytes"
	"context"
)

// VisitMatch visits the original population's postings for an exact byte value.
// Valid absence is empty. Missing or malformed native indexes return an error.
func (s *IndexedSnapshot) VisitMatch(ctx context.Context, name, value []byte, visit func(*SnapshotEntry) error) error {
	if err := s.begin(ctx); err != nil {
		return err
	}
	defer func() { s.busy = false }()
	_, d, err := s.findData(ctx, name, value, false)
	if err != nil {
		return err
	}
	return s.visitPostings(ctx, d, visit)
}

// VisitField applies accept to each distinct value from a declared FIELD head,
// then visits that value's original postings. An entry with multiple accepted
// values can be visited multiple times; this is not a sorted/deduplicated union.
// accept receives borrowed bytes and must not reenter the snapshot.
func (s *IndexedSnapshot) VisitField(ctx context.Context, name []byte, accept func([]byte) (bool, error), visit func(*SnapshotEntry) error) error {
	if err := s.begin(ctx); err != nil {
		return err
	}
	defer func() { s.busy = false }()
	off, ok := s.fields[string(name)]
	if !ok {
		return ErrSnapshotUndeclared
	}
	for off != 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		d, err := s.readData(off)
		if err != nil {
			return err
		}
		if d.nextFieldOffset != 0 && d.nextFieldOffset >= off {
			return snapshotCorrupt("invalid FIELD DATA chain progress")
		}
		accepted := false
		err = s.reader.visitDataPayloadWithHeader(off, d, func(payload []byte) error {
			if len(payload) <= len(name) || !bytes.Equal(payload[:len(name)], name) || payload[len(name)] != '=' {
				return snapshotCorrupt("FIELD DATA payload mismatch")
			}
			var err error
			accepted, err = accept(payload[len(name)+1:])
			return err
		})
		if err != nil {
			return err
		}
		if accepted {
			if err := s.visitPostings(ctx, d, visit); err != nil {
				return err
			}
		}
		off = d.nextFieldOffset
	}
	return ctx.Err()
}

// VisitEntries streams the captured global population in native array order.
// It uses O(entries) work and bounded scratch space, without eager offsets.
func (s *IndexedSnapshot) VisitEntries(ctx context.Context, visit func(*SnapshotEntry) error) error {
	if err := s.begin(ctx); err != nil {
		return err
	}
	defer func() { s.busy = false }()
	return s.visitArrays(ctx, s.reader.header.entryArrayOffset, s.reader.header.nEntries, 0, false, visit)
}

func (s *IndexedSnapshot) visitPostings(ctx context.Context, d dataHeader, visit func(*SnapshotEntry) error) error {
	if d.nEntries == 0 {
		return ctx.Err()
	}
	if d.entryOffset > s.maxEntry {
		return ctx.Err()
	}
	if err := s.visitEntry(ctx, d.entryOffset, visit); err != nil {
		return err
	}
	// A DATA header may span independently refreshed windows during append.
	// A missing array still proves there was none in the captured population:
	// append-only writers never remove an existing pointer.
	if d.nEntries == 1 || d.entryArrayOffset == 0 {
		return nil
	}
	return s.visitArrays(ctx, d.entryArrayOffset, d.nEntries-1, d.entryOffset, true, visit)
}

func (s *IndexedSnapshot) visitArrays(ctx context.Context, off, count, previous uint64, clip bool, visit func(*SnapshotEntry) error) error {
	size := s.reader.offsetArrayItemSize()
	for count > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		if clip && (off == 0 || off > s.maxObject) {
			return nil
		}
		a, capacity, err := s.arrayHeader(off)
		if err != nil {
			return err
		}
		used := minUint64(count, capacity)
		for pos := uint64(0); pos < used; {
			if err := ctx.Err(); err != nil {
				return err
			}
			chunk := minUint64(used-pos, uint64(len(s.postingBuf))/size)
			buf := s.postingBuf[:chunk*size]
			if err := s.reader.readAt(buf, off+offsetArrayObjectHeaderSize+pos*size); err != nil {
				return err
			}
			for i := uint64(0); i < chunk; i++ {
				entry := entryOffsetArrayItem(buf[i*size:], size)
				if clip && (entry == 0 || entry > s.maxEntry) {
					return nil
				}
				if entry <= previous || entry > s.maxEntry {
					return snapshotCorrupt("invalid entry-array member")
				}
				if err := s.visitEntry(ctx, entry, visit); err != nil {
					return err
				}
				previous = entry
			}
			pos += chunk
		}
		count -= used
		off = a.nextArrayOffset
	}
	return ctx.Err()
}

func (s *IndexedSnapshot) visitEntry(ctx context.Context, off uint64, visit func(*SnapshotEntry) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.checkOffset(off); err != nil {
		return err
	}
	h, err := s.reader.readEntryHeaderAt(off)
	if err != nil {
		return err
	}
	if err := s.checkObject(off, h.object.size); err != nil {
		return err
	}
	if (h.object.size-entryObjectHeaderSize)%s.reader.entryItemSize() != 0 {
		return snapshotCorrupt("invalid ENTRY items")
	}
	e := &s.entry
	e.Seqnum, e.Realtime, e.Monotonic, e.BootID = h.seqnum, h.realtime, h.monotonic, h.bootID
	e.offset, e.size, e.ctx, e.active = off, h.object.size, ctx, true
	defer func() { e.active = false; e.ctx = nil }()
	if err := visit(e); err != nil {
		return err
	}
	return ctx.Err()
}

// VisitPayloads visits complete KEY=value payloads. Bytes are valid only until
// the callback returns. Nested payload/snapshot operations are rejected.
func (e *SnapshotEntry) VisitPayloads(visit func([]byte) error) error {
	if !e.active {
		return ErrSnapshotEntryExpired
	}
	if e.inPayload {
		return ErrSnapshotBusy
	}
	e.inPayload = true
	defer func() { e.inPayload = false }()
	s, r := e.snapshot, e.snapshot.reader
	size := r.entryItemSize()
	count := (e.size - entryObjectHeaderSize) / size
	for pos := uint64(0); pos < count; {
		if err := e.ctx.Err(); err != nil {
			return err
		}
		chunk := minUint64(count-pos, uint64(len(s.payloadItems))/size)
		buf := s.payloadItems[:chunk*size]
		if err := r.readAt(buf, e.offset+entryObjectHeaderSize+pos*size); err != nil {
			return err
		}
		for i := uint64(0); i < chunk; i++ {
			if err := e.ctx.Err(); err != nil {
				return err
			}
			off := entryOffsetArrayItem(buf[i*size:], r.offsetArrayItemSize())
			d, err := s.readData(off)
			if err != nil {
				return err
			}
			if err := r.visitDataPayloadWithHeader(off, d, visit); err != nil {
				return err
			}
		}
		pos += chunk
	}
	return e.ctx.Err()
}
