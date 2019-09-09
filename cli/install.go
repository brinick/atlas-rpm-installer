package cli

import "flag"

type installOpts struct {
	Release   string
	Branch    string
	Platform  string
	Timestamp string
	Project   string
}

func (i *installOpts) flags() {
	flag.StringVar(&i.Release, "install.release", "", "The release to install")
	flag.StringVar(&i.Project, "install.project", "", "The project to install")
}
