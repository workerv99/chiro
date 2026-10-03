package svc

import (
	"errors"
	"testing"
)

func TestAllocateCascade(t *testing.T) {
	tests := []struct {
		name         string
		installments []CascadeInstallment
		amount       float64
		wantAllocs   []CascadeAllocation
		wantStarting float64
		wantExcess   float64
		wantErr      bool
	}{
		{
			name:         "exact pay",
			installments: []CascadeInstallment{{ID: "a", Number: 1, Amount: 100, Paid: 0}},
			amount:       100,
			wantAllocs:   []CascadeAllocation{{ID: "a", Number: 1, Amount: 100}},
			wantStarting: 100,
			wantExcess:   0,
		},
		{
			name:         "partial pay stays on same installment",
			installments: []CascadeInstallment{{ID: "a", Number: 1, Amount: 100, Paid: 0}},
			amount:       40,
			wantAllocs:   []CascadeAllocation{{ID: "a", Number: 1, Amount: 40}},
			wantStarting: 100,
			wantExcess:   0,
		},
		{
			name: "excess across two installments",
			installments: []CascadeInstallment{
				{ID: "a", Number: 1, Amount: 100, Paid: 0},
				{ID: "b", Number: 2, Amount: 100, Paid: 0},
			},
			amount: 150,
			wantAllocs: []CascadeAllocation{
				{ID: "a", Number: 1, Amount: 100},
				{ID: "b", Number: 2, Amount: 50},
			},
			wantStarting: 100,
			wantExcess:   50,
		},
		{
			name: "starting installment already partially paid (100/40, pay 80)",
			installments: []CascadeInstallment{
				{ID: "a", Number: 1, Amount: 100, Paid: 40},
				{ID: "b", Number: 2, Amount: 100, Paid: 0},
			},
			amount: 80,
			wantAllocs: []CascadeAllocation{
				{ID: "a", Number: 1, Amount: 60},
				{ID: "b", Number: 2, Amount: 20},
			},
			wantStarting: 60,
			wantExcess:   20,
		},
		{
			name: "skip already-paid installments",
			installments: []CascadeInstallment{
				{ID: "a", Number: 1, Amount: 100, Paid: 100},
				{ID: "b", Number: 2, Amount: 50, Paid: 0},
			},
			amount:       50,
			wantAllocs:   []CascadeAllocation{{ID: "b", Number: 2, Amount: 50}},
			wantStarting: 0,
			wantExcess:   50,
		},
		{
			name:         "zero amount errors",
			installments: []CascadeInstallment{{ID: "a", Number: 1, Amount: 100, Paid: 0}},
			amount:       0,
			wantErr:      true,
		},
		{
			name:         "negative amount errors",
			installments: []CascadeInstallment{{ID: "a", Number: 1, Amount: 100, Paid: 0}},
			amount:       -10,
			wantErr:      true,
		},
		{
			name:         "overpay beyond total remaining errors",
			installments: []CascadeInstallment{{ID: "a", Number: 1, Amount: 100, Paid: 0}},
			amount:       150,
			wantErr:      true,
		},
		{
			name: "overpay across multiple installments errors",
			installments: []CascadeInstallment{
				{ID: "a", Number: 1, Amount: 100, Paid: 0},
				{ID: "b", Number: 2, Amount: 50, Paid: 0},
			},
			amount:  200,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allocs, startingRemaining, excess, err := AllocateCascade(tt.installments, tt.amount)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("AllocateCascade(%v, %v): want error, got nil", tt.installments, tt.amount)
				}
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("AllocateCascade(%v, %v): want *ValidationError, got %T", tt.installments, tt.amount, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("AllocateCascade(%v, %v): unexpected error: %v", tt.installments, tt.amount, err)
			}
			if startingRemaining != tt.wantStarting {
				t.Errorf("startingRemaining = %v, want %v", startingRemaining, tt.wantStarting)
			}
			if excess != tt.wantExcess {
				t.Errorf("excess = %v, want %v", excess, tt.wantExcess)
			}
			if len(allocs) != len(tt.wantAllocs) {
				t.Fatalf("allocs = %v, want %v", allocs, tt.wantAllocs)
			}
			for i, a := range allocs {
				if a != tt.wantAllocs[i] {
					t.Errorf("allocs[%d] = %v, want %v", i, a, tt.wantAllocs[i])
				}
			}
		})
	}
}

func TestReversePaid(t *testing.T) {
	tests := []struct {
		name         string
		paid, alloc  float64
		want         float64
		wantClearDay bool
	}{
		{"full allocation zeroes the installment", 32.5, 32.5, 0, true},
		{"partial allocation leaves earlier payments", 50, 17.5, 32.5, false},
		{"over-reversal clamps at zero", 10, 15, 0, true},
		{"float noise rounds away", 0.3, 0.1 + 0.2, 0, true},
	}
	for _, tt := range tests {
		got, clear := ReversePaid(tt.paid, tt.alloc)
		if got != tt.want || clear != tt.wantClearDay {
			t.Errorf("%s: ReversePaid(%v, %v) = (%v, %v), want (%v, %v)",
				tt.name, tt.paid, tt.alloc, got, clear, tt.want, tt.wantClearDay)
		}
	}
}
