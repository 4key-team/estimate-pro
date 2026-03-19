package pagination

import (
	"net/http/httptest"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
	}{
		{"valid", "?page=2&limit=10", 2, 10},
		{"defaults", "", 1, 20},
		{"page zero", "?page=0&limit=10", 1, 10},
		{"negative page", "?page=-1&limit=10", 1, 10},
		{"limit zero", "?page=1&limit=0", 1, 20},
		{"limit over 100", "?page=1&limit=200", 1, 20},
		{"limit exactly 100", "?page=1&limit=100", 1, 100},
		{"non-numeric", "?page=abc&limit=xyz", 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/"+tt.query, nil)
			p := Parse(req)
			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", p.Limit, tt.wantLimit)
			}
		})
	}
}

func TestParams_Offset(t *testing.T) {
	tests := []struct {
		page, limit, want int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 10, 20},
		{1, 100, 0},
		{5, 5, 20},
	}

	for _, tt := range tests {
		p := Params{Page: tt.page, Limit: tt.limit}
		if got := p.Offset(); got != tt.want {
			t.Errorf("Params{%d,%d}.Offset() = %d, want %d", tt.page, tt.limit, got, tt.want)
		}
	}
}
