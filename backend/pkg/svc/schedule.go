// Package svc contiene la lógica de dominio portada de utils/repo y utils/sync.
package svc

import (
	"crypto/rand"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Round2 redondea a 2 decimales (como Math.round(x*100)/100).
func Round2(v float64) float64 { return math.Round(v*100) / 100 }

// GenID genera un id con prefijo, port de utils/ids.ts genId().
func GenID(prefix string) string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	suffix := fmt.Sprintf("%x", b)
	return prefix + "_" + suffix
}

// TodayISO devuelve la fecha de hoy en UTC como YYYY-MM-DD.
func TodayISO() string { return time.Now().UTC().Format("2006-01-02") }

// ParseISO parsea YYYY-MM-DD en UTC.
func ParseISO(s string) (time.Time, error) { return time.Parse("2006-01-02", s) }

// AdvanceByFreq avanza una fecha YYYY-MM-DD según la frecuencia k períodos.
// Port de utils/repo/installments.ts advanceByFreq (aritmética en UTC).
func AdvanceByFreq(dateStr string, freq string, k int) (string, error) {
	d, err := ParseISO(dateStr)
	if err != nil {
		return "", err
	}
	var out time.Time
	switch freq {
	case "weekly":
		out = d.AddDate(0, 0, 7*k)
	case "biweekly":
		out = d.AddDate(0, 0, 14*k)
	default: // monthly
		out = d.AddDate(0, k, 0)
	}
	return out.Format("2006-01-02"), nil
}

// AdvanceBillDate avanza la próxima fecha de una factura recurrente.
func AdvanceBillDate(dateStr string, frequency string) (string, error) {
	d, err := ParseISO(dateStr)
	if err != nil {
		return "", err
	}
	var out time.Time
	switch frequency {
	case "daily":
		out = d.AddDate(0, 0, 1)
	case "weekly":
		out = d.AddDate(0, 0, 7)
	case "yearly":
		out = d.AddDate(1, 0, 0)
	default: // monthly
		out = d.AddDate(0, 1, 0)
	}
	return out.Format("2006-01-02"), nil
}

// SplitAmounts reparte un total en n cuotas iguales (la última absorbe el redondeo).
func SplitAmounts(total float64, n int) []float64 {
	if n < 1 {
		n = 1
	}
	per := Round2(total / float64(n))
	arr := make([]float64, n)
	for i := range arr {
		arr[i] = per
	}
	arr[n-1] = Round2(total - per*float64(n-1))
	return arr
}

// LoanTotal calcula el total a devolver según el tipo de interés.
// simple   → monto * (1 + tasa/100)
// compound → monto * (1 + tasa/100)^meses
func LoanTotal(amount, rate float64, months int, interestType string) float64 {
	total := amount
	if rate > 0 {
		switch interestType {
		case "compound":
			if months > 0 {
				total = amount * math.Pow(1+rate/100, float64(months))
			} else {
				total = amount + amount*rate/100
			}
		default:
			total = amount + amount*rate/100
		}
	}
	return Round2(total)
}

// InstallmentIn es la forma de una cuota leída de la base.
type InstallmentIn struct {
	InstallmentID string  `json:"installment_id"`
	LoanID        string  `json:"loan_id"`
	Number        int     `json:"number"`
	DueDate       string  `json:"due_date"`
	Amount        float64 `json:"amount"`
	PaidDate      *string `json:"paid_date"`
	PaidAmount    float64 `json:"paid_amount"`
}

// ComputeEffectiveAmounts aplica el arrastre de deuda vencida parcialmente pagada.
// Port de utils/repo/installments.ts computeEffectiveAmounts.
func ComputeEffectiveAmounts(installments []InstallmentIn, today string) []InstallmentIn {
	carry := 0.0
	for i := range installments {
		inst := &installments[i]
		effective := Round2(inst.Amount + carry)
		isOverdue := inst.DueDate < today
		fullyPaid := inst.PaidAmount >= inst.Amount
		hasPartial := inst.PaidAmount > 0 && !fullyPaid
		remaining := math.Max(0, Round2(effective-inst.PaidAmount))
		if isOverdue && hasPartial {
			carry = remaining
		} else {
			carry = 0
		}
		inst.Amount = effective
	}
	return installments
}

// InstallmentID genera el id determinista de una cuota (ins_<loan>_<número>).
func InstallmentID(loanID string, number int) string {
	return "ins_" + loanID + "_" + strconv.Itoa(number)
}

// Money formatea un monto sin símbolo de moneda.
func Money(n float64) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintf("%.2f", n), "0"), ".")
}
