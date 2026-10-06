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
	off, d, err := s.findData(ctx, name, value, false)
	if err != nil {
		return err
	}
	if off == 0 {
		return ctx.Err()
	}
	return s.visitPostings(ctx, off, d, visit)
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
		err = s.reader.visitDataPayloadWithHeader(off, d.object, func(payload []byte) error {
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
			if err := s.visitPostings(ctx, off, d, visit); err != nil {
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

func (s *IndexedSnapshot) visitPostings(ctx context.Context, off uint64, d dataHeader, visit func(*SnapshotEntry) error) error {
	if d.nEntries == 0 {
		return snapshotCorrupt("DATA has no committed posting")
	}
	if d.entryOffset > s.maxEntry {
		return ctx.Err()
	}
	if err := s.visitEntry(ctx, d.entryOffset, visit); err != nil {
		return err
	}
	if d.nEntries == 1 {
		return nil
	}
	if d.entryArrayOffset == 0 {
		var err error
		d.entryArrayOffset, err = s.refreshPostingOffset(off+48, 8)
		if err != nil {
			return err
		}
	}
	return s.visitArrays(ctx, d.entryArrayOffset, d.nEntries-1, d.entryOffset, true, visit)
}

// Cached windows can pair a new count with an old zero pointer or slot.
// Writers publish these offsets before the count, so a fresh zero is corrupt.
// Read only that scalar, bypassing cached windows; future offsets still clip.
func (s *IndexedSnapshot) refreshPostingOffset(off, size uint64) (uint64, error) {
	if off < s.reader.header.headerSize || off > s.objectEnd || size > s.objectEnd-off {
		return 0, snapshotCorrupt("posting scalar outside committed bounds")
	}
	var buf [8]byte
	if err := readFileAtFull(s.reader.file, buf[:size], off); err != nil {
		return 0, err
	}
	value := entryOffsetArrayItem(buf[:], size)
	if value == 0 {
		return 0, snapshotCorrupt("missing committed posting offset")
	}
	return value, nil
}

func (s *IndexedSnapshot) visitArrays(ctx context.Context, off, count, previous uint64, clip bool, visit func(*SnapshotEntry) error) error {
	for count > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		if clip && off > s.maxObject {
			return nil
		}
		a, capacity, err := s.arrayHeader(off)
		if err != nil {
			return err
		}
		used := minUint64(count, capacity)
		var clipped bool
		previous, clipped, err = s.visitArrayItems(ctx, off, used, previous, clip, visit)
		if err != nil || clipped {
			return err
		}
		count -= used
		if clip && count > 0 && a.nextArrayOffset == 0 {
			a.nextArrayOffset, err = s.refreshPostingOffset(off+16, 8)
			if err != nil {
				return err
			}
			if a.nextArrayOffset <= off {
				return snapshotCorrupt("invalid entry-array progress")
			}
		}
		off = a.nextArrayOffset
	}
	return ctx.Err()
}

// visitArrayItems reports clipping separately from exhausting this array.
func (s *IndexedSnapshot) visitArrayItems(ctx context.Context, off, used, previous uint64, clip bool, visit func(*SnapshotEntry) error) (uint64, bool, error) {
	size := s.reader.offsetArrayItemSize()
	for pos := uint64(0); pos < used; {
		if err := ctx.Err(); err != nil {
			return previous, false, err
		}
		chunk := minUint64(used-pos, uint64(len(s.postingBuf))/size)
		buf := s.postingBuf[:chunk*size]
		if err := s.reader.readAt(buf, off+offsetArrayObjectHeaderSize+pos*size); err != nil {
			return previous, false, err
		}
		for i := uint64(0); i < chunk; i++ {
			entry := entryOffsetArrayItem(buf[i*size:], size)
			if clip && entry == 0 {
				var err error
				entry, err = s.refreshPostingOffset(off+offsetArrayObjectHeaderSize+(pos+i)*size, size)
				if err != nil {
					return previous, false, err
				}
			}
			if clip && entry > s.maxEntry {
				return previous, true, nil
			}
			if entry <= previous || entry > s.maxEntry {
				return previous, false, snapshotCorrupt("invalid entry-array member")
			}
			if err := s.visitEntry(ctx, entry, visit); err != nil {
				return previous, false, err
			}
			previous = entry
		}
		pos += chunk
	}
	return previous, false, nil
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
			header, err := s.readPayloadHeader(off)
			if err != nil {
				return err
			}
			if err := r.visitDataPayloadWithHeader(off, header, visit); err != nil {
				return err
			}
		}
		pos += chunk
	}
	return e.ctx.Err()
}
