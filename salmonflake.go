package salmonflake

import (
	"fmt"
	"time"

	"github.com/sdual/salmonflake/config"
	"github.com/sdual/salmonflake/dtype"
)

// Salmonflake is a distributed parallel Snowflake ID generator.
type Salmonflake struct {
	start     dtype.Time
	elapsed   dtype.Time
	sequence  dtype.Sequence
	machineID dtype.MachineID
}

func New(conf config.Config) Salmonflake {
	if err := validate(conf); err != nil {
		panic("initialization error")
	}

	if conf.Start.IsZero() {
		conf.Start = config.DefaultStart
	}

	// TODO: set parameters.
	return Salmonflake{
		start:     uint64(conf.Start.Unix()),
		elapsed:   0,
		sequence:  0,
		machineID: 0,
	}
}

func validate(conf config.Config) error {
	if conf.Start.After(time.Now()) {
		return fmt.Errorf(
			"the start time must be before the current time. start time: %s",
			conf.Start.Format("2006-01-02T15:04:05Z07:00"),
		)
	}
	return nil
}

// NextID generates a next unique ID.
func (s *Salmonflake) NextID() (uint64, error) {
	// TODO: generates a next unique ID.
	return 0, nil
}
