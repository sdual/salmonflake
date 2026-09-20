package salmonflake

import (
	"testing"

	"github.com/sdual/salmonflake/config"
)

func TestInitializeSalmonflakeWithWrongConf(t *testing.T) {
	conf := config.Config{
		MachineID: "Machine",
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected initialization to panic")
		}
	}()

	_ = New(conf)
}
