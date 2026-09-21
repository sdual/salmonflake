package salmonflake

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sdual/salmonflake/config"
	"github.com/sdual/salmonflake/dtype"
)

const (
	sequenceBits = 12
	machineBits  = 10
	maxSequence  = 1<<sequenceBits - 1
	maxMachineID = 1<<machineBits - 1
	maxElapsed   = 1<<41 - 1
)

// Salmonflake is a concurrency-safe Snowflake ID generator.
// A Salmonflake must not be copied after first use. Each active generator
// must have a distinct machine ID and share the same epoch.
type Salmonflake struct {
	mu        sync.Mutex
	start     int64
	elapsed   dtype.Time
	sequence  dtype.Sequence
	machineID dtype.MachineID
	started   bool
	now       func() time.Time
	sleep     func(time.Duration)
}

// New constructs a generator and returns an error if cfg is invalid.
// MachineID must be a decimal integer between 0 and 1023.
func New(cfg config.Config) (*Salmonflake, error) {
	if cfg.Start.IsZero() {
		cfg.Start = config.DefaultStart
	}
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	machineID, err := strconv.ParseUint(cfg.MachineID, 10, machineBits)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MachineID: %w", err)
	}

	return &Salmonflake{
		start:     cfg.Start.UnixMilli(),
		machineID: dtype.MachineID(machineID),
		now:       time.Now,
		sleep:     time.Sleep,
	}, nil
}

func validate(conf config.Config) error {
	now := time.Now()
	if conf.Start.After(now) {
		return fmt.Errorf(
			"the start time must be before the current time. start time: %s",
			conf.Start.Format("2006-01-02T15:04:05Z07:00"),
		)
	}
	if conf.Start.Before(now.Add(-time.Duration(maxElapsed) * time.Millisecond)) {
		return fmt.Errorf("the start time exceeds the 41-bit timestamp range")
	}
	if _, err := strconv.ParseUint(conf.MachineID, 10, machineBits); err != nil {
		return fmt.Errorf("machine ID must be a decimal integer between 0 and %d: %w", maxMachineID, err)
	}
	return nil
}

// NextID generates an ID with a 41-bit millisecond timestamp, a 10-bit machine
// ID, and a 12-bit sequence. It waits for the next millisecond when the sequence
// is exhausted, and returns an error if the clock moves backwards or the
// timestamp no longer fits in 41 bits.
func (s *Salmonflake) NextID() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.now == nil {
		return 0, fmt.Errorf("generator must be initialized with New")
	}
	for {
		timestamp := s.now().UnixMilli()
		if timestamp < s.start || (s.started && timestamp < s.start+int64(s.elapsed)) {
			return 0, fmt.Errorf("clock moved backwards")
		}
		elapsed := uint64(timestamp) - uint64(s.start)
		if elapsed > maxElapsed {
			return 0, fmt.Errorf("timestamp exceeds the 41-bit range")
		}
		if s.started && elapsed == s.elapsed {
			if s.sequence == maxSequence {
				s.sleep(time.Millisecond)
				continue
			}
			s.sequence++
		} else {
			s.sequence = 0
		}
		s.elapsed = elapsed
		s.started = true
		return elapsed<<(machineBits+sequenceBits) |
			uint64(s.machineID)<<sequenceBits | uint64(s.sequence), nil
	}
}
