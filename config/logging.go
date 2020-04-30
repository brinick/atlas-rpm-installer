package config

import (
	"flag"
	"fmt"
	"strings"
)

// LoggingOpts are options for the application logging
type LoggingOpts struct {
	Level   string
	Format  string
	OutFile string
	Client  string
}

func (l *LoggingOpts) flags() {
	flag.StringVar(
		&l.Level,
		"log.level",
		"info",
		"Log level to use",
	)
	flag.StringVar(
		&l.Format,
		"log.format",
		"text",
		"Log format to use",
	)
	flag.StringVar(
		&l.OutFile,
		"log.file",
		"",
		"Path to the log file to which to output (default none i.e. use stdout/err)",
	)
	flag.StringVar(
		&l.Client,
		"log.client",
		"logrus",
		"Type of logging client to use",
	)
}

func (l *LoggingOpts) validate() error {
	return nil
}

func (l *LoggingOpts) String() string {
	outFileText := fmt.Sprintf("   - OutFile: %s", l.OutFile)
	if l.OutFile == "" {
		outFileText = fmt.Sprint("   - OutFile: (none, goes to stdout/err)", l.OutFile)
	}
	return strings.Join(
		[]string{
			"- Logging Options:",
			fmt.Sprintf("   - Client: %sLogger", strings.Title(l.Client)),
			fmt.Sprintf("   - Level: %s", l.Level),
			fmt.Sprintf("   - Format: %s", l.Format),
			outFileText,
		},
		"\n",
	)
}
