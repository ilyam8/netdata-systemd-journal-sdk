package journal

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func snapshotFixture(t testing.TB, compact bool, compression, count int) (string, *Writer) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "snapshot.journal")
	opts := testOptions()
	opts.Compact = compact
	opts.Compression = compression
	opts.FieldNamePolicy = FieldNamePolicyRaw
	w, err := Create(path, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if !w.closed {
			if err := w.Close(); err != nil {
				t.Error(err)
			}
		}
	})
	appendSnapshotRows(t, w, 0, count)
	return path, w
}

func appendSnapshotRows(t testing.TB, w *Writer, start, count int) {
	t.Helper()
	for i := start; i < start+count; i++ {
		fields := []Field{StringField("SCHEMA", "1"), StringField("BUCKET", fmt.Sprint(i/10)), StringField("ROW", fmt.Sprint(i)), StringField("MESSAGE", strings.Repeat("data", 300)+fmt.Sprint(i)), {Name: "RAW\x00\xff", Value: []byte{0, 255, byte(i)}}}
		if i%3 == 0 {
			fields = append(fields, StringField("BUCKET", "shared"))
		}
		if err := w.Append(fields, EntryOptions{RealtimeUsec: uint64(1_000_000 + i), MonotonicUsec: uint64(i + 1)}); err != nil {
			t.Fatal(err)
		}
	}
}

func openTestSnapshot(t testing.TB, path string, opts IndexedSnapshotOptions) *IndexedSnapshot {
	t.Helper()
	s, err := OpenIndexedSnapshot(context.Background(), path, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return s
}

func TestIndexedSnapshotPopulationAndPayloads(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, compression := range []int{CompressionNone, CompressionZSTD, CompressionXZ, CompressionLZ4} {
			for _, mode := range []ReaderAccessMode{ReaderAccessReadAt, ReaderAccessMmap} {
				t.Run(fmt.Sprintf("compact=%v/compression=%d/access=%d", compact, compression, mode), func(t *testing.T) {
					path, w := snapshotFixture(t, compact, compression, 117)
					name := []byte("BUCKET")
					value := []byte("1")
					s := openTestSnapshot(t, path, IndexedSnapshotOptions{Reader: DefaultReaderOptions().WithAccessMode(mode).WithWindowSize(4096).WithMaxWindows(1), CaptureFields: [][]byte{name, []byte("ABSENT")}, CaptureValues: []Field{{Name: "SCHEMA", Value: value}, StringField("ABSENT", "x")}})
					name[0] = 'X'
					value[0] = '2' // declarations must own their bytes
					appendSnapshotRows(t, w, 117, 128)
					if s.EntryCount() != 117 {
						t.Fatalf("count=%d", s.EntryCount())
					}
					cv, err := s.CapturedValue([]byte("SCHEMA"), []byte("1"))
					if err != nil || !cv.Present || cv.EntryCount != 117 {
						t.Fatalf("count=%+v, %v", cv, err)
					}
					cv, err = s.CapturedValue([]byte("ABSENT"), []byte("x"))
					if err != nil || cv.Present || cv.EntryCount != 0 {
						t.Fatalf("absence=%+v %v", cv, err)
					}
					if _, err := s.CapturedValue([]byte("SCHEMA"), []byte("2")); !errors.Is(err, ErrSnapshotUndeclared) {
						t.Fatal(err)
					}
					var rows []uint64
					visit := func(e *SnapshotEntry) error {
						rows = append(rows, e.Seqnum)
						found := false
						err := e.VisitPayloads(func(p []byte) error {
							if bytes.HasPrefix(p, []byte("ROW=")) {
								found = true
								want := fmt.Sprintf("ROW=%d", e.Seqnum-1)
								if string(p) != want {
									return fmt.Errorf("got %q want %q", p, want)
								}
							}
							return nil
						})
						if err == nil && !found {
							return errors.New("missing ROW")
						}
						return err
					}
					if err := s.VisitMatch(context.Background(), []byte("BUCKET"), []byte("11"), visit); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(rows, []uint64{111, 112, 113, 114, 115, 116, 117}) {
						t.Fatal(rows)
					}
					rows = nil
					if err := s.VisitMatch(context.Background(), []byte("BUCKET"), []byte("20"), visit); err != nil || len(rows) != 0 {
						t.Fatalf("new value leaked: %v %v", rows, err)
					}
					if err := s.VisitMatch(context.Background(), []byte("RAW\x00\xff"), []byte{0, 255, 4}, visit); err != nil || !reflect.DeepEqual(rows, []uint64{5}) {
						t.Fatalf("binary=%v %v", rows, err)
					}
					rows = nil
					if err := s.VisitEntries(context.Background(), visit); err != nil {
						t.Fatal(err)
					}
					if len(rows) != 117 {
						t.Fatal(len(rows))
					}
					for i, v := range rows {
						if v != uint64(i+1) {
							t.Fatal(rows)
						}
					}
					rows = nil
					if err := s.VisitField(context.Background(), []byte("BUCKET"), func([]byte) (bool, error) { return true, nil }, visit); err != nil {
						t.Fatal(err)
					}
					want := make([]uint64, 0, 156)
					for i := 0; i < 117; i++ {
						want = append(want, uint64(i+1))
						if i%3 == 0 {
							want = append(want, uint64(i+1))
						}
					}
					sort.Slice(rows, func(i, j int) bool { return rows[i] < rows[j] })
					sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
					if !reflect.DeepEqual(rows, want) {
						t.Fatalf("multivalue population=%v want %v", rows, want)
					}
					if err := s.VisitField(context.Background(), []byte("ABSENT"), func([]byte) (bool, error) { t.Fatal("absent predicate called"); return true, nil }, visit); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestIndexedSnapshotEmptyAndHistoricalHeader(t *testing.T) {
	for _, count := range []int{0, 33} {
		for _, size := range []uint64{208, 240, 256, 264, 272} {
			t.Run(fmt.Sprintf("count=%d/header=%d", count, size), func(t *testing.T) {
				path, w := snapshotFixture(t, false, CompressionNone, count)
				if err := w.Sync(); err != nil {
					t.Fatal(err)
				}
				if err := w.closeArena(); err != nil {
					t.Fatal(err)
				} // modify only the test header's declared version
				var b [8]byte
				binary.LittleEndian.PutUint64(b[:], size)
				if _, err := w.file.WriteAt(b[:], 88); err != nil {
					t.Fatal(err)
				}
				s := openTestSnapshot(t, path, IndexedSnapshotOptions{CaptureFields: [][]byte{[]byte("BUCKET")}, CaptureValues: []Field{StringField("SCHEMA", "1")}})
				appendSnapshotRows(t, w, count, 10)
				var got int
				if err := s.VisitEntries(context.Background(), func(*SnapshotEntry) error { got++; return nil }); err != nil {
					t.Fatal(err)
				}
				if got != count {
					t.Fatalf("got=%d want=%d", got, count)
				}
				got = 0
				if err := s.VisitMatch(context.Background(), []byte("SCHEMA"), []byte("1"), func(*SnapshotEntry) error { got++; return nil }); err != nil {
					t.Fatal(err)
				}
				if got != count {
					t.Fatalf("match got=%d want=%d", got, count)
				}
			})
		}
	}
}

func TestIndexedSnapshotCallbacksAndCancellation(t *testing.T) {
	path, _ := snapshotFixture(t, false, CompressionNone, 20)
	s := openTestSnapshot(t, path, IndexedSnapshotOptions{CaptureFields: [][]byte{[]byte("BUCKET")}})
	ctx, cancel := context.WithCancel(context.Background())
	var retained *SnapshotEntry
	err := s.VisitEntries(ctx, func(e *SnapshotEntry) error {
		retained = e
		if err := s.Close(); !errors.Is(err, ErrSnapshotBusy) {
			t.Fatal(err)
		}
		if err := s.VisitEntries(ctx, func(*SnapshotEntry) error { return nil }); !errors.Is(err, ErrSnapshotBusy) {
			t.Fatal(err)
		}
		err := e.VisitPayloads(func([]byte) error {
			if err := e.VisitPayloads(func([]byte) error { return nil }); !errors.Is(err, ErrSnapshotBusy) {
				t.Fatal(err)
			}
			cancel()
			return nil
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		return err
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := retained.VisitPayloads(func([]byte) error { return nil }); !errors.Is(err, ErrSnapshotEntryExpired) {
		t.Fatal(err)
	}
	sentinel := errors.New("callback failed")
	if err := s.VisitEntries(context.Background(), func(*SnapshotEntry) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	if err := s.VisitField(context.Background(), []byte("BUCKET"), func([]byte) (bool, error) { return false, sentinel }, func(*SnapshotEntry) error { return nil }); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	if _, err := OpenIndexedSnapshot(ctx, path, IndexedSnapshotOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.VisitEntries(context.Background(), func(*SnapshotEntry) error { return nil }); !errors.Is(err, ErrSnapshotClosed) {
		t.Fatal(err)
	}
}

func TestIndexedSnapshotSystemdCompressedFile(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "systemd", "test-data", "no-rtc", "system.journal.zst")
	r, err := OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	var expected []uint64
	var message []byte
	var matched []uint64
	for r.Next() == nil {
		entry, err := r.GetEntry()
		if err != nil {
			t.Fatal(err)
		}
		expected = append(expected, entry.Seqnum)
		if message == nil && entry.Fields["MESSAGE"] != nil {
			message = bytes.Clone(entry.Fields["MESSAGE"])
		}
		if bytes.Equal(entry.Fields["MESSAGE"], message) {
			matched = append(matched, entry.Seqnum)
		}
	}
	s := openTestSnapshot(t, path, IndexedSnapshotOptions{CaptureFields: [][]byte{[]byte("MESSAGE")}, CaptureValues: []Field{{Name: "MESSAGE", Value: message}}})
	var got []uint64
	visit := func(e *SnapshotEntry) error { got = append(got, e.Seqnum); return nil }
	if err := s.VisitEntries(context.Background(), visit); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("global got%v want%v", got, expected)
	}
	got = nil
	if err := s.VisitMatch(context.Background(), []byte("MESSAGE"), message, visit); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, matched) {
		t.Fatalf("match got%v want%v", got, matched)
	}
	cv, err := s.CapturedValue([]byte("MESSAGE"), message)
	if err != nil || !cv.Present || cv.EntryCount != uint64(len(matched)) {
		t.Fatalf("count=%v %v", cv, err)
	}
	got = nil
	if err := s.VisitField(context.Background(), []byte("MESSAGE"), func(v []byte) (bool, error) { return bytes.Equal(v, message), nil }, visit); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, matched) {
		t.Fatalf("field got%v want%v", got, matched)
	}
}

func TestIndexedSnapshotRejectsUnpublishedTailAndMissingIndexes(t *testing.T) {
	for _, mutation := range []string{"count", "tail", "index", "array-cycle"} {
		t.Run(mutation, func(t *testing.T) {
			path, w := snapshotFixture(t, false, CompressionNone, 5)
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			f, err := os.OpenFile(path, os.O_RDWR, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			var offset int64
			var value uint64
			switch mutation {
			case "count":
				offset = 152
				value = 4
			case "tail":
				offset = 264
				value = w.header.tailEntryOffset + 8
			case "index":
				offset = 104
				value = 0
			case "array-cycle":
				offset = int64(w.header.entryArrayOffset + 16)
				value = w.header.entryArrayOffset
			}
			var b [8]byte
			binary.LittleEndian.PutUint64(b[:], value)
			if _, err := f.WriteAt(b[:], offset); err != nil {
				t.Fatal(err)
			}
			s, err := OpenIndexedSnapshot(context.Background(), path, IndexedSnapshotOptions{CaptureValues: []Field{StringField("SCHEMA", "1")}})
			if err == nil {
				s.Close()
				t.Fatal("corrupt capture succeeded")
			}
		})
	}
}

func TestIndexedSnapshotAppendDuringTraversal(t *testing.T) {
	for _, compact := range []bool{false, true} {
		t.Run(fmt.Sprint(compact), func(t *testing.T) {
			path, w := snapshotFixture(t, compact, CompressionNone, 2000)
			s := openTestSnapshot(t, path, IndexedSnapshotOptions{Reader: DefaultReaderOptions().WithWindowSize(4096).WithMaxWindows(1), CaptureFields: [][]byte{[]byte("BUCKET")}, CaptureValues: []Field{StringField("SCHEMA", "1")}})
			done := make(chan error, 1)
			go func() {
				for i := 0; i < 1500; i++ {
					if err := w.Append([]Field{StringField("SCHEMA", "1"), StringField("BUCKET", "0"), StringField("LARGE", strings.Repeat("growing", 1000))}, EntryOptions{MonotonicUsec: uint64(3000 + i), RealtimeUsec: uint64(5_000_000 + i)}); err != nil {
						done <- err
						return
					}
				}
				done <- nil
			}()
			for i := 0; i < 5; i++ {
				var count int
				err := s.VisitMatch(context.Background(), []byte("SCHEMA"), []byte("1"), func(e *SnapshotEntry) error { count++; return e.VisitPayloads(func([]byte) error { return nil }) })
				if err != nil || count != 2000 {
					t.Errorf("population=%d %v", count, err)
					break
				}
			}
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			var got int
			if err := s.VisitField(context.Background(), []byte("BUCKET"), func(v []byte) (bool, error) { return bytes.Equal(v, []byte("0")), nil }, func(*SnapshotEntry) error { got++; return nil }); err != nil || got != 10 {
				t.Fatalf("field=%d %v", got, err)
			}
		})
	}
}

func TestIndexedSnapshotConcurrentFirstPostingArray(t *testing.T) {
	for _, mode := range []ReaderAccessMode{ReaderAccessMmap, ReaderAccessReadAt} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			path, w := snapshotFixture(t, false, CompressionNone, 80)
			s := openTestSnapshot(t, path, IndexedSnapshotOptions{Reader: DefaultReaderOptions().WithAccessMode(mode).WithWindowSize(4096).WithMaxWindows(1)})
			done := make(chan error, 1)
			go func() {
				for i := 0; i < 2400; i++ {
					if err := w.Append([]Field{StringField("ROW", fmt.Sprint(i%80))}, EntryOptions{RealtimeUsec: uint64(5_000_000 + i), MonotonicUsec: uint64(1000 + i)}); err != nil {
						done <- err
						return
					}
				}
				done <- nil
			}()
			for i := 0; i < 800; i++ {
				got := 0
				err := s.VisitMatch(context.Background(), []byte("ROW"), []byte(fmt.Sprint(i%80)), func(e *SnapshotEntry) error {
					got++
					if e.Seqnum != uint64(i%80+1) {
						return fmt.Errorf("leaked seqnum %d", e.Seqnum)
					}
					return nil
				})
				if err != nil || got != 1 {
					t.Errorf("count=%d err=%v", got, err)
					break
				}
			}
			if err := <-done; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestIndexedSnapshotLegacyBootMetadata(t *testing.T) {
	for _, modern := range []bool{false, true} {
		t.Run(fmt.Sprint(modern), func(t *testing.T) {
			path, w := snapshotFixture(t, false, CompressionNone, 3)
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			flags := binary.LittleEndian.Uint32(b[8:])
			if !modern {
				flags &^= compatibleTailEntryBootID
			}
			binary.LittleEndian.PutUint32(b[8:], flags)
			b[56] ^= 0xff
			binary.LittleEndian.PutUint64(b[200:], 999)
			if err := os.WriteFile(path, b, 0600); err != nil {
				t.Fatal(err)
			}
			s, err := OpenIndexedSnapshot(context.Background(), path, IndexedSnapshotOptions{})
			if modern {
				if err == nil {
					s.Close()
					t.Fatal("modern mismatched tail accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if err := VerifyIndex(context.Background(), path); err != nil {
				t.Fatal(err)
			}
			if err := s.VisitEntries(context.Background(), func(e *SnapshotEntry) error {
				if e.BootID != testBootID || e.Monotonic != e.Seqnum {
					return fmt.Errorf("entry metadata was replaced by legacy header")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestIndexedSnapshotRejectsFieldWithoutData(t *testing.T) {
	path, w := snapshotFixture(t, false, CompressionNone, 3)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for off := binary.LittleEndian.Uint64(b[88:]); off <= binary.LittleEndian.Uint64(b[136:]); off += align8(binary.LittleEndian.Uint64(b[off+8:])) {
		size := binary.LittleEndian.Uint64(b[off+8:])
		if b[off] == objectTypeField && string(b[off+fieldObjectHeaderSize:off+size]) == "SCHEMA" {
			binary.LittleEndian.PutUint64(b[off+32:], 0)
			break
		}
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	s, err := OpenIndexedSnapshot(context.Background(), path, IndexedSnapshotOptions{CaptureFields: [][]byte{[]byte("SCHEMA")}})
	if err == nil {
		s.Close()
		t.Fatal("missing FIELD data chain reported as absent")
	}
}

func TestIndexedSnapshotArchivedStateIsFrozen(t *testing.T) {
	for _, mode := range []ReaderAccessMode{ReaderAccessReadAt, ReaderAccessMmap} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			path, writer := snapshotFixture(t, false, CompressionNone, 1)
			opts := IndexedSnapshotOptions{Reader: DefaultReaderOptions().WithAccessMode(mode)}
			active := openTestSnapshot(t, path, opts)
			if active.IsArchived() {
				t.Fatal("active capture reported archived")
			}
			if err := writer.ArchiveTo(path); err != nil {
				t.Fatal(err)
			}
			if active.IsArchived() {
				t.Fatal("archive changed the previously captured state")
			}
			archived := openTestSnapshot(t, path, opts)
			if !archived.IsArchived() {
				t.Fatal("new capture did not report archived")
			}
		})
	}
	path, writer := snapshotFixture(t, false, CompressionNone, 1)
	if err := writer.CloseOffline(); err != nil {
		t.Fatal(err)
	}
	if openTestSnapshot(t, path, IndexedSnapshotOptions{}).IsArchived() {
		t.Fatal("offline capture reported archived")
	}
}

func TestIndexedSnapshotRejectsMalformedPayloadObjects(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, mode := range []ReaderAccessMode{ReaderAccessReadAt, ReaderAccessMmap} {
			for _, mutation := range []string{"type", "short-object", "short-data", "extent", "overflow", "unaligned", "past-tail", "before-header", "compression"} {
				t.Run(fmt.Sprintf("compact=%v/access=%d/%s", compact, mode, mutation), func(t *testing.T) {
					path, writer := snapshotFixture(t, compact, CompressionNone, 1)
					if err := writer.Close(); err != nil {
						t.Fatal(err)
					}
					file, err := os.OpenFile(path, os.O_RDWR, 0)
					if err != nil {
						t.Fatal(err)
					}
					defer file.Close()
					itemOffset := writer.header.tailEntryOffset + entryObjectHeaderSize
					var item [8]byte
					if _, err := file.ReadAt(item[:], int64(itemOffset)); err != nil {
						t.Fatal(err)
					}
					dataOffset := binary.LittleEndian.Uint64(item[:])
					payloadOffset := uint64(dataObjectHeaderSize)
					if compact {
						dataOffset = uint64(binary.LittleEndian.Uint32(item[:]))
						payloadOffset = compactDataObjectHeaderSize
					}
					patchOffset := dataOffset + 8
					patch := make([]byte, 8)
					switch mutation {
					case "type":
						patchOffset, patch = dataOffset, []byte{objectTypeField}
					case "short-object":
						binary.LittleEndian.PutUint64(patch, objectHeaderSize-1)
					case "short-data":
						binary.LittleEndian.PutUint64(patch, payloadOffset-1)
					case "extent":
						binary.LittleEndian.PutUint64(patch, writer.appendOffset)
					case "overflow":
						binary.LittleEndian.PutUint64(patch, ^uint64(0))
					case "compression":
						patchOffset, patch = dataOffset+1, []byte{objectCompressedZSTD}
					default:
						patchOffset = itemOffset
						offset := dataOffset + 1
						if mutation == "past-tail" {
							offset = writer.header.tailObjectOffset + 8
						} else if mutation == "before-header" {
							offset = writer.header.headerSize - 8
						}
						binary.LittleEndian.PutUint64(patch, offset)
						if compact {
							patch = patch[:4]
						}
					}
					if _, err := file.WriteAt(patch, int64(patchOffset)); err != nil {
						t.Fatal(err)
					}
					snapshot := openTestSnapshot(t, path, IndexedSnapshotOptions{Reader: DefaultReaderOptions().WithAccessMode(mode).WithWindowSize(4096).WithMaxWindows(1)})
					called := false
					err = snapshot.VisitEntries(context.Background(), func(entry *SnapshotEntry) error {
						return entry.VisitPayloads(func([]byte) error { called = true; return nil })
					})
					if err == nil || called {
						t.Fatalf("malformed first DATA: err=%v, callback called=%v", err, called)
					}
				})
			}
		}
	}
}
