package flags

import (
	"reflect"
	"testing"
)

func TestMultiTagCachedFallbackAndSetMany(t *testing.T) {
	mt := newMultiTag(`bad-tag-without-colon`)
	cache := mt.cached()

	if cache == nil {
		t.Fatalf("expected non-nil cache map")
	}

	if len(cache) != 0 {
		t.Fatalf("expected empty cache for invalid tag")
	}

	mt.SetMany("k", []string{"a", "b"})
	if got := mt.GetMany("k"); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("unexpected SetMany/GetMany result: %#v", got)
	}
}
