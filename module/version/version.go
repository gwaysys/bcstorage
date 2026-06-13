package version

import (
	"github.com/gwaysys/goapp/version"
)

func BuildVersion() string {
	return "v0.0.1-" + version.GitCommit
}
