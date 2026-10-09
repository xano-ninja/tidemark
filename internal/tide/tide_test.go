package tide

import (
	"testing"
	"time"
)

func TestStationLocal(t *testing.T) {
	tests := []struct {
		name string
		tz   string
		utc  time.Time
		want string
	}{
		{"santos", "America/Sao_Paulo", time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC), "09:00"},
		{"seattle", "America/Los_Angeles", time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), "04:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Station{Timezone: tt.tz}
			got, err := s.Local(Extreme{Time: tt.utc, Kind: Low})
			if err != nil {
				t.Fatal(err)
			}
			if g := got.Format("15:04"); g != tt.want {
				t.Errorf("got %s, want %s", g, tt.want)
			}
		})
	}
}

func TestStationLocalBadZone(t *testing.T) {
	if _, err := (Station{Timezone: "Nowhere/Land"}).Local(Extreme{}); err == nil {
		t.Fatal("expected error")
	}
}
