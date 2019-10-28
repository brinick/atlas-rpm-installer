package main

import (
	"context"
	_ "expvar" // register the /debug/vars endpoint for metrics
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	installer "github.com/brinick/atlas-rpm-installer"
	"github.com/brinick/atlas-rpm-installer/config"

	"github.com/brinick/atlas-rpm-installer/pkg/ayum"
	"github.com/brinick/atlas-rpm-installer/pkg/filesystem"
	"github.com/brinick/atlas-rpm-installer/pkg/filesystem/cvmfs"
	"github.com/brinick/atlas-rpm-installer/pkg/rpm"

	"github.com/brinick/logging"
)

func init() {
	// TODO: Set up the expvar end point variables to monitor

}

var (
	// ExitCode contains the mapping of int exit codes
	// to a name that explains what it means
	ExitCode = struct {
		OK              int
		ParserError     int
		PreInstallError int
		InstallerError  int
		SignalEvent     int
	}{0, 1, 2, 3, 4}
)

func fsSelector(installdir string) func(string) bool {
	return func(name string) bool {
		return strings.HasPrefix(installdir, name)
	}
}

// TODO:
// We need to log to stdout/err (with color) and also
// to file for ayum output (ayum also should go to stdout)
func main() {
	// Set up to trap term and int signals
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	// By default this is configured to stdout
	mainlog := logging.NewClient("logrus")

	cfg, err := config.New()
	if err != nil {
		mainlog.Error("Error parsing/validating command line", logging.F("err", err))
		os.Exit(ExitCode.ParserError)
	}

	mainlog.Debug(cfg.Dump())

	/*
		nightlyInstallDir := filepath.Join(
			cfg.Dirs.InstallBase,
			cfg.Install.Branch,
			cfg.Install.Timestamp,
		)
	*/

	// --------------------------------------------------------------
	// Construct the bits we need for installation
	// --------------------------------------------------------------
	// --- 1. File system transactioner
	var fsTransactioner filesystem.Transactioner
	shouldInstallOn := fsSelector(cfg.Dirs.InstallBase)
	switch {
	case shouldInstallOn("cvmfs"):
		fsTransactioner = cvmfs.New(&cfg.CVMFS.Opts)
	default:
		mainlog.Error("Unknown install directory file system")
		os.Exit(ExitCode.PreInstallError)
	}

	// --- 2. An Ayum instance
	// Make a specific logger for ayum  - sending to logfile
	name := fmt.Sprintf("%s.ayum.log", cfg.Install.Opts)
	ayumlogFile := filepath.Join(cfg.Dirs.Logs, name)
	ayumlog := logging.NewClient("logrus")
	ayumlog.Configure(&logging.Config{Outfile: ayumlogFile})

	ayumer := ayum.New(&cfg.Ayum.Opts, ayumlog)

	// --- 3. An RPM finder for locating our RPMs and their dependencies
	rpmFinder := rpm.Finder{cfg.Dirs.RPMSrcBase}

	// --------------------------------------------------------------
	// Now we go ahead and install...
	// --------------------------------------------------------------

	// Timeout and abort the install attempt if used provided flag is present
	if cfg.Global.TimeOut > 0 {
		var timeoutFn context.CancelFunc
		duration := time.Duration(cfg.Global.TimeOut) * time.Second
		ctx, timeoutFn = context.WithTimeout(ctx, duration)
		defer timeoutFn()
	}

	inst := installer.New(&cfg.Install.Opts, fsTransactioner, ayumer, rpmFinder, mainlog)

	go inst.Execute(ctx)

	select {
	case sig := <-signalChan:
		cancelCtx()
		<-inst.Done()
		// log about the signal in ayum log also?
	case <-inst.Done():
		//
	case <-ctx.Done():
		// timeout
	}

	// TODO: this should return an error not a string
	if inst.Error() != "" {
		mainlog.Error("Install not OK", logging.F("err", inst.Error()))
		os.Exit(ExitCode.InstallerError)
	}

	mainlog.Info("Install OK")
	os.Exit(ExitCode.OK)
}
