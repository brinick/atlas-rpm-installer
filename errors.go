package installer

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------

// RPMFinderError represents any problem finding RPMs or any problem
// related to found RPMs, such as having zero size
type RPMFinderError struct {
	err error
}

func (r RPMFinderError) Error() string {
	return fmt.Sprintf("failed to get RPMs to install: %v", r.err)

}

// ---------------------------------------------------------------------

// NewMultiError returns a MultiError, a wrapper for more than one error
func NewMultiError(errs ...error) MultiError {
	return MultiError{errs}
}

// MultiError represents an error comprising multiple errors
type MultiError struct {
	errs []error
}

func (m MultiError) Error() string {
	var out []string
	for _, err := range m.errs {
		out = append(out, fmt.Sprintf("%v", err))
	}
	return strings.Join(out, "\n")
}

func (m MultiError) add(e error) {
	if e != nil {
		m.errs = append(m.errs, e)
	}
}

func (m MultiError) length() int {
	return len(m.errs)
}

// ---------------------------------------------------------------------

// NewInstallError returns a new installation error
func NewInstallError(errs ...error) InstallError {
	return InstallError{
		MultiError{
			errs,
		},
	}
}

// InstallError represents an error that occured during installation attempts
type InstallError struct {
	MultiError
}

// ---------------------------------------------------------------------

// PanicRecoverError represents the panic error which was recovered
type PanicRecoverError struct {
	msg string
}

func (p PanicRecoverError) Error() string {
	return p.msg
}

// ---------------------------------------------------------------------

// TransactionError is the base error for transaction errors
type TransactionError struct {
	err error
}

func (t TransactionError) Error() string {
	return fmt.Sprintf("%v", t.err)
}

// NewTransactionOpenError returns a TransactionOpenError
func NewTransactionOpenError(e error) TransactionOpenError {
	return TransactionOpenError{
		TransactionError{
			err: e,
		},
	}
}

// NewTransactionCloseError returns a TransactionCloseError
func NewTransactionCloseError(e error) TransactionCloseError {
	return TransactionCloseError{
		TransactionError{
			err: e,
		},
	}
}

// NewTransactionAbortError returns a TransactionAbortError
func NewTransactionAbortError(e error) TransactionAbortError {
	return TransactionAbortError{
		TransactionError{
			err: e,
		},
	}
}

// TransactionOpenError represents an error opening a transaction
type TransactionOpenError struct {
	TransactionError
}

func (t TransactionOpenError) Error() string {
	return fmt.Sprintf("failed to open transaction: %s", t.TransactionError.Error())
}

// TransactionCloseError represents an error closing a transaction
type TransactionCloseError struct {
	err error
}

func (t TransactionCloseError) Error() string {
	return fmt.Sprintf("failed to close transaction: %v", t.err)
}

// TransactionAbortError represents an error aborting a transaction
type TransactionAbortError struct {
	err error
}

func (t TransactionAbortError) Error() string {
	return fmt.Sprintf("failed to abort transaction: %v", t.err)
}

// ---------------------------------------------------------------------

// AyumCopyLogError represents an error during the copying of the ayum log
type AyumCopyLogError struct {
	msg string
}

func (a AyumCopyLogError) Error() string {
	return a.msg
}

// ---------------------------------------------------------------------
