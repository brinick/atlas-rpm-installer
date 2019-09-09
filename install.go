package installer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/brinick/atlas-rpm-installer/pkg/fs"
	"github.com/brinick/atlas-rpm-installer/pkg/rpm"
	"github.com/brinick/logging"
)

type ayumer interface {
	Download(context.Context, time.Duration)
	PreConfigure(string)
	Configure(string) error
	AddRemoteRepos(rpm.Repos) error
	CleanAll(string) error
	Install(context.Context, ...string) error
}

type rpmFinder interface {
	Find(string, string) (rpm.RPMs, error)
}

// New returns an installer that can perform an install.
// The type of the installer returned depends on the file
// system of the target install directory (CVMFS, AFS, ...).
func New(
	args *Args, log logging.Logger, t fs.Transactioner, ay ayumer, finder rpmFinder,
) (*Installer, error) {
	return &Installer{
		branch:      args.Branch,
		platform:    args.Platform,
		timestamp:   args.Timestamp,
		project:     args.Project,
		rpmSrcDir:   filepath.Join(args.RPMSrcBaseDir, args.Branch, args.Platform, args.Timestamp),
		installdir:  args.Installdir,
		workdir:     args.Workdir,
		log:         log,
		transaction: t,
		ayum:        ay,
		rpms:        finder,
		abortChan:   make(chan struct{}),
		doneChan:    make(chan struct{}),
	}, nil
}

// Installer is the main data structure for installing RPMs
type Installer struct {
	branch      string
	platform    string
	project     string
	timestamp   string
	workdir     string
	installdir  string
	rpmSrcDir   string
	log         logging.Logger
	abortChan   chan struct{}
	doneChan    chan struct{}
	transaction fs.Transactioner
	ayum        ayumer
	rpms        rpmFinder
	err         error
}

// Err returns any installer error
func (inst *Installer) Err() error {
	return inst.err
}

func (inst *Installer) done() <-chan struct{} {
	return inst.doneChan
}

func (inst *Installer) setDone() {
	// Close of a closed channel panics, hence this check
	_, isOpen := <-inst.doneChan
	if isOpen {
		close(inst.doneChan)
	}
}

func (inst *Installer) aborted() <-chan struct{} {
	return inst.abortChan
}

func (inst *Installer) setAbort() {
	// Close of a closed channel panics, hence this check
	_, isOpen := <-inst.abortChan
	if isOpen {
		close(inst.abortChan)
	}
}

func (inst *Installer) exec(ctx context.Context) error {
	go func() {
		defer inst.setDone()

		rpms, err := inst.rpms.Find(inst.project, inst.platform)
		if err != nil {
			inst.err = err
			return
		}

		// Download and configure ayum
		inst.ayum.Download(ctx)
		inst.ayum.PreConfigure(inst.installdir)
		inst.ayum.AddRemoteRepos(inst.getRemoteRepos())
		inst.ayum.Configure(inst.installdir)
		inst.ayum.CleanAll("atlas-offline-nightly")
		inst.ayum.Install(ctx, rpms.Names()...)
	}()

	select {
	case <-inst.done():
		//
	case <-ctx.Done():
		<-inst.done()
	}
	return nil
}

func (inst *Installer) doExec(ctx context.Context) {
	defer func() {
		// TODO: get error from close and email/log etc if failed to close
		inst.transaction.Close()
		inst.setDone()
	}()

	if inst.err = inst.transaction.Open(ctx); inst.err != nil {
		inst.log.Error("Unable to open file system transaction", logging.F("err", inst.err))
	}

	inst.err = inst.exec(ctx)
}

// Execute will perform the install
func (inst *Installer) Execute(ctx context.Context) error {
	// Set up to trap term and int signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	// Launch the install in the background
	go inst.doExec(ctx)

	// Wait for either the install to be done,
	// or a signal to be trapped (in which case we abort the install)
	select {
	case <-inst.done():
		// we're done here, let's go home
	case <-ctx.Done():
		<-inst.done()
	case sig := <-signalChan:
		inst.log.Info("Signal was caught, aborting install", logging.F("sig", sig))
		inst.setAbort()
		inst.log.Debug("Waiting for installer to exit")
		<-inst.done()
		return fmt.Errorf("signal %s was caught, install aborted", sig)
	}

	return nil
}
