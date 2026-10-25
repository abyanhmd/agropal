package post

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestSlugFor(t *testing.T) {
	id := uuid.MustParse("0198a1b2-c3d4-7e5f-8a6b-9c0d1e2f3a4b")

	cases := []struct {
		title string
		want  string
	}{
		{"Best cover crops for clay soil?", "best-cover-crops-for-clay-soil-1e2f3a4b"},
		{"  Drip   irrigation vs. soaker hoses ", "drip-irrigation-vs-soaker-hoses-1e2f3a4b"},
		{"100% organic!!!", "100-organic-1e2f3a4b"},
		{"Café — Ünïcode", "caf-n-code-1e2f3a4b"},
		{"???", "post-1e2f3a4b"},
		{"", "post-1e2f3a4b"},
	}
	for _, tc := range cases {
		if got := slugFor(tc.title, id); got != tc.want {
			t.Errorf("slugFor(%q) = %q, want %q", tc.title, got, tc.want)
		}
	}

	t.Run("a long title is cut and still ends cleanly", func(t *testing.T) {
		got := slugFor(strings.Repeat("word ", 60), id)
		if len(got) > maxSlugBase+len("-1e2f3a4b") || strings.Contains(got, "--") {
			t.Errorf("slugFor(long) = %q (%d chars), want a short slug without doubled dashes", got, len(got))
		}
	})

	t.Run("the same title on two posts gives two slugs", func(t *testing.T) {
		a, b := slugFor("Hello", uuid.New()), slugFor("Hello", uuid.New())
		if a == b {
			t.Errorf("both posts got %q", a)
		}
	})
}
