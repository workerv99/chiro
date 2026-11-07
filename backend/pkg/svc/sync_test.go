package svc

import (
	"testing"
	"time"
)

func TestIsProActive(t *testing.T) {
	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	tests := []struct {
		name    string
		plan    string
		status  string
		expires *time.Time
		want    bool
	}{
		{"pro activo vigente", "pro", "active", &future, true},
		{"pro activo sin vencimiento", "pro", "active", nil, true},
		{"pro vencido", "pro", "active", &past, false},
		{"pro vence justo ahora", "pro", "active", &now, false},
		{"pro cancelado", "pro", "cancelled", &future, false},
		{"free activo", "free", "active", nil, false},
		{"plan vacío", "", "", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsProActive(tt.plan, tt.status, tt.expires, now); got != tt.want {
				t.Errorf("IsProActive=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestNextSince(t *testing.T) {
	tests := []struct {
		name        string
		since       int64
		pages       []TablePage
		wantSince   int64
		wantHasMore bool
	}{
		{"sin filas conserva since", 100, nil, 100, false},
		{"sin truncar usa el máximo", 100, []TablePage{
			{Count: 2, Limit: 5, LastUpdatedAt: 150},
			{Count: 1, Limit: 5, LastUpdatedAt: 300},
		}, 300, false},
		{"tabla truncada limita el cursor al mínimo", 100, []TablePage{
			{Count: 5, Limit: 5, LastUpdatedAt: 200},
			{Count: 2, Limit: 5, LastUpdatedAt: 900},
		}, 200, true},
		{"varias truncadas toma la menor", 100, []TablePage{
			{Count: 5, Limit: 5, LastUpdatedAt: 400},
			{Count: 5, Limit: 5, LastUpdatedAt: 250},
		}, 250, true},
		{"tabla vacía se ignora", 100, []TablePage{
			{Count: 0, Limit: 5, LastUpdatedAt: 0},
			{Count: 3, Limit: 5, LastUpdatedAt: 180},
		}, 180, false},
		{"nunca retrocede de since", 500, []TablePage{
			{Count: 5, Limit: 5, LastUpdatedAt: 500},
		}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSince, gotMore := NextSince(tt.since, tt.pages)
			if gotSince != tt.wantSince || gotMore != tt.wantHasMore {
				t.Errorf("NextSince=(%d,%v), want (%d,%v)", gotSince, gotMore, tt.wantSince, tt.wantHasMore)
			}
		})
	}
}
