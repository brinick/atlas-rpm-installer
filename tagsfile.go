package installer

type tagsfile struct {
	src string
	tmp string
}

func writeTagsFile(src string, tmp string, []string) error {

	// Push out a new tags file (local copy) that has no bad entries
	src, err := os.Open(inst.tags.Src())
	if err != nil {
		return fmt.Errorf("unable to open tags file for reading %s (%w)", inst.tags.Src(), err)
	}
	defer src.Close()

	tmpFile := filepath.Join(os.Getenv("HOME"), "AMItags")
	tmp, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {

	}

	defer tmp.Close()

	if err = copyLines(src, tmp, lineReject(ignore)); err != nil {
		return err
	}

	dir := fs.Directory(
		inst.opts.InstallBaseDir,
		inst.opts.Branch,
		inst.opts.Timestamp,
	)

	projdir := dir.Append(inst.opts.Project)
	subdirs, err := projdir.SubDirs()
	if err != nil {
		return fmt.Errorf("failed to list sub-dirs of %s (%w)", projdir.Path, err)
	}

	if len(*subdirs) != 1 {
		return fmt.Errorf(
			"expected project dir (%s) to contain a single subdir, found %d",
			projdir.Path,
			len(*subdirs),
		)
	}

	baseRelease := subdirs.Names()[0]

	// TODO: push these into a configurable step
	entries := dir.Entries().Not(".cvmfscatalog*", "*.ayum.log")
	var lines []string
	for _, entry := range *entries {
		tfe := tagsfile.Entry{
			Label:    "VO-atlas-nightly",
			Branch:   inst.opts.Branch,
			Datetime: inst.opts.Timestamp,
			Project:  inst.opts.Project,
			BaseRel:  baseRelease,
			Platform: inst.opts.Platform,
		}

		lines = append(lines, tfe.String())
	}

	// output the file
	return writeLines(tmp, lines)
}