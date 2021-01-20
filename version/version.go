package version

import (
	"runtime"

	log "github.com/sirupsen/logrus"
)

var (
	// Release version of the service. Bump it up during new version release
	Release = "1.0.7"
	// Commit hash provided during build
	Commit = "Unknown"
	// BuildTime provided during build
	BuildTime = "Unknown"
	// GO provides golang version
	GO = runtime.Version()
	// Compiler info
	Compiler = runtime.Compiler
	// OS Info
	OS = runtime.GOOS
	// Arch info
	Arch = runtime.GOARCH
)

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
			"architecture": "Arch",
		},
	).Infof("Running Argo Messaging v%s (%s/%s)", Release, OS, Arch)
}

// Model struct holds version information about the binary build
type Model struct {
	BuildTime string `xml:"build_time" json:"build_time"`
	GO        string `xml:"golang" json:"golang"`
	Compiler  string `xml:"compiler" json:"compiler"`
	OS        string `xml:"os" json:"os"`
	Arch      string `xml:"architecture" json:"architecture"`
}
