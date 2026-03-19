package domain

import (
	"math"
	"testing"
)

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"draft is valid", StatusDraft, true},
		{"submitted is valid", StatusSubmitted, true},
		{"empty is invalid", Status(""), false},
		{"unknown is invalid", Status("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("Status.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstimation_IsSubmitted(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"draft", StatusDraft, false},
		{"submitted", StatusSubmitted, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Estimation{Status: tt.status}
			if got := e.IsSubmitted(); got != tt.want {
				t.Errorf("IsSubmitted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstimationItem_PERTHours(t *testing.T) {
	tests := []struct {
		name   string
		min    float64
		likely float64
		max    float64
		want   float64
	}{
		{
			name:   "symmetric estimate",
			min:    2, likely: 4, max: 6,
			want: 4.0, // (2+16+6)/6 = 4
		},
		{
			name:   "skewed optimistic",
			min:    1, likely: 2, max: 9,
			want: 3.0, // (1+8+9)/6 = 3
		},
		{
			name:   "all zeros",
			min:    0, likely: 0, max: 0,
			want: 0.0,
		},
		{
			name:   "same values",
			min:    5, likely: 5, max: 5,
			want: 5.0,
		},
		{
			name:   "real-world estimate",
			min:    4, likely: 8, max: 20,
			want: 9.333333333333334, // (4+32+20)/6
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &EstimationItem{
				MinHours:    tt.min,
				LikelyHours: tt.likely,
				MaxHours:    tt.max,
			}
			got := item.PERTHours()
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("PERTHours() = %v, want %v", got, tt.want)
			}
		})
	}
}
