package config

import "time"

var DefaultStart = time.Date(2014, 9, 1, 0, 0, 0, 0, time.UTC)

const (
	TimeBitLen     = 41 // Bit length of time.
	SequenceBitLne = 12 // Bit length of sequence number.
	MachineBitLen  = 10 // Bit length of machine id.
)

// Config is a config struct for ID generator.
type Config struct {
	Start     time.Time
	MachineID string
}
