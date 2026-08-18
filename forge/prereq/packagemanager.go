package prereq

import "github.com/mgmaster24/config-gen-tools/forge/runner"

// PackageManager identifies a system package manager forge tools know how to
// suggest install commands for.
type PackageManager string

const (
	PMBrew   PackageManager = "brew"
	PMApt    PackageManager = "apt"
	PMDnf    PackageManager = "dnf"
	PMPacman PackageManager = "pacman"
	PMWinget PackageManager = "winget"
	PMScoop  PackageManager = "scoop"
	PMChoco  PackageManager = "choco"
)

// InstallHint is one "run this command with this package manager"
// suggestion attached to a Check.
type InstallHint struct {
	Manager PackageManager
	Command string
}

// pmProbeOrder lists, per GOOS, which package managers are worth probing
// for and in what priority order.
var pmProbeOrder = map[string][]PackageManager{
	"darwin":  {PMBrew},
	"linux":   {PMApt, PMDnf, PMPacman, PMBrew},
	"windows": {PMWinget, PMScoop, PMChoco},
}

// pmProbeBinary is the binary on PATH that indicates a PackageManager is
// installed.
var pmProbeBinary = map[PackageManager]string{
	PMBrew:   "brew",
	PMApt:    "apt-get",
	PMDnf:    "dnf",
	PMPacman: "pacman",
	PMWinget: "winget",
	PMScoop:  "scoop",
	PMChoco:  "choco",
}

// DetectPackageManagers returns every PackageManager relevant to goos that
// is actually present on PATH, in priority order.
func DetectPackageManagers(r runner.Runner, goos string) []PackageManager {
	var found []PackageManager
	for _, pm := range pmProbeOrder[goos] {
		if _, ok := r.LookPath(pmProbeBinary[pm]); ok {
			found = append(found, pm)
		}
	}
	return found
}
