package pkginstaller

import (
	"fmt"

	"github.com/brinick/atlas-rpm-installer/pkg/pkginstaller/ayum"
	"github.com/brinick/atlas-rpm-installer/pkg/pkginstaller/dnf"
)

func choose(name string) func(opts, logpath) PkgInstaller {
	switch name {
	case "ayum":
		return ayum.New
	case "dnf":
		panic("dnf installer not yet implemented")
	default:
		panic(fmt.Sprintf("%s: unknown package installer"))
	}
}

type PkgInstaller interface {
}
