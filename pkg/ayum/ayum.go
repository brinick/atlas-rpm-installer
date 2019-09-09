package ayum

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"

	"github.com/brinick/logging"
	"github.com/brinick/shell"
)

// Package is an RPM package
type Package struct {
	Name    string
	Version string
}

// Opts configures the ayum instance
type Opts struct {
	Repo            string
	Dir             string
	GitCloneTimeout int
}

// New creates a new Ayum instance
func New(log logging.Logger, opts *Opts) *Ayum {
	// func New(srcRepo, localdir string, log logging.Logger) *Ayum {
	a := &Ayum{
		SrcRepo: opts.Repo,
		Dir:     opts.Dir,
		Log:     log,
	}

	a.Binary = filepath.Join(a.Dir, "ayum/ayum")
	return a
}

// Ayum is the ayum wrapper
type Ayum struct {
	SrcRepo string
	Dir     string
	Binary  string
	Log     logging.Logger
}

// Download clones the ayum source git repository.
// The timeout parameter is the maximum duration that
// this operation may take before interruption.
// If set to <= 0, no timeout is applied.
func (a *Ayum) Download(ctx context.Context, timeout time.Duration) error {
	if timeout > 0 {
		var cancelFn context.CancelFunc
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
		defer cancelFn()
	}

	dir := filepath.Join(filepath.Dir(a.Dir), "ayum")
	os.RemoveAll(dir)

	tgtDir := a.Dir
	isBare := false
	opts := &git.CloneOptions{URL: a.SrcRepo}
	_, err := git.PlainCloneContext(ctx, tgtDir, isBare, opts)

	switch err {
	case nil:
		return nil
	case context.DeadlineExceeded:
		err = fmt.Errorf("ayum repo git clone took too long and was killed")
	case context.Canceled:
		err = fmt.Errorf("ayum repo git clone is cancelled")
	default:
		err = errors.Wrap(err, "ayum repo git clone download failed")
	}

	return err
}

// PreConfigure will copy, for cache nightly installations,
// the stable base release .rmpdb directory to the install directory
// to allow dependencies to be found.
func (a *Ayum) PreConfigure(installdir, stableRelBase string) error {
	branch := filepath.Base(installdir)
	isCacheNightly := (strings.Count(branch, ".")) > 2
	if !isCacheNightly {
		return nil
	}

	tokens := strings.Split(branch, ".")
	baseRelease := strings.Join(tokens[:2], ".") // e.g. 21.2

	stableRelSrc := filepath.Join(stableRelBase, baseRelease)
	exists, err := pathExists(stableRelSrc)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%s: stable release dir does not exist", stableRelSrc)
	}

	src := filepath.Join(stableRelSrc, ".rpmdb")
	dst := filepath.Join(installdir, ".rpmdb")
	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	return copyDir(src, dst)
}

type repository interface {
	Filename() string
	String() string
}

// AddRemoteRepos configures the ayum installation with the provided remote repositories
func (a *Ayum) AddRemoteRepos(repos []repository) error {
	for _, repo := range repos {
		repoConf := filepath.Join(a.Dir, "ayum/etc/yum.repos.d", repo.Filename())
		if err := ioutil.WriteFile(repoConf, []byte(repo.String()), 0774); err != nil {
			return errors.Wrapf(err, "could not configure remote repo %s", repo.Filename())
		}
	}

	return nil
}

// Configure configures the yum.conf file with the given install directory path
func (a *Ayum) Configure(installdir string) error {
	cfgExe := filepath.Join(a.Dir, "configure.ayum")
	yumConf := filepath.Join(a.Dir, "yum.conf")

	// Remove the "AYUM package location" line before
	// redirecting to the yum.conf file
	cmd := fmt.Sprintf(
		"%s -i %s -D | grep -v 'AYUM package location' > %s",
		cfgExe,
		installdir,
		yumConf,
	)

	result := shell.Run(cmd)
	// TODO: examine result...
	return result.Error()
}

// CleanAll runs an ayum clean all on the given repository
func (a *Ayum) CleanAll(repoName string) error {
	cmds := []string{
		a.configureShellCmds(),
		fmt.Sprintf("%s --enablerepo=%s clean all", a.Binary, repoName),
	}

	delay := 60
	result := shell.Run(strings.Join(cmds, ";"), shell.Timeout(time.Duration(delay)))
	if result.TimedOut {
		return fmt.Errorf("ayum clean all command failed to complete within %ds", delay)
	}
	return nil
}

type rpm string

func (r *rpm) removeFileExt() {
	*r = rpm(strings.TrimSuffix(string(*r), filepath.Ext(string(*r))))
}

type rpms []string

func (r *rpms) removeFileExt() {
	// Remove the .rpm file extension
	for i, el := range *r {
		*r[i] = strings.TrimSuffix(el, filepath.Ext(el))
	}
}

// Install will install the provided RPMs. Firstly, establish if
// any are already installed, and if so run an ayum reinstall.
// Otherwise just plain install. If any install or reinstall exits
// with non-zero exitcode, stop and return an error.
func (a *Ayum) Install(ctx context.Context, rpmsToInstall ...string) error {
	if len(rpmsToInstall) == 0 {
		return nil
	}

	r := rpms{rpmsToInstall}

	installed, err := a.listInstalledPackages()
	if err != nil {
		return errors.Wrap(err, "unable to install rpms, failed to retrieve installed packages")
	}

	type installFn func(...string) error

	var toReinstall, toInstall []string
	for _, rpm := range rpms {

	}

	return nil
}

func (a *Ayum) installRPMs(rpms ...string) error {
	return nil
}

func (a *Ayum) reinstallRPMs(rpms ...string) error {
	return nil
}

// configureShellCmds returns the commands to execute prior
// to any ayum commands, so that the environement is correctly
// configured
func (a *Ayum) configureShellCmds() string {
	cmds := []string{
		"cd %s",
		"shopt -s expand_aliases",
		"source ayum/setup.sh",
	}
	return strings.Join(cmds, ";")
}

// Split an input list of RPM names into two: one list
// containing RPMS already locally installed, the other list for
// RPMS not yet installed.
func (a *Ayum) categoriseByInstallStatus(rpms ...string) error {
	packages, err := a.listInstalledPackages()
	if err != nil {

	}

	if packages == nil {
		// No local installed packages

	}
}

func (a *Ayum) listInstalledPackages() ([]*Package, error) {
	timeout := 60 // seconds
	result := a._execListInstalledCmd(timeout)

	// Check for problems...
	if result.TimedOut {
		return nil, fmt.Errorf("ayum list installed: command timed out (%ds)", timeout)
	}

	exitcode := result.ExitCode()

	// No problems found...
	if exitcode == 0 {
		packages := a._parseInstalledPackages(result.Stdout().Text())
		return packages, nil
	}

	if result.Stdout().Text() == result.Stderr().Text() {
		a.Log.Info("No local packages installed")
		return nil, nil
	}

	a.Log.Error(
		"Unable to retrieve locally installed package list",
		logging.F("ec", result.ExitCode()),
	)

	for _, line := range result.Stderr().Lines() {
		a.Log.Error(line)
	}

	return nil, fmt.Errorf("ayum list installed - command failed")
}

func (a *Ayum) _parseInstalledPackages(packagesText string) []*Package {
	lines := strings.Split(packagesText, "installed")

	var packages []*Package
	for _, line := range lines {
		tokens := strings.Fields(line)
		if len(tokens) != 2 {
			a.Log.Info("ayum list installed - skipping unexpected line", logging.F("l", line))
			continue
		}

		name, version := tokens[0], tokens[1]
		// RPMs are labelled as <name>.noarch so we remove this last part
		name = strings.Replace(name, ".noarch", "", 1)
		packages = append(packages, &Package{name, version})
	}

	return packages
}

func (a *Ayum) _execListInstalledCmd(timeoutSecs int) *shell.Result {
	cmds := []string{
		a.configureShellCmds(),
		fmt.Sprintf("%s -q list installed", a.Binary),
	}

	opt := shell.Timeout(time.Duration(timeoutSecs))
	return shell.Run(strings.Join(cmds, ";"), opt)
}

func (a *Ayum) isInstalled(rpm string) bool {
	return false
}

// ------------------------------------------------------------------
// -- File utility functions ----------------------------------------
// ------------------------------------------------------------------

// pathExists checks if the given path exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// copyDir copies a whole directory recursively
func copyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}
