package ayum

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4"
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

	// dir := filepath.Join(filepath.Dir(a.Dir), "ayum")
	os.RemoveAll(cmd.tgtDir)

	isBare := false
	opts := &git.CloneOptions{URL: cmd.srcRepo}
	_, err := git.PlainCloneContext(ctx, cmd.tgtDir, isBare, opts)

	switch err {
	case nil:
		return nil
	case context.DeadlineExceeded:
		return errors.New("ayum repo git clone took too long and was killed")
	case context.Canceled:
		return errors.New("ayum repo git clone is cancelled")
	default:
		return fmt.Errorf("ayum repo git clone download failed (%w)", err)
	}
}
