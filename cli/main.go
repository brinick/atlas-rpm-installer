package cli

import "flag"

func Parse() (*Args, error) {
	var args Args
	args.Parse()
	args.Validate()
	return args
}

type Args struct {
	Install *installOpts
	Global  *globalOpts
	Ayum    *ayumOpts
	CVMFS   *cvmfsOpts
}

func (a *Args) flags() {
	a.Global.flags()
	a.Install.flags()
	a.CVMFS.flags()
	a.Ayum.flags()
	flag.Parse()
}

func (a *Args) Validate() {

}

type cvmfsOpts struct {
	// Name of the nightly repo
	NightlyRepo string

	// Path where we store fixed number software releases
	StableRepoDir string

	// Gateway Machine to access CVMFS
	GatewayNode string

	// How many times we try to open our own CVMFS transaction
	MaxTransitionAttempts int
}

func (c *cvmfsOpts) flags() {
	flag.StringVar(&c.NightlyRepo)
}

func (c *cvmfsOpts) validate() {
	//
}

type installOpts struct {
	Release   string
	Branch    string
	Platform  string
	Timestamp string
	Project   string
}

type ayumOpts struct {
	Repo            string
	GitCloneTimeOut int
}
type globalOpts struct {
	TimeOut int
}
