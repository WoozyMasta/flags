// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
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

var (
	readBuildInfoOnce sync.Once
	cachedVersionInfo VersionInfo
)

// VersionInfo represents build/version metadata of the running binary.
type VersionInfo struct {
	// RevisionTime is VCS revision time in UTC when available.
	RevisionTime time.Time
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
	// GoVersion is the Go toolchain version used for build.
	GoVersion string
	// URL is repository URL inferred from module path unless overridden.
	URL string
	// GOOS is build target operating system.
	GOOS string
	// GOARCH is build target architecture.
	GOARCH string
	// Modified reports whether source tree was dirty at build time.
	Modified bool
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

	prevHelpColorEnabled := p.helpColorEnabled
	p.helpColorEnabled = (p.Options&ColorHelp) != None && DetectColorSupport(w)
	defer func() {
		p.helpColorEnabled = prevHelpColorEnabled
	}()

	basePrefix := ""
	padToTerminalWidth := false
	terminalColumns := 0
	if (p.Options&ColorHelp) != None && p.helpColorEnabled {
		basePrefix = helpStylePrefix(p.helpColorScheme.BaseText)
		if basePrefix != "" {
			_, _ = io.WriteString(w, basePrefix)
		}
		if p.helpColorScheme.BaseText.UseBG {
			terminalColumns = p.helpColumns()
			padToTerminalWidth = terminalColumns > 0
		}
	}

	version := info.Version
	if version == "" {
		version = p.i18nText("version.value.unknown", "unknown")
	}

	commit := info.Revision
	if commit == "" {
		commit = p.i18nText("version.value.unknown", "unknown")
	}

	built := p.i18nText("version.value.unknown", "unknown")
	if !info.RevisionTime.IsZero() {
		built = info.RevisionTime.Format(time.RFC3339)
	}

	file := info.File
	if file == "" {
		file = p.i18nText("version.value.unknown", "unknown")
	}

	url := info.URL
	if url == "" {
		url = p.i18nText("version.value.unknown", "unknown")
	}

	path := info.Path
	if path == "" {
		path = p.i18nText("version.value.unknown", "unknown")
	}

	module := info.ModulePath
	if module == "" {
		module = p.i18nText("version.value.unknown", "unknown")
	}

	goVersion := info.GoVersion
	if goVersion == "" {
		goVersion = p.i18nText("version.value.unknown", "unknown")
	}

	targetUnknown := p.i18nText("version.value.unknown", "unknown")
	target := targetUnknown + "/" + targetUnknown
	if info.GOOS != "" || info.GOARCH != "" {
		goos := info.GOOS
		goarch := info.GOARCH
		if goos == "" {
			goos = targetUnknown
		}
		if goarch == "" {
			goarch = targetUnknown
		}
		target = goos + "/" + goarch
	}

	fileLabel := p.i18nText("version.label.file", "file")
	versionLabel := p.i18nText("version.label.version", "version")
	commitLabel := p.i18nText("version.label.commit", "commit")
	builtLabel := p.i18nText("version.label.built", "built")
	urlLabel := p.i18nText("version.label.url", "url")
	pathLabel := p.i18nText("version.label.path", "path")
	moduleLabel := p.i18nText("version.label.module", "module")
	modifiedLabel := p.i18nText("version.label.modified", "modified")
	goLabel := p.i18nText("version.label.go", "go")
	targetLabel := p.i18nText("version.label.target", "target")

	writeVersionLine := func(label string, value string) {
		padded := label + ":"
		if pad := 9 - textWidth(padded); pad > 0 {
			padded += strings.Repeat(" ", pad)
		}

		colored := p.colorizeHelp(padded, p.helpColorScheme.VersionLabel)
		coloredValue := p.colorizeHelp(value, p.helpColorScheme.VersionValue)

		_, _ = io.WriteString(w, colored)
		_, _ = io.WriteString(w, " ")
		_, _ = io.WriteString(w, coloredValue)

		if padToTerminalWidth {
			lineWidth := textWidth(padded) + 1 + textWidth(value)
			if pad := terminalColumns - lineWidth; pad > 0 {
				_, _ = io.WriteString(w, strings.Repeat(" ", pad))
			}
		}

		_, _ = io.WriteString(w, "\n")
	}

	if (fields & VersionFieldFile) != 0 {
		writeVersionLine(fileLabel, file)
	}
	if (fields & VersionFieldVersion) != 0 {
		writeVersionLine(versionLabel, version)
	}
	if (fields & VersionFieldCommit) != 0 {
		writeVersionLine(commitLabel, commit)
	}
	if (fields & VersionFieldBuilt) != 0 {
		writeVersionLine(builtLabel, built)
	}
	if (fields & VersionFieldURL) != 0 {
		writeVersionLine(urlLabel, url)
	}
	if (fields & VersionFieldPath) != 0 {
		writeVersionLine(pathLabel, path)
	}
	if (fields & VersionFieldModule) != 0 {
		writeVersionLine(moduleLabel, module)
	}
	if (fields & VersionFieldModified) != 0 {
		writeVersionLine(modifiedLabel, strconv.FormatBool(info.Modified))
	}
	if (fields & VersionFieldGoVersion) != 0 {
		writeVersionLine(goLabel, goVersion)
	}
	if (fields & VersionFieldTarget) != 0 {
		writeVersionLine(targetLabel, target)
	}

	if (p.Options&ColorHelp) != None && p.helpColorEnabled {
		writeANSIReset(w)
	}
}

// ReadVersionInfo reads build metadata of the running binary.
func ReadVersionInfo() VersionInfo {
	readBuildInfoOnce.Do(func() {
		info := VersionInfo{
			GOOS:   runtime.GOOS,
			GOARCH: runtime.GOARCH,
		}

		bi, ok := debug.ReadBuildInfo()
		if ok && bi != nil {
			info.Path = bi.Path
			info.ModulePath = bi.Main.Path
			info.Version = bi.Main.Version
			info.GoVersion = bi.GoVersion

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
		}

		cachedVersionInfo = info
	})

	info := cachedVersionInfo
	if len(os.Args) > 0 {
		info.File = os.Args[0]
	}

	return info
}
