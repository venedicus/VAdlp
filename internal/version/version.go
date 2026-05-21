package version

import "fmt"

var (
	Version   = "dev"
	Commit    = ""
	BuildDate = ""
)

func Label() string {
	switch {
	case Commit != "" && BuildDate != "":
		return fmt.Sprintf("%s (%s, %s)", Version, Commit, BuildDate)
	case Commit != "":
		return fmt.Sprintf("%s (%s)", Version, Commit)
	default:
		return Version
	}
}
