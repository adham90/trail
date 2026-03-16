package cmd

import (
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

var version = buildVersion()

var rootCmd = &cobra.Command{
	Use:   "trail",
	Short: "A CLI planning tool for Claude Code",
	Long:  "trail keeps persistent plan files that bridge context between Claude Code sessions.",
}

func init() {
	rootCmd.Version = version
}

// buildVersion returns the module version embedded by Go at build time.
// Tagged releases (via go install ...@v1.0.0) return the clean tag.
// Local builds return dev-<commit>[-dirty].
func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	// Clean tagged version (e.g., "v0.1.0") — not pseudo-version, not "(devel)"
	v := info.Main.Version
	if v != "" && v != "(devel)" && !strings.HasPrefix(v, "v0.0.0-") {
		return v
	}

	// Fall back to git commit from VCS info
	var revision, dirty string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				dirty = "-dirty"
			}
		}
	}
	if revision != "" {
		if len(revision) > 8 {
			revision = revision[:8]
		}
		return "dev-" + revision + dirty
	}

	return "dev"
}

func Execute() error {
	return rootCmd.Execute()
}
