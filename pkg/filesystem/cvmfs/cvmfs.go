package cvmfs

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
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

// New will create a transaction object and call
// its open() method. The transaction Close() method should
// be deferred immediately after calling this, assuming
// no error was returned.
func New(opts *Opts) *Transaction {
	return &Transaction{
		Repo:     opts.NightlyRepo,
		Binary:   opts.Binary,
		Node:     opts.GatewayNode,
		attempts: opts.MaxTransactionAttempts,
	}
}

// Transaction represents a CVMFS transaction
type Transaction struct {
	Binary   string
	Repo     string
	Node     string
	attempts int
	ongoing  bool
}

// Open will create a new transaction. If one
// is already ongoing on this node, it will return
// an error
func (t *Transaction) Open(ctx context.Context) error {
	if t.ongoing {
		return nil
	}

	var err error

	for t.attempts > 0 {
		err = shellWithContext(ctx, t.Binary, "transaction", t.Repo)

		// We break and return if no error returned (transaction opened ok),
		// or the error is a context cancel/deadline related one. Any other error
		// implies trying again to open the transaction.
		if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			break
		}

		t.attempts--
	}

	// TODO: log output?
	return err
}

// Close will exit the transaction after publishing
func (t *Transaction) Close(ctx context.Context) error {
	if !t.ongoing {
		return nil
	}

	var err error

	attempts := 3
	for attempts > 0 {
		// TODO: can publish take a legitimately long time?
		err := shellWithContext(ctx, t.Binary, "publish", t.Repo)
		if err == nil {
			break
		}

		attempts--
	}

	// TODO: nested catalog

	return err
}

// Abort will halt the ongoing transaction forcefully
// exiting without publishing
func (t *Transaction) Abort() error {
	if !t.ongoing {
		return nil
	}

	return shellWithContext(context.TODO(), t.Binary, "abort", "-f", t.Repo)
}
