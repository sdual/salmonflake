package salmonflake

import (
	"fmt"
	"time"

	"github.com/sdual/salmonflake/dtype"
)

// Salmonflake is a distributed parallel Snowflake ID generator.
type Salmonflake struct {
	startTime   dtype.Time
	elapsedTime dtype.Time
	sequence    dtype.Sequence
	machineID   dtype.MachineID
}

type Config struct {
	startTime time.Time
	MachineID string
}

func New(conf Config) {
	if err := validate(conf), err != nil {
		panic("initialization error")
	}
}

func validate(conf Config) error {
	if conf.startTime.After(time.Now()) {
		return fmt.Errorf("startTime must be before the current time. startTime: %w", conf.startTime)
	}
	return nil
}

// NextID generates a next unique ID.
func (s *Salmonflake) NextID() (uint64, error) {
	return 0, nil
}
