// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"os"
	"testing"
	"time"
)

func TestDocNowFallsBackOnInvalidSourceDateEpoch(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("SOURCE_DATE_EPOCH", "not-a-unix-timestamp")

	before := time.Now()
	got := docNow()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Fatalf("expected docNow to fall back to the current time, got %v (want between %v and %v)", got, before, after)
	}
}

func TestDocNowUsesValidSourceDateEpoch(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("SOURCE_DATE_EPOCH", "1700000000")

	got := docNow()
	want := time.Unix(1700000000, 0).UTC()

	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestDocNowUsesCurrentTimeWhenUnset(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Unsetenv("SOURCE_DATE_EPOCH")

	before := time.Now()
	got := docNow()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Fatalf("expected docNow to return the current time, got %v (want between %v and %v)", got, before, after)
	}
}
