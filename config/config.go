// Package config retrieves configuration options.
package config

import (
	"path"
	"runtime/debug"
)

// Name returns the go.mod package name.
func Name() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return path.Base(buildInfo.Main.Path)
}

// Version returns the module version or git SHA commit id.
func Version() string {
	rev := "n/a"
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return rev
	}
	for _, k := range buildInfo.Settings {
		if k.Key == "vcs.revision" {
			return "(" + k.Value + ")"
		}
	}
	return buildInfo.Main.Version
}

// Repo returns the repository path.
func Repo() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return buildInfo.Main.Path
}
