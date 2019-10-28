package installer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// rejectLine is a simple search through the provided
// search patterns for a match against the given line.
// Boolean returned indicates if a match was found.
func rejectLine(line string) bool {

	for _, sp := range searchPatts {
		if strings.Contains(line, sp) {
			return true
		}
	}
	return false
}

func lineReject(search []string) lineRejector {
	return func(line string) bool {
		for _, patt := range search {
			if strings.Contains(line, patt) {
				return true
			}
		}
		return false
	}
}

type lineRejector func(string) bool

// copyLines transfers those lines from the src to the tgt
// that pass the reject function
func copyLines(src io.Reader, tgt io.Writer, reject lineRejector) error {
	reader := bufio.NewReader(src)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return fmt.Errorf("failed to read line (%w)", err)
			}
		}

		if !reject(line) {
			_, err = tgt.Write([]byte(line))
			if err != nil {
				return fmt.Errorf("failed to write line (%w)", err)
			}
		}
	}

	return nil
}

func writeLines(w io.StringWriter, lines []string) error {
	var err error
	for _, line := range lines {
		if _, err = w.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("unable to write line %s (%w)", line, err)
		}
	}

	return nil
}

func removeDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove dir %s (%w)", dir, err)
		}
	}

	return nil
}
