package ayum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/brinick/fs"
	"github.com/brinick/shell"
)

type configurer interface {
	PreConfigure(string) error
	Configure(context.Context) error
}

type cmdConfigure struct {
	installDir string
	runner     ayumCmdRunner
}

// PreConfigure will copy, for cache nightly installations,
// the stable base release .rmpdb directory to the install directory
// to allow dependencies to be found.
func (c *cmdConfigure) PreConfigure(stableRelBase string) error {
	branch := filepath.Base(c.installDir)
	isCacheNightly := (strings.Count(branch, ".")) > 2
	if !isCacheNightly {
		return nil
	}

	tokens := strings.Split(branch, ".")
	baseRelease := strings.Join(tokens[:2], ".") // e.g. 21.2

	stableRelSrc := filepath.Join(stableRelBase, baseRelease)
	exists, err := fs.Exists(stableRelSrc)
	if err != nil {
		return fmt.Errorf("Unable to check existance of dir %s (%w)", stableRelSrc, err)
	}

	if !exists {
		return fmt.Errorf("%s: stable release dir does not exist", stableRelSrc)
	}

	dst := filepath.Join(c.installDir, ".rpmdb")
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("unable to remove directory tree %s (%w)", dst, err)
	}

	return fs.Dir(stableRelSrc, ".rpmdb").CopyTo(dst)
}

// Configure configures the yum.conf file with the given install directory path
func (c *cmdConfigure) Configure(ctx context.Context) error {
	err := c.runner.Run(shell.Context(ctx))

	// TODO: add logging, check errors
	return err
}
