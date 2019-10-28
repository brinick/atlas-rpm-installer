package filesystem

import (
	"context"
)

// Transactioner defines the interface for file system transactions
type Transactioner interface {
	Open(context.Context) error
	Close(context.Context) error
	Abort() error
}
