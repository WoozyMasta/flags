package flags

import (
	"fmt"
	"testing"
)

func TestRequiredOptionSliceMinFail(t *testing.T) {
	var opts struct {
		Tag []string `long:"tag" required:"2"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"--tag", "one"})

	assertError(
		t,
		err,
		ErrRequired,
		fmt.Sprintf(
			"the required flag `%stag (at least 2 values, but got only 1)` was not specified",
			defaultLongOptDelimiter,
		),
	)
}

func TestRequiredOptionSliceMinPass(t *testing.T) {
	var opts struct {
		Tag []string `long:"tag" required:"2"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"--tag", "one", "--tag", "two"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	assertStringArray(t, opts.Tag, []string{"one", "two"})
}

func TestRequiredOptionSliceRangeFail(t *testing.T) {
	var opts struct {
		Tag []string `long:"tag" required:"1-2"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"--tag", "one", "--tag", "two", "--tag", "three"})

	assertError(
		t,
		err,
		ErrRequired,
		fmt.Sprintf(
			"the required flag `%stag (at most 2 values, but got 3)` was not specified",
			defaultLongOptDelimiter,
		),
	)
}

func TestRequiredOptionSliceZeroRangeAllowsMissing(t *testing.T) {
	var opts struct {
		Tag []string `long:"tag" required:"0-0"`
	}

	p := NewParser(&opts, None)
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

func TestRequiredOptionSliceZeroRangeRejectsValue(t *testing.T) {
	var opts struct {
		Tag []string `long:"tag" required:"0-0"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"--tag", "one"})

	assertError(
		t,
		err,
		ErrRequired,
		fmt.Sprintf(
			"the required flag `%stag (zero values)` was not specified",
			defaultLongOptDelimiter,
		),
	)
}

func TestRequiredOptionMapMinPass(t *testing.T) {
	var opts struct {
		Label map[string]string `long:"label" required:"2"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"--label", "one=1", "--label", "two=2"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(opts.Label) != 2 {
		t.Fatalf("expected two labels, got %d", len(opts.Label))
	}
}

func TestRequiredOptionScalarKeepsBooleanOne(t *testing.T) {
	var opts struct {
		Name string `long:"name" required:"1"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs(nil)

	assertError(
		t,
		err,
		ErrRequired,
		fmt.Sprintf("the required flag `%sname` was not specified", defaultLongOptDelimiter),
	)
}

func TestRequiredOptionNumericRejectsScalar(t *testing.T) {
	var opts struct {
		Count int `long:"count" required:"2"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs(nil)

	assertErrorTypeAndMessageContains(
		t,
		err,
		ErrInvalidTag,
		"numeric required ranges are only supported for slice or map options",
	)
}

func TestRequiredOptionInvalidRange(t *testing.T) {
	var opts struct {
		Tag []string `long:"tag" required:"3-2"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs(nil)

	assertErrorTypeAndMessageContains(t, err, ErrInvalidTag, "invalid required range `3-2`")
}

func TestOptionSetRequiredRange(t *testing.T) {
	var opts struct {
		Tag []string `long:"tag"`
	}

	p := NewParser(&opts, None)
	opt := p.FindOptionByLongName("tag")
	if opt == nil {
		t.Fatalf("expected option")
	}
	if err := opt.SetRequiredRange(2, -1); err != nil {
		t.Fatalf("unexpected SetRequiredRange error: %v", err)
	}

	_, err := p.ParseArgs([]string{"--tag", "one"})
	assertError(
		t,
		err,
		ErrRequired,
		fmt.Sprintf(
			"the required flag `%stag (at least 2 values, but got only 1)` was not specified",
			defaultLongOptDelimiter,
		),
	)
}
