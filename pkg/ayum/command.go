package ayum

import (
	"fmt"
	"strings"
	"time"

	"github.com/brinick/logging"
	"github.com/brinick/shell"
)

/*
type ayumCmdRunner interface {
	Description() string
	Command() string
	SetCommand(string)
	Run(...shell.Option)
	Result() *shell.Result
	Error() error // shortcut to the shell.Result.Error
}
*/

type shellRunner interface {
	Run(string, ...shell.Option) shellResulter
}

type outputter interface {
	Stdout() *shell.Output
	Stderr() *shell.Output
}

type resultAnalyser interface {
	IsError() bool
	Crashed() bool
	CrashReason() string
	Canceled() bool
	TimedOut() bool
}

type shellResulter interface {
	outputter
	resultAnalyser
	Duration() float64
	ExitCode() int
	Err() error
}

type ayumCommand struct {
	label    string
	preCmds  []string
	cmd      string
	postCmds []string
	timeout  int
	result   shellResulter
	runner   shellRunner
}

func (ac *ayumCommand) Run(opts ...shell.Option) {
	if ac.timeout > 0 {
		opts = append(opts, shell.Timeout(time.Duration(ac.timeout)*time.Second))
	}

	if ac.runner == nil {
		// Use the default shell runner
		ac.result = shell.Run(ac.cmd, opts...)
	} else {
		ac.result = ac.runner.Run(ac.cmd, opts...)
	}
}

// Ran indicates if this command already executed
func (ac *ayumCommand) Ran() bool {
	return ac.result != nil
}

// Result retrieves the Result object after running the command
// func (ac *ayumCommand) Result() *shell.Result {
func (ac *ayumCommand) Result() shellResulter {
	return ac.result
}

func (ac *ayumCommand) Command() string {
	return ac.cmd
}

func (ac *ayumCommand) SetCommand(c string) {
	ac.cmd = c
}

func (ac *ayumCommand) Err() error {
	if ac.result != nil {
		return ac.result.Err()
	}
	return nil
}

// outcome should only be called if the command failed.
// It indicates briefly what the failure mode was.
func (ac *ayumCommand) outcome() string {
	var o string
	switch {
	case ac.result.TimedOut():
		o = "timedout"
	case ac.result.Canceled():
		o = "aborted"
	case ac.result.Crashed():
		o = "crashed"
	default:
		o = "failed"
	}

	return o
}

func (ac *ayumCommand) ok() bool {
	return ac.result != nil && !ac.result.IsError() && ac.result.ExitCode() == 0
}

func (ac *ayumCommand) duration() float64 {
	return ac.result.Duration()
}

func (ac *ayumCommand) command() string {
	return strings.Join(
		[]string{
			strings.Join(ac.preCmds, ";"),
			ac.cmd,
			strings.Join(ac.postCmds, ";"),
		},
		";",
	)
}

// ayumEnv returns the commands to execute prior to any ayum commands,
// so that the environement is correctly configured
func ayumEnv(ayumdir string) []string {
	return []string{
		fmt.Sprintf("cd %s", ayumdir),
		"shopt -s expand_aliases",
		"source ayum/setup.sh",
	}
}

// doPostMortem examines the result of having run the ayumCommand,
// outputs to the provided logger and returns any error
func doPostMortem(cmd *ayumCommand, log logging.Logger) error {
	if !cmd.Ran() {
		return nil
	}

	log.InfoL(cmd.Result().Stdout().Lines())

	var err error

	if !cmd.ok() {
		err = cmd.Result().Err()
		outcome := cmd.outcome()
		fields := []logging.Field{
			logging.ErrField(err),
			logging.F("outcome", outcome),
			logging.F("cmd", cmd.label),
		}

		// Particular case
		if cmd.Result().Crashed() {
			field := logging.F("crashReason", cmd.Result().CrashReason())
			fields = append(fields, field)
		}

		log.Error("ayum command failure", fields...)
		log.ErrorL(cmd.Result().Stderr().Lines())
	}

	return err
}
