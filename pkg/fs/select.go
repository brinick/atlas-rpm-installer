package fs

import (
	"context"
	"strings"

	"github.com/brinick/atlas-rpm-installer/pkg/fs/cvmfs"
)

// Transactioner defines the interface for file system transactions
type Transactioner interface {
	Open(context.Context) error
	Close() error
	Abort() error
}

// Select chooses the appropriate file system
// installer, based on the prefix of the install directory
func Select(installdir string) Transactioner {
	is := func(name string) bool {
		return strings.HasPrefix(installdir, name)
	}

	switch {
	case is("/cvmfs"):
		return cvmfs.Transaction
		// return cvmfs.New(installdir)
	case is("/afs"):
		//
	default:
		//
	}

	return nil
}
