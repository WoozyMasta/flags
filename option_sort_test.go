package flags

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestOptionSortDefaultDeclarationOrder(t *testing.T) {
	var opts struct {
		B string `long:"b"`
		A string `long:"a"`
	}

	p := NewParser(&opts, None)
	groups := p.Groups()
	if len(groups) == 0 {
		t.Fatalf("expected at least one group")
	}

	options := groups[0].Options()
	if len(options) != 2 {
		t.Fatalf("expected two options, got %d", len(options))
	}

	if options[0].LongName != "b" || options[1].LongName != "a" {
		t.Fatalf("expected declaration order [b a], got [%s %s]", options[0].LongName, options[1].LongName)
	}
}

func TestOptionSortByNameAscDesc(t *testing.T) {
	var opts struct {
		B string `long:"b"`
		A string `long:"a"`
	}

	p := NewParser(&opts, None)
	groups := p.Groups()
	if len(groups) == 0 {
		t.Fatalf("expected at least one group")
	}
	g := groups[0]

	p.SetOptionSort(OptionSortByNameAsc)
	options := g.Options()
	if options[0].LongName != "a" || options[1].LongName != "b" {
		t.Fatalf("expected name asc [a b], got [%s %s]", options[0].LongName, options[1].LongName)
	}

	p.SetOptionSort(OptionSortByNameDesc)
	options = g.Options()
	if options[0].LongName != "b" || options[1].LongName != "a" {
		t.Fatalf("expected name desc [b a], got [%s %s]", options[0].LongName, options[1].LongName)
	}
}

func TestOptionSortOrderPriority(t *testing.T) {
	var opts struct {
		Mid    string `long:"mid"`
		TopA   string `long:"top-a" order:"100"`
		Bottom string `long:"bottom" order:"-100"`
		TopB   string `long:"top-b" order:"10"`
	}

	p := NewParser(&opts, None)
	p.SetOptionSort(OptionSortByNameAsc)

	groups := p.Groups()
	if len(groups) == 0 {
		t.Fatalf("expected at least one group")
	}

	options := groups[0].Options()
	got := []string{options[0].LongName, options[1].LongName, options[2].LongName, options[3].LongName}
	expected := []string{"top-a", "top-b", "mid", "bottom"}

	assertStringArray(t, got, expected)
}

func TestOptionSortByTypeAndCustomTypeOrder(t *testing.T) {
	var opts struct {
		S string        `long:"string"`
		B bool          `long:"bool"`
		N int           `long:"number"`
		D time.Duration `long:"duration"`
		L []string      `long:"list"`
		C func()        `long:"custom"`
	}

	p := NewParser(&opts, None)
	p.SetOptionSort(OptionSortByType)

	groups := p.Groups()
	if len(groups) == 0 {
		t.Fatalf("expected at least one group")
	}
	g := groups[0]

	options := g.Options()
	got := []string{
		options[0].LongName,
		options[1].LongName,
		options[2].LongName,
		options[3].LongName,
		options[4].LongName,
		options[5].LongName,
	}
	expected := []string{"bool", "number", "string", "duration", "list", "custom"}
	assertStringArray(t, got, expected)

	if err := p.SetOptionTypeOrder([]OptionTypeClass{OptionTypeString, OptionTypeBool}); err != nil {
		t.Fatalf("unexpected custom type order error: %v", err)
	}
	options = g.Options()
	got = []string{
		options[0].LongName,
		options[1].LongName,
	}
	expected = []string{"string", "bool"}
	assertStringArray(t, got, expected)
}

func TestOptionSortSetOptionTypeOrderValidation(t *testing.T) {
	p := NewNamedParser("test", None)

	if err := p.SetOptionTypeOrder([]OptionTypeClass{OptionTypeString, OptionTypeString}); err == nil {
		t.Fatalf("expected duplicate type order validation error")
	}
}

func TestOptionSortInvalidOrderTag(t *testing.T) {
	var opts struct {
		Value string `long:"value" order:"abc"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestOptionSortAppliedToHelp(t *testing.T) {
	var opts struct {
		B string `long:"b"`
		A string `long:"a"`
	}

	p := NewParser(&opts, None)
	p.SetOptionSort(OptionSortByNameAsc)

	var out bytes.Buffer
	p.WriteHelp(&out)
	help := out.String()

	aIdx := strings.Index(help, defaultLongOptDelimiter+"a")
	bIdx := strings.Index(help, defaultLongOptDelimiter+"b")
	if aIdx < 0 || bIdx < 0 {
		t.Fatalf("expected both options in help output")
	}
	if aIdx > bIdx {
		t.Fatalf("expected option a before b in help output")
	}
}

func TestOptionSortAppliedToCompletion(t *testing.T) {
	var opts struct {
		Mid    string `long:"mid"`
		Top    string `long:"top" order:"10"`
		Bottom string `long:"bottom" order:"-10"`
	}

	p := NewParser(&opts, None)
	c := &completion{parser: p}

	input := defaultLongOptDelimiter
	items := c.complete([]string{input})
	if len(items) < 3 {
		t.Fatalf("expected at least three completion items, got %d", len(items))
	}

	got := []string{items[0].Item, items[1].Item, items[2].Item}
	expected := []string{
		defaultLongOptDelimiter + "top",
		defaultLongOptDelimiter + "mid",
		defaultLongOptDelimiter + "bottom",
	}

	assertStringArray(t, got, expected)
}
