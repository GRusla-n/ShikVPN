package version

import "fmt"

// Set via ldflags at build time.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func String() string {
	return fmt.Sprintf("ShikVPN %s (commit %s, built %s)", Version, Commit, BuildDate)
}
