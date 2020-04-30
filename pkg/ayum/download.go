package ayum

import (
	"context"
	"fmt"
	"os"
	"time"

	git "gopkg.in/src-d/go-git.v4"
)

type downloader interface {
	Download(context.Context) error
}

type cmdDownload struct {
	srcRepo string
	tgtDir  string
	timeout int
}

// Download clones the ayum source git repository.
// The DownloadTimeout option is the maximum seconds that
// this operation may take before interruption.
// If set to <= 0, no timeout is applied.
func (cmd *cmdDownload) Download(ctx context.Context) error {
	if cmd.timeout > 0 {
		var cancelFn context.CancelFunc
		duration := time.Duration(cmd.timeout) * time.Second
		ctx, cancelFn = context.WithTimeout(ctx, duration)
		defer cancelFn()
	}

	os.RemoveAll(cmd.tgtDir)

	isBare := false
	opts := &git.CloneOptions{URL: cmd.srcRepo}
	_, err := git.PlainCloneContext(ctx, cmd.tgtDir, isBare, opts)

	if err == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		err = ctx.Err()
	default:
	}

	return fmt.Errorf("ayum repo download failed (%w)", err)
}
