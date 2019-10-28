package config

import (
	"flag"
	"fmt"
	"path/filepath"
)

// New creates a new Config instance
func New() (*Config, error) {
	var c Config
	c.instantiate()
	c.flags()
	err := c.parse()
	return &c, err
}

// Config is the full set of available command line args
type Config struct {
	Admin   *adminOpts
	Ayum    *ayumOpts
	CVMFS   *cvmfsOpts
	Dirs    *dirsOpts
	EOS     *eosOpts
	Global  *globalOpts
	Install *installOpts
}

// Dump returns the config as a string representation
func (c *Config) Dump() string {
	// TODO: write this
	return ""
}

// parse parses and validates the command line,
// returning an error if appropriate.
func (c *Config) parse() error {
	flag.Parse()
	c.postConfig()
	return c.validate()
}

// instantiate initialises the config member structs
func (c *Config) instantiate() {
	c.Admin = &adminOpts{}
	c.Ayum = &ayumOpts{}
	c.CVMFS = &cvmfsOpts{}
	c.Dirs = &dirsOpts{}
	c.EOS = &eosOpts{}
	c.Global = &globalOpts{}
	c.Install = &installOpts{}
}

// flags defines the CLI flags for this config
func (c *Config) flags() {
	c.Admin.flags()
	c.Ayum.flags()
	c.CVMFS.flags()
	c.Dirs.flags()
	c.EOS.flags()
	c.Global.flags()
	c.Install.flags()
}

// postConfig adapts some variables that depend on others
func (c *Config) postConfig() error {
	if c.Dirs.InstallBase == "" {
		// Nothing was given, so we define the base install dir
		c.Dirs.InstallBase = fmt.Sprintf("/cvmfs/%s/repo/sw", c.CVMFS.NightlyRepo)
	}

	if c.Dirs.RPMSrcBase == "" {
		c.Dirs.RPMSrcBase = filepath.Join(c.EOS.NightlyBaseDir, c.Install.Release)
	}

	if c.Ayum.InstallDir == "" {
		c.Ayum.InstallDir = filepath.Join(c.Dirs.InstallBase, c.Install.Branch)
	}

	if c.Ayum.AyumDir == "" {
		c.Ayum.AyumDir = c.Dirs.WorkBase
	}

	if err := c.ensureAbsPaths(); err != nil {
		return err
	}

	// Logs dir should sit in the work base directory,
	// so we ensure that here
	c.Dirs.Logs = filepath.Join(c.Dirs.WorkBase, "logs")
	return nil
}

func (c *Config) ensureAbsPaths() error {
	// Make sure we are dealing with absolute paths
	var err error
	c.Dirs.InstallBase, err = filepath.Abs(c.Dirs.InstallBase)
	if err != nil {
		return err
	}

	c.Ayum.InstallDir, err = filepath.Abs(c.Ayum.InstallDir)
	if err != nil {
		return err
	}
	c.Dirs.RPMSrcBase, err = filepath.Abs(c.Dirs.RPMSrcBase)
	if err != nil {
		return err
	}

	c.Dirs.WorkBase, err = filepath.Abs(c.Dirs.WorkBase)
	if err != nil {
		return err
	}

	return nil
}

// Validate validates the parsed args
func (c *Config) validate() error {
	type validateFn func() error
	for _, fn := range []validateFn{
		c.Admin.validate,
		c.Ayum.validate,
		c.CVMFS.validate,
		c.Dirs.validate,
		c.EOS.validate,
		c.Global.validate,
		c.Install.validate,
	} {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}
