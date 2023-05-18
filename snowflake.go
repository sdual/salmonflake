package salmonflake

import (
	"github.com/sdual/salmonflake/dtype"
)

// Salmonflake is a distributed parallel Snowflake ID generator.
type Salmonflake struct {
	startTime   dtype.Time
	elapsedTime dtype.Time
	sequence    dtype.Sequence
	machineID   dtype.MachineID
}

// NextID generates a next unique ID.
func (s *Salmonflake) NextID() (uint64, error) {
	return 0, nil
}
