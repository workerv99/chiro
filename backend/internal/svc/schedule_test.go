package svc

import (
	"math"
	"testing"
)

func TestRound2(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{1.004, 1.00},
		{2.555, 2.56},
		{0.1 + 0.2, 0.3},
		{0, 0},
		{100.005, 100.01},
		{99.995, 100.00},
	}
	for _, tt := range tests {
		got := Round2(tt.in)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Round2(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestSplitAmounts(t *testing.T) {
	tests := []struct {
		total float64
		n     int
		sum   float64 // sum of all parts should equal total (within rounding)
	}{
		{100, 3, 100},
		{100, 1, 100},
		{100, 10, 100},
		{33.33, 3, 33.33},
		{1, 7, 1},
		{0, 5, 0},
	}
	for _, tt := range tests {
		parts := SplitAmounts(tt.total, tt.n)
		if len(parts) != tt.n {
			t.Errorf("SplitAmounts(%v, %d): got %d parts, want %d", tt.total, tt.n, len(parts), tt.n)
			continue
		}
		sum := 0.0
		for _, p := range parts {
			sum += p
		}
		if math.Abs(sum-tt.sum) > 0.01 {
			t.Errorf("SplitAmounts(%v, %d): sum = %v, want %v", tt.total, tt.n, sum, tt.sum)
		}
	}
}

func TestSplitAmounts_LastAbsorbsRounding(t *testing.T) {
	// 100 / 3 = 33.33 each, last absorbs: 33.34
	parts := SplitAmounts(100, 3)
	if parts[0] != 33.33 || parts[1] != 33.33 || parts[2] != 33.34 {
		t.Errorf("SplitAmounts(100, 3) = %v, want [33.33 33.33 33.34]", parts)
	}
}

func TestLoanTotal_Simple(t *testing.T) {
	// 1000 * (1 + 10/100) = 1100
	got := LoanTotal(1000, 10, 12, "simple")
	if got != 1100 {
		t.Errorf("LoanTotal(1000, 10, 12, simple) = %v, want 1100", got)
	}
}

func TestLoanTotal_Compound(t *testing.T) {
	// 1000 * (1 + 10/100)^12 = 1000 * 1.1^12 ≈ 3138.43
	got := LoanTotal(1000, 10, 12, "compound")
	want := 3138.43
	if math.Abs(got-want) > 0.01 {
		t.Errorf("LoanTotal(1000, 10, 12, compound) = %v, want %v", got, want)
	}
}

func TestLoanTotal_ZeroRate(t *testing.T) {
	got := LoanTotal(500, 0, 6, "simple")
	if got != 500 {
		t.Errorf("LoanTotal(500, 0, 6, simple) = %v, want 500", got)
	}
}

func TestAdvanceByFreq(t *testing.T) {
	tests := []struct {
		date, freq string
		k          int
		want       string
	}{
		{"2026-01-01", "monthly", 1, "2026-02-01"},
		{"2026-01-01", "monthly", 12, "2027-01-01"},
		{"2026-01-01", "weekly", 1, "2026-01-08"},
		{"2026-01-01", "biweekly", 1, "2026-01-15"},
		{"2026-01-31", "monthly", 1, "2026-03-03"}, // Go AddDate normalizes: Jan 31 + 1 month = Mar 3
		{"2026-01-01", "monthly", 0, "2026-01-01"},
	}
	for _, tt := range tests {
		got, err := AdvanceByFreq(tt.date, tt.freq, tt.k)
		if err != nil {
			t.Errorf("AdvanceByFreq(%s, %s, %d): unexpected error: %v", tt.date, tt.freq, tt.k, err)
			continue
		}
		if got != tt.want {
			t.Errorf("AdvanceByFreq(%s, %s, %d) = %s, want %s", tt.date, tt.freq, tt.k, got, tt.want)
		}
	}
}

func TestComputeEffectiveAmounts_NoCarry(t *testing.T) {
	// No overdue: no carry
	insts := []InstallmentIn{
		{InstallmentID: "1", DueDate: "2026-02-01", Amount: 100, PaidAmount: 0},
		{InstallmentID: "2", DueDate: "2026-03-01", Amount: 100, PaidAmount: 0},
	}
	result := ComputeEffectiveAmounts(insts, "2026-01-15")
	if result[0].Amount != 100 || result[1].Amount != 100 {
		t.Errorf("expected no carry, got %v", result)
	}
}

func TestComputeEffectiveAmounts_OverduePartialCarry(t *testing.T) {
	// Inst 1: due 2026-02-01, paid 50 (partial, overdue as of 2026-03-01)
	// Inst 2: due 2026-03-01, unpaid — should carry 50
	insts := []InstallmentIn{
		{InstallmentID: "1", DueDate: "2026-02-01", Amount: 100, PaidAmount: 50},
		{InstallmentID: "2", DueDate: "2026-03-01", Amount: 100, PaidAmount: 0},
	}
	result := ComputeEffectiveAmounts(insts, "2026-03-15")
	// Inst 1: effective = 100 (no prior carry), overdue, partial → carry = 50
	// Inst 2: effective = 100 + 50 = 150
	if result[0].Amount != 100 {
		t.Errorf("inst 1: got amount %v, want 100", result[0].Amount)
	}
	if result[1].Amount != 150 {
		t.Errorf("inst 2: got amount %v, want 150", result[1].Amount)
	}
}

func TestComputeEffectiveAmounts_OverdueFullNoCarry(t *testing.T) {
	// Inst 1: overdue, fully paid → no carry
	insts := []InstallmentIn{
		{InstallmentID: "1", DueDate: "2026-02-01", Amount: 100, PaidAmount: 100},
		{InstallmentID: "2", DueDate: "2026-03-01", Amount: 100, PaidAmount: 0},
	}
	result := ComputeEffectiveAmounts(insts, "2026-03-15")
	if result[1].Amount != 100 {
		t.Errorf("inst 2: got amount %v, want 100 (no carry from fully paid)", result[1].Amount)
	}
}
