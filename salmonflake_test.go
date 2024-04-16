package salmonflake

import (
	"testing"

	"github.com/sdual/salmonflake/config"
)

func TestInitializeSalmonflakeWithWrongConf(t *testing.T) {
	conf := config.Config{
		MachineID: "Machine",
	}
	// TODO: check initalization is falied.
	sf, err := New(conf)
	if err != nil {

	}
}
