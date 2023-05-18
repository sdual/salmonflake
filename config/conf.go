package config

import "time"

var DefaultStart = time.Date(2014, 9, 1, 0, 0, 0, 0, time.UTC)

// Config is a config struct for ID generator.
type Config struct {
	Start     time.Time
	MachineID string
}
