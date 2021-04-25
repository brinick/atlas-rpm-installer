package ayum

import (
	"fmt"
	"path/filepath"

	"github.com/brinick/atlas-rpm-installer/pkg/metric"
	"github.com/brinick/logging"
)

// The queue to which metrics may be pushed, if requested
var metrics *metric.Queue

// MetricQueue defines the interface for fetching metrics from the queue
type MetricQueue interface {
	Items() <-chan string
}

// Metrics returns a queue of metrics, encoded to a particular format.
// The queue is nil i.e. no metrics are output, unless the ayum struct is
// created with the Opts.MonitoringFormat set to the name of the
// metric encoder to use.
func Metrics() MetricQueue {
	return metrics
}

// New creates a new Ayum instance
func New(opts *Opts, log logging.Logger) *Ayum {
	if opts.MonitoringFormat != "" {
		metrics = metric.NewQueue(opts.MonitoringFormat, 100)
	}

	binary := filepath.Join(opts.AyumDir, "ayum/ayum")

	preCmds := opts.PreCommands
	if len(preCmds) == 0 {
		// default
		preCmds = ayumEnv(opts.AyumDir)
	}

	// default postCommand is to do nothing
	postCmds := opts.PostCommands

	configureExe := filepath.Join(opts.AyumDir, "configure.ayum")
	yumConf := filepath.Join(opts.AyumDir, "yum.conf")

	a := &Ayum{
		Dir:        opts.AyumDir,
		Binary:     binary,
		InstallDir: opts.InstallDir,
		log:        log,
		downloader: &cmdDownload{
			srcRepo: opts.SrcRepo,
			tgtDir:  opts.AyumDir,
			timeout: opts.DownloadTimeout,
		},
		rpmRepoAdder: &rpmRepoAdd{
			basedir: opts.AyumDir,
		},
		configurer: &cmdConfigure{
			installDir: opts.InstallDir,
			log:        log,
			cmd: &ayumCommand{
				label:   "ayum configure",
				timeout: opts.Timeout,
				preCmds: preCmds,
				cmd: fmt.Sprintf(
					"%s -i %s -D | grep -v 'AYUM package location' > %s",
					configureExe,
					opts.InstallDir,
					yumConf,
				),
				postCmds: postCmds,
			},
		},
		installer: &cmdInstall{
			log: log,
			lister: &cmdList{
				log: log,
				cmd: &ayumCommand{
					label:    "ayum list",
					timeout:  opts.Timeout,
					preCmds:  preCmds,
					cmd:      fmt.Sprintf("%s -q list installed", binary),
					postCmds: postCmds,
				},
			},
			rpmInstaller: &ayumCommand{
				preCmds:  preCmds,
				timeout:  opts.InstallTimeout,
				cmd:      fmt.Sprintf("%s -y install ", binary) + "%s",
				postCmds: postCmds,
			},
			rpmReinstaller: &ayumCommand{
				preCmds:  preCmds,
				timeout:  opts.InstallTimeout,
				cmd:      fmt.Sprintf("%s -y reinstall ", binary) + "%s",
				postCmds: postCmds,
			},
		},
		cleaner: &cmdClean{
			log: log,
			// timeout: opts.Timeout,
			cmd: &ayumCommand{
				label:   "ayum clean all",
				preCmds: preCmds,
				cmd: binary +
					"--enablerepo=%s" + // repo name to clean, filled in later
					"clean all",
			},
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

	// PreCommands is a list of commands to run prior to all ayum subcommands
	PreCommands []string

	// PostCommands is a list of commands to run after all ayum subcommands
	PostCommands []string

	// If empty string, do no monitoring, else it is the name of
	// the monitoring format to use (statsd for the moment)
	MonitoringFormat string
}

// Ayum is the ayum wrapper
type Ayum struct {
	downloader
	configurer
	cleaner
	rpmRepoAdder
	installer

	// Binary is the path to the ayum executable
	Binary string

	// Dir is the root directory of the ayum installation
	Dir string

	// InstallDir is the root dir of the install path
	InstallDir string

	// log is a logger instance
	log logging.Logger
}

// Name returns the name of the ayum executable
func (a *Ayum) Name() string {
	return "ayum"
}

// Log retrieves the logging instance to which ayum output is sent
func (a *Ayum) Log() logging.Logger {
	return a.log
}
