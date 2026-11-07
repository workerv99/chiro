package store

import "testing"

func TestNormalizeForzaUserID(t *testing.T) {
	row := map[string]any{"id": "x", "user_id": "otro-usuario", "name": "n"}
	out := Normalize(row, "yo")
	if out["user_id"] != "yo" {
		t.Errorf("user_id=%v, want yo", out["user_id"])
	}
	if row["user_id"] != "otro-usuario" {
		t.Error("Normalize no debe mutar la fila de entrada")
	}
}

func TestNormalizeDefaults(t *testing.T) {
	out := Normalize(map[string]any{"id": "x"}, "yo")
	if out["user_id"] != "yo" || out["deleted"] != 0 || out["updated_at"] == nil {
		t.Errorf("defaults faltantes: %v", out)
	}
}
