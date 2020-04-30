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
// the src is first copied to the given bckupDir directory, and the edits
// are made on the copy. It is up to the client to request a Write
// to push these changes back to the src.
func New(src string, bckupDir string) *TagsFile {
	now := time.Now().UnixNano()
	return &TagsFile{
		src: fs.NewFile(src),
		bck: fs.NewFile(filepath.Join(bckupDir, fmt.Sprintf("AMItags.%d", now))),
	}
}

// ---------------------------------------------------------------------

// Entries represents a list of Entry instances
type Entries []*Entry

// Size returns the number of Entry elements in this slice
func (e *Entries) Size() int {
	return len(*e)
}

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
func (e *Entries) Remove(values []string) {
	var newE = (*e)[:0]
	for _, entry := range *e {
		if !entry.contains(values) {
			newE = append(newE, entry)
		}
	}

	*e = newE
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
	src *fs.File

	// Last modification time of the src tags file
	srcModTime time.Time

	// The copy of the src file, on which edits are made
	bck *fs.File

	// The list of entries in the tags file
	entries *Entries
}

// Size returns the number of entries/lines in this tagsfile
// If the file has not yet been loaded, 0 will be returned.
func (t *TagsFile) Size() int {
	if t.entries == nil {
		return 0
	}
	return t.entries.Size()
}

// Src returns the path to the source tags file
func (t *TagsFile) Src() *fs.File {
	return t.src
}

// Remove will delete any tag file entries that contain
// any of the passed in strings
func (t *TagsFile) Remove(values ...string) error {
	if len(values) == 0 {
		return nil
	}

	if t.entries == nil {
		return fmt.Errorf("tagsfile not loaded yet, cannot remove entries")
	}

	t.entries.Remove(values)
	return nil
}

// GetEntries returns the Entries object containing
// the list of tags file entries
func (t *TagsFile) GetEntries() *Entries {
	return t.entries
}

// Add appends a single tagsfile Entry onto this tags file
func (t *TagsFile) Add(entry *Entry) error {
	if entry == nil {
		return nil
	}

	// Have not yet loaded the file
	if t.entries == nil {
		if err := t.load(); err != nil {
			return err
		}
	}

	t.entries.Add(entry)
	return nil
}

// Append will add the given entries to the end of the tags file
func (t *TagsFile) Append(entries *Entries) error {
	if entries == nil || len(*entries) == 0 {
		return nil
	}

	// Have not yet loaded the file
	if t.entries == nil {
		if err := t.load(); err != nil {
			return err
		}
	}

	t.entries.Append(entries)
	return nil
}

// Save will write out the entries in memory to the source tags file
func (t *TagsFile) Save() error {
	// First, dump the in-memory entries to the temp file.
	// Then copy that file to original tags file src.
	data := t.entries.AsLines()
	// TODO: save empty file will fail?
	if err := t.bck.WriteLines(data); err != nil {
		return fmt.Errorf("failed to open file %s for writing (%w)", t.bck.Path, err)
	}

	// Check now that the original file has not changed since
	// we first loaded it
	mod, err := t.src.ModTime()
	if err != nil {
		return fmt.Errorf("unable to check if tags file has been updated (%w)", err)
	}

	if mod.After(t.srcModTime) {
		return fmt.Errorf("source tags file (%s) has changed, will not overwrite it", t.src)
	}

	return t.bck.RenameTo(t.src.Path)
}

func (t *TagsFile) load() error {
	if err := t.checkSrcExists(); err != nil {
		return err
	}

	if err := t.saveSrcModTime(); err != nil {
		return err
	}

	if err := t.backupSrc(); err != nil {
		return err
	}

	fd, err := os.Open(t.bck.Path)
	if err != nil {
		return fmt.Errorf("unable to open backup tags file %s (%w)", t.bck, err)
	}

	defer fd.Close()

	entries, err := t.getEntriesFromFile(fd)
	if err != nil {
		return err
	}

	t.entries = entries
	return nil
}

func (t *TagsFile) checkSrcExists() error {
	exists, err := t.src.Exists()
	if err != nil {
		return fmt.Errorf("unable to check if tagsfile exists: %w", err)
	}

	if !exists {
		return fs.InexistantError{Path: t.src.Path}
	}

	return nil
}

func (t *TagsFile) backupSrc() error {
	if err := t.src.ExportTo(t.bck.Path); err != nil {
		return fmt.Errorf("failed to back up src tags file (%s) to %s: %w", t.src.Path, t.bck.Path, err)
	}

	return nil
}

func (t *TagsFile) saveSrcModTime() error {
	md, err := t.src.ModTime()
	if err != nil {
		err = fmt.Errorf("failed to get src tags file last modification time (%v)", err)
		return err
	}

	t.srcModTime = *md
	return nil
}

func (t *TagsFile) getEntriesFromFile(fd *os.File) (*Entries, error) {
	var (
		entries Entries
		err     error
	)
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := t.createEntry(line)
		if err != nil {
			break
		}

		entries = append(entries, entry)
	}

	return &entries, err
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
