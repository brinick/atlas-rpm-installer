package tagsfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brinick/fs"
)

// The field separator in tag file lines
var defaultEntrySeparator = ";"

// SetFieldSeparator sets the field separator in tag file lines
func SetFieldSeparator(s string) {
	if len(s) == 1 {
		defaultEntrySeparator = s
	}
}

// New creates a new tags file instance based on src.
// If editing of the file is requested (writing to, deleting entries),
// the src is first copied to the given tmp directory, and the edits
// are made on the copy. It is up to the client to request a Write
// to push these changes to the src.
func New(src string, tmp string) *TagsFile {
	now := time.Now().UnixNano()
	return &TagsFile{
		src: src,
		bck: filepath.Join(tmp, fmt.Sprintf("AMItags.%d", now)),
	}
}

// ---------------------------------------------------------------------

// Entries represents a list of Entry instances
type Entries []*Entry

// Add will append Entry instances to the current list
func (e *Entries) Add(entries ...*Entry) {
	*e = append(*e, entries...)
}

// Append will add the Entry instances in an Entries object
// on the end of the current list
func (e *Entries) Append(entries *Entries) {
	*e = append(*e, *entries...)
}

// Remove will delete any entry where a given field
// matches against one of the passed in strings
func (e *Entries) Remove(values []string) error {
	for _, entry := range *e {
		if entry.contains(values) {

		}
	}

	return nil
}

// AsLines returns the entries as a list of strings
func (e *Entries) AsLines() []string {
	var lines []string
	for _, entry := range *e {
		lines = append(lines, entry.String())
	}

	return lines
}

// Entry is a line in a tags file
type Entry struct {
	Label       string
	Branch      string
	Datetime    string
	Project     string
	NextRel     string
	Platform    string
	isValidated bool
}

func (e *Entry) validate() {
	if e.isValidated {
		return
	}

	// TODO: validate...
}

func (e *Entry) contains(vals []string) bool {
	eStr := e.String()
	for _, val := range vals {
		if strings.Contains(eStr, val) {
			return true
		}
	}

	return false
}

func (e *Entry) String() string {
	return strings.Join(
		[]string{
			e.Label,
			e.Branch,
			e.Datetime,
			fmt.Sprintf("%s-%s", e.Project, e.NextRel),
			e.Platform,
		},
		defaultEntrySeparator,
	)
}

// ----------------------------------------------------------------------

// TagsFile represents a tags file
type TagsFile struct {
	// The source of the tags file
	src string

	// The copy of the src file, on which edits are made
	bck string

	// The list of entries in the tags file
	entries *Entries
}

// createEntry creates a new tags file Entry from a given file line of text
func (t *TagsFile) createEntry(line string) (*Entry, error) {
	expected := []string{
		"vo-label",
		"branch",
		"timestamp",
		"project-nextRelease",
		"platform",
	}

	fields := strings.Split(line, defaultEntrySeparator)
	if len(fields) != len(expected) {
		err := fmt.Errorf(
			"badly formatted tags file line, expected %s, got %s",
			strings.Join(expected, defaultEntrySeparator),
			line,
		)
		return nil, err
	}

	toks := strings.Split(fields[3], "-")
	if len(toks) != 2 {
		err := fmt.Errorf(
			"badly formatted field, expected %s, got %s",
			expected[3],
			fields[3],
		)
		return nil, err
	}

	return &Entry{
		Label:    fields[0],
		Branch:   fields[1],
		Datetime: fields[2],
		Project:  toks[0],
		NextRel:  toks[1],
		Platform: fields[4],
	}, nil
}

// Src returns the path to the source tags file
func (t *TagsFile) Src() string {
	return t.src
}

// Remove will delete any tag file entries that contain
// any of the passed in strings
func (t *TagsFile) Remove(values ...string) error {
	if len(values) == 0 {
		return nil
	}

	// Have not yet loaded the file
	if t.entries == nil {
		if err := t.load(); err != nil {
			return fmt.Errorf("failed to load tags file (%w)", err)
		}
	}

	return t.entries.Remove(values)
}

// GetEntries returns the Entries object containing
// the list of tags file entries
func (t *TagsFile) GetEntries() *Entries {
	return t.entries
}

// Append will add the given entries to the end of the tags file
func (t *TagsFile) Append(entries *Entries) error {
	if entries == nil || len(*entries) == 0 {
		return nil
	}

	// Have not yet loaded the file
	if t.entries == nil {
		if err := t.load(); err != nil {
			return fmt.Errorf("failed to load tags file (%w)", err)
		}
	}

	t.entries.Append(entries)
	return nil
}

// Save will write out the entries in memory to the source tags file
func (t *TagsFile) Save() error {
	// TODO: check that the original src file has not
	// been updated compared to when we first coped it

	// First, dump the in-memory entries to the temp file.
	// Then copy that file to original tags file src.
	tmp := fs.File(t.bck)
	data := t.entries.AsLines()
	if err := tmp.WriteLines(data, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return fmt.Errorf("failed to open file %s for writing (%w)", t.bck, err)
	}

	return tmp.CopyTo(t.src)
}

func (t *TagsFile) load() error {
	if err := fs.File(t.src).CopyTo(t.bck); err != nil {
		return fmt.Errorf("failed to copy src tags file to %s (%w)", t.bck, err)
	}

	fd, err := os.Open(t.bck)
	if err != nil {
		return fmt.Errorf("unable to open tags file %s (%w)", t.bck, err)
	}

	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := t.createEntry(line)
		if err != nil {
			return err
		}

		*t.entries = append(*t.entries, entry)
	}

	return nil
}
