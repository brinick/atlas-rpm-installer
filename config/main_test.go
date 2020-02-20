package config

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	os.Args = []string{
		"installer",
		"--global.timeout", "5",
		"-cvmfs.max-transaction-attempts", "12",
	}
	c, _ := New()
	if c.Global.TimeOut != 5 {
		t.Errorf("global timeout is %ds, expected 5s", c.Global.TimeOut)
	}

	if c.CVMFS.MaxTransactionAttempts != 12 {
		t.Errorf("cvmfs max transition attempts is %d, expected 12", c.CVMFS.MaxTransactionAttempts)
	}
}
