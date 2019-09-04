package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	//Global holds some meta information
	Global *global

	// Ayum holds info for configuring the ayum installation
	Ayum *ayum

	// Paths configures directories in which to do work
	Paths *paths

	// CVMFS configures related info
	CVMFS *cvmfs

	// Repo configures RPM repository locations
	Repo *rpmRepo

	// EOS configures EOS paths
	EOS *eos
)

func init() {
	Global = &global{
		// After how many seconds should the entire install attempt stop (0 = no timeout)
		Timeout:            0,
		SudoUser:           "cvatlasnightlies",
		TagsFieldSeparator: ";",

		// Who to send emails on failure
		EmailOnFail: []string{
			"oana.boeriu@cern.ch",
		},
	}

	Paths = &paths{
		WorkBase: os.Getenv("HOME"),
	}

	Ayum = &ayum{
		// Src repo for ayum
		Repo: "https://gitlab.cern.ch/atlas-sit/ayum.git",

		// Abort download attempt after this many seconds
		GitCloneTimeout: 60,
	}

	CVMFS = &cvmfs{
		NightlyRepo:           "atlas-nightlies.cern.ch",
		StableRepoDir:         "/cvmfs/atlas.cern.ch/repo/sw/software",
		GatewayNode:           "lxcvmfs78.cern.ch",
		MaxTransitionAttempts: 3,
	}

	EOS = &eos{
		BaseDir:        "/eos/project/a/atlas-software-dist/www/RPMs/",
		NightlyBaseDir: "/eos/project/a/atlas-software-dist/www/RPMs/nightlies",
	}

	Repo = &rpmRepo{
		Base: "http://cern.ch/atlas-software-dist-eos/RPMs/",
		TDAQ: &tdaqRepo{Base: "http://cern.ch/atlas-tdaq-sw/yum"},
		LCG:  "http://lcgpackages.web.cern.ch/lcgpackages/rpms",
	}

	// Post configuration - values that depend on other config values
	Paths.InstallBase = fmt.Sprintf("/cvmfs/%s/repo/sw", CVMFS.NightlyRepo)
	Paths.Logs = filepath.Join(Paths.WorkBase, "logs")

	Repo.NightlyBase = filepath.Join(Repo.Base, "nightlies")
	Repo.Data = filepath.Join(Repo.Base, "data")

	Repo.LocalSimulation = filepath.Join(Repo.Base, "local/simulation")
	Repo.LocalAtlasDQM = filepath.Join(Repo.Base, "local/atlasdqm")
	Repo.LocalAtlasTest = filepath.Join(Repo.Base, "local/atlas_test")
	Repo.LocalSMH = filepath.Join(Repo.Base, "local/smh")
	Repo.Local = Repo.LocalSimulation

	Repo.TDAQ.Nightly = filepath.Join(Repo.TDAQ.Base, "nightly")
	Repo.TDAQ.Testing = filepath.Join(Repo.TDAQ.Base, "testing")
	Repo.TDAQ.Centos7 = filepath.Join(Repo.TDAQ.Base, "centos7")
	Repo.TDAQCommon.Testing = filepath.Join(Repo.TDAQ.Base, "tdaq-common/testing")
	Repo.TDAQCommon.Centos7 = filepath.Join(Repo.TDAQ.Base, "tdaq-common/centos7")

	Repo.DQMCommon.Testing = filepath.Join(Repo.TDAQ.Base, "dqm-common/testing")
	Repo.DQMCommon.Centos7 = filepath.Join(Repo.TDAQ.Base, "dqm-common/centos7")

}

// ----------------------------------------------------------------------

type ayum struct {
	Repo            string
	GitCloneTimeout int
}

type global struct {
	Timeout  int
	SudoUser string

	// The character used to separate the fields in tag file entries
	TagsFieldSeparator string

	// Who to email upon problems
	EmailOnFail []string
}

type paths struct {
	InstallBase string
	WorkBase    string
	Logs        string
}

type rpmRepo struct {
	Base        string
	NightlyBase string

	TDAQ       *tdaqRepo
	TDAQCommon *tdaqcommonRepo
	DQMCommon  *dqmcommonRepo

	LocalSimulation string
	LocalAtlasDQM   string
	LocalAtlasTest  string
	LocalSMH        string
	Local           string

	Data string
	LCG  string
}

type tdaqRepo struct {
	Base    string
	Nightly string
	Testing string
	Centos7 string
}

type tdaqcommonRepo struct {
	Testing string
	Centos7 string
}

type dqmcommonRepo struct {
	Testing string
	Centos7 string
}

type eos struct {
	// Directory root below which we store RPMs
	BaseDir string

	// Directory root below which we store nightly RPMs
	NightlyBaseDir string
}

type cvmfs struct {
	// Name of the nightly repo
	NightlyRepo string

	// Path where we store fixed number software releases
	StableRepoDir string

	// Gateway Machine to access CVMFS
	GatewayNode string

	// How many times we try to open our own CVMFS transaction
	MaxTransitionAttempts int
}
