package version

import (
	"os/exec"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	// Release version of the service. Bump it up during new version release
	Release = "1.5.0"
	// Commit hash provided during build
	Commit = "Unknown"
	// BuildTime provided during build
	BuildTime = "Unknown"
	// GO provides golang version
	GO = runtime.Version()
	// Distro provides extra information regarding the os distribution
	Distro = "Unknown"
	// Compiler info
	Compiler = runtime.Compiler
	// OS Info
	OS = runtime.GOOS
	// Arch info
	Arch = runtime.GOARCH
)

func init() {
	Distro = distro()
}

// LogInfo can be used to log version information during startup or termination etc...
func LogInfo() {

	log.WithFields(
		log.Fields{
			"type":         "service_log",
			"release":      Release,
			"commit":       Commit,
			"build_time":   BuildTime,
			"go":           GO,
			"compiler":     Compiler,
			"os":           OS,
			"distro":       Distro,
			"architecture": "Arch",
		},
	).Infof("Running Argo Messaging v%s (%s/%s)", Release, OS, Arch)
}

func distro() string {
	// command line to execute in order to retrieve the os distribution in linux systems(ubuntu/centos)
	if OS == "linux" {
		cmd := exec.Command("cat", "/etc/redhat-release")
		stdout, err := cmd.Output()
		if err != nil {
			cmd = exec.Command("lsb_release", "-ds")
			stdout, err = cmd.Output()
			if err != nil {
				return ""
			}
		}
		return strings.TrimSuffix(string(stdout), "\n")

		// command line to execute in order to retrieve the os distribution in osx
	} else if OS == "darwin" {
		cmd := exec.Command("sw_vers", "-productName")
		stdout, err := cmd.Output()
		if err != nil {
			return ""
		}

		pName := strings.TrimSuffix(string(stdout), "\n")

		cmd = exec.Command("sw_vers", "-productVersion")
		stdout, err = cmd.Output()
		if err != nil {
			return ""
		}

		pVersion := strings.TrimSuffix(string(stdout), "\n")

		return pName + " " + pVersion
	}

	return ""
}

// Model struct holds version information about the binary build
type Model struct {
	BuildTime string `xml:"build_time" json:"build_time"`
	GO        string `xml:"golang" json:"golang"`
	Compiler  string `xml:"compiler" json:"compiler"`
	OS        string `xml:"os" json:"os"`
	Arch      string `xml:"architecture" json:"architecture"`
	Release   string `xml:"release,omitempty" json:"release,omitempty"`
	Distro    string `xml:"distro,omitempty" json:"distro,omitempty"`
	Hostname  string `xml:"hostname,omitempty" json:"hostname,omitempty"`
}
