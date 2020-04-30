package tagsfile_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/brinick/atlas-rpm-installer/pkg/tagsfile"
)

func tempDir() string {
	dir, err := ioutil.TempDir("", "atlas_rpm_installer_tagsfile")
	if err != nil {
		panic(
			fmt.Sprintf(
				"unable to make a temporary directory: %v",
				err,
			),
		)
	}

	return dir
}

func tempFile() *os.File {
	d := tempDir()
	f, err := ioutil.TempFile(d, "temp")
	if err != nil {
		panic(err)
	}
	return f
}

func TestRemoveEntry(t *testing.T) {
	f := tempFile()
	defer os.Remove(f.Name())
	f.Close()

	d := tempDir()
	defer os.RemoveAll(d)

	tf := tagsfile.New(f.Name(), d)

	if err := tf.Remove("this should provoke an error"); err == nil {
		t.Errorf("removing an element on an empty tagsfile should return an error, but did not")
	}

	var entries tagsfile.Entries
	entries.Add(&tagsfile.Entry{Label: "be the change"})
	entries.Add(&tagsfile.Entry{Label: "you wish to see in the world"})
	if err := tf.Append(&entries); err != nil {
		t.Fatal(err)
	}

	if tf.Size() != 2 {
		t.Errorf("incorrect tagsfile size, expected 2, got %d", tf.Size())
	}

	if err := tf.Remove("change"); err != nil {
		t.Errorf("unable to remove entry from tags file, got error: %v", err)
	}

	if tf.Size() != 1 {
		t.Errorf("incorrect tagsfile size, expected 1, got %d", tf.Size())
	}

	if err := tf.Remove("this"); err != nil {
		t.Errorf("unable to remove entry from tags file, got error: %v", err)
	}

	if tf.Size() != 1 {
		t.Errorf("incorrect tagsfile size, expected 1, got %d", tf.Size())
	}

	if err := tf.Remove("to see"); err != nil {
		t.Errorf("unable to remove entry from tags file, got error: %v", err)
	}

	if tf.Size() != 0 {
		t.Errorf("incorrect tagsfile size, expected 0, got %d", tf.Size())
	}
}
