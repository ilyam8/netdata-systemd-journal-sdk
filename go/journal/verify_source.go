package journal

import (
	"context"
	"encoding/binary"
	"fmt"
)

type verifyByteSource interface {
	Len() uint64
	Slice(offset, size uint64) ([]byte, error)
}

type readerVerifySource struct {
	reader *Reader
}

func (s readerVerifySource) Len() uint64 {
	// Verifier readers are opened with snapshot bounds; this is the immutable
	// byte view that structural and sealed verification operate on.
	return s.reader.fileSize
}

func (s readerVerifySource) Slice(offset, size uint64) ([]byte, error) {
	end, ok := checkedAdd(offset, size)
	if !ok {
		return nil, fmt.Errorf("slice %d..+%d overflows", offset, size)
	}
	if end > s.Len() {
		return nil, fmt.Errorf("slice %d..%d exceeds file bounds", offset, end)
	}
	if size > uint64(int(^uint(0)>>1)) {
		return nil, fmt.Errorf("slice %d..%d exceeds platform bounds", offset, end)
	}
	buf := make([]byte, int(size))
	if err := s.reader.readAt(buf, offset); err != nil {
		return nil, err
	}
	return buf, nil
}

func verifySourceByte(source verifyByteSource, offset uint64) (byte, error) {
	buf, err := source.Slice(offset, 1)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func verifySourceU32(source verifyByteSource, offset uint64) (uint32, error) {
	buf, err := source.Slice(offset, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf), nil
}

func verifySourceU64(source verifyByteSource, offset uint64) (uint64, error) {
	buf, err := source.Slice(offset, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf), nil
}

func verifySourceUUID(source verifyByteSource, offset uint64) (UUID, error) {
	buf, err := source.Slice(offset, 16)
	if err != nil {
		return UUID{}, err
	}
	var id UUID
	copy(id[:], buf)
	return id, nil
}

func verifySourceHeader(source verifyByteSource) (journalHeader, error) {
	size := minUint64(headerSize, source.Len())
	if size < headerMinSize {
		return journalHeader{}, errInvalidJournal
	}
	buf, err := source.Slice(0, size)
	if err != nil {
		return journalHeader{}, err
	}
	return parseHeader(buf)
}

func verifySourceHasHeaderField(source verifyByteSource, headerSize uint64, end int) bool {
	return headerSize >= uint64(end) && source.Len() >= uint64(end)
}

// indexVerifySource amortizes scalar parsing without retaining the whole file.
// A refill allocates a new window so previously returned payload slices remain
// valid until their caller finishes hashing them.
type indexVerifySource struct {
	reader *Reader
	ctx    context.Context
	base   uint64
	window []byte
}

func (s *indexVerifySource) Len() uint64 { return s.reader.fileSize }

func (s *indexVerifySource) Slice(offset, size uint64) ([]byte, error) {
	if err := s.ctx.Err(); err != nil {
		return nil, err
	}
	end, ok := checkedAdd(offset, size)
	if !ok || end > s.Len() {
		return nil, fmt.Errorf("verification read exceeds file bounds")
	}
	if offset >= s.base && end-s.base <= uint64(len(s.window)) {
		return s.window[offset-s.base : end-s.base], nil
	}
	const windowSize = 256 * 1024
	base := offset - offset%windowSize
	length := minUint64(windowSize, s.Len()-base)
	if end-base > length {
		// Unusually large objects get their own read, leaving the scalar window in
		// place for their headers. This avoids copying the same large payload twice.
		return (readerVerifySource{reader: s.reader}).Slice(offset, size)
	}
	window := make([]byte, int(length))
	if err := s.reader.readAt(window, base); err != nil {
		return nil, err
	}
	s.base, s.window = base, window
	return window[offset-base : end-base], nil
}
