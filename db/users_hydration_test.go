package db

import "testing"

func TestUserClientStatusFromValues(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
		want   *UserClientStatus
	}{
		{
			name:   "empty status",
			values: map[string]string{},
		},
		{
			name: "populated status",
			values: map[string]string{
				"s": "2",
				"m": "7",
				"c": "Playing",
			},
			want: &UserClientStatus{Status: 2, Mode: 7, Content: "Playing"},
		},
		{
			name:   "invalid values use defaults",
			values: map[string]string{"s": "invalid", "m": "invalid"},
			want:   &UserClientStatus{Status: 0, Mode: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := userClientStatusFromValues(tt.values)

			if tt.want == nil {
				if got != nil {
					t.Fatalf("userClientStatusFromValues() = %#v, want nil", got)
				}

				return
			}

			if got == nil || *got != *tt.want {
				t.Fatalf("userClientStatusFromValues() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
