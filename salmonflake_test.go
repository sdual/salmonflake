package salmonflake

import (
	"sync"
	"testing"
	"time"

	"github.com/sdual/salmonflake/config"
)

func TestNewInvalidConfig(t *testing.T) {
	for _, tc := range []struct {
		name string
		conf config.Config
	}{
		{"empty machine", config.Config{}},
		{"non-numeric machine", config.Config{MachineID: "Machine"}},
		{"negative machine", config.Config{MachineID: "-1"}},
		{"machine overflow", config.Config{MachineID: "1024"}},
		{"future epoch", config.Config{MachineID: "0", Start: time.Now().Add(time.Hour)}},
		{"expired epoch", config.Config{MachineID: "0", Start: time.Now().Add(-time.Duration(maxElapsed+1000) * time.Millisecond)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := New(tc.conf)
			if err == nil {
				t.Fatal("expected initialization error")
			}
			if s != nil {
				t.Fatal("expected nil generator for invalid config")
			}
		})
	}
}

func TestNewDefaults(t *testing.T) {
	s, err := New(config.Config{MachineID: "1023"})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if s.start != config.DefaultStart.UnixMilli() || s.machineID != 1023 {
		t.Fatalf("unexpected initialization: start=%d machine=%d", s.start, s.machineID)
	}
}

func TestNextIDLayoutAndSequence(t *testing.T) {
	epoch := time.Now().Add(-time.Hour).Truncate(time.Millisecond)
	s, err := New(config.Config{Start: epoch, MachineID: "42"})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	now := epoch.Add(123 * time.Millisecond)
	s.now = func() time.Time { return now }
	for sequence := uint64(0); sequence < 3; sequence++ {
		id, err := s.NextID()
		want := uint64(123)<<22 | uint64(42)<<12 | sequence
		if err != nil || id != want {
			t.Fatalf("NextID() = %d, %v; want %d", id, err, want)
		}
	}
	now = now.Add(time.Millisecond)
	id, err := s.NextID()
	if want := uint64(124)<<22 | uint64(42)<<12; err != nil || id != want {
		t.Fatalf("NextID() after advancing clock = %d, %v; want %d", id, err, want)
	}
}

func TestNextIDSequenceExhaustion(t *testing.T) {
	epoch := time.Now().Add(-time.Hour).Truncate(time.Millisecond)
	s, err := New(config.Config{Start: epoch, MachineID: "0"})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	now := epoch
	s.now = func() time.Time { return now }
	waits := 0
	s.sleep = func(d time.Duration) {
		waits++
		now = now.Add(d)
	}
	for want := uint64(0); want < 4096; want++ {
		id, err := s.NextID()
		if err != nil || id != want {
			t.Fatalf("NextID() = %d, %v; want %d", id, err, want)
		}
	}
	id, err := s.NextID()
	if err != nil || id != 1<<22 || waits != 1 {
		t.Fatalf("NextID() at exhaustion = %d, %v; waits=%d", id, err, waits)
	}
}

func TestNextIDClockRollback(t *testing.T) {
	epoch := time.Now().Add(-time.Hour).Truncate(time.Millisecond)
	s, err := New(config.Config{Start: epoch, MachineID: "1"})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	now := epoch.Add(time.Second)
	s.now = func() time.Time { return now }
	first, err := s.NextID()
	if err != nil {
		t.Fatalf("NextID() failed: %v", err)
	}
	for _, backwards := range []time.Time{now.Add(-time.Millisecond), epoch.Add(-time.Millisecond)} {
		now = backwards
		if _, err := s.NextID(); err == nil {
			t.Fatal("expected clock rollback error")
		}
	}
	now = epoch.Add(time.Second)
	id, err := s.NextID()
	if err != nil || id != first+1 {
		t.Fatalf("NextID() after recovery = %d, %v; want %d", id, err, first+1)
	}
}

func TestNextIDTimestampLimit(t *testing.T) {
	epoch := time.Now().Add(-time.Hour).Truncate(time.Millisecond)
	s, err := New(config.Config{Start: epoch, MachineID: "1023"})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	now := epoch.Add(time.Duration(maxElapsed) * time.Millisecond)
	s.now = func() time.Time { return now }
	var id uint64
	for i := 0; i < 4096; i++ {
		var err error
		id, err = s.NextID()
		if err != nil {
			t.Fatalf("NextID() at sequence %d failed: %v", i, err)
		}
	}
	if id != 1<<63-1 {
		t.Fatalf("last possible ID = %d; want %d", id, uint64(1<<63-1))
	}
	s.sleep = func(d time.Duration) { now = now.Add(d) }
	if _, err := s.NextID(); err == nil {
		t.Fatal("expected timestamp overflow error")
	}
}

func TestNextIDDistinctMachines(t *testing.T) {
	epoch := time.Now().Add(-time.Hour)
	ids := make(map[uint64]bool)
	for _, machine := range []string{"0", "1", "1023"} {
		s, err := New(config.Config{Start: epoch, MachineID: machine})
		if err != nil {
			t.Fatalf("New() for machine %s failed: %v", machine, err)
		}
		s.now = func() time.Time { return epoch.Add(time.Second) }
		id, err := s.NextID()
		if err != nil || ids[id] {
			t.Fatalf("duplicate or invalid ID for machine %s: %d, %v", machine, id, err)
		}
		ids[id] = true
	}
}

func TestNextIDConcurrent(t *testing.T) {
	const workers, count = 16, 1000
	s, err := New(config.Config{MachineID: "1"})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	ids := make(chan uint64, workers*count)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var previous uint64
			for j := 0; j < count; j++ {
				id, err := s.NextID()
				if err != nil {
					t.Errorf("NextID() at iteration %d failed: %v", j, err)
					return
				}
				if j > 0 && id <= previous {
					t.Errorf("IDs are not increasing: %d <= %d", id, previous)
				}
				previous = id
				ids <- id
			}
		}()
	}
	wg.Wait()
	close(ids)
	seen := make(map[uint64]bool, workers*count)
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate ID: %d", id)
		}
		seen[id] = true
	}
	if len(seen) != workers*count {
		t.Fatalf("generated %d IDs; want %d", len(seen), workers*count)
	}
}

func TestNextIDUninitialized(t *testing.T) {
	var s Salmonflake
	if _, err := s.NextID(); err == nil {
		t.Fatal("expected error for uninitialized generator")
	}
}
