package ayum

import (
	"fmt"
	"io/ioutil"

	"github.com/brinick/logging"
	"github.com/brinick/shell"
)

func makeTempDir(prefix string) (string, error) {
	tempdir, err := ioutil.TempDir("", fmt.Sprintf("ayum.%s", prefix))
	if err != nil {
		return "", err
	}

	return tempdir, nil
}

func defaultOpts() *Opts {
	tempdir, err := makeTempDir("")
	if err != nil {
		return nil
	}
	return &Opts{
		SrcRepo:         "https://gitlab.cern.ch/atlas-sit/ayum.git",
		AyumDir:         tempdir,
		DownloadTimeout: 30,
	}
}

func makeAyum(opts *Opts) (*Ayum, error) {
	if opts == nil {
		opts = defaultOpts()
	}

	return New(opts, logging.NullLogger{}), nil
}

// ----------------------------------------------------

type fakeResult struct{}

type fakeRunner struct {
}

func (fr *fakeRunner) Run(cmd string, opts ...shell.Option) *fakeResult {
	return &fakeResult{}
}
