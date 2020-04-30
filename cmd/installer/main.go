package main

//TODO: .keep file in nightlies to indicate we don't delete after 30 days

import (
	"context"
	_ "expvar" // register the /debug/vars endpoint for metrics
	"fmt"
	"io/ioutil"
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
	"github.com/brinick/atlas-rpm-installer/pkg/filesystem/afs"
	"github.com/brinick/atlas-rpm-installer/pkg/filesystem/cvmfs"
	"github.com/brinick/atlas-rpm-installer/pkg/filesystem/localfs"
	"github.com/brinick/atlas-rpm-installer/pkg/notify"
	"github.com/brinick/atlas-rpm-installer/pkg/rpm"
	"github.com/brinick/atlas-rpm-installer/pkg/tagsfile"

	"github.com/brinick/logging"
)

var (
	version string

	startEpoch = time.Now().Unix()

	// ExitCode contains the mapping of int exit codes
	// to a name that explains what it means
	ExitCode = struct {
		OK              int
		ParserError     int
		PreInstallError int
		InstallerError  int
		SignalEvent     int
	}{0, 1, 2, 3, 4}

	// Load up the configuration
	cfg = getConfig()
)

func main() {
	var log logging.Logger

	// Time the execution of this main function i.e. of the whole install process
	defer func(start time.Time) {
		d := time.Since(start)
		if log != nil {
			log.Debug(
				"Execution time",
				logging.F("secs", d.Seconds),
			)
		}
	}(time.Now())

	// Trap int/term signals
	signalChan := trap()
	defer close(signalChan)

	// Init the app and ayum loggers
	var outfile = strings.TrimSpace(cfg.Logging.OutFile)

	// Replace special text markers, if they exist, with their corresponding values
	outfile = strings.ReplaceAll(outfile, "%branch", cfg.Install.Branch)
	outfile = strings.ReplaceAll(outfile, "%platform", cfg.Install.Platform)
	outfile = strings.ReplaceAll(outfile, "%project", cfg.Install.Project)
	outfile = strings.ReplaceAll(outfile, "%timestamp", cfg.Install.Timestamp)
	outfile = strings.ReplaceAll(outfile, "%start", fmt.Sprintf("%d", startEpoch))

	// Now, make it into an absolute path
	outfilePath := ""
	ayumOutFilePath := ""
	if len(outfile) > 0 {
		outfilePath = filepath.Join(cfg.Dirs.Logs, outfile+".log")
		ayumOutFilePath = filepath.Join(cfg.Dirs.Logs, outfile+".ayum.log")
	}

	// The main install log
	log, err := logging.NewClient(
		cfg.Logging.Client,
		&logging.Config{
			LogLevel:  cfg.Logging.Level,
			OutFormat: cfg.Logging.Format,
			Outfile:   outfilePath,
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] unable to configure logging (%v)\n", err)
		os.Exit(ExitCode.ParserError)
	}

	// The ayum-specific log
	ayumlog, err := logging.NewClient(
		cfg.Logging.Client,
		&logging.Config{
			LogLevel:  cfg.Logging.Level,
			OutFormat: cfg.Logging.Format,
			Outfile:   ayumOutFilePath,
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] unable to configure ayum logging (%v)\n", err)
		os.Exit(ExitCode.ParserError)
	}

	log.Debug(fmt.Sprintf("\n--- Configuration Dump ---\n\n%s\n", cfg.String()))

	fsTransactioner := makeTransactioner(fsSelector(cfg.Dirs.InstallBase), log)
	os.Exit(0)

	// Make a temporary directory for storing the tagsfile editable copy
	tmpDir, err := ioutil.TempDir("", "AMITags")
	if err != nil {
		log.Fatal("failed to create temporary directory for storing tagsfile", logging.ErrField(err))
	}

	// Clean up the temp tagsfile dir before we exit
	defer os.RemoveAll(tmpDir)

	// Instantiate the installer with all the required plumbing
	inst := installer.New(
		// installation options
		&cfg.Install.Opts,

		// file system transaction handler
		fsTransactioner,

		// ayum handler
		ayum.New(&cfg.Ayum.Opts, ayumlog),

		// rpm/dependency finder
		rpm.NewFinder(cfg.Dirs.RPMSrcBase),

		// tagsfile updater
		tagsfile.New(cfg.Install.TagsFile, tmpDir),

		// Use the same log handler everywhere
		log,
	)

	// Prepare a context to allow for cancelling the installation
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	// Add an install global timeout, if requested
	if cfg.Global.TimeOut > 0 {
		var timeoutFn context.CancelFunc
		duration := time.Duration(cfg.Global.TimeOut) * time.Second
		ctx, timeoutFn = context.WithTimeout(ctx, duration)
		defer timeoutFn()
	}

	// Launch the install in the background
	// go inst.Execute(ctx)

	// And now, we wait...
	select {
	case sig := <-signalChan:
		log.Info("Signal trapped", logging.F("sig", sig))
		log.Info("Install ABORT")
		cancelCtx()
		<-inst.Done()

		// Do not _not_ send email...er...ok, so send email
		if !cfg.Admin.DontSendEmail {
			log.Info("Email requested on failure, sending...")
			email := notify.NewEmail(cfg.Admin.EmailFrom, cfg.Admin.EmailTo)
			res := email.WithTimeout(30).Send(
				fmt.Sprintf("ABORTED: nightly install %s", inst.NightlyID()),
				fmt.Sprintf(
					"The install process with PID %d was terminated "+
						"(signal %s was trapped)\n"+
						"The nightly:\n"+
						"\t%s\n"+
						"was thus not installed.\n\n"+
						"Full output available in the log file:\n"+
						"%s\n",
					os.Getpid(),
					sig,
					inst.NightlyID(),
					log.Path(),
				),
			)

			if res.IsError() {
				log.Error(
					"Failed to send email about nightly install being aborted",
					logging.ErrField(res.Err()),
				)
			}
		}

		os.Exit(ExitCode.SignalEvent)

	case <-inst.Done():
		log.Info("Installation done, checking outcome")
	case <-ctx.Done():
		log.Info("Context is done, installation was stopped")
		<-inst.Done()
	}

	// All ok, no errors, exit normally
	if !inst.IsError() {
		log.Info("Install OK")
		os.Exit(ExitCode.OK)
	}

	// TODO: push errors to metrics counts

	// Something went wrong, log errors and send notification, if configured
	log.Error("Install FAIL")

	for _, err := range *inst.Err() {
		log.Error(fmt.Sprintf("%v", err))
	}

	// Send an email, if requested
	if !cfg.Admin.DontSendEmail {
		log.Info("Email requested on failure, sending...")
		errs := *inst.Err()
		res := notify.NewEmail(cfg.Admin.EmailFrom, cfg.Admin.EmailTo).Send(
			fmt.Sprintf("FAILED: nightly install %s", inst.NightlyID()),
			fmt.Sprintf(
				"The installation for nightly:\n"+
					"\t%s\n"+
					"failed. %d error(s) reported:\n\n"+
					"%s",
				inst.NightlyID(),
				len(errs),
				errs.String(),
			),
		)

		if res.IsError() {
			log.Error(
				"Failed to send email about nightly install failure",
				logging.ErrField(res.Err()),
			)
		}
	}

	os.Exit(ExitCode.InstallerError)
}

// makeTransactioner instantiates the appropriate file system transactioner
func makeTransactioner(is func(string) bool, log logging.Logger) filesystem.Transactioner {
	var t filesystem.Transactioner

	switch {
	case is("/cvmfs"):
		t = cvmfs.NewTransaction(&cfg.CVMFS.Opts, log)
	case is("/afs"):
		t = afs.NewTransaction(&cfg.AFS.Opts, log)
	default:
		t = localfs.NewTransaction(&cfg.LocalFS.Opts, log)
	}

	return t
}

func fsSelector(installdir string) func(string) bool {
	return func(name string) bool {
		return strings.HasPrefix(installdir, name)
	}
}

func getConfig() *config.Config {
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing configuration: %v\n", err)
		os.Exit(ExitCode.ParserError)
	}

	return cfg
}

// trap sets up a channel to trap term and int signals
func trap() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	return signalChan
}
