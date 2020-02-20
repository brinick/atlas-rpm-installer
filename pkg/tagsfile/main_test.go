package tagsfile

import (
	"strings"
	"testing"
)

func TestEntryContains(t *testing.T) {
	e := Entry{
		Label:    "VO-atlas-nightly",
		Branch:   "21.3",
		Datetime: "2019-10-27T0347",
		Project:  "AnalysisBase",
		NextRel:  "21.3.16",
		Platform: "x86_64-centos7-gcc8-opt",
	}

	var entryTests = []struct {
		name   string
		search []string
		expect bool
	}{
		{name: "empty list", expect: false},
		{name: "single item empty string", search: []string{""}, expect: true},
		{name: "single item matching within word", search: []string{"centos"}, expect: true},
		{name: "single item matching exact word", search: []string{"21.3.16"}, expect: true},
		{name: "single item matching across fields", search: []string{"nightly;21.3"}, expect: true},
		{name: "single item not matching", search: []string{"Missing"}, expect: false},
		{name: "first item ok, next not matching", search: []string{"", "Missing"}, expect: true},
		{name: "first item not matching next ok", search: []string{"Missing", ""}, expect: true},
		{name: "only items not matching", search: []string{"Missing", "Missing2"}, expect: false},
	}

	for _, tt := range entryTests {
		t.Run(tt.name, func(t *testing.T) {
			if e.contains(tt.search) != tt.expect {
				t.Errorf(
					"%s should return %t, got %t",
					tt.name,
					tt.expect,
					!tt.expect,
				)
			}
		})
	}
}

func TestCreateBadEntry(t *testing.T) {
	tf := TagsFile{}
	_, err := tf.createEntry("bad-line")
	if err == nil {
		t.Errorf("badly formatted entry line should provoke error, got nil")
	}

	msg := err.Error()
	if !strings.HasPrefix(msg, "badly formatted") {
		t.Errorf("badly formatted entry line should provoke bad format error, got %s", msg)

	}

	line := strings.Join(
		[]string{
			"VO-atlas-nightly",
			"21.3",
			"2019-10-27T0347",
			"AnalysisBase-21.3.16",
			"x86_64-centos7-gcc8-opt",
		},
		";",
	)
	_, err = tf.createEntry(line)
	if !strings.HasPrefix(err.Error(), "badly formatted field") {
		t.Errorf("Error is %s", err.Error())
	}
}
