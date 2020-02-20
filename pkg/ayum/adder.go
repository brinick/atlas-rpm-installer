package ayum

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/brinick/atlas-rpm-installer/pkg/rpm"
)

type rpmRepoAdder interface {
	AddRemoteRepos([]*rpm.Repo) error
}

type rpmRepoAdd struct {
	basedir string
}

// AddRemoteRepos configures the ayum installation with the provided remote repositories
func (r *rpmRepoAdd) AddRemoteRepos(repos []*rpm.Repo) error {
	for _, repo := range repos {
		repoConf := filepath.Join(r.basedir, "ayum/etc/yum.repos.d", repo.Filename())
		if err := ioutil.WriteFile(repoConf, []byte(repo.String()), 0774); err != nil {
			return fmt.Errorf("could not configure remote repo %s (%w)", repo.Filename(), err)
		}
	}

	return nil
}
