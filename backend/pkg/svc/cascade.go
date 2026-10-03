package svc

import "fmt"

// CascadeInstallment es el estado mínimo de una cuota candidata a recibir
// parte de un abono en cascada.
type CascadeInstallment struct {
	ID     string
	Number int
	Amount float64
	Paid   float64
}

// CascadeAllocation es el monto aplicado a una cuota puntual dentro de un
// reparto en cascada.
type CascadeAllocation struct {
	ID     string
	Number int
	Amount float64
}

// ValidationError señala un error de validación de negocio (mapeable a 400
// en el handler HTTP) en vez de un error interno/de base de datos.
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

// AllocateCascade reparte `amount` entre `installments` (ordenadas por
// número, empezando por la cuota inicial) sin pisar lo ya pagado: llena la
// cuota inicial primero y vuelca el excedente a las siguientes. Es una
// función pura (sin acceso a base de datos) para poder testear la
// matemática del reparto de forma aislada.
//
// Retorna:
//   - allocs: el monto aplicado a cada cuota afectada.
//   - startingRemaining: el saldo pendiente de la cuota inicial ANTES del
//     abono (amount - paid de la primera cuota, nunca negativo).
//   - excess: lo que sobró del abono una vez cubierto startingRemaining.
//   - err: *ValidationError si amount<=0 o si amount excede el saldo total
//     pendiente de las cuotas recibidas (a partir de la inicial).
func AllocateCascade(installments []CascadeInstallment, amount float64) (allocs []CascadeAllocation, startingRemaining, excess float64, err error) {
	if verr := ValidateAmountPositive(amount); verr != nil {
		return nil, 0, 0, &ValidationError{Msg: verr.Error()}
	}
	amt := Round2(amount)

	if len(installments) > 0 {
		startingRemaining = max(0, Round2(installments[0].Amount-installments[0].Paid))
	}

	var totalRemaining float64
	for _, i := range installments {
		totalRemaining = Round2(totalRemaining + max(0, Round2(i.Amount-i.Paid)))
	}
	const epsilon = 0.005
	if amt > totalRemaining+epsilon {
		return nil, 0, 0, &ValidationError{
			Msg: fmt.Sprintf("el monto excede el saldo pendiente del préstamo ($%.2f)", totalRemaining),
		}
	}

	remaining := amt
	allocs = make([]CascadeAllocation, 0)
	for _, i := range installments {
		if remaining <= 0 {
			break
		}
		space := Round2(i.Amount - i.Paid)
		if space <= 0 {
			continue
		}
		give := Round2(min(remaining, space))
		remaining = Round2(remaining - give)
		allocs = append(allocs, CascadeAllocation{ID: i.ID, Number: i.Number, Amount: give})
	}

	excess = max(0, Round2(amt-startingRemaining))
	return allocs, startingRemaining, excess, nil
}

// ReversePaid resta `alloc` del monto pagado de una cuota al revertir un pago.
// Devuelve el nuevo paid_amount (nunca negativo) y si hay que limpiar paid_date
// (cuando la cuota queda sin nada pagado).
func ReversePaid(paid, alloc float64) (newPaid float64, clearDate bool) {
	newPaid = max(0, Round2(paid-alloc))
	return newPaid, newPaid <= 0
}
