package post

import (
	"strings"

	"github.com/google/uuid"
)

const maxSlugBase = 80

// slugFor builds the address of a post from its title, e.g. "Best cover crops?" becomes
// "best-cover-crops-3f9a1c2e". The tail comes from the post's own id, so two posts with the same
// title never share a slug and nothing has to check for a clash.
func slugFor(title string, postID uuid.UUID) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			dash = false
		} else {
			dash = true
		}
	}

	base := b.String()
	if len(base) > maxSlugBase {
		base = strings.TrimRight(base[:maxSlugBase], "-")
	}
	if base == "" {
		base = "post"
	}

	// The last bytes of a v7 id are random; the first ones are the time and would repeat
	id := strings.ReplaceAll(postID.String(), "-", "")
	return base + "-" + id[len(id)-8:]
}
