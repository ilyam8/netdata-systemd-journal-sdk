package journal

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDeclaredArenaBoundsBeforeReuse(t *testing.T) {
	mutations := map[string]func([]byte){
		"zero arena":                     func(b []byte) { binary.LittleEndian.PutUint64(b[96:], 0) },
		"entry array outside arena":      func(b []byte) { binary.LittleEndian.PutUint64(b[176:], binary.LittleEndian.Uint64(b[136:])+8) },
		"tail entry outside arena":       func(b []byte) { binary.LittleEndian.PutUint64(b[264:], binary.LittleEndian.Uint64(b[136:])+8) },
		"tail array outside arena":       func(b []byte) { binary.LittleEndian.PutUint32(b[256:], uint32(binary.LittleEndian.Uint64(b[136:])+8)) },
		"tail array count outside arena": func(b []byte) { binary.LittleEndian.PutUint32(b[260:], ^uint32(0)) },
		"hash offset outside arena": func(b []byte) {
			binary.LittleEndian.PutUint64(b[104:], binary.LittleEndian.Uint64(b[136:])+objectHeaderSize+8)
		},
		"cut tail header": func(b []byte) {
			tail := binary.LittleEndian.Uint64(b[136:])
			binary.LittleEndian.PutUint64(b[96:], tail+8-headerSize)
		},
		"cut tail object": func(b []byte) {
			tail := binary.LittleEndian.Uint64(b[136:])
			end := tail + binary.LittleEndian.Uint64(b[tail+8:])
			binary.LittleEndian.PutUint64(b[96:], end-1-headerSize)
		},
		"arena beyond file": func(b []byte) { binary.LittleEndian.PutUint64(b[96:], uint64(len(b))) },
		"arena overflow":    func(b []byte) { binary.LittleEndian.PutUint64(b[96:], ^uint64(0)) },
		"tail size outside arena": func(b []byte) {
			tail := binary.LittleEndian.Uint64(b[136:])
			size := binary.LittleEndian.Uint64(b[tail+8:])
			binary.LittleEndian.PutUint64(b[96:], tail+align8(size)-headerSize)
			binary.LittleEndian.PutUint64(b[tail+8:], size+8)
		},
		"data table outside arena":  func(b []byte) { binary.LittleEndian.PutUint64(b[112:], uint64(len(b))) },
		"field table outside arena": func(b []byte) { binary.LittleEndian.PutUint64(b[128:], uint64(len(b))) },
		"data table overflow":       func(b []byte) { binary.LittleEndian.PutUint64(b[112:], ^uint64(0)-15) },
		"field table overflow":      func(b []byte) { binary.LittleEndian.PutUint64(b[128:], ^uint64(0)-15) },
	}
	for _, compact := range []bool{false, true} {
		for name, mutate := range mutations {
			t.Run(fmt.Sprintf("compact=%v/%s", compact, name), func(t *testing.T) {
				path := verifyIndexFixture(t, 1, compact)
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				mutate(before)
				if err := os.WriteFile(path, before, 0600); err != nil {
					t.Fatal(err)
				}
				if err := VerifyIndex(context.Background(), path); err == nil {
					t.Error("strict verification accepted invalid arena")
				}
				snapshot, err := OpenIndexedSnapshot(context.Background(), path, IndexedSnapshotOptions{})
				if snapshot != nil {
					snapshot.Close()
				}
				if err == nil {
					t.Error("snapshot accepted invalid arena")
				}
				assertRejectedHeaderReuse(t, before)
			})
		}
	}
}

func TestDeclaredArenaAllowsPhysicalPadding(t *testing.T) {
	for _, compact := range []bool{false, true} {
		t.Run(fmt.Sprint(compact), func(t *testing.T) {
			path := verifyIndexFixture(t, 1, compact)
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			tail := binary.LittleEndian.Uint64(b[136:])
			end := tail + align8(binary.LittleEndian.Uint64(b[tail+8:]))
			binary.LittleEndian.PutUint64(b[96:], end-headerSize)
			if err := os.WriteFile(path, b, 0600); err != nil {
				t.Fatal(err)
			}
			if err := VerifyIndex(context.Background(), path); err != nil {
				t.Fatal(err)
			}
			w, err := OpenWithOptions(path, testOptions())
			if err != nil {
				t.Fatal(err)
			}
			defer w.Close()
			s := openTestSnapshot(t, path, IndexedSnapshotOptions{})
			if err := w.Append([]Field{StringField("MESSAGE", "later")}, testEntryOptions(2)); err != nil {
				t.Fatal(err)
			}
			if err := w.CloseOffline(); err != nil {
				t.Fatal(err)
			}
			if err := VerifyIndex(context.Background(), path); err != nil {
				t.Fatal(err)
			}
			visited := 0
			if err := s.VisitEntries(context.Background(), func(*SnapshotEntry) error { visited++; return nil }); err != nil {
				t.Fatal(err)
			}
			if visited != 1 {
				t.Fatalf("frozen visits = %d", visited)
			}
		})
	}
}

func TestAppendOpenPreservesPhysicalPadding(t *testing.T) {
	for _, compact := range []bool{false, true} {
		t.Run(fmt.Sprint(compact), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "padding.journal")
			opts := testOptions()
			opts.Compact = compact
			w, err := Create(path, opts)
			if err != nil {
				t.Fatal(err)
			}
			defer w.Close()
			fields := []Field{StringField("MESSAGE", "same")}
			for i := uint64(1); i <= 3; i++ {
				if err := w.Append(fields, testEntryOptions(i)); err != nil {
					t.Fatal(err)
				}
			}
			if err := w.CloseOffline(); err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			tail := binary.LittleEndian.Uint64(b[136:])
			end := tail + binary.LittleEndian.Uint64(b[tail+8:])
			if compact && end == align8(end) {
				t.Fatal("fixture must retain final object alignment padding")
			}
			binary.LittleEndian.PutUint64(b[96:], end-headerSize)
			if err := os.WriteFile(path, b, 0600); err != nil {
				t.Fatal(err)
			}
			if err := VerifyIndex(context.Background(), path); err != nil {
				t.Fatal(err)
			}
			for _, appendEntry := range []bool{false, true} {
				w, err := OpenWithOptions(path, opts)
				if err != nil {
					t.Fatal(err)
				}
				defer w.Close()
				if appendEntry {
					if err := w.Append(fields, testEntryOptions(4)); err != nil {
						t.Fatal(err)
					}
				}
				if err := w.CloseOffline(); err != nil {
					t.Fatal(err)
				}
				stat, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				if stat.Size() != int64(len(b)) {
					t.Errorf("append=%v: physical size = %d, want %d", appendEntry, stat.Size(), len(b))
				}
				if err := VerifyIndex(context.Background(), path); err != nil {
					t.Errorf("append=%v: %v", appendEntry, err)
				}
			}
		})
	}
}

func TestEmptyInheritedTailSequence(t *testing.T) {
	for _, compact := range []bool{false, true} {
		t.Run(fmt.Sprint(compact), func(t *testing.T) {
			path := verifyIndexFixture(t, 0, compact)
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			binary.LittleEndian.PutUint64(b[160:], 42)
			if err := os.WriteFile(path, b, 0600); err != nil {
				t.Fatal(err)
			}
			if err := VerifyIndex(context.Background(), path); err != nil {
				t.Fatal(err)
			}
			s := openTestSnapshot(t, path, IndexedSnapshotOptions{})
			assertEmpty := func() {
				t.Helper()
				if s.EntryCount() != 0 {
					t.Fatal("nonempty captured population")
				}
				if err := s.VisitEntries(context.Background(), func(*SnapshotEntry) error { t.Fatal("visited empty snapshot"); return nil }); err != nil {
					t.Fatal(err)
				}
			}
			assertEmpty()
			w, err := OpenWithOptions(path, testOptions())
			if err != nil {
				t.Fatal(err)
			}
			defer w.Close()
			if err := w.Append([]Field{StringField("MESSAGE", "first")}, testEntryOptions(1)); err != nil {
				t.Fatal(err)
			}
			if w.header.headEntrySeqnum != 43 {
				t.Fatalf("new head seq = %d", w.header.headEntrySeqnum)
			}
			if err := w.CloseOffline(); err != nil {
				t.Fatal(err)
			}
			if err := VerifyIndex(context.Background(), path); err != nil {
				t.Fatal(err)
			}
			assertEmpty()
		})
	}
}

func TestEmptyCurrentEntryMetadataRejected(t *testing.T) {
	fields := map[string]int{"head sequence": 168, "head realtime": 184, "tail realtime": 192, "tail monotonic": 200, "entry array": 176, "tail array": 256, "tail array count": 260, "tail entry": 264, "tail boot": 56}
	for _, compact := range []bool{false, true} {
		for name, offset := range fields {
			t.Run(fmt.Sprintf("compact=%v/%s", compact, name), func(t *testing.T) {
				path := verifyIndexFixture(t, 0, compact)
				b, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				b[offset] = 1
				if err := os.WriteFile(path, b, 0600); err != nil {
					t.Fatal(err)
				}
				if err := VerifyIndex(context.Background(), path); err == nil {
					t.Error("strict verification accepted empty ENTRY metadata")
				}
				s, err := OpenIndexedSnapshot(context.Background(), path, IndexedSnapshotOptions{})
				if s != nil {
					s.Close()
				}
				if err == nil {
					t.Error("snapshot accepted empty ENTRY metadata")
				}
				assertRejectedHeaderReuse(t, b)
				// Compatibility verification historically permits these head/time/boot values.
				if name == "head sequence" || name == "head realtime" || name == "tail realtime" || name == "tail monotonic" || name == "tail boot" {
					if err := VerifyFile(path); err != nil {
						t.Fatalf("compatibility contract changed: %v", err)
					}
				}
			})
		}
	}
}

func TestEmptyHistoricalHeaderPresence(t *testing.T) {
	for _, size := range []uint64{208, 216, 240, 256, 264, 272} {
		for _, modernBoot := range []bool{false, true} {
			t.Run(fmt.Sprintf("size=%d/modern=%v", size, modernBoot), func(t *testing.T) {
				h := journalHeader{signature: [8]byte{'L', 'P', 'K', 'S', 'H', 'H', 'R', 'H'}, headerSize: size, tailEntrySeqnum: 42, tailEntryBootID: testBootID}
				if modernBoot {
					h.compatibleFlags = compatibleTailEntryBootID
				}
				b := make([]byte, headerSize)
				putHeader(b, h)
				// Bytes after the on-disk header are physical padding, not newer fields.
				for i := size; i < uint64(len(b)); i++ {
					b[i] = 0xff
				}
				path := filepath.Join(t.TempDir(), "historical.journal")
				if err := os.WriteFile(path, b, 0600); err != nil {
					t.Fatal(err)
				}
				err := VerifyIndex(context.Background(), path)
				if (err != nil) != (modernBoot && size >= 272) {
					t.Errorf("strict = %v, modern=%v", err, modernBoot)
				}
				s, err := OpenIndexedSnapshot(context.Background(), path, IndexedSnapshotOptions{})
				if s != nil {
					s.Close()
				}
				if (err != nil) != (modernBoot && size >= 272) {
					t.Errorf("snapshot = %v, modern=%v", err, modernBoot)
				}
			})
		}
	}
}

func assertRejectedHeaderReuse(t *testing.T, before []byte) {
	t.Helper()
	for _, guarded := range []bool{false, true} {
		t.Run(fmt.Sprintf("guarded=%v", guarded), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "reuse.journal")
			if err := os.WriteFile(path, before, 0600); err != nil {
				t.Fatal(err)
			}
			var err error
			if guarded {
				err = VerifyIndex(context.Background(), path)
			}
			if err == nil {
				var w *Writer
				w, err = OpenWithOptions(path, testOptions())
				if w != nil {
					w.Close()
				}
			}
			if err == nil {
				t.Error("reuse accepted invalid header")
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Error("rejected reuse changed journal bytes")
			}
		})
	}
}
