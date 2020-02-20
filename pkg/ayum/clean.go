package ayum

import (
	"context"

	"github.com/brinick/shell"
)

type cleaner interface {
	CleanAll(context.Context, string) error
}

type cmdClean struct {
	timeout int
	runner  ayumCmdRunner
}

// CleanAll runs an ayum clean all on the given repository
func (c *cmdClean) CleanAll(ctx context.Context, repoName string) error {
	err := c.runner.Run(shell.Context(ctx))

	// TODO: check error and log messages, return correctly
	return err
}
