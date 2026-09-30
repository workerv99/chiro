package svc

import "testing"

func TestCheckLimit(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		limit    int
		current  int
		wantErr  bool
	}{
		{"unlimited plan allows any count", "expense", -1, 1000, false},
		{"expense under limit", "expense", 50, 10, false},
		{"expense at limit", "expense", 50, 50, true},
		{"expense over limit", "expense", 50, 51, true},
		{"account under limit", "account", 3, 2, false},
		{"account at limit", "account", 3, 3, true},
		{"loan under limit", "loan", 10, 9, false},
		{"loan at limit", "loan", 10, 10, true},
		{"unknown resource has no limit", "widget", 5, 100, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckLimit(tt.resource, tt.limit, tt.current)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckLimit(%q, %d, %d): err=%v, wantErr=%v", tt.resource, tt.limit, tt.current, err, tt.wantErr)
			}
		})
	}
}
