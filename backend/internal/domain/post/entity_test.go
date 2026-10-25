package post

import (
	"agropal/internal/domain/shared"
	"testing"
)

func TestValidSort(t *testing.T) {
	cases := []struct {
		name    string
		sort    *shared.SortQuery
		allowed map[string]struct{}
		want    bool
	}{
		{"no sort", nil, PostSortColumns, true},
		{"allowed column ascending", &shared.SortQuery{Column: "like_count", Dir: "asc"}, PostSortColumns, true},
		{"allowed column descending", &shared.SortQuery{Column: "created_at", Dir: "desc"}, PostSortColumns, true},
		{"unknown column", &shared.SortQuery{Column: "password", Dir: "asc"}, PostSortColumns, false},
		{"injected column", &shared.SortQuery{Column: "created_at; DROP TABLE posts", Dir: "asc"}, PostSortColumns, false},
		{"injected direction", &shared.SortQuery{Column: "created_at", Dir: "asc, (SELECT 1)"}, PostSortColumns, false},
		{"missing direction", &shared.SortQuery{Column: "created_at"}, PostSortColumns, false},
		{"column of the other entity", &shared.SortQuery{Column: "comment_count", Dir: "asc"}, CommentSortColumns, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidSort(tc.sort, tc.allowed); got != tc.want {
				t.Errorf("ValidSort(%+v) = %v, want %v", tc.sort, got, tc.want)
			}
		})
	}
}
