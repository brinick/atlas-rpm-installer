package ayum

import (
	"context"
	"fmt"
	"time"

	"github.com/brinick/shell"
)

func configureCommand(configureExe, installDir, yumConf string) string {
	return fmt.Sprintf(
		"%s -i %s -D | grep -v 'AYUM package location' > %s",
		configureExe,
		installDir,
		yumConf,
	)
}

type configurer interface {
	Configure(context.Context) error
}

type cmdConfigure struct {
	cmd     []string
	timeout int
}

// Configure configures the yum.conf file with the given install directory path
func (c *cmdConfigure) Configure(ctx context.Context) error {
	cmd := ayumCommand{
		label: "ayum configure",
		cmd:   c.cmd,
		opts: []shell.Option{
			shell.Context(ctx),
			shell.Timeout(time.Duration(c.timeout) * time.Second),
		},
	}

	cmd.run()

	// TODO: add logging, check errors
	return nil
}
