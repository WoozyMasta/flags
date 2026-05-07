// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"strings"
	"testing"
)

// assertCompletionFunc runs the completion engine with the given args against
// parser p and returns the Item strings of all returned completions.
func assertCompletionFunc(t *testing.T, p *Parser, args []string) []string {
	t.Helper()

	c := &completion{parser: p}
	res := c.complete(args)
	out := make([]string, len(res))
	for i, r := range res {
		out[i] = r.Item
	}
	return out
}

func assertContains(t *testing.T, got []string, want string) {
	t.Helper()
	for _, s := range got {
		if s == want {
			return
		}
	}
	t.Errorf("expected %q in completions %v", want, got)
}

func assertNotContains(t *testing.T, got []string, want string) {
	t.Helper()
	for _, s := range got {
		if s == want {
			t.Errorf("did not expect %q in completions %v", want, got)
			return
		}
	}
}

// TestCompletionFuncOption verifies that SetCompletionFunc on an Option is
// invoked during completion and its results are returned.
func TestCompletionFuncOption(t *testing.T) {
	var opts struct {
		Zone string `long:"zone"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByLongName("zone")
	if opt == nil {
		t.Fatal("option --zone not found")
	}

	zones := []string{"us-east-1", "us-west-2", "eu-central-1"}
	opt.SetCompletionFunc(func(match string) []Completion {
		var ret []Completion
		for _, z := range zones {
			if strings.HasPrefix(z, match) {
				ret = append(ret, Completion{Item: z})
			}
		}
		return ret
	})

	got := assertCompletionFunc(t, p, []string{"--zone", "us"})
	assertContains(t, got, "us-east-1")
	assertContains(t, got, "us-west-2")
	assertNotContains(t, got, "eu-central-1")
}

// TestCompletionFuncOptionEmptyMatch verifies that an empty match returns all
// completions from the callback.
func TestCompletionFuncOptionEmptyMatch(t *testing.T) {
	var opts struct {
		Env string `long:"env"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByLongName("env")

	envs := []string{"dev", "staging", "prod"}
	opt.SetCompletionFunc(func(match string) []Completion {
		ret := make([]Completion, 0, len(envs))
		for _, e := range envs {
			if strings.HasPrefix(e, match) {
				ret = append(ret, Completion{Item: e})
			}
		}
		return ret
	})

	got := assertCompletionFunc(t, p, []string{"--env", ""})
	assertContains(t, got, "dev")
	assertContains(t, got, "staging")
	assertContains(t, got, "prod")
}

// TestCompletionFuncPriorityOverChoices verifies that a registered callback
// takes priority over static Choices.
func TestCompletionFuncPriorityOverChoices(t *testing.T) {
	var opts struct {
		Mode string `long:"mode" choices:"fast;safe"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByLongName("mode")

	// Callback returns something completely different.
	opt.SetCompletionFunc(func(match string) []Completion {
		return []Completion{{Item: "dynamic"}}
	})

	got := assertCompletionFunc(t, p, []string{"--mode", ""})
	assertContains(t, got, "dynamic")
	assertNotContains(t, got, "fast")
	assertNotContains(t, got, "safe")
}

// TestCompletionFuncNilRemovesCallback verifies that passing nil clears a
// previously registered callback and falls back to Choices.
func TestCompletionFuncNilRemovesCallback(t *testing.T) {
	var opts struct {
		Mode string `long:"mode" choices:"fast;safe"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByLongName("mode")

	opt.SetCompletionFunc(func(match string) []Completion {
		return []Completion{{Item: "dynamic"}}
	})
	opt.SetCompletionFunc(nil)

	got := assertCompletionFunc(t, p, []string{"--mode", ""})
	assertContains(t, got, "fast")
	assertContains(t, got, "safe")
	assertNotContains(t, got, "dynamic")
}

// TestCompletionFuncArg verifies that SetCompletionFunc on a positional Arg
// is invoked during positional completion.
func TestCompletionFuncArg(t *testing.T) {
	var opts struct {
		Positional struct {
			Target string `positional-arg-name:"target"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, Default)

	var targetArg *Arg
	for _, a := range p.Command.args {
		if a.Name == "target" {
			targetArg = a
			break
		}
	}
	if targetArg == nil {
		t.Fatal("positional arg 'target' not found")
	}

	targets := []string{"prod", "staging", "dev"}
	targetArg.SetCompletionFunc(func(match string) []Completion {
		var ret []Completion
		for _, tgt := range targets {
			if strings.HasPrefix(tgt, match) {
				ret = append(ret, Completion{Item: tgt})
			}
		}
		return ret
	})

	got := assertCompletionFunc(t, p, []string{"p"})
	assertContains(t, got, "prod")
	assertNotContains(t, got, "staging")
	assertNotContains(t, got, "dev")
}
