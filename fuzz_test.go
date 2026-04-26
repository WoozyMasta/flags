// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func fuzzArgs(raw string) []string {
	if len(raw) > 512 {
		raw = raw[:512]
	}

	parts := strings.Fields(raw)
	if len(parts) > 32 {
		parts = parts[:32]
	}

	args := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ReplaceAll(part, "\x00", "")
		if part == "" {
			continue
		}
		if len(part) > 96 {
			part = part[:96]
		}
		args = append(args, part)
	}

	return args
}

func fuzzShell(b byte) CompletionShell {
	switch b % 4 {
	case 0:
		return CompletionShellBash
	case 1:
		return CompletionShellZsh
	case 2:
		return CompletionShellPwsh
	default:
		return CompletionShell("unknown-shell")
	}
}

func FuzzParseArgsMainFlow(f *testing.F) {
	f.Add("")
	f.Add("--verbose --count 2 deploy --force")
	f.Add("--mode safe --label a:b target artifact")
	f.Add("-- --literal --tokens")

	f.Fuzz(func(t *testing.T, raw string) {
		t.Setenv("GO_FLAGS_COMPLETION", "")

		var opts struct {
			Verbose bool     `short:"v" long:"verbose" description:"Verbose mode"`
			Count   int      `long:"count" description:"Counter"`
			Mode    string   `long:"mode" choices:"safe;fast;slow" description:"Mode"`
			Labels  []string `long:"label" description:"Labels"`
			Exec    []string `long:"exec" terminator:";" description:"Terminated args"`

			Deploy struct {
				Force bool `long:"force" description:"Force deploy"`
			} `command:"deploy" description:"Deploy command"`
		}

		p := NewNamedParser("fuzz-main", None|PassDoubleDash|AllowBoolValues|IgnoreUnknown)
		if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
			t.Fatalf("add group: %v", err)
		}

		args := fuzzArgs(raw)
		_, _ = p.ParseArgs(args)
		_, _ = p.ParseArgs(args)
		p.WriteHelp(&bytes.Buffer{})
	})
}

func FuzzParserRetagAndParse(f *testing.F) {
	f.Add("x-", "--value demo run")
	f.Add("cli-", "-v")
	f.Add("", "run")

	f.Fuzz(func(t *testing.T, prefix string, raw string) {
		t.Setenv("GO_FLAGS_COMPLETION", "")

		if len(prefix) > 8 {
			prefix = prefix[:8]
		}

		var opts struct {
			Verbose bool   `short:"v" long:"verbose"`
			Value   string `long:"value"`
			Run     struct {
				Dry bool `long:"dry-run"`
			} `command:"run"`
		}

		p := NewNamedParser("fuzz-retag", None|IgnoreUnknown)
		if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
			t.Fatalf("add group: %v", err)
		}

		_ = p.SetTagPrefix(prefix)
		_ = p.SetFlagTags(NewFlagTagsWithPrefix(prefix))
		_, _ = p.ParseArgs(fuzzArgs(raw))
	})
}

func FuzzWriteNamedCompletion(f *testing.F) {
	f.Add("app", byte(0))
	f.Add("demo-tool", byte(1))
	f.Add("cli", byte(2))
	f.Add("", byte(3))

	f.Fuzz(func(t *testing.T, commandName string, shellByte byte) {
		p := NewNamedParser("fuzz-completion", None)
		var out bytes.Buffer

		shell := fuzzShell(shellByte)
		_ = p.WriteNamedCompletion(&out, shell, commandName)
		_ = p.WriteCompletion(&out, CompletionShellBash)
		_ = p.WriteAutoCompletion(&out)
	})
}

func FuzzMultiTagParse(f *testing.F) {
	f.Add(`long:"value" short:"v"`)
	f.Add(`required:"true" choices:"a;b;c"`)
	f.Add(`bad`)

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 512 {
			raw = raw[:512]
		}

		mtag := newMultiTag(raw)
		_ = mtag.Parse()
		_ = mtag.Get("long")
		_ = mtag.GetMany("choice")

		key := fmt.Sprintf("k%d", len(raw)%7)
		mtag.Set(key, "value")
		mtag.SetMany(key, []string{"a", "b"})
	})
}

func FuzzINIReadWriteRoundTrip(f *testing.F) {
	f.Add("[Application Options]\nmode = safe\ncount = 2\n")
	f.Add("[Application Options]\nlabel = a:b\nlabel = c:d\n")
	f.Add("[deploy]\nforce = true\n")

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 2048 {
			raw = raw[:2048]
		}
		raw = strings.ReplaceAll(raw, "\x00", "")

		var opts struct {
			Mode  string   `long:"mode" choices:"safe;fast;slow" description:"Mode"`
			Count int      `long:"count" default:"1" description:"Count"`
			Label []string `long:"label" description:"Labels"`
			Path  string   `long:"path" description:"Path"`

			Deploy struct {
				Force bool `long:"force" description:"Force deploy"`
			} `command:"deploy" description:"Deploy command"`
		}

		p := NewNamedParser("fuzz-ini", None)
		if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
			t.Fatalf("add group: %v", err)
		}

		ip := NewIniParser(p)
		_ = ip.Parse(strings.NewReader(raw))

		var out bytes.Buffer
		ip.Write(&out, IniDefault|IniIncludeDefaults|IniCommentDefaults)
		ip.WriteExample(&out)

		var opts2 struct {
			Mode  string   `long:"mode" choices:"safe;fast;slow"`
			Count int      `long:"count" default:"1"`
			Label []string `long:"label"`
			Path  string   `long:"path"`

			Deploy struct {
				Force bool `long:"force"`
			} `command:"deploy"`
		}

		p2 := NewNamedParser("fuzz-ini-2", None)
		if _, err := p2.AddGroup("Application Options", "", &opts2); err != nil {
			t.Fatalf("add group 2: %v", err)
		}

		_ = NewIniParser(p2).Parse(strings.NewReader(out.String()))
	})
}

func FuzzHelpRenderingFlow(f *testing.F) {
	f.Add("--verbose build", byte(80), byte(0), byte(0), "simple desc")
	f.Add("--name alpha --count 7", byte(0), byte(6), byte(2), "long long long text")
	f.Add("deploy --force", byte(120), byte(2), byte(3), "unicode пример текста")

	f.Fuzz(func(t *testing.T, raw string, width byte, indent byte, sortMode byte, desc string) {
		t.Setenv("GO_FLAGS_COMPLETION", "")

		if len(desc) > 512 {
			desc = desc[:512]
		}
		desc = strings.ReplaceAll(desc, "\x00", "")

		var opts struct {
			Verbose bool     `short:"v" long:"verbose" description:"Verbose mode"`
			Name    string   `long:"name" description:"Name"`
			Count   int      `long:"count" description:"Count"`
			Choice  string   `long:"choice" choices:"alpha;bravo;charlie;delta" description:"Choice"`
			Labels  []string `long:"label" description:"Label values"`

			Deploy struct {
				Force bool `long:"force" description:"Force deploy"`
			} `command:"deploy" description:"Deploy command"`
		}

		p := NewNamedParser(
			"fuzz-help",
			None|ShowRepeatableInHelp|ShowChoiceListInHelp|DetectShellFlagStyle|DetectShellEnvStyle,
		)
		if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
			t.Fatalf("add group: %v", err)
		}

		_ = p.SetHelpWidth(int(width))
		_ = p.SetCommandOptionIndent(int(indent % 12))
		p.SetOptionSort(OptionSortMode(sortMode % 4))

		if opt := p.FindOptionByLongName("name"); opt != nil {
			opt.SetDescription(desc)
		}

		_, _ = p.ParseArgs(fuzzArgs(raw))

		var b bytes.Buffer
		p.WriteHelp(&b)
	})
}

func FuzzParserOptionsBitmask(f *testing.F) {
	f.Add(uint64(0), "--name demo")
	f.Add(uint64(Default), "--help")
	f.Add(uint64(HelpCommands), "completion")
	f.Add(^uint64(0), "--unknown --value=1")

	f.Fuzz(func(t *testing.T, mask uint64, raw string) {
		t.Setenv("GO_FLAGS_COMPLETION", "")

		var opts struct {
			Name  string `long:"name" description:"Name"`
			Value int    `long:"value" description:"Value"`
			Req   string `long:"req" required:"true" description:"Required"`

			Run struct {
				Dry bool `long:"dry-run" description:"Dry run"`
			} `command:"run" description:"Run command"`
		}

		p := NewNamedParser("fuzz-bits", Options(mask))
		if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
			t.Fatalf("add group: %v", err)
		}

		p.EnsureBuiltinOptions()
		_ = p.EnsureBuiltinCommands()
		_ = p.Validate()

		_, _ = p.ParseArgs(fuzzArgs(raw))
		p.WriteHelp(&bytes.Buffer{})
		p.WriteVersion(&bytes.Buffer{}, p.versionFields)
	})
}

func FuzzLegacyRepeatableTagsCompat(f *testing.F) {
	f.Add("--mode safe --token abc --profile dev")
	f.Add("-P qa --mode fast")
	f.Add("--legacy one")

	f.Fuzz(func(t *testing.T, raw string) {
		t.Setenv("GO_FLAGS_COMPLETION", "")

		var opts struct {
			Mode string `long:"mode" choice:"safe" choice:"fast" choice:"slow" description:"Legacy choice tag"`
			// Keep legacy repeatable alias tags intentionally in this fuzz target
			// to exercise backward compatibility paths.
			Profile string `long:"profile" alias:"legacy-profile" short-alias:"P" long-alias:"prof" description:"Legacy alias tags"`
			Token   string `long:"token" description:"Token"`
		}

		p := NewNamedParser("fuzz-legacy-tags", None|IgnoreUnknown)
		if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
			t.Fatalf("add group: %v", err)
		}

		_, _ = p.ParseArgs(fuzzArgs(raw))
		p.WriteHelp(&bytes.Buffer{})
	})
}
