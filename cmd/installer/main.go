package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	installer "github.com/brinick/atlas-rpm-installer"
	"github.com/brinick/atlas-rpm-installer/cli"
	"github.com/brinick/atlas-rpm-installer/config"
	"github.com/brinick/atlas-rpm-installer/pkg/ayum"
	"github.com/brinick/atlas-rpm-installer/pkg/fs"
	"github.com/brinick/atlas-rpm-installer/pkg/rpm"

	"github.com/brinick/logging"
)

var (
	// ExitCode contains the mapping of int exit codes
	// to a name that explains what it means
	ExitCode = struct {
		OK             int
		ParserError    int
		InstallerError int
		SignalEvent    int
	}{0, 1, 2, 3}
)

// TODO:
// We need to log to stdout/err (with color) and also
// to file for ayum output etc
func main() {
	// By default this is configured to stdout
	mainlog := logging.NewClient("logrus")

	args, err := cli.Parse()
	if err != nil {
		mainlog.Error("Command line args parse error", logging.F("err", err))
		os.Exit(ExitCode.ParserError)
	}

	// Choose the appropriate filesystem transactioner
	// based on the installdir path prefix.
	fsTransactioner := fs.Select(args.Paths.InstallDir)

	// Make a specific logger for ayum  - sending to logfile
	ayumlog := logging.NewClient("logrus")
	ayumlog.Configure(&logging.Config{
		Outfile: filepath.Join(config.Paths.Logs, fmt.Sprintf("%s.ayum.log", args.ID())),
	})
	ayumer := ayum.New(args.Ayum.Repo, args.Installdir, ayumlog)

	rpmFinder := rpm.Finder{args.RPMSrcBaseDir}

	inst, err := installer.New(&args, mainlog, fsTransactioner, ayumer, rpmFinder)
	if err != nil {
		mainlog.Error("Failed to create an Installer", logging.F("err", err))
		os.Exit(ExitCode.InstallerError)
	}

	ctx := context.Background()

	// Timeout and abort the install attempt if used provided flag is present
	if args.Timeout > 0 {
		var timeoutFn context.CancelFunc
		ctx, timeoutFn = context.WithTimeout(ctx, time.Duration(args.Timeout)*time.Second)
		defer timeoutFn()
	}

	if err := inst.Execute(ctx); err != nil {
		mainlog.Error("Installer failed to execute", logging.F("err", err))
		os.Exit(ExitCode.InstallerError)
	}

	os.Exit(ExitCode.OK)
}
