package config

import "time"

// DefaultStart is the epoch used when Config.Start is zero.
var DefaultStart = time.Date(2014, 9, 1, 0, 0, 0, 0, time.UTC)

// Config is a config struct for ID generator.
type Config struct {
	// Start is the shared epoch, with millisecond precision (zero uses DefaultStart).
	Start time.Time
	// MachineID is a decimal integer from 0 to 1023, unique to each active generator.
	MachineID string
}
