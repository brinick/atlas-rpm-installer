package ayum

import (
	"fmt"
	"strings"

	"github.com/brinick/shell"
)

type ayumCommand struct {
	label  string
	cmd    []string
	opts   []shell.Option
	result *shell.Result
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

func (ac *ayumCommand) run() {
	ac.result = shell.Run(strings.Join(ac.cmd, ";"), ac.opts...)
}

// Result retrieves the Result object after running the command
func (ac *ayumCommand) Result() *shell.Result {
	return ac.result
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

func wrapCommand(preCmds []string, postCmds []string) func(string) []string {
	return func(cmd string) []string {
		return append(
			preCmds,
			append(
				[]string{cmd},
				postCmds...,
			)...,
		)
	}
}
