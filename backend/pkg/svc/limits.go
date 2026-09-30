package svc

import "fmt"

// CheckLimit verifica si current alcanzó o superó limit para un resource
// ("expense", "account" o "loan"). limit < 0 significa plan ilimitado (Pro).
// Un resource desconocido no tiene límite definido.
func CheckLimit(resource string, limit, current int) error {
	if limit < 0 || current < limit {
		return nil
	}
	switch resource {
	case "expense":
		return fmt.Errorf("límite de %d gastos/mes alcanzado. Actualiza a Pro para gastos ilimitados", limit)
	case "account":
		return fmt.Errorf("límite de %d cuentas alcanzado. Actualiza a Pro", limit)
	case "loan":
		return fmt.Errorf("límite de %d préstamos alcanzado. Actualiza a Pro", limit)
	default:
		return nil
	}
}
