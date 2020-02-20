package cvmfs

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/brinick/atlas-rpm-installer/pkg/filesystem"
)

// Opts configures the CVMFS transaction
type Opts struct {
	// Path to the CVMFS server binary
	Binary string

	// Name of the nightly repo
	NightlyRepo string

	// Gateway Machine to access CVMFS
	GatewayNode string

	// How many times we try to open our own CVMFS transaction
	MaxTransactionAttempts int
}

func shellWithContext(ctx context.Context, cmd string, args ...string) error {
	c := exec.CommandContext(ctx, cmd, args...)
	return c.Run()
}

var (
	// ErrTooManyAttempts is the error returned once the maximum number
	// of allowed open transaction attempts is reached
	ErrTooManyAttempts = fmt.Errorf("Too many attempts made to open transaction")
)

// NewTransaction will create a transaction object and call
// its open() method. The transaction Close() method should
// be deferred immediately after calling this, assuming
// no error was returned.
func NewTransaction(opts *Opts) *Transaction {
	t := Transaction{
		Repo:     opts.NightlyRepo,
		Binary:   opts.Binary,
		Node:     opts.GatewayNode,
		attempts: opts.MaxTransactionAttempts,
	}

	t.Transaction.Starter = &t
	t.Transaction.Stopper = &t
	return &t
}

// Transaction represents a CVMFS transaction
type Transaction struct {
	filesystem.Transaction
	Binary   string
	Repo     string
	Node     string
	attempts int
}

// Attempts provides the number of tries allowed for opening the transaction
func (t *Transaction) Attempts() int {
	return t.attempts
}

// Start will open a new transaction. If one is already ongoing on
// this node, it will return an error
func (t *Transaction) Start(ctx context.Context) error {
	// TODO: log output?
	return shellWithContext(ctx, t.Binary, "transaction", t.Repo)
}

// Stop will exit the transaction after publishing
func (t *Transaction) Stop(ctx context.Context) error {
	err := shellWithContext(ctx, t.Binary, "publish", t.Repo)

	// TODO: Create the nested catalog

	return err
}

// Kill will halt the ongoing transaction forcefully
// exiting without publishing
func (t *Transaction) Kill(ctx context.Context) error {
	return shellWithContext(ctx, t.Binary, "abort", "-f", t.Repo)
}
