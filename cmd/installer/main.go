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
	"github.com/brinick/atlas-rpm-installer/pkg/tagsfile"

	"github.com/brinick/logging"
)

func init() {
	// Should we set up and expose expvar end points for monitoring?
	if cfg.Admin.Monitor {

	}
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

	// Declare a default null logger
	log logging.Logger = logging.NewNullLogger(nil)

	// Load up the configuration
	cfg = getConfig()
)

// TODO: We need to log to stdout/err (with color) and also
// to file for ayum output (ayum also should go to stdout)
func main() {
	defer timeIt(time.Now(), "main")

	// Trap int/term signals
	signalChan := trap()
	defer close(signalChan)

	log = getLogger(cfg.Logging)

	ayumLog := filepath.Join(cfg.Dirs.Logs, fmt.Sprintf("%s.ayum.log", cfg.Install.Opts))

	// Instantiate the installer with all the required plumbing
	inst := installer.New(
		// installation options
		&cfg.Install.Opts,

		// file system transaction
		makeTransactioner(fsSelector(cfg.Dirs.InstallBase)),

		// ayum handler
		makeAyumer(ayumLog),

		// rpm/dependency finder
		rpm.NewFinder(cfg.Dirs.RPMSrcBase),

		// tagsfile updater
		tagsfile.New(cfg.Install.TagsFile, os.Getenv("HOME")),

		// Write to the same log file
		log,
	)

	// Prepare a context to allow for cancelling the installation
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	// Add an install timeout, if requested
	if cfg.Global.TimeOut > 0 {
		var timeoutFn context.CancelFunc
		duration := time.Duration(cfg.Global.TimeOut) * time.Second
		ctx, timeoutFn = context.WithTimeout(ctx, duration)
		defer timeoutFn()
	}

	// Launch the install in the background
	go inst.Execute(ctx)

	// And wait...
	select {
	case sig := <-signalChan:
		log.Info("Signal trapped, shutting down", logging.F("sig", sig))
		cancelCtx()
		<-inst.Done()
		// TODO: log about the signal in ayum log also?
	case <-inst.Done():
		//
	case <-ctx.Done():
		// timeout
	}

	// TODO: this should return an error not a string
	if inst.Error() != "" {
		log.Error("Install FAIL", logging.F("err", inst.Error()))
		os.Exit(ExitCode.InstallerError)
	}

	log.Info("Install OK")
	os.Exit(ExitCode.OK)
}

func makeAyumer(logPath string) *ayum.Ayum {
	// Make a specific logger for ayum  - sending to logfile
	ayumlog := logging.NewClient("logrus", nil)
	ayumlog.Configure(&logging.Config{Outfile: logPath})
	return ayum.New(&cfg.Ayum.Opts, ayumlog)
}

func makeTransactioner(shouldInstallOn func(string) bool) filesystem.Transactioner {
	var fsTransactioner filesystem.Transactioner
	switch {
	case shouldInstallOn("cvmfs"):
		fsTransactioner = cvmfs.NewTransaction(&cfg.CVMFS.Opts)
	default:
		log.Error("Unknown install directory file system")
		os.Exit(ExitCode.PreInstallError)
	}

	return fsTransactioner
}

func fsSelector(installdir string) func(string) bool {
	return func(name string) bool {
		return strings.HasPrefix(installdir, name)
	}
}

func getLogger(opts *config.LoggingOpts) logging.Logger {
	return logging.NewClient(
		opts.Client,
		&logging.Config{
			LogLevel:  opts.Level,
			OutFormat: opts.Format,
		},
	)
}

func getConfig() *config.Config {
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing/validating command line: %v", err)
		os.Exit(ExitCode.ParserError)
	}

	return cfg
}

func trap() chan os.Signal {
	// Set up to trap term and int signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	return signalChan
}

func timeIt(start time.Time, name string) {
	d := time.Since(start)
	log.Debug(
		"Execution time",
		logging.F("name", name),
		logging.F("secs", d.Seconds),
	)
}
