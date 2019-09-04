package cli

import (
	"flag"
)

type ayum struct {
	Repo            string
	GitCloneTimeout int
}

func ayum() *ayum {
	flag.StringVar(&ayum.Repo, "ayum.repo")

}
