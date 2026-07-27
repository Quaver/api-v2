package enums

import "testing"

func TestGetSpeedRate(t *testing.T) {
	tests := []struct {
		name string
		mods Mods
		want int
	}{
		{name: "none", mods: 0, want: 100},
		{name: "0.5x", mods: ModSpeed05X, want: 50},
		{name: "0.55x", mods: ModSpeed055X, want: 55},
		{name: "0.6x", mods: ModSpeed06X, want: 60},
		{name: "0.65x", mods: ModSpeed065X, want: 65},
		{name: "0.7x", mods: ModSpeed07X, want: 70},
		{name: "0.75x", mods: ModSpeed075X, want: 75},
		{name: "0.8x", mods: ModSpeed08X, want: 80},
		{name: "0.85x", mods: ModSpeed085X, want: 85},
		{name: "0.9x", mods: ModSpeed09X, want: 90},
		{name: "0.95x", mods: ModSpeed095X, want: 95},
		{name: "1.05x", mods: ModSpeed105X, want: 105},
		{name: "1.1x", mods: ModSpeed11X, want: 110},
		{name: "1.15x", mods: ModSpeed115X, want: 115},
		{name: "1.2x", mods: ModSpeed12X, want: 120},
		{name: "1.25x", mods: ModSpeed125X, want: 125},
		{name: "1.3x", mods: ModSpeed13X, want: 130},
		{name: "1.35x", mods: ModSpeed135X, want: 135},
		{name: "1.4x", mods: ModSpeed14X, want: 140},
		{name: "1.45x", mods: ModSpeed145X, want: 145},
		{name: "1.5x", mods: ModSpeed15X, want: 150},
		{name: "1.55x", mods: ModSpeed155X, want: 155},
		{name: "1.6x", mods: ModSpeed16X, want: 160},
		{name: "1.65x", mods: ModSpeed165X, want: 165},
		{name: "1.7x", mods: ModSpeed17X, want: 170},
		{name: "1.75x", mods: ModSpeed175X, want: 175},
		{name: "1.8x", mods: ModSpeed18X, want: 180},
		{name: "1.85x", mods: ModSpeed185X, want: 185},
		{name: "1.9x", mods: ModSpeed19X, want: 190},
		{name: "1.95x", mods: ModSpeed195X, want: 195},
		{name: "2.0x", mods: ModSpeed20X, want: 200},
		{name: "non-speed", mods: ModMirror, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSpeedRate(tt.mods); got != tt.want {
				t.Fatalf("GetSpeedRate(%v) = %v, want %v", tt.mods, got, tt.want)
			}
		})
	}
}
