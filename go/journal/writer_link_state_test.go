package journal

import (
	"path/filepath"
	"testing"
)

func TestNewDataStartsWithEmptyResolvedLinkState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new-data-link-state.journal")
	w, err := Create(path, testOptions())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	offset, _, state, err := w.addData([]byte("MESSAGE=new-data-link-state"))
	if err != nil {
		t.Fatalf("addData() error = %v", err)
	}
	if offset == 0 {
		t.Fatal("addData() offset is zero")
	}
	if state != (resolvedDataLinkState{}) {
		t.Fatalf("new DATA link state = %+v, want empty", state)
	}
}

func TestResolvedLinkStateTracksRegularAndCompactTailGrowth(t *testing.T) {
	payload := []byte("MESSAGE=reused-link-state")
	for _, compact := range []bool{false, true} {
		t.Run(map[bool]string{false: "regular", true: "compact"}[compact], func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "link-state.journal")
			opts := testOptions()
			opts.Compact = compact
			w, err := Create(path, opts)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			var headArrayOffset uint64
			for entryCount := uint64(1); entryCount <= 6; entryCount++ {
				if err := w.AppendRaw([][]byte{payload}, testEntryOptions(entryCount)); err != nil {
					t.Fatalf("AppendRaw(%d) error = %v", entryCount, err)
				}
				_, state, ok, err := w.findData(w.hash(payload), payload)
				if err != nil {
					t.Fatalf("findData(%d) error = %v", entryCount, err)
				}
				if !ok {
					t.Fatalf("findData(%d) did not find DATA", entryCount)
				}
				if state.nEntries != entryCount {
					t.Fatalf("entry %d state nEntries = %d", entryCount, state.nEntries)
				}
				if entryCount == 1 {
					if state.entryArrayOffset != 0 || state.compactTailOffset != 0 || state.compactTailEntries != 0 {
						t.Fatalf("first-link state = %+v", state)
					}
					continue
				}
				if state.entryArrayOffset == 0 {
					t.Fatalf("entry %d array offset is zero", entryCount)
				}
				if headArrayOffset == 0 {
					headArrayOffset = state.entryArrayOffset
				} else if state.entryArrayOffset != headArrayOffset {
					t.Fatalf("entry %d head array changed from %d to %d", entryCount, headArrayOffset, state.entryArrayOffset)
				}
				if !compact {
					if state.compactTailOffset != 0 || state.compactTailEntries != 0 {
						t.Fatalf("regular state has compact tail: %+v", state)
					}
					continue
				}
				if entryCount <= 5 {
					if state.compactTailOffset != headArrayOffset || state.compactTailEntries != entryCount-1 {
						t.Fatalf("compact entry %d tail state = %+v", entryCount, state)
					}
				} else if state.compactTailOffset == headArrayOffset || state.compactTailEntries != 1 {
					t.Fatalf("compact grown tail state = %+v", state)
				}
			}

			if err := w.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
			if got := readJournalSnapshot(t, path).dataByPayload[string(payload)].header.nEntries; got != 6 {
				t.Fatalf("snapshot nEntries = %d, want 6", got)
			}
		})
	}
}

func TestResolvedLinkStateReconstructsAfterReopen(t *testing.T) {
	payload := []byte("MESSAGE=reopened-link-state")
	for _, compact := range []bool{false, true} {
		t.Run(map[bool]string{false: "regular", true: "compact"}[compact], func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "reopened-link-state.journal")
			opts := testOptions()
			opts.Compact = compact
			w, err := Create(path, opts)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			for entry := uint64(1); entry <= 5; entry++ {
				if err := w.AppendRaw([][]byte{payload}, testEntryOptions(entry)); err != nil {
					t.Fatalf("AppendRaw(%d) error = %v", entry, err)
				}
			}
			if err := w.Close(); err != nil {
				t.Fatalf("Close(first) error = %v", err)
			}

			w, err = Open(path)
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			if len(w.entryItemsScratch) != 0 {
				t.Fatalf("reopened writer retained %d entry items", len(w.entryItemsScratch))
			}
			if err := w.AppendRaw([][]byte{payload}, testEntryOptions(6)); err != nil {
				t.Fatalf("AppendRaw(reopened) error = %v", err)
			}
			_, state, ok, err := w.findData(w.hash(payload), payload)
			if err != nil || !ok {
				t.Fatalf("findData(reopened) state/error = %+v/%v", state, err)
			}
			if state.nEntries != 6 || state.entryArrayOffset == 0 {
				t.Fatalf("reopened state = %+v", state)
			}
			if compact && (state.compactTailOffset == 0 || state.compactTailEntries == 0) {
				t.Fatalf("reopened compact state = %+v", state)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("Close(reopened) error = %v", err)
			}
		})
	}
}

func TestDedupeEntryItemsRetainsResolvedState(t *testing.T) {
	first := resolvedDataLinkState{nEntries: 7, entryArrayOffset: 128}
	items := []entryItem{
		{offset: 64, hash: 1, linkState: first},
		{offset: 64, hash: 1, linkState: resolvedDataLinkState{nEntries: 99}},
	}
	items = dedupeEntryItems(items)
	if len(items) != 1 {
		t.Fatalf("deduped item count = %d, want 1", len(items))
	}
	if items[0].linkState != first {
		t.Fatalf("deduped state = %+v, want %+v", items[0].linkState, first)
	}
}

func TestCompactTailFallsBackWithoutMappedArena(t *testing.T) {
	path := filepath.Join(t.TempDir(), "compact-tail-portable-fallback.journal")
	opts := testOptions()
	opts.Compact = true
	w, err := Create(path, opts)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	payload := []byte("MESSAGE=compact-tail-portable-fallback")

	for entry := uint64(1); entry <= 2; entry++ {
		if err := w.AppendRaw([][]byte{payload}, testEntryOptions(entry)); err != nil {
			t.Fatalf("AppendRaw(%d) error = %v", entry, err)
		}
	}
	if err := w.closeArena(); err != nil {
		t.Fatalf("closeArena() error = %v", err)
	}
	if err := w.AppendRaw([][]byte{payload}, testEntryOptions(3)); err != nil {
		t.Fatalf("AppendRaw(fallback) error = %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if got := readJournalSnapshot(t, path).dataByPayload[string(payload)].header.nEntries; got != 3 {
		t.Fatalf("snapshot nEntries = %d, want 3", got)
	}
}

func TestInvalidCompactTailStateFallsBackToAuthoritativeChain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "compact-tail-fallback.journal")
	opts := testOptions()
	opts.Compact = true
	w, err := Create(path, opts)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	payload := []byte("MESSAGE=compact-tail-fallback")

	for entry := uint64(1); entry <= 2; entry++ {
		if err := w.AppendRaw([][]byte{payload}, testEntryOptions(entry)); err != nil {
			t.Fatalf("AppendRaw(%d) error = %v", entry, err)
		}
	}
	dataOffset, originalState, ok, err := w.findData(w.hash(payload), payload)
	if err != nil || !ok {
		t.Fatalf("findData() offset/state/error = %d/%+v/%v", dataOffset, originalState, err)
	}
	if originalState.entryArrayOffset == 0 {
		t.Fatal("promoted DATA entry array is zero")
	}
	if err := w.writeCompactDataTail(dataOffset, 8, 1); err != nil {
		t.Fatalf("writeCompactDataTail(corrupt) error = %v", err)
	}

	if err := w.AppendRaw([][]byte{payload}, testEntryOptions(3)); err != nil {
		t.Fatalf("AppendRaw(fallback) error = %v", err)
	}
	_, repairedState, ok, err := w.findData(w.hash(payload), payload)
	if err != nil || !ok {
		t.Fatalf("findData(repaired) state/error = %+v/%v", repairedState, err)
	}
	if repairedState.nEntries != 3 ||
		repairedState.entryArrayOffset != originalState.entryArrayOffset ||
		repairedState.compactTailOffset != originalState.entryArrayOffset ||
		repairedState.compactTailEntries != 2 {
		t.Fatalf("repaired state = %+v", repairedState)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if got := readJournalSnapshot(t, path).dataByPayload[string(payload)].header.nEntries; got != 3 {
		t.Fatalf("snapshot nEntries = %d, want 3", got)
	}
}
