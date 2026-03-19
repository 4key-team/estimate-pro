package parser

import (
	"math"
	"testing"
)

func TestParse_ValidCSV(t *testing.T) {
	input := `Backend API,4,8,16
Frontend UI,2,4,8
Database migrations,1,2,4`

	items, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	tests := []struct {
		taskName    string
		min, likely, max float64
		sortOrder   int
	}{
		{"Backend API", 4, 8, 16, 0},
		{"Frontend UI", 2, 4, 8, 1},
		{"Database migrations", 1, 2, 4, 2},
	}

	for i, tt := range tests {
		item := items[i]
		if item.TaskName != tt.taskName {
			t.Errorf("item[%d].TaskName = %q, want %q", i, item.TaskName, tt.taskName)
		}
		if item.MinHours != tt.min {
			t.Errorf("item[%d].MinHours = %v, want %v", i, item.MinHours, tt.min)
		}
		if item.LikelyHours != tt.likely {
			t.Errorf("item[%d].LikelyHours = %v, want %v", i, item.LikelyHours, tt.likely)
		}
		if item.MaxHours != tt.max {
			t.Errorf("item[%d].MaxHours = %v, want %v", i, item.MaxHours, tt.max)
		}
		if item.SortOrder != tt.sortOrder {
			t.Errorf("item[%d].SortOrder = %d, want %d", i, item.SortOrder, tt.sortOrder)
		}
	}
}

func TestParse_WithHeader(t *testing.T) {
	input := `task_name,min_hours,likely_hours,max_hours
Backend API,4,8,16
Frontend UI,2,4,8`

	items, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items (header skipped), got %d", len(items))
	}
	if items[0].TaskName != "Backend API" {
		t.Errorf("first item = %q, want Backend API", items[0].TaskName)
	}
}

func TestParse_WithNotes(t *testing.T) {
	input := `Backend API,4,8,16,includes auth endpoints
Frontend UI,2,4,8`

	items, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Note != "includes auth endpoints" {
		t.Errorf("note = %q, want %q", items[0].Note, "includes auth endpoints")
	}
	if items[1].Note != "" {
		t.Errorf("note should be empty, got %q", items[1].Note)
	}
}

func TestParse_DecimalHours(t *testing.T) {
	input := `Task A,1.5,3.25,8.75`

	items, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(items[0].MinHours-1.5) > 1e-9 {
		t.Errorf("MinHours = %v, want 1.5", items[0].MinHours)
	}
	if math.Abs(items[0].LikelyHours-3.25) > 1e-9 {
		t.Errorf("LikelyHours = %v, want 3.25", items[0].LikelyHours)
	}
}

func TestParse_SkipsEmptyLines(t *testing.T) {
	input := `Backend API,4,8,16

Frontend UI,2,4,8

`

	items, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items (empty lines skipped), got %d", len(items))
	}
}

func TestParse_EmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParse_InvalidHours(t *testing.T) {
	input := `Backend API,abc,8,16`

	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected error for non-numeric hours")
	}
}

func TestParse_TooFewColumns(t *testing.T) {
	input := `Backend API,4,8`

	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected error for too few columns")
	}
}

func TestParse_NegativeHours(t *testing.T) {
	input := `Backend API,-1,8,16`

	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected error for negative hours")
	}
}

func TestParse_MinGreaterThanLikely(t *testing.T) {
	input := `Backend API,10,8,16`

	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected error when min > likely")
	}
}

func TestParse_LikelyGreaterThanMax(t *testing.T) {
	input := `Backend API,4,20,16`

	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected error when likely > max")
	}
}

func TestParse_TrimSpaces(t *testing.T) {
	input := `  Backend API  ,  4  ,  8  ,  16  `

	items, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items[0].TaskName != "Backend API" {
		t.Errorf("TaskName = %q, want %q", items[0].TaskName, "Backend API")
	}
}
