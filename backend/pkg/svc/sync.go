package svc

import "time"

// IsProActive indica si la suscripción da derecho a la sincronización en la
// nube: plan "pro", estado "active" y no vencida. Un vencimiento nulo se
// considera sin caducidad.
func IsProActive(plan, status string, expiresAt *time.Time, now time.Time) bool {
	if plan != "pro" || status != "active" {
		return false
	}
	return expiresAt == nil || now.Before(*expiresAt)
}

// TablePage resume el resultado de pull de una tabla: filas devueltas, el
// límite aplicado y el updated_at de la última fila (orden ascendente).
type TablePage struct {
	Count         int
	Limit         int
	LastUpdatedAt int64
}

// NextSince calcula el cursor del siguiente pull y si quedan más filas.
//
// Semántica: el pull es INCLUSIVO (updated_at >= since) para no perder filas
// que comparten el mismo timestamp en el borde de una página. El cliente puede
// recibir filas repetidas; es idempotente porque el merge es "último gana".
//   - Si una tabla se truncó por el LIMIT (Count >= Limit), el cursor es el
//     MÍNIMO de los LastUpdatedAt de las tablas truncadas: avanzar más allá
//     saltaría filas aún no entregadas de esa tabla.
//   - Si ninguna se truncó, el cursor es el máximo updated_at visto.
//   - El cursor nunca retrocede por debajo de since.
//
// Limitación: si una tabla tiene >= Limit filas con exactamente el mismo
// updated_at igual a since, el cursor no avanza; el cliente debe subir limit.
func NextSince(since int64, pages []TablePage) (int64, bool) {
	var maxSeen int64 = since
	minTrunc := int64(-1)
	for _, p := range pages {
		if p.Count == 0 {
			continue
		}
		if p.LastUpdatedAt > maxSeen {
			maxSeen = p.LastUpdatedAt
		}
		if p.Count >= p.Limit && (minTrunc < 0 || p.LastUpdatedAt < minTrunc) {
			minTrunc = p.LastUpdatedAt
		}
	}
	if minTrunc >= 0 {
		if minTrunc < since {
			minTrunc = since
		}
		return minTrunc, true
	}
	return maxSeen, false
}
