package svc

import "testing"

func TestValidateDate(t *testing.T) {
	tests := []struct {
		in    string
		valid bool
	}{
		{"2026-01-01", true},
		{"2026-12-31", true},
		{"", true}, // optional
		{"2026-13-01", false},
		{"2026-00-01", false},
		{"not-a-date", false},
		{"2026/01/01", false},
		{"2026-02-30", false},
	}
	for _, tt := range tests {
		err := ValidateDate(tt.in)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateDate(%q): err=%v, valid=%v", tt.in, err, tt.valid)
		}
	}
}

func TestValidateInterestType(t *testing.T) {
	tests := []struct {
		in    string
		valid bool
	}{
		{"", true},
		{"simple", true},
		{"compound", true},
		{"SIMPLE", false},
		{"compound2", false},
		{"annual", false},
	}
	for _, tt := range tests {
		err := ValidateInterestType(tt.in)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateInterestType(%q): err=%v, valid=%v", tt.in, err, tt.valid)
		}
	}
}

func TestValidateFrequency(t *testing.T) {
	tests := []struct {
		in    string
		valid bool
	}{
		{"", true},
		{"monthly", true},
		{"biweekly", true},
		{"weekly", true},
		{"daily", false},
		{"yearly", false},
		{"MONTHLY", false},
	}
	for _, tt := range tests {
		err := ValidateFrequency(tt.in)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateFrequency(%q): err=%v, valid=%v", tt.in, err, tt.valid)
		}
	}
}

func TestValidateAmountPositive(t *testing.T) {
	tests := []struct {
		in    float64
		valid bool
	}{
		{100, true},
		{0.01, true},
		{0, false},
		{-1, false},
		{-100.50, false},
	}
	for _, tt := range tests {
		err := ValidateAmountPositive(tt.in)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateAmountPositive(%v): err=%v, valid=%v", tt.in, err, tt.valid)
		}
	}
}
