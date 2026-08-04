package svc

import (
	"fmt"
	"time"
)

// ValidateDate comprueba que un string sea YYYY-MM-DD válido.
func ValidateDate(s string) error {
	if s == "" {
		return nil // fechas opcionales
	}
	_, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("fecha inválida %q (usa YYYY-MM-DD)", s)
	}
	return nil
}

// ValidateInterestType comprueba que el tipo de interés sea válido.
func ValidateInterestType(s string) error {
	if s == "" || s == "simple" || s == "compound" {
		return nil
	}
	return fmt.Errorf("interest_type inválido: %q (simple|compound)", s)
}

// ValidateFrequency comprueba que la frecuencia de cuota sea válida.
func ValidateFrequency(s string) error {
	if s == "" || s == "monthly" || s == "biweekly" || s == "weekly" {
		return nil
	}
	return fmt.Errorf("frequency inválido: %q (monthly|biweekly|weekly)", s)
}

// ValidateAmountPositive comprueba que un monto sea > 0.
func ValidateAmountPositive(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount debe ser > 0, recibido: %v", amount)
	}
	return nil
}
