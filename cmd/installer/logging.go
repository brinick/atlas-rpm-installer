package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/brinick/atlas-rpm-installer/config"

	"github.com/brinick/logging"
)

func createLogger(logfile string, opts *config.LoggingOpts) logging.Logger {
	log, err := logging.NewClient(
		opts.Client,
		&logging.Config{
			LogLevel:  opts.Level,
			OutFormat: opts.Format,
			Outfile:   logfile,
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] unable to create logger at %s (%v)\n", logfile, err)
		os.Exit(ExitCode.ParserError)
	}

	return log
}

// makeLogName replaces special text markers, if they exist, with
// their corresponding values within the tpl logname
func makeLogName(tpl string, inst *config.InstallOpts, instStart int64) string {
	for oldStr, newStr := range map[string]string{
		"%branch":    inst.Branch,
		"%platform":  inst.Platform,
		"%project":   inst.Project,
		"%timestamp": inst.Timestamp,
		"%start":     fmt.Sprintf("%d", instStart),
	} {
		tpl = strings.ReplaceAll(tpl, oldStr, newStr)
	}

	return strings.TrimSpace(tpl)
}

func initLogging(logpath string, opts *cfg.Logging) (logging.Logger, logging.Logger) {
	var (
		log           logging.Logger
		pkgManagerLog logging.Logger
	)

	// Init the app and ayum loggers

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

	return log, ayumlog
}

func initMainLog() logging.Logger {

}
