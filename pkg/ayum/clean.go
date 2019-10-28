package ayum

import (
	"context"
	"fmt"
	"time"

	"github.com/brinick/shell"
)

type cleaner interface {
	CleanAll(context.Context, string) error
}

type cmdClean struct {
	ayumExe  string
	preCmds  []string
	postCmds []string
	timeout  int
}

// CleanAll runs an ayum clean all on the given repository
func (c *cmdClean) CleanAll(ctx context.Context, repoName string) error {
	cmd := append(
		c.preCmds,
		append(
			[]string{
				fmt.Sprintf("%s --enablerepo=%s clean all", c.ayumExe, repoName),
			},
			c.postCmds...,
		)...,
	)

	ac := ayumCommand{
		cmd: cmd,
		opts: []shell.Option{
			shell.Context(ctx),
			shell.Timeout(time.Duration(c.timeout) * time.Second),
		},
	}

	ac.run()

	// TODO: check error and log messages, return correctly
	return nil
}
