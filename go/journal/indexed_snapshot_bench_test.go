package journal

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

// This fixture separates open cost from selection and broad payload work.
func BenchmarkIndexedSnapshot(b *testing.B) {
	path := filepath.Join(b.TempDir(), "bench.journal")
	w, err := Create(path, testOptions())
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 100000; i++ {
		if err := w.Append([]Field{StringField("BUCKET", fmt.Sprint(i/100)), StringField("ROW", fmt.Sprint(i)), StringField("SCHEMA", "1")}, EntryOptions{RealtimeUsec: uint64(i + 1), MonotonicUsec: uint64(i + 1)}); err != nil {
			b.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	opts := IndexedSnapshotOptions{CaptureFields: [][]byte{[]byte("BUCKET")}, CaptureValues: []Field{StringField("SCHEMA", "1")}}
	for _, name := range []string{"open", "exact", "field", "all-payloads", "eager-open"} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if name == "eager-open" {
					r, err := OpenFile(path)
					if err != nil {
						b.Fatal(err)
					}
					if err := r.Close(); err != nil {
						b.Fatal(err)
					}
					continue
				}
				s, err := OpenIndexedSnapshot(ctx, path, opts)
				if err != nil {
					b.Fatal(err)
				}
				count := 0
				visit := func(e *SnapshotEntry) error { count++; return e.VisitPayloads(func([]byte) error { return nil }) }
				switch name {
				case "exact":
					err = s.VisitMatch(ctx, []byte("BUCKET"), []byte("500"), visit)
				case "field":
					err = s.VisitField(ctx, []byte("BUCKET"), func(v []byte) (bool, error) { return string(v) == "500", nil }, visit)
				case "all-payloads":
					err = s.VisitEntries(ctx, visit)
				}
				if err != nil {
					b.Fatal(err)
				}
				if name == "exact" || name == "field" {
					if count != 100 {
						b.Fatalf("count=%d", count)
					}
				}
				if name == "all-payloads" && count != 100000 {
					b.Fatal(count)
				}
				if err := s.Close(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
