package ayum

import (
	"fmt"
	"strings"
	"time"

	"github.com/brinick/shell"
)

type ayumCmdRunner interface {
	Command() string
	SetCommand(string)
	Run(...shell.Option) error
	Result() *shell.Result
}

type ayumCommand struct {
	label    string
	preCmds  []string
	cmd      string
	postCmds []string
	timeout  int
	result   *shell.Result
}

func (ac *ayumCommand) Run(opts ...shell.Option) error {
	if ac.timeout > 0 {
		opts = append(opts, shell.Timeout(time.Duration(ac.timeout)*time.Second))
	}
	ac.result = shell.Run(ac.command(), opts...)
	return ac.result.Error()
}

// Result retrieves the Result object after running the command
func (ac *ayumCommand) Result() *shell.Result {
	return ac.result
}

func (ac *ayumCommand) Command() string {
	return ac.cmd
}

func (ac *ayumCommand) SetCommand(c string) {
	ac.cmd = c
}

func (ac *ayumCommand) outcome() string {
	var o string
	switch {
	case ac.result.TimedOut:
		o = "timedout"
	case ac.result.Cancelled:
		o = "aborted"
	case ac.result.Crashed:
		o = "crashed"
		// TODO: crash reason
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
