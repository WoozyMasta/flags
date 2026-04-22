// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package flags

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// VersionFields configures which metadata fields are rendered by WriteVersion.
type VersionFields uint

const (
	// VersionFieldFile is executable path/name (`os.Args[0]`).
	VersionFieldFile VersionFields = 1 << iota
	// VersionFieldVersion is module/application version.
	VersionFieldVersion
	// VersionFieldCommit is VCS revision (commit SHA).
	VersionFieldCommit
	// VersionFieldBuilt is VCS revision time.
	VersionFieldBuilt
	// VersionFieldURL is repository URL.
	VersionFieldURL
	// VersionFieldPath is main package path from build info.
	VersionFieldPath
	// VersionFieldModule is main module path from build info.
	VersionFieldModule
	// VersionFieldModified is dirty tree marker from build info.
	VersionFieldModified
	// VersionFieldGoVersion is Go toolchain version used for build.
	VersionFieldGoVersion
	// VersionFieldTarget is build target in GOOS/GOARCH form.
	VersionFieldTarget
)

const (
	// VersionFieldsCore is compact default output.
	VersionFieldsCore VersionFields = VersionFieldFile |
		VersionFieldVersion |
		VersionFieldCommit |
		VersionFieldBuilt |
		VersionFieldURL
	// VersionFieldsAll enables all known version fields.
	VersionFieldsAll VersionFields = VersionFieldFile |
		VersionFieldVersion |
		VersionFieldCommit |
		VersionFieldBuilt |
		VersionFieldURL |
		VersionFieldPath |
		VersionFieldModule |
		VersionFieldModified |
		VersionFieldGoVersion |
		VersionFieldTarget
)

// VersionInfo represents build/version metadata of the running binary.
type VersionInfo struct {
	// File is executable path/name (usually os.Args[0]).
	File string
	// Path is the main package path of the running binary.
	Path string
	// ModulePath is the main module path.
	ModulePath string
	// Version is the module version.
	Version string
	// Revision is VCS revision (commit SHA).
	Revision string
	// RevisionTime is VCS revision time in UTC when available.
	RevisionTime time.Time
	// Modified reports whether source tree was dirty at build time.
	Modified bool
	// GoVersion is the Go toolchain version used for build.
	GoVersion string
	// URL is repository URL inferred from module path unless overridden.
	URL string
	// GOOS is build target operating system.
	GOOS string
	// GOARCH is build target architecture.
	GOARCH string
}

// VersionInfo returns detected build metadata merged with parser-level overrides.
func (p *Parser) VersionInfo() VersionInfo {
	base := ReadVersionInfo()

	if p.versionInfo.File != "" {
		base.File = p.versionInfo.File
	}
	if p.versionInfo.Path != "" {
		base.Path = p.versionInfo.Path
	}
	if p.versionInfo.ModulePath != "" {
		base.ModulePath = p.versionInfo.ModulePath
	}
	if p.versionInfo.Version != "" {
		base.Version = p.versionInfo.Version
	}
	if p.versionInfo.Revision != "" {
		base.Revision = p.versionInfo.Revision
	}
	if !p.versionInfo.RevisionTime.IsZero() {
		base.RevisionTime = p.versionInfo.RevisionTime
	}
	if p.versionInfo.GoVersion != "" {
		base.GoVersion = p.versionInfo.GoVersion
	}
	if p.versionInfo.URL != "" {
		base.URL = p.versionInfo.URL
	}
	if p.versionInfo.GOOS != "" {
		base.GOOS = p.versionInfo.GOOS
	}
	if p.versionInfo.GOARCH != "" {
		base.GOARCH = p.versionInfo.GOARCH
	}
	if p.versionInfo.Modified {
		base.Modified = true
	}

	if base.File == "" && len(os.Args) > 0 {
		base.File = os.Args[0]
	}
	if base.Path == "" {
		base.Path = p.Name
	}

	return base
}

// SetVersionInfo sets parser-level version metadata overrides.
func (p *Parser) SetVersionInfo(info VersionInfo) {
	p.versionInfo = info
}

// SetVersion sets parser-level version override.
func (p *Parser) SetVersion(version string) {
	p.versionInfo.Version = version
}

// SetVersionCommit sets parser-level VCS revision override.
func (p *Parser) SetVersionCommit(revision string) {
	p.versionInfo.Revision = revision
}

// SetVersionTime sets parser-level VCS revision time override.
func (p *Parser) SetVersionTime(t time.Time) {
	p.versionInfo.RevisionTime = t.UTC()
}

// SetVersionURL sets parser-level repository URL override.
func (p *Parser) SetVersionURL(url string) {
	p.versionInfo.URL = url
}

// SetVersionTarget sets parser-level GOOS/GOARCH overrides.
func (p *Parser) SetVersionTarget(goos string, goarch string) {
	p.versionInfo.GOOS = goos
	p.versionInfo.GOARCH = goarch
}

// SetVersionFields configures fields rendered by built-in version output.
func (p *Parser) SetVersionFields(fields VersionFields) {
	p.versionFields = fields
}

// WriteVersion writes version/build metadata in human-readable format.
func (p *Parser) WriteVersion(w io.Writer, fields VersionFields) {
	if fields == 0 {
		fields = p.versionFields
		if fields == 0 {
			fields = VersionFieldsCore
		}
	}

	info := p.VersionInfo()

	version := info.Version
	if version == "" {
		version = "unknown"
	}

	commit := info.Revision
	if commit == "" {
		commit = "unknown"
	}

	built := "unknown"
	if !info.RevisionTime.IsZero() {
		built = info.RevisionTime.Format(time.RFC3339)
	}

	file := info.File
	if file == "" {
		file = "unknown"
	}

	url := info.URL
	if url == "" {
		url = "unknown"
	}

	path := info.Path
	if path == "" {
		path = "unknown"
	}

	module := info.ModulePath
	if module == "" {
		module = "unknown"
	}

	goVersion := info.GoVersion
	if goVersion == "" {
		goVersion = "unknown"
	}

	target := "unknown/unknown"
	if info.GOOS != "" || info.GOARCH != "" {
		goos := info.GOOS
		goarch := info.GOARCH
		if goos == "" {
			goos = "unknown"
		}
		if goarch == "" {
			goarch = "unknown"
		}
		target = goos + "/" + goarch
	}

	if (fields & VersionFieldFile) != 0 {
		_, _ = fmt.Fprintf(w, "file:     %s\n", file)
	}
	if (fields & VersionFieldVersion) != 0 {
		_, _ = fmt.Fprintf(w, "version:  %s\n", version)
	}
	if (fields & VersionFieldCommit) != 0 {
		_, _ = fmt.Fprintf(w, "commit:   %s\n", commit)
	}
	if (fields & VersionFieldBuilt) != 0 {
		_, _ = fmt.Fprintf(w, "built:    %s\n", built)
	}
	if (fields & VersionFieldURL) != 0 {
		_, _ = fmt.Fprintf(w, "url:      %s\n", url)
	}
	if (fields & VersionFieldPath) != 0 {
		_, _ = fmt.Fprintf(w, "path:     %s\n", path)
	}
	if (fields & VersionFieldModule) != 0 {
		_, _ = fmt.Fprintf(w, "module:   %s\n", module)
	}
	if (fields & VersionFieldModified) != 0 {
		_, _ = fmt.Fprintf(w, "modified: %t\n", info.Modified)
	}
	if (fields & VersionFieldGoVersion) != 0 {
		_, _ = fmt.Fprintf(w, "go:       %s\n", goVersion)
	}
	if (fields & VersionFieldTarget) != 0 {
		_, _ = fmt.Fprintf(w, "target:   %s\n", target)
	}
}

// ReadVersionInfo reads build metadata of the running binary.
func ReadVersionInfo() VersionInfo {
	info := VersionInfo{}
	if len(os.Args) > 0 {
		info.File = os.Args[0]
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil {
		return info
	}

	info.Path = bi.Path
	info.ModulePath = bi.Main.Path
	info.Version = bi.Main.Version
	info.GoVersion = bi.GoVersion
	info.GOOS = runtime.GOOS
	info.GOARCH = runtime.GOARCH

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			info.Revision = s.Value
		case "vcs.time":
			if t, err := time.Parse(time.RFC3339, s.Value); err == nil {
				info.RevisionTime = t.UTC()
			}
		case "vcs.modified":
			info.Modified = s.Value == "true"
		case "GOOS":
			info.GOOS = s.Value
		case "GOARCH":
			info.GOARCH = s.Value
		}
	}

	if info.ModulePath != "" {
		info.URL = "https://" + strings.TrimPrefix(info.ModulePath, "https://")
	}

	return info
}
