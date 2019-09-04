package rpm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	rpm "github.com/cavaliercoder/go-rpm"
	"github.com/pkg/errors"
)

// Repos is a collection of RPM repo instances
type Repos []Repo

// Repo represents an RPM repository
type Repo struct {
	Name    string
	Label   string
	URL     string
	Prefix  string
	Enabled bool
}

// Filename returns the file name into which this repo will write its description
func (r *Repo) Filename() string {
	return fmt.Sprintf("%s.repo", r.Label)
}
func (r *Repo) String() string {
	var tokens []string
	tokens = append(tokens, fmt.Sprintf("[%s]", r.Label))
	tokens = append(tokens, fmt.Sprintf("name=%s", r.Name))
	tokens = append(tokens, fmt.Sprintf("baseurl=%s", r.URL))
	tokens = append(tokens, fmt.Sprintf("enabled=%t", r.Enabled))
	if len(r.Prefix) > 0 {
		tokens = append(tokens, fmt.Sprintf("prefix=%s", r.Prefix))
	}
	return strings.Join(tokens, "\n") + "\n"
}

// ---------------------------------------------------------------------

// Finder is the RPM locator thingy
type Finder struct {
	Basedir string
}

func (f *Finder) findTopRPM(project, platform string) (string, error) {
	// Find the top RPM which we need to install (with its dependencies)
	fname := fmt.Sprintf("%s_*_%s.rpm", project, platform)
	fpath := filepath.Join(f.Basedir, fname)
	matches, err := filepath.Glob(fpath)
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no top RPM found to install (%s)", fpath)
	}

	return matches[0], nil
}

// Find is the method that finds RPMs
func (f *Finder) Find(project, platform string) ([]*RPM, error) {
	path, err := f.findTopRPM(project, platform)
	if err != nil {
		return nil, err
	}

	topRPM, err := New(path)
	if topRPM.Size == 0 {
		return nil, fmt.Errorf("Top RPM has zero size: %s", path)
	}

	deps, err := topRPM.LocalDependencies()
	if err != nil {
		return nil, err
	}

	// Ensure that no dependencies have zero size, else fail
	var zero []string
	for _, dep := range deps {
		if dep.Size == 0 {
			zero = append(zero, filepath.Base(dep.Path))
		}
	}

	if len(zero) > 0 {
		return nil, fmt.Errorf("rpm dependencies in %s found with zero size: %s", path, strings.Join(zero, ", "))
	}

	return append([]*RPM{topRPM}, deps...), nil
}

// ---------------------------------------------------------------------

// New creates an RPM instance for the RPM at path
func New(path string) (*RPM, error) {
	size, err := fileSize(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rpm file size")
	}

	return &RPM{Path: path, Size: size}, nil
}

// ---------------------------------------------------------------------

// RPMs is a collection of RPM instances
type RPMs []*RPM

// Paths returns the paths to each of the RPM instances
func (r *RPMs) Paths() []string {
	var paths []string
	for _, rpm := range *r {
		paths = append(paths, rpm.Path)
	}

	return paths
}

// Names returns the names of each of the RPM instances
func (r *RPMs) Names() []string {
	var names []string
	for _, rpm := range *r {
		names = append(names, filepath.Base(rpm.Path))
	}

	return names
}

// ---------------------------------------------------------------------

// RPM is the basic wrapper around the given RPM path
type RPM struct {
	Path string
	Size int64
}

// LocalDependencies finds only those dependencies
// that are in the same directory as the RPM
func (r *RPM) LocalDependencies() ([]*RPM, error) {
	deps, err := listDeps(r.Path)
	if err != nil {
		return nil, err
	}

	deps, err = listDir(filepath.Dir(r.Path), deps)
	if err != nil {
		return nil, err
	}

	var localdeps []*RPM
	for _, dep := range deps {
		depPath := filepath.Join(filepath.Dir(r.Path), dep)
		fi, err := os.Stat(depPath)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get file size for dependency %s", depPath)
		}
		depSize := fi.Size()
		localdeps = append(localdeps, &RPM{depPath, depSize})
	}

	return localdeps, nil
}

// listDeps is a helper function to get the names of
// dependencies of a given starting root RPM
func listDeps(path string) ([]string, error) {
	p, err := rpm.OpenPackageFile(path)
	if err != nil {
		return nil, err
	}

	deps := p.Requires()
	names := make([]string, len(deps))
	for _, dep := range deps {
		names = append(names, dep.Name())
	}

	return names, nil
}

func listDir(dir string, filenames []string) ([]string, error) {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	lut := toLUT(filenames)

	var found []string
	for _, entry := range entries {
		name := entry.Name()
		if _, keyExists := lut[name]; keyExists && !entry.IsDir() {
			found = append(found, name)
		}
	}

	return found, nil
}

func toLUT(items []string) map[string]struct{} {
	var m map[string]struct{}
	for _, item := range items {
		m[item] = struct{}{}
	}

	return m
}

func fileSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return fi.Size(), nil
}
