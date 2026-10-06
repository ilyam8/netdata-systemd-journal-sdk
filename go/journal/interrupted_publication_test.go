package journal

import (
	"context"
	"path/filepath"
	"testing"
)

// Model a process stopping between posting and header publication. Closing the
// arena directly preserves the synthetic interrupted state without clean-close
// metadata. Reopening alone is intentionally not a graph-certification step.
func TestVerifyIndexRejectsInterruptedPublicationAfterReopen(t *testing.T) {
	for _, stage := range []string{"postings-before-header", "tail-before-count"} {
		t.Run(stage, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "interrupted.journal")
			options := Options{Compact: true, MachineID: UUID{1}, BootID: UUID{2}}
			writer, err := Create(path, options)
			if err != nil {
				t.Fatal(err)
			}
			appendRow := func(sequence uint64) {
				t.Helper()
				if err := writer.Append([]Field{StringField("MESSAGE", "repeated")}, EntryOptions{
					RealtimeUsec: 1_700_000_000_000_000 + sequence, MonotonicUsec: sequence,
				}); err != nil {
					t.Fatal(err)
				}
			}
			for sequence := uint64(1); sequence <= 5; sequence++ {
				appendRow(sequence)
			}
			previous := writer.header
			appendRow(6)
			if stage == "postings-before-header" {
				header := writer.header
				header.nEntries = previous.nEntries
				header.tailEntryOffset = previous.tailEntryOffset
				header.tailEntrySeqnum = previous.tailEntrySeqnum
				header.tailEntryRealtime = previous.tailEntryRealtime
				header.tailEntryMonotonic = previous.tailEntryMonotonic
				header.tailEntryArrayOffset = previous.tailEntryArrayOffset
				header.tailEntryArrayNEntries = previous.tailEntryArrayNEntries
				buffer := make([]byte, headerSize)
				putHeader(buffer, header)
				if err := writer.writeAt(0, buffer); err != nil {
					t.Fatal(err)
				}
			} else if err := writer.writeUint64At(152, previous.nEntries); err != nil {
				t.Fatal(err)
			}
			if err := writer.closeArena(); err != nil {
				t.Fatal(err)
			}
			if err := writer.file.Close(); err != nil {
				t.Fatal(err)
			}
			if err := VerifyIndex(context.Background(), path); err == nil {
				t.Fatal("interrupted publication passed strict verification")
			}

			// Demonstrate why callers must verify before reuse. A later successful
			// append/close does not repair orphan postings from an earlier process.
			writer, err = OpenWithOptions(path, options)
			if err != nil {
				t.Fatal(err)
			}
			appendRow(7)
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			if err := VerifyIndex(context.Background(), path); err == nil {
				t.Fatal("reopen and append hid the interrupted publication")
			}
		})
	}
}
