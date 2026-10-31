package svc

import (
	"errors"
	"testing"
)

func threeInstallments() []CascadeInstallment {
	return []CascadeInstallment{
		{ID: "a", Number: 1, Amount: 100, Paid: 100},
		{ID: "b", Number: 2, Amount: 100, Paid: 50},
		{ID: "c", Number: 3, Amount: 100, Paid: 0},
	}
}

func TestReallocatePayment(t *testing.T) {
	// El pago original (150) cubrió a:100 y b:50.
	old := []CascadeAllocation{{ID: "a", Number: 1, Amount: 100}, {ID: "b", Number: 2, Amount: 50}}

	t.Run("decrease frees later installments", func(t *testing.T) {
		reversed, allocs, err := ReallocatePayment(threeInstallments(), old, "a", 120)
		if err != nil {
			t.Fatal(err)
		}
		if reversed[0].Paid != 0 || reversed[1].Paid != 0 {
			t.Fatalf("reversed paid = %v", reversed)
		}
		want := []CascadeAllocation{{ID: "a", Number: 1, Amount: 100}, {ID: "b", Number: 2, Amount: 20}}
		if len(allocs) != len(want) || allocs[0] != want[0] || allocs[1] != want[1] {
			t.Fatalf("allocs = %v, want %v", allocs, want)
		}
	})

	t.Run("increase spills into the next installment", func(t *testing.T) {
		_, allocs, err := ReallocatePayment(threeInstallments(), old, "a", 250)
		if err != nil {
			t.Fatal(err)
		}
		if len(allocs) != 3 || allocs[2].ID != "c" || allocs[2].Amount != 50 {
			t.Fatalf("allocs = %v", allocs)
		}
	})

	t.Run("exceeding the loan balance is a validation error", func(t *testing.T) {
		_, _, err := ReallocatePayment(threeInstallments(), old, "a", 300.01)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("err = %v, want ValidationError", err)
		}
	})

	t.Run("non positive amount is rejected", func(t *testing.T) {
		_, _, err := ReallocatePayment(threeInstallments(), old, "a", 0)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("err = %v, want ValidationError", err)
		}
	})

	t.Run("does not mutate the input", func(t *testing.T) {
		in := threeInstallments()
		if _, _, err := ReallocatePayment(in, old, "a", 50); err != nil {
			t.Fatal(err)
		}
		if in[0].Paid != 100 || in[1].Paid != 50 {
			t.Fatalf("input mutated: %v", in)
		}
	})

	t.Run("later payments are left alone", func(t *testing.T) {
		// Otro pago posterior dejó c:30; editar el primero no lo toca.
		in := threeInstallments()
		in[2].Paid = 30
		reversed, allocs, err := ReallocatePayment(in, old, "a", 150)
		if err != nil {
			t.Fatal(err)
		}
		if reversed[2].Paid != 30 {
			t.Fatalf("c paid = %v", reversed[2].Paid)
		}
		if len(allocs) != 2 {
			t.Fatalf("allocs = %v", allocs)
		}
	})

	t.Run("missing starting installment falls back to the first allocated one", func(t *testing.T) {
		_, allocs, err := ReallocatePayment(threeInstallments(), old, "gone", 60)
		if err != nil {
			t.Fatal(err)
		}
		if len(allocs) != 1 || allocs[0].ID != "a" || allocs[0].Amount != 60 {
			t.Fatalf("allocs = %v", allocs)
		}
	})
}
