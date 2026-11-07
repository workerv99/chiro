package svc

import "fmt"

// CheckLimit verifica si current alcanzó o superó limit para un resource
// ("expense", "account" o "loan"). limit < 0 significa sin límite.
// Un resource desconocido no tiene límite definido.
func CheckLimit(resource string, limit, current int) error {
	if limit < 0 || current < limit {
		return nil
	}
	switch resource {
	case "expense":
		return fmt.Errorf("límite de %d gastos/mes alcanzado en tu plan", limit)
	case "account":
		return fmt.Errorf("límite de %d cuentas alcanzado en tu plan", limit)
	case "loan":
		return fmt.Errorf("límite de %d préstamos alcanzado en tu plan", limit)
	default:
		return nil
	}
}
