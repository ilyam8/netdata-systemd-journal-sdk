package journal

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func verifyIndexFixture(t testing.TB, count int, compact bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "index.journal")
	opts := testOptions()
	opts.Compact = compact
	w, err := Create(path, opts)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < count; i++ {
		if err := w.Append([]Field{StringField("COMMON", "value"), StringField("OTHER", "value")}, EntryOptions{RealtimeUsec: uint64(i + 1), MonotonicUsec: uint64(i + 1)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestVerifyIndexRejectsMissingLastReverseLink(t *testing.T) {
	path := verifyIndexFixture(t, 2, false)
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for off := binary.LittleEndian.Uint64(buf[88:96]); off < uint64(len(buf)); {
		if buf[off] == objectTypeData {
			binary.LittleEndian.PutUint64(buf[off+48:], 0)
			binary.LittleEndian.PutUint64(buf[off+56:], 1)
			break
		}
		size := binary.LittleEndian.Uint64(buf[off+8:])
		if size == 0 {
			t.Fatal("missing DATA")
		}
		off += align8(size)
	}
	if err := os.WriteFile(path, buf, 0600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFile(path); err != nil {
		t.Fatalf("compatibility verification changed: %v", err)
	}
	if err := VerifyIndex(context.Background(), path); err == nil {
		t.Fatal("missing last ENTRY reverse posting was accepted")
	}
}

func TestVerifyIndexValidLayouts(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, count := range []int{0, 1, 2, 4, 5, 100} {
			t.Run(fmt.Sprintf("compact=%v/count=%d", compact, count), func(t *testing.T) {
				if err := VerifyIndex(context.Background(), verifyIndexFixture(t, count, compact)); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func verifyIndexObjects(buf []byte, typ byte) []uint64 {
	var offsets []uint64
	tail := binary.LittleEndian.Uint64(buf[136:])
	for off := binary.LittleEndian.Uint64(buf[88:]); off <= tail; {
		if buf[off] == typ {
			offsets = append(offsets, off)
		}
		off += align8(binary.LittleEndian.Uint64(buf[off+8:]))
	}
	return offsets
}

func TestVerifyIndexRejectsBrokenCoverage(t *testing.T) {
	tests := map[string]func([]byte){
		"orphan-entry":         func(b []byte) { binary.LittleEndian.PutUint64(b[152:], 1) },
		"missing-global-entry": func(b []byte) { a := binary.LittleEndian.Uint64(b[176:]); binary.LittleEndian.PutUint64(b[a+32:], 0) },
		"extra-global-posting": func(b []byte) {
			a := binary.LittleEndian.Uint64(b[176:])
			e := verifyIndexObjects(b, objectTypeEntry)
			binary.LittleEndian.PutUint64(b[a+40:], e[1])
		},
		"missing-data-hash": func(b []byte) {
			d := verifyIndexObjects(b, objectTypeData)[0]
			n := binary.LittleEndian.Uint64(b[112:]) / hashItemSize
			h := binary.LittleEndian.Uint64(b[d+16:])
			base := binary.LittleEndian.Uint64(b[104:]) + (h%n)*hashItemSize
			clear(b[base : base+16])
		},
		"missing-field-hash": func(b []byte) {
			f := verifyIndexObjects(b, objectTypeField)[0]
			n := binary.LittleEndian.Uint64(b[128:]) / hashItemSize
			h := binary.LittleEndian.Uint64(b[f+16:])
			base := binary.LittleEndian.Uint64(b[120:]) + (h%n)*hashItemSize
			clear(b[base : base+16])
		},
		"missing-field-chain": func(b []byte) {
			f := verifyIndexObjects(b, objectTypeField)[0]
			binary.LittleEndian.PutUint64(b[f+32:], 0)
		},
		"field-wrong-name": func(b []byte) {
			fs := verifyIndexObjects(b, objectTypeField)
			a := binary.LittleEndian.Uint64(b[fs[0]+32:])
			c := binary.LittleEndian.Uint64(b[fs[1]+32:])
			binary.LittleEndian.PutUint64(b[fs[0]+32:], c)
			binary.LittleEndian.PutUint64(b[fs[1]+32:], a)
		},
		"data-posting-order": func(b []byte) {
			d := verifyIndexObjects(b, objectTypeData)[0]
			a := binary.LittleEndian.Uint64(b[d+48:])
			binary.LittleEndian.PutUint64(b[a+24:], binary.LittleEndian.Uint64(b[d+40:]))
		},
		"unpublished-data-posting": func(b []byte) {
			d := verifyIndexObjects(b, objectTypeData)[0]
			a := binary.LittleEndian.Uint64(b[d+48:])
			binary.LittleEndian.PutUint64(b[a+32:], binary.LittleEndian.Uint64(b[d+40:]))
		},
		"tail-array-hint": func(b []byte) { binary.LittleEndian.PutUint32(b[260:], 1) },
		"entry-duplicate-data": func(b []byte) {
			e := verifyIndexObjects(b, objectTypeEntry)[1]
			binary.LittleEndian.PutUint64(b[e+80:], binary.LittleEndian.Uint64(b[e+64:]))
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			path := verifyIndexFixture(t, 2, false)
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mutate(b)
			if err := os.WriteFile(path, b, 0600); err != nil {
				t.Fatal(err)
			}
			if err := VerifyIndex(context.Background(), path); err == nil {
				t.Fatal("damaged index accepted")
			}
		})
	}
}

func TestVerifyIndexCompressionAndHistorical(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, compression := range []int{CompressionNone, CompressionZSTD, CompressionXZ, CompressionLZ4} {
			t.Run(fmt.Sprintf("compact=%v/compression=%d", compact, compression), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "compressed.journal")
				opts := testOptions()
				opts.Compact = compact
				opts.Compression = compression
				opts.CompressThresholdBytes = 1
				w, err := Create(path, opts)
				if err != nil {
					t.Fatal(err)
				}
				for i := 1; i <= 6; i++ {
					if err := w.Append([]Field{StringField("MESSAGE", strings.Repeat("repeated ", 200)), {Name: "BINARY", Value: []byte{0, 255, 10}}}, testEntryOptions(uint64(i))); err != nil {
						t.Fatal(err)
					}
				}
				if err := w.Close(); err != nil {
					t.Fatal(err)
				}
				if err := VerifyIndex(context.Background(), path); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
	path := filepath.Join("..", "..", "fixtures", "systemd", "test-data", "no-rtc", "system.journal.zst")
	if err := VerifyIndex(context.Background(), path); err != nil {
		t.Fatalf("historical systemd fixture: %v", err)
	}
}

func TestVerifyIndexCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := VerifyIndex(ctx, "missing.journal"); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func BenchmarkVerifyIndexRepeatedValue(b *testing.B) {
	for _, count := range []int{1000, 10000, 100000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			path := verifyIndexFixture(b, count, false)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := VerifyIndex(context.Background(), path); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestVerifyIndexRejectsDuplicateDataPayload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "duplicate.journal")
	w, err := Create(path, testOptions())
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range []string{"first", "other"} {
		if err := w.Append([]Field{StringField("FIELD", value)}, testEntryOptions(uint64(i+1))); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ds := verifyIndexObjects(b, objectTypeData)
	size := binary.LittleEndian.Uint64(b[ds[0]+8:])
	copy(b[ds[1]+dataObjectHeaderSize:ds[1]+size], b[ds[0]+dataObjectHeaderSize:ds[0]+size])
	h := binary.LittleEndian.Uint64(b[ds[0]+16:])
	binary.LittleEndian.PutUint64(b[ds[1]+16:], h)
	// Keep both equal DATA objects indexed and independently posted. An exact
	// lookup can return only one even though both reverse relationships are valid.
	base := binary.LittleEndian.Uint64(b[104:])
	tableSize := binary.LittleEndian.Uint64(b[112:])
	clear(b[base : base+tableSize])
	bucket := base + (h%(tableSize/hashItemSize))*hashItemSize
	binary.LittleEndian.PutUint64(b[bucket:], ds[0])
	binary.LittleEndian.PutUint64(b[bucket+8:], ds[1])
	binary.LittleEndian.PutUint64(b[ds[0]+24:], ds[1])
	binary.LittleEndian.PutUint64(b[ds[1]+24:], 0)
	e := verifyIndexObjects(b, objectTypeEntry)[1]
	binary.LittleEndian.PutUint64(b[e+72:], h)
	binary.LittleEndian.PutUint64(b[e+56:], h)
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	err = VerifyIndex(context.Background(), path)
	if err == nil || !strings.Contains(err.Error(), "duplicate DATA payload") {
		t.Fatalf("expected duplicate DATA rejection, got %v", err)
	}
}

func TestVerifyIndexRejectsReverseOnlyPosting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reverse.journal")
	w, err := Create(path, testOptions())
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range []string{"a", "b", "c"} {
		if err := w.Append([]Field{StringField("FIELD", value)}, testEntryOptions(uint64(i+1))); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ds := verifyIndexObjects(b, objectTypeData)
	es := verifyIndexObjects(b, objectTypeEntry)
	binary.LittleEndian.PutUint64(b[ds[2]+40:], es[1])
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFile(path); err != nil {
		t.Fatalf("compatibility verifier changed: %v", err)
	}
	if err := VerifyIndex(context.Background(), path); err == nil {
		t.Fatal("reverse-only posting was accepted")
	}
}

// The deterministic context interrupts inside the graph walk, without a
// wall-clock race or a second goroutine.
type verifyIndexCancelContext struct {
	context.Context
	checks int
}

func (c *verifyIndexCancelContext) Err() error {
	c.checks--
	if c.checks <= 0 {
		return context.Canceled
	}
	return nil
}

func TestVerifyIndexCancellationDuringWalk(t *testing.T) {
	path := verifyIndexFixture(t, 1000, false)
	ctx := &verifyIndexCancelContext{Context: context.Background(), checks: 1000}
	if err := VerifyIndex(ctx, path); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func BenchmarkVerifyIndexRetainedFixture(b *testing.B) {
	path := os.Getenv("JOURNAL_VERIFY_INDEX_BENCH_PATH")
	if path == "" {
		b.Skip("set JOURNAL_VERIFY_INDEX_BENCH_PATH to an immutable journal")
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := VerifyIndex(context.Background(), path); err != nil {
			b.Fatal(err)
		}
	}
}

func TestVerifyIndexRejectsUncommittedData(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, committed := range []bool{false, true} {
			t.Run(fmt.Sprintf("compact=%v/committed=%v", compact, committed), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "uncommitted.journal")
				opts := testOptions()
				opts.Compact = compact
				w, err := Create(path, opts)
				if err != nil {
					t.Fatal(err)
				}
				if committed {
					if err := w.Append([]Field{StringField("COMMITTED", "row")}, testEntryOptions(1)); err != nil {
						t.Fatal(err)
					}
				}
				// Reproduce the actual append prefix: DATA and FIELD indexes exist, but
				// ENTRY publication and the first DATA reverse posting have not happened.
				if _, _, _, err := w.addData([]byte("UNCOMMITTED=value")); err != nil {
					t.Fatal(err)
				}
				// Persist header counters deterministically; this models the prefix graph,
				// rather than injecting a process crash or invoking the public append API.
				if err := w.Close(); err != nil {
					t.Fatal(err)
				}
				if err := VerifyFile(path); err != nil {
					t.Fatalf("compatibility verifier changed: %v", err)
				}
				if err := VerifyIndex(context.Background(), path); err == nil {
					t.Fatal("uncommitted zero-reference DATA accepted")
				}
			})
		}
	}
}

func TestVerifyIndexRejectsUncommittedField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uncommitted-field.journal")
	w, err := Create(path, testOptions())
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := w.addData([]byte("UNCOMMITTED=value")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	field := verifyIndexObjects(b, objectTypeField)[0]
	// addField publishes the FIELD hash membership before linkDataToField sets
	// this head. Restore that exact prefix, with its DATA still uncommitted.
	binary.LittleEndian.PutUint64(b[field+32:], 0)
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFile(path); err != nil {
		t.Fatalf("compatibility verifier changed: %v", err)
	}
	err = VerifyIndex(context.Background(), path)
	if err == nil || !strings.Contains(err.Error(), "FIELD has no DATA") {
		t.Fatalf("expected empty FIELD rejection, got %v", err)
	}
}
