package ayum

import (
	"context"
	"fmt"

	"github.com/brinick/logging"
	"github.com/brinick/shell"
)

type cleaner interface {
	CleanAll(context.Context, string) error
}

type cmdClean struct {
	// timeout int
	log logging.Logger
	cmd *ayumCommand
}

// CleanAll runs an ayum clean all on the repository of the given name
func (c *cmdClean) CleanAll(ctx context.Context, name string) error {
	// Update the command string
	c.cmd.cmd = fmt.Sprintf(c.cmd.cmd, name)

	// Run the command
	c.cmd.Run(shell.Context(ctx))
	return doPostMortem(c.cmd, c.log)
}
