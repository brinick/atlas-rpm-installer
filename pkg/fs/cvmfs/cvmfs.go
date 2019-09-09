package cvmfs

import (
	"context"
	"fmt"
	"os/exec"
)

// Opts configures the CVMFS transaction
type Opts struct {
	// Path to the CVMFS server binary
	Binary string

	// Name of the nightly repo
	NightlyRepo string

	// Path where we store fixed number software releases
	StableRepoDir string

	// Gateway Machine to access CVMFS
	GatewayNode string

	// How many times we try to open our own CVMFS transaction
	MaxTransitionAttempts int
}

func shell(cmd string, args ...string) error {
	exec.Command(cmd, args...)
	return nil
}

func shellWithContext(ctx context.Context, cmd string, args ...string) error {
	c := exec.CommandContext(ctx, cmd, args...)
	return c.Run()
}

var (
	// ErrTransactionOngoing is the error returned if an attempt to
	// open a second transaction is made
	ErrTransactionOngoing = fmt.Errorf("transaction already ongoing")
)

// New will create a transaction object and call
// its open() method. The transaction Close() method should
// be deferred immediately after calling this, assuming
// no error was returned.
func New(opts *Opts) (*Transaction, error) {
	t := &Transaction{
		Repo: opts.NightlyRepo,
	}

	return t, t.open()
}

// Transaction represents a CVMFS transaction
type Transaction struct {
	Binary  string
	Repo    string
	ongoing bool
}

// Open will create a new transaction. If one
// is already ongoing on this node, it will return
// an error
func (t *Transaction) Open(ctx context.Context) error {
	if !t.ongoing {
		return ErrTransactionOngoing
	}

	err := shellWithContext(context.TODO(), t.Binary, "transaction", t.Repo)

	return shell(t.Binary, "transaction", t.Repo)
}

// Close will exit the transaction after publishing
func (t *Transaction) Close() error {
	if !t.ongoing {
		return nil
	}

	// exec stop command
	return shell(t.Binary, "publish", t.Repo)
}

// Abort will halt the ongoing transaction forcefully
// exiting without publishing
func (t *Transaction) Abort() error {
	if !t.ongoing {
		return nil
	}

	return shell(t.Binary, "abort", "-f", t.Repo)
}
