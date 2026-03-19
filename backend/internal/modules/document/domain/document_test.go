package domain

import "testing"

func TestFileType_IsValid(t *testing.T) {
	tests := []struct {
		ft   FileType
		want bool
	}{
		{FileTypePDF, true},
		{FileTypeDOCX, true},
		{FileTypeXLSX, true},
		{FileTypeMD, true},
		{FileTypeTXT, true},
		{FileTypeCSV, true},
		{"exe", false},
		{"jpg", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(string(tt.ft), func(t *testing.T) {
			if got := tt.ft.IsValid(); got != tt.want {
				t.Errorf("FileType(%q).IsValid() = %v, want %v", tt.ft, got, tt.want)
			}
		})
	}
}

func TestMaxFileSize(t *testing.T) {
	const expected int64 = 50 * 1024 * 1024
	if MaxFileSize != expected {
		t.Errorf("MaxFileSize = %d, want %d", MaxFileSize, expected)
	}
}
