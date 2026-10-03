package journal

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCloseWithoutRetentionPreservesExpiredArchive(t *testing.T) {
	for _, systemdNaming := range []bool{false, true} {
		name := "chain"
		if systemdNaming {
			name = "systemd"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			config := LogConfig{Source: "system", Options: testOptions(), StrictSystemdNaming: systemdNaming}
			initial, err := NewLog(root, config)
			if err != nil {
				t.Fatal(err)
			}
			saved := time.Now()
			opts := EntryOptions{RealtimeUsec: uint64(saved.UnixMicro()), MonotonicUsec: 1}
			if err := initial.Append([]Field{StringField("MESSAGE", "older")}, opts); err != nil {
				t.Fatal(err)
			}
			if err := initial.Close(); err != nil {
				t.Fatal(err)
			}
			config.OpenMode = LogOpenEager
			config.RetentionPolicy = RetentionPolicy{}.WithMaxAge(time.Second).WithMaxBytes(1 << 30)
			log, err := NewLog(root, config)
			if err != nil {
				t.Fatal(err)
			}
			if len(journalFiles(t, log.JournalDirectory())) != 2 {
				t.Fatal("old archive should survive before its expiry")
			}
			opts.MonotonicUsec = 2
			if err := log.Append([]Field{StringField("MESSAGE", "current")}, opts); err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Until(saved.Add(time.Second + 50*time.Millisecond)))
			if err := log.CloseWithoutRetention(); err != nil {
				t.Fatal(err)
			}
			if err := log.CloseWithoutRetention(); err != nil {
				t.Fatal(err)
			}
			if err := log.Close(); err != nil {
				t.Fatal(err)
			}
			files := journalFiles(t, log.JournalDirectory())
			if len(files) != 2 {
				t.Fatalf("retained archives = %d, want 2", len(files))
			}
			for _, path := range files {
				snapshot := readJournalSnapshot(t, path)
				if snapshot.header.state != stateArchived {
					t.Fatalf("state = %d, want archived", snapshot.header.state)
				}
			}
			config.RetentionPolicy = RetentionPolicy{}.WithMaxAge(time.Hour).WithMaxBytes(1 << 30)
			relaxed, err := NewLog(root, config)
			if err != nil {
				t.Fatal(err)
			}
			if len(journalFiles(t, relaxed.JournalDirectory())) != 3 {
				t.Fatal("relaxed policy should preserve both archives")
			}
			if err := relaxed.CloseWithoutRetention(); err != nil {
				t.Fatal(err)
			}
			config.RetentionPolicy = RetentionPolicy{}.WithMaxBytes(1)
			tightened, err := NewLog(root, config)
			if err != nil {
				t.Fatal(err)
			}
			if len(journalFiles(t, tightened.JournalDirectory())) != 1 {
				t.Fatal("new policy should remove archives and protect fresh active")
			}
			if err := tightened.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCloseWithoutRetentionSkipsArtifactAccounting(t *testing.T) {
	failAccounting := false
	accounting := LogArtifactSizeFunc(func(string) (uint64, error) {
		if failAccounting {
			return 0, errors.New("retention callback must not run")
		}
		return 0, nil
	})
	log, dir := newTestLog(t, LogConfig{Source: "system", Options: testOptions(),
		RetentionPolicy: RetentionPolicy{}.WithMaxBytes(1), ArtifactSizer: accounting})
	if err := log.Append([]Field{StringField("MESSAGE", "retained")}, EntryOptions{RealtimeUsec: 1_700_000_000_000_000, MonotonicUsec: 1}); err != nil {
		t.Fatal(err)
	}
	// A retention-only failure must not stop the archive/release variant.
	failAccounting = true
	if err := log.CloseWithoutRetention(); err != nil {
		t.Fatal(err)
	}
	if len(journalFiles(t, dir)) != 1 {
		t.Fatal("archive should remain readable")
	}
}

func TestCloseWithoutRetentionEmptyAndLazy(t *testing.T) {
	for _, eager := range []bool{false, true} {
		mode := LogOpenLazy
		if eager {
			mode = LogOpenEager
		}
		log, dir := newTestLog(t, LogConfig{Source: "system", Options: testOptions(), OpenMode: mode})
		if err := log.CloseWithoutRetention(); err != nil {
			t.Fatal(err)
		}
		if err := log.CloseWithoutRetention(); err != nil {
			t.Fatal(err)
		}
		wantFiles := 0
		if eager {
			wantFiles = 1
		} // Chain naming preserves the existing empty-archive behavior.
		if got := len(journalFiles(t, dir)); got != wantFiles {
			t.Fatalf("empty/lazy archive count = %d, want %d", got, wantFiles)
		}
		if err := log.Append([]Field{StringField("MESSAGE", "after close")}, EntryOptions{MonotonicUsec: 1}); err == nil {
			t.Fatal("append should reject a closed log")
		}
	}
}

func TestCloseWithoutRetentionReportsArchiveFailure(t *testing.T) {
	log, dir := newTestLog(t, LogConfig{Source: "system", Options: testOptions(), StrictSystemdNaming: true})
	if err := log.Append([]Field{StringField("MESSAGE", "retained")}, EntryOptions{RealtimeUsec: 1_700_000_000_000_000, MonotonicUsec: 1}); err != nil {
		t.Fatal(err)
	}
	moved := filepath.Join(filepath.Dir(dir), "moved-machine")
	if err := os.Rename(dir, moved); err != nil {
		t.Fatal(err)
	}
	closeErr := log.CloseWithoutRetention()
	// Restore our own fixture before assertions and release the surviving writer.
	if err := os.Rename(moved, dir); err != nil {
		t.Fatal(err)
	}
	if closeErr == nil {
		t.Fatal("archive publication failure must be returned")
	}
	if err := log.CloseWithoutRetention(); err != nil {
		t.Fatal(err)
	}
}
