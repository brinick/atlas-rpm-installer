package tagsfile

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// New creates a new tags file based on src. It immediately copies
// this src to a backup, and any entries added will act on
// the backup version.
func New(src string) (*tagsFile, error) {
	t := &tagsFile{
		src: src,
	}

	return t, nil
}

// ---------------------------------------------------------------------

// Entry is a line in a tags file
type Entry struct {
	Label       string
	Branch      string
	Datetime    string
	Project     string
	BaseRel     string
	Platform    string
	Separator   string
	isValidated bool
}

func (e *Entry) validate() {
	if e.isValidated {
		return
	}

	// validate...
}

func (e *Entry) String() string {
	e.Separator = strings.TrimSpace(e.Separator)
	if e.Separator == "" {
		e.Separator = defaultSeparator
	}
	return strings.Join(
		[]string{e.Label,
			e.Branch,
			e.Datetime,
			fmt.Sprintf("%s-%s", e.Project, e.BaseRel),
			e.Platform,
		},
		e.Separator,
	)
}

// ----------------------------------------------------------------------

type tagsFile struct {
	src     string
	bck     string
	ignore  []string
	entries []string
}

func (t *tagsFile) Src() string {
	return t.src
}

func (t *tagsFile) load() error {
	fd, err := os.Open(t.Src)
	if err != nil {
		return fmt.Errorf("unable to open tags file (%w)", err)
	}

	defer fd.Close()

	rejectLine := func(s string) bool {
		for _, i := range t.ignore {
			if strings.Contains(s, i) {
				return true
			}
		}

		return false
	}

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		if !rejectLine(line) {
			t.entries = append(t.entries, line)
		}

	}
	contents, err := ioutil.ReadFile(t.src)
	if err != nil {
		return err
	}

}

func (t *tagsFile) Add(entries []*Entry) error {
	if len(entries) == 0 {
		return nil
	}

	return nil
}

func openFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}

}
