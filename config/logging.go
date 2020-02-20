package config

import (
	"flag"
	"fmt"
	"strings"
)

// LoggingOpts are options for the application logging
type LoggingOpts struct {
	Level  string
	Format string
	File   string
	Client string
}

func (l *LoggingOpts) flags() {
	flag.StringVar(
		&l.Level,
		"log.level",
		"info",
		"Log level to use (default: info, alternatives: info, error)",
	)
	flag.StringVar(
		&l.Format,
		"log.format",
		"text",
		"Log format to use (default: text, alternatives: json)",
	)
	flag.StringVar(
		&l.File,
		"log.file",
		"",
		"Path to the log file to which to output (default: none i.e. use stdout/err)",
	)
	flag.StringVar(
		&l.Client,
		"log.type",
		"logrus",
		"Type of logging client to use (default: logrus, alternatives: null)",
	)
}

func (l *LoggingOpts) validate() error {
	return nil
}

func (l *LoggingOpts) String() string {
	return strings.Join(
		[]string{
			"- Logging Options:",
			fmt.Sprintf("   - Client: %sLogger", strings.Title(l.Client)),
			fmt.Sprintf("   - Level: %s", l.Level),
			fmt.Sprintf("   - Format: %s", l.Format),
			fmt.Sprintf("   - File: %s", l.File),
		},
		"\n",
	)
}
