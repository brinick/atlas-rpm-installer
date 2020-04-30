package ayum

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

func TestDownload(t *testing.T) {
	a, err := makeAyum(nil)
	if err != nil {
		t.Fatalf("failed to make ayum instance (%v)\n", err)
	}

	err = a.Download(context.Background())
	if err != nil {
		t.Fatalf("ayum download returned an error: %v\n", err)
	}

	now := time.Now()
	entries, err := ioutil.ReadDir(a.Dir)
	if err != nil {
		t.Fatalf("unable to read ayum dir %s (%v)\n", a.Dir, err)
	}

	// very crude check on the downloaded ayum directory
	var msg = fmt.Sprintf(
		"unable to find expected file 'ayum' in download dir (%s)",
		a.Dir,
	)

	for _, entry := range entries {
		if entry.Name() == "ayum" {
			modtime := entry.ModTime()
			if now.Sub(modtime).Seconds() < 1 {
				// file exists and is recent (< 1s), all ok
				return
			}

			msg = fmt.Sprintf(
				"downloaded ayum repo contains an old 'ayum' file (modtime: %s)",
				modtime,
			)

			break
		}
	}

	t.Error(msg)

}

func TestDownloadWithTimeout(t *testing.T) {
	a, err := makeAyum(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = a.Download(ctx)

	if !errors.Is(errors.Unwrap(err), context.DeadlineExceeded) {
		t.Errorf("expected a context deadline exceeded error, got => %v", err)
	}
}

func TestDownloadWithCancel(t *testing.T) {
	a, err := makeAyum(nil)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err = a.Download(ctx)
	if !errors.Is(errors.Unwrap(err), context.Canceled) {
		t.Errorf("expected a context canceled error, got => %v", err)
	}

}
