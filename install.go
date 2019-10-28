package installer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/brinick/atlas-rpm-installer/pkg/filesystem"
	"github.com/brinick/atlas-rpm-installer/pkg/rpm"
	"github.com/brinick/atlas-rpm-installer/pkg/tagsfile"
	"github.com/brinick/fs"
	"github.com/brinick/logging"
)

// TODO: are we consistent with setting the Installer err field?

var (
	ErrTransactionClose = fmt.Errorf("failed to close transaction")
)

type ayumer interface {
	Download(context.Context) error
	PreConfigure(string) error
	Configure(context.Context) error
	AddRemoteRepos([]*rpm.Repo) error
	CleanAll(context.Context, string) error
	Install(context.Context, ...string) error
}

type rpmFinder interface {
	Find(string, string) (*rpm.RPMs, error)
}

// Opts configures the Installer
type Opts struct {
	Branch string

	Platform  string
	Timestamp string
	Project   string

	// Base directory below which we install
	InstallBaseDir string

	// Directory where we do our work
	WorkBaseDir string

	// Directory where the stable releases
	// repository is located (to get dependencies)
	StableReleasesDir string

	// TagsFile is the path to the tags file
	TagsFile string
}

func (o *Opts) String() string {
	now := strconv.FormatInt(time.Now().Unix(), 10)
	return strings.Join([]string{
		o.Branch,
		o.Project,
		o.Platform,
		o.Timestamp,
		now,
	}, "__")
}

// New returns an installer that can perform an install.
func New(opts *Opts, t filesystem.Transactioner, ay ayumer, finder rpmFinder, log logging.Logger) *Installer {
	return &Installer{
		opts:        opts,
		log:         log,
		transaction: t,
		ayum:        ay,
		rpms:        finder,
		doneChan:    make(chan struct{}),
		err:         &Errors{},
	}
}

// ---------------------------------------------------------------------

// Errors represents a slice of error values
type Errors []error

// Append appends the error
func (e *Errors) Append(err error) {
	*e = append(*e, err)
}

func (e *Errors) String() string {
	return ""
}

func (e *Errors) Error() string {
	return ""
}

// ---------------------------------------------------------------------

// Installer is the main data structure for installing RPMs
type Installer struct {
	opts        *Opts
	transaction filesystem.Transactioner
	ayum        ayumer
	rpms        rpmFinder
	log         logging.Logger
	aborted     bool
	doneChan    chan struct{}
	err         *Errors
}

// Error returns any installer error (or internal ayum error)
func (inst *Installer) Error() string {
	return fmt.Sprintf("%v", *inst.err)
}

// Done returns a channel to wait for the installer to be done
func (inst *Installer) Done() <-chan struct{} {
	return inst.doneChan
}

// Execute will perform the install
func (inst *Installer) Execute(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			inst.log.Info("Recovered from panic", logging.F("err", r))
			inst.err.Append(fmt.Errorf("%v", r))
		}
		inst.setDone()
	}()

	// stop on error
	if err := inst.openTransaction(ctx); err != nil {
		inst.err.Append(err)
		inst.aborted = true
		return
	}

	defer func() {
		if err := inst.closeTransaction(ctx); err != nil {
			// TODO: warn via email/slack etc about transaction close failure
			inst.err.Append(fmt.Errorf("failed to close transaction (%w)", err))
		}
	}()

	// Launch the install in the background
	go func() {
		if err := inst.doInstall(ctx); err != nil {
			inst.err.Append(err)
		}
	}()

	// Wait for either the install to be done,
	// or a signal to be trapped (in which case we abort the install)
	select {
	case <-inst.Done():
		// we're done here, let's go home
	case <-ctx.Done():
		inst.aborted = true
		<-inst.Done()
	}
}

func (inst *Installer) setDone() {
	// Close of a closed channel panics, hence this check
	_, isOpen := <-inst.doneChan
	if isOpen {
		close(inst.doneChan)
	}
}

// Aborted indicates if this installer was stopped prematurely,
// likely because a context was canceled
func (inst *Installer) Aborted() bool {
	return inst.aborted
}

func (inst *Installer) getRPMsList() ([]*rpm.RPMs, error) {
	rpms, err := inst.rpms.Find(inst.opts.Project, inst.opts.Platform)
	if err != nil {
		return nil, err
	}

	// Split RPMs
	isCacheNightly := (strings.Count(inst.opts.Branch, ".")) > 2
	if isCacheNightly {
		return []*rpm.RPMs{rpms}, nil
	}

	var offline, hlt rpm.RPMs
	for _, r := range *rpms {
		// TODO: finish this
		if strings.HasPrefix(r.Name(), "AtlasHLT") {
			hlt = rpm.RPMs{r}
		} else {
			offline = rpm.RPMs{r}
		}

	}

	list := []*rpm.RPMs{&offline, &hlt}
	return list, nil
}

func (inst *Installer) doInstall(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			inst.log.Error("Recovered from panic", logging.F("err", r))
		}

		inst.setDone()
	}()

	rpmsList, err := inst.getRPMsList()
	if err != nil {
		return err
	}

	if err := inst.ayumPrepare(ctx); err != nil {
		return err
	}

	var installOK bool
	for _, rpms := range rpmsList {
		if err = inst.installRPMs(ctx, rpms); err == nil {
			installOK = true
		}
	}

	if !installOK {
		inst.cleanDirs()
		// send email
	}

	// copy ayum log

	return nil
}

// installRPMs installs a given set of RPMs
func (inst *Installer) installRPMs(ctx context.Context, rpms *rpm.RPMs) error {
	if err := inst.ayum.Install(ctx, rpms.Names()...); err != nil {
		return err
	}

	inst.cleanDirs()
	inst.ayum.CleanAll(ctx, "atlas-offline-nightly")

	writeTagsFile(inst.opts.TagsFile)
}

func (inst *Installer) writeTagsFile(ignore []string) error {
	// Push out a new tags file (local copy) that has no bad entries
	src, err := os.Open(inst.opts.TagsFile)
	if err != nil {
		return fmt.Errorf(
			"unable to open tags file for reading %s (%w)",
			inst.opts.TagsFile,
			err,
		)
	}
	defer src.Close()

	tmpFile := filepath.Join(os.Getenv("HOME"), "AMItags")
	tmp, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf(
			"failed to open temp tags file for writing %s (%w)",
			tmpFile,
			err,
		)
	}

	defer tmp.Close()

	if err = copyLines(src, tmp, lineReject(ignore)); err != nil {
		return err
	}

	dir := fs.Dir(
		inst.opts.InstallBaseDir,
		inst.opts.Branch,
		inst.opts.Timestamp,
	)

	projdir := dir.Append(inst.opts.Project)
	subdirs, err := projdir.SubDirs()
	if err != nil {
		return fmt.Errorf("failed to list sub-dirs of %s (%w)", projdir.Path, err)
	}

	if len(*subdirs) != 1 {
		return fmt.Errorf(
			"expected project dir (%s) to contain a single subdir, found %d",
			projdir.Path,
			len(*subdirs),
		)
	}

	baseRelease := subdirs.Names()[0]

	// TODO: push these into a configurable step
	entries := dir.Entries().Not(".cvmfscatalog*", "*.ayum.log")
	var lines []string
	for _, entry := range *entries {
		tfe := tagsfile.Entry{
			Label:    "VO-atlas-nightly",
			Branch:   inst.opts.Branch,
			Datetime: inst.opts.Timestamp,
			Project:  inst.opts.Project,
			BaseRel:  baseRelease,
			Platform: inst.opts.Platform,
		}

		lines = append(lines, tfe.String())
	}

	// output the file
	return writeLines(tmp, lines)
}

// cleanDirs removes certain install directories, post install
func (inst *Installer) cleanDirs() error {
	installdir := filepath.Join(inst.opts.InstallBaseDir, inst.opts.Branch)

	toDelete := []string{".yumcache"}
	if inst.opts.Branch == "master" || inst.opts.Branch == "master-GAUDI" {
		toDelete = append(toDelete, "tdaq", "tdaq-common", "dqm-common")
	}

	for i, d := range toDelete {
		toDelete[i] = filepath.Join(installdir, d)
	}

	return fs.Dirs(toDelete...).Remove()
}

// ayumPrepare downloads ayum and configures it, ready for installing RPMs
func (inst *Installer) ayumPrepare(ctx context.Context) error {
	var err error
	if err = inst.ayum.Download(ctx); err != nil {
		return err
	}

	if err = inst.ayum.PreConfigure(inst.opts.StableReleasesDir); err != nil {
		return err
	}

	if err = inst.ayum.AddRemoteRepos(inst.getRemoteRepos()); err != nil {
		return err
	}

	if err = inst.ayum.Configure(ctx); err != nil {
		return err
	}

	if err = inst.ayum.CleanAll(ctx, "atlas-offline-nightly"); err != nil {
		return err
	}

	return nil
}

// openTransacation tries to open the appropriate file-system transaction
func (inst *Installer) openTransaction(ctx context.Context) error {
	err := inst.transaction.Open(ctx)
	if err == nil {
		return nil
	}

	// Is a context error
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		inst.log.Info("Context done, aborting file-system transaction open", logging.ErrField(err))
		return fmt.Errorf("context done, aborting transaction open (%w)", err)
	}

	inst.log.Error("Unable to open file system transaction", logging.ErrField(err))
	return fmt.Errorf("failed to open file system transaction (%w)", err)
}

func (inst *Installer) closeTransaction(ctx context.Context) error {
	// TODO: get error from close and email/log etc if failed to close
	return inst.transaction.Close(ctx)
}
