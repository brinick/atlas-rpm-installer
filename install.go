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

// ----------------------------------------------------------------------
// - Interfaces to decouple dependencies for injection
// ----------------------------------------------------------------------

type downloader interface {
	Download(context.Context) error
}
type configurer interface {
	PreConfigure(string) error
	Configure(context.Context) error
}

type cleaner interface {
	CleanAll(context.Context, string) error
}

type installer interface {
	Install(context.Context, ...string) error
}

type rpmRepoAdder interface {
	AddRemoteRepos([]*rpm.Repo) error
}

type ayumer interface {
	downloader
	configurer
	rpmRepoAdder
	cleaner
	installer
	Log() logging.Logger
}

type rpmRepoer interface {
	Filename() string
	String() string
}

type rpmFinder interface {
	Find(string, string) (*rpm.RPMs, error)
	SrcDir() string
}

type tagsFiler interface {
	Src() *fs.File
	Remove(...string) error
	Append(*tagsfile.Entries) error
	Save() error
}

// --------------------------------------------------------------------

// Opts configures the Installer
type Opts struct {
	Branch    string `json:"branch"`
	Platform  string `json:"platform"`
	Timestamp string `json:"timestamp"`
	Project   string `json:"project"`

	// Base directory below which we install
	InstallBaseDir string `json:"install_base_dir"`

	// Directory where we do our work
	WorkBaseDir string `json:"work_base_dir"`

	// Directory where the stable releases
	// repository is located (to get dependencies)
	StableReleasesDir string `json:"stable_releases_dir"`

	// TagsFile is the path to the tags file
	TagsFile string `json:"tagsfile"`
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

// ---------------------------------------------------------------------

// New returns an installer that can perform an install.
func New(
	opts *Opts,
	t filesystem.Transactioner,
	ay ayumer,
	finder rpmFinder,
	tags tagsFiler,
	log logging.Logger,
) *Installer {
	return &Installer{
		opts:        opts,
		log:         log,
		transaction: t,
		ayum:        ay,
		rpms:        finder,
		tags:        tags,
		doneChan:    make(chan struct{}),
		err:         &Errors{},
	}
}

// ---------------------------------------------------------------------

// Errors represents a slice of error values
type Errors []error

// Append appends the error
func (e *Errors) Append(err error) {
	if err != nil {
		*e = append(*e, err)
	}
}

func (e *Errors) String() string {
	var o []string
	for _, err := range *e {
		o = append(o, fmt.Sprintf("%v", err))
	}
	return strings.Join(o, "\n")
}

// ---------------------------------------------------------------------

// Installer is the main data structure for installing RPMs
type Installer struct {
	opts        *Opts
	transaction filesystem.Transactioner
	ayum        ayumer
	rpms        rpmFinder
	log         logging.Logger
	tags        tagsFiler
	aborted     bool
	doneChan    chan struct{}
	err         *Errors
}

// IsError indicates if any errors have occured
func (inst *Installer) IsError() bool {
	return len(*inst.err) > 0
}

// Err returns the installer Errors instance
func (inst *Installer) Err() *Errors {
	return inst.err
}

// Done returns a channel to wait for the installer to be done
func (inst *Installer) Done() <-chan struct{} {
	return inst.doneChan
}

// Aborted indicates if this installer was stopped prematurely,
// likely because a context was canceled
func (inst *Installer) Aborted() bool {
	return inst.aborted
}

// Execute will perform the install
func (inst *Installer) Execute(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			inst.log.Info("Recovered from panic", logging.F("err", r))
			inst.err.Append(PanicRecoverError{fmt.Sprintf("%v", r)})
		}
		inst.setDone()
	}()

	// stop on error
	if err := inst.openTransaction(ctx); err != nil {
		inst.err.Append(NewTransactionOpenError(err))
		inst.aborted = true
		return
	}

	defer func() {
		inst.err.Append(inst.copyAyumLog())

		inst.endTransaction(ctx)
	}()

	// Launch the install in the background
	go func() {
		defer inst.setDone()
		if err := inst.doInstall(ctx); err != nil {
			inst.err.Append(err)
		}
	}()

	// Wait for either the install or the context to be done
	select {
	case <-inst.Done():
		// we're done here, let's go home
	case <-ctx.Done():
		inst.aborted = true
		<-inst.Done()
	}
}

func (inst *Installer) endTransaction(ctx context.Context) {
	// TODO: how to check if the transaction is still open at the end, which
	// will mess with future installation attempts.

	// Should we end by abort, or by normal close?
	shouldAbort := inst.IsError() &&
		!(len(*inst.err) == 1 && errors.Is((*inst.err)[0], AyumCopyLogError{}))

	switch shouldAbort {
	case true:
		if err := inst.abortTransaction(ctx); err != nil {
			inst.err.Append(NewTransactionAbortError(err))
		}
	case false:
		if err := inst.closeTransaction(ctx); err != nil {
			inst.err.Append(NewTransactionCloseError(err))
		}
	}

}

// NightlyID returns a string that identifies this given nightly branch
func (inst *Installer) NightlyID() string {
	return fmt.Sprintf(
		"%s_%s_%s",
		inst.opts.Branch,
		inst.opts.Project,
		inst.opts.Platform,
	)
}

// NightlyInstallDir returns the full path to the installation directory
// for this nightly
func (inst *Installer) NightlyInstallDir() string {
	return filepath.Join(
		inst.opts.InstallBaseDir,
		inst.NightlyID(),
		inst.opts.Timestamp,
	)
}

func (inst *Installer) copyAyumLog() error {
	ayumLog := inst.ayum.Log().Path()
	tgtDir := inst.NightlyInstallDir()

	if err := os.MkdirAll(tgtDir, 0755); err != nil {
		return AyumCopyLogError{
			msg: fmt.Sprintf("cannot create directory %s (%v)", tgtDir, err),
		}
	}

	if err := fs.CopyFile(ayumLog, tgtDir); err != nil {
		return AyumCopyLogError{
			msg: fmt.Sprintf(
				"cannot copy ayum log (%s) to directory %s (%v)",
				ayumLog,
				tgtDir,
				err,
			),
		}
	}

	return nil
}

func (inst *Installer) setDone() {
	// Close of a closed channel panics, hence this check in case we
	// already called this function previously
	_, isOpen := <-inst.doneChan
	if isOpen {
		close(inst.doneChan)
	}
}

func (inst *Installer) getRPMsList() ([]*rpm.RPMs, error) {
	rpms, err := inst.rpms.Find(inst.opts.Project, inst.opts.Platform)
	if err != nil {
		return nil, err
	}

	if inst.isCacheNightly() {
		return []*rpm.RPMs{rpms}, nil
	}

	return inst.splitRPMsList(rpms), nil
}

func (inst *Installer) isCacheNightly() bool {
	return (strings.Count(inst.opts.Branch, ".")) > 2
}

// If we are installing a full nightly release, we should split the list
// of RPMs into two: those for the offline and those for AtlasHLT.
func (inst *Installer) splitRPMsList(rpms *rpm.RPMs) []*rpm.RPMs {
	var offline, hlt rpm.RPMs
	var foundOffline, foundHLT bool

	for _, r := range *rpms {
		if r.NameStartsWith("AtlasHLT") {
			hlt = append(hlt, r)
			continue
		}

		offline = append(offline, r)
		foundOffline = foundOffline || r.NameStartsWith("AtlasOffline")
	}

	if foundOffline && foundHLT {
		return []*rpm.RPMs{&offline, &hlt}
	}

	return []*rpm.RPMs{rpms}
}

func (inst *Installer) doInstall(ctx context.Context) error {
	// 1. Get the RPMs that should be installed
	rpmsList, err := inst.getRPMs(ctx)
	if err != nil {
		return err
	}

	// 2. Download and configure ayum
	if err = inst.configure(ctx); err != nil {
		return err
	}

	// TODO: check that the number of RPMs in EOS nightly dir matches the number
	// installed in our install directory

	// 3. Use ayum to (re)install the RPMs
	var installErr = NewInstallError()
	for _, rpms := range rpmsList {
		installErr.add(inst.installRPMs(ctx, rpms))

		// Stop if the context is done, and return its error
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	nErrs := installErr.length()
	nInstalls := len(rpmsList)

	switch {
	case nErrs == 0:
		return nil
	case nErrs < nInstalls:
		return installErr
	default:
		// No installs were successful
		if err := inst.cleanDirs(ctx); err != nil {
			return NewMultiError(installErr, err)
		}

		return installErr
	}
}

func (inst *Installer) getRPMs(ctx context.Context) ([]*rpm.RPMs, error) {
	var (
		err      error
		done     = make(chan struct{})
		rpmsList []*rpm.RPMs
	)

	go func() {
		defer close(done)
		rpmsList, err = inst.getRPMsList()
	}()

	select {
	case <-done:
		if err != nil {
			return nil, RPMFinderError{err}
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return rpmsList, nil
}

// installRPMs installs a given set of RPMs
func (inst *Installer) installRPMs(ctx context.Context, rpms *rpm.RPMs) error {
	if err := inst.ayum.Install(ctx, rpms.Names()...); err != nil {
		return err
	}

	if err := inst.cleanDirs(ctx); err != nil {
		return err
	}

	// TODO: configure this name
	if err := inst.ayum.CleanAll(ctx, "atlas-offline-nightly"); err != nil {
		return err
	}

	inst.log.Info("Everything complete!")
	return inst.writeTagsFile()
}

func (inst *Installer) writeTagsFile() error {
	inst.log.Info("Writing tags file", logging.F("tgt", inst.tags.Src()))
	nightlyDir, err := fs.NewDir(inst.NightlyInstallDir())
	if err != nil {
		return err
	}

	projectDirs, err := nightlyDir.SubDirs()
	if err != nil {
		return fmt.Errorf("failed to list sub-dirs of %s (%w)", nightlyDir.Path, err)
	}

	projdir := nightlyDir.Append(inst.opts.Project)
	projSubdirs, err := projdir.SubDirs()
	if err != nil {
		return fmt.Errorf("failed to list sub-dirs of %s (%w)", projdir.Path, err)
	}

	if len(*projSubdirs) != 1 {
		return fmt.Errorf(
			"expected project dir (%s) to contain a single subdir, found %d",
			projdir.Path,
			len(*projSubdirs),
		)
	}

	nextRelease := projSubdirs.Names()[0]

	var entries *tagsfile.Entries
	for _, project := range *projectDirs {
		entries.Add(
			&tagsfile.Entry{
				Label:    "VO-atlas-nightly",
				Branch:   inst.opts.Branch,
				Datetime: inst.opts.Timestamp,
				Project:  project.Name(),
				NextRel:  nextRelease,
				Platform: inst.opts.Platform,
			},
		)
	}

	if err := inst.tags.Append(entries); err != nil {
		return err
	}

	inst.tags.Remove(".cvmfscatalog", ".ayum.log")
	return inst.tags.Save()
}

// cleanDirs removes certain install directories, post install
func (inst *Installer) cleanDirs(ctx context.Context) error {
	var (
		err  error
		done = make(chan struct{})
	)

	go func() {
		defer close(done)
		installdir := filepath.Join(inst.opts.InstallBaseDir, inst.NightlyID())

		toDelete := []string{".yumcache"}
		if inst.opts.Branch == "master" || inst.opts.Branch == "master-GAUDI" {
			toDelete = append(toDelete, "tdaq", "tdaq-common", "dqm-common")
		}

		for i, d := range toDelete {
			toDelete[i] = filepath.Join(installdir, d)
		}

		err = fs.Dirs(toDelete...).Remove()
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-done:
	}

	return err
}

// configure readies the installer for installing RPMs,
// by downloading ayum and configuring it.
func (inst *Installer) configure(ctx context.Context) error {
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

	// TODO: configure the name of this repo
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
	return inst.transaction.Close(ctx)
}

func (inst *Installer) abortTransaction(ctx context.Context) error {
	return inst.transaction.Kill(ctx)
}
