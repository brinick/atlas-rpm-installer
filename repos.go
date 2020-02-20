package installer

import (
	"path/filepath"

	"github.com/brinick/atlas-rpm-installer/pkg/rpm"
)

const (
	baseURL     = "http://cern.ch/atlas-software-dist-eos/RPMs"
	tdaqBaseURL = "http://cern.ch/atlas-tdaq-sw/yum"
	lcgURL      = "http://lcgpackages.web.cern.ch/lcgpackages/rpms"
)

func (inst *Installer) getRemoteRepos() []*rpm.Repo {
	return []*rpm.Repo{
		&rpm.Repo{
			Label: "atlas-offline-data",
			Name:  "ATLAS offline data packages",
			URL:   filepath.Join(baseURL, "data"),
		},

		&rpm.Repo{
			Label:  "lcg",
			Name:   "LCG Repository",
			URL:    lcgURL,
			Prefix: filepath.Join(inst.opts.InstallBaseDir, "sw/lcg/releases"),
		},

		&rpm.Repo{
			Label: "tdaq-nightly",
			Name:  "Nightly snapshots of TDAQ releases",
			URL:   filepath.Join(tdaqBaseURL, "tdaq/nightly"),
		},

		&rpm.Repo{
			Label: "tdaq-testing",
			Name:  "Non-official updates and patches for TDAQ releases",
			URL:   filepath.Join(tdaqBaseURL, "tdaq/testing"),
		},

		&rpm.Repo{
			Label: "dqm-common-testing",
			Name:  "dqm-common projects",
			URL:   filepath.Join(tdaqBaseURL, "dqm-common/testing"),
		},

		&rpm.Repo{
			Label: "dqm-common-testing",
			Name:  "dqm-common projects centos7",
			URL:   filepath.Join(tdaqBaseURL, "dqm-common/centos7"),
		},

		&rpm.Repo{
			Label: "tdaq-common-testing",
			Name:  "Non-official updates and patches for TDAQ releases",
			URL:   filepath.Join(tdaqBaseURL, "tdaq-common/testing"),
		},

		&rpm.Repo{
			Label: "tdaq-common-testing",
			Name:  "tdaq-common projects centos7",
			URL:   filepath.Join(tdaqBaseURL, "tdaq-common/centos7"),
		},

		&rpm.Repo{
			Label:  "atlas-offline-nightly",
			Name:   "ATLAS offline nightly releases",
			URL:    inst.rpms.SrcDir(),
			Prefix: filepath.Join(inst.opts.InstallBaseDir, inst.opts.Timestamp),
		},
	}
}
