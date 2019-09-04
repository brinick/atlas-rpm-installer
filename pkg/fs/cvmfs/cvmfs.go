package cvmfs

import (
	"context"
	"fmt"
	"os/exec"
)

var cvmfsBin = "/usr/bin/cvmfs_server"

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
func New(repo string) (*Transaction, error) {
	t := &Transaction{
		Repo: repo,
	}

	return t, t.open()
}

// Transaction represents a CVMFS transaction
type Transaction struct {
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

	err := shellWithContext(context.TODO(), cvmfsBin, "transaction", t.Repo)

	return shell(cvmfsBin, "transaction", t.Repo)
}

// Close will exit the transaction after publishing
func (t *Transaction) Close() error {
	if !t.ongoing {
		return nil
	}

	// exec stop command
	return shell(cvmfsBin, "publish", t.Repo)
}

// Abort will halt the ongoing transaction forcefully
// exiting without publishing
func (t *Transaction) Abort() error {
	if !t.ongoing {
		return nil
	}

	return shell(cvmfsBin, "abort", "-f", t.Repo)
}
