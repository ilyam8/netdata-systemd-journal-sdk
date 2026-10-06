// livegrowth is a synthetic writer/reader for the deterministic growth matrix.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"github.com/netdata/systemd-journal-sdk/go/journal"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: livegrowth read|write FILE [compact]")
	}
	if os.Args[1] == "read" {
		r, err := journal.OpenFile(os.Args[2])
		if err != nil {
			return err
		}
		defer r.Close()
		var hashes []string
		for {
			ok, err := r.Step()
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			payload, found, err := r.GetEntryPayload([]byte("MESSAGE"))
			if err != nil {
				return err
			}
			if !found {
				return fmt.Errorf("missing MESSAGE")
			}
			hashes = append(hashes, fmt.Sprintf("%x", sha256.Sum256(payload)))
		}
		return json.NewEncoder(os.Stdout).Encode(hashes)
	}
	if os.Args[1] != "write" {
		return fmt.Errorf("unknown operation %q", os.Args[1])
	}
	w, err := journal.Create(os.Args[2], journal.Options{
		MachineID: journal.UUID{1}, BootID: journal.UUID{2}, SeqnumID: journal.UUID{3}, FileID: journal.UUID{4},
		Compact: len(os.Args) == 4 && os.Args[3] == "compact", DataHashTableBuckets: 64, FieldHashTableBuckets: 16,
	})
	if err != nil {
		return err
	}
	defer w.Close()
	ack := bufio.NewScanner(os.Stdin)
	for i, payload := range [][]byte{[]byte("MESSAGE=seed"), append([]byte("MESSAGE="), bytes.Repeat([]byte{'x'}, 9*1024*1024-8)...)} {
		if err := w.AppendRaw([][]byte{payload}, journal.EntryOptions{RealtimeUsec: 1_700_000_000_000_001 + uint64(i), MonotonicUsec: uint64(i + 1)}); err != nil {
			return err
		}
		if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"stage": []string{"seed", "committed"}[i], "entries": i + 1}); err != nil {
			return err
		}
		if !ack.Scan() {
			return fmt.Errorf("missing reader acknowledgement: %v", ack.Err())
		}
	}
	return w.Close()
}
