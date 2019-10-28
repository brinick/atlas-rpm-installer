package ayum

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/brinick/fs"
	"github.com/brinick/logging"
)

// New creates a new Ayum instance
func New(opts *Opts, log logging.Logger) *Ayum {
	binary := filepath.Join(opts.AyumDir, "ayum/ayum")
	preCmds := ayumEnv(opts.AyumDir)
	postCmds := []string{}

	configureExe := filepath.Join(opts.AyumDir, "configure.ayum")
	yumConf := filepath.Join(opts.AyumDir, "yum.conf")

	wrap := wrapCommand(preCmds, postCmds)

	a := &Ayum{
		Dir:        opts.AyumDir,
		Binary:     binary,
		InstallDir: opts.InstallDir,
		Log:        log,
		downloader: &cmdDownload{
			srcRepo: opts.SrcRepo,
			timeout: opts.DownloadTimeout,
		},
		configurer: &cmdConfigure{
			cmd:     wrap(configureCommand(configureExe, opts.InstallDir, yumConf)),
			timeout: opts.Timeout,
		},
		installer: &cmdInstall{
			lister: &cmdList{
				cmd:     wrap(listCommand(binary)),
				timeout: opts.Timeout,
				log:     log,
			},
			ayumExe: binary,
			preCmds: preCmds,
			timeout: opts.InstallTimeout,
			log:     log,
		},
		cleaner: &cmdClean{
			ayumExe: binary,
			preCmds: preCmds,
			timeout: opts.Timeout,
		},
	}

	return a
}

// Opts configures the ayum instance
type Opts struct {
	SrcRepo    string
	AyumDir    string
	InstallDir string

	// Timeout is the general maximum number of seconds allowed
	// to perform an ayum command
	Timeout int

	// DownloadTimeout is the maximum number of seconds allowed
	// to clone locally the ayum source git repo
	DownloadTimeout int

	// InstallTimeout is the maximum number of seconds allowed
	// in the install attempt
	InstallTimeout int
}

// Ayum is the ayum wrapper
type Ayum struct {
	downloader
	configurer
	cleaner
	installer

	// Binary is the path to the ayum executable
	Binary string

	// Dir is the root directory of the ayum installation
	Dir string

	// InstallDir is the root dir of the install path
	InstallDir string

	// Log is a logger instance
	Log logging.Logger
}

// PreConfigure will copy, for cache nightly installations,
// the stable base release .rmpdb directory to the install directory
// to allow dependencies to be found.
func (a *Ayum) PreConfigure(stableRelBase string) error {
	branch := filepath.Base(a.InstallDir)
	isCacheNightly := (strings.Count(branch, ".")) > 2
	if !isCacheNightly {
		return nil
	}

	tokens := strings.Split(branch, ".")
	baseRelease := strings.Join(tokens[:2], ".") // e.g. 21.2

	stableRelSrc := filepath.Join(stableRelBase, baseRelease)
	// stableRelSrc := fs.Directory(stableRelBase).Append(baseRelease)
	exists, err := fs.Exists(stableRelSrc)
	if err != nil {
		return fmt.Errorf("Unable to check existance of dir %s (%w)", stableRelSrc, err)
	}

	if !exists {
		return fmt.Errorf("%s: stable release dir does not exist", stableRelSrc)
	}

	dst := filepath.Join(a.InstallDir, ".rpmdb")
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("unable to remove directory tree %s (%w)", dst, err)
	}

	src := fs.Dir(stableRelSrc, ".rpmdb")
	return src.CopyTo(dst)
}

type repoer interface {
	Filename() string
	String() string
}

// AddRemoteRepos configures the ayum installation with the provided remote repositories
// func (a *Ayum) AddRemoteRepos(repos []*rpm.Repo) error {
func (a *Ayum) AddRemoteRepos(repos []repoer) error {
	for _, repo := range repos {
		repoConf := filepath.Join(a.Dir, "ayum/etc/yum.repos.d", repo.Filename())
		if err := ioutil.WriteFile(repoConf, []byte(repo.String()), 0774); err != nil {
			return fmt.Errorf("could not configure remote repo %s (%w)", repo.Filename(), err)
		}
	}

	return nil
}
