package db

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/enums"
)

func setScoreModColumnsReadyForTest(t *testing.T, ready bool) {
	t.Helper()

	original := config.Instance
	configured := &config.Config{ScoreModColumnsReady: ready}
	config.Instance = configured

	t.Cleanup(func() {
		config.Instance = original
	})
}

func TestGetModifierScoreFilter(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		mods      int64
		ready     bool
		wantQuery string
		wantArgs  []any
		wantErr   error
	}{
		{
			name:      "none",
			alias:     "scores",
			mods:      0,
			wantQuery: "AND scores.mods = ? ",
			wantArgs:  []any{int64(0)},
		},
		{
			name:      "legacy supported modifier",
			alias:     "s",
			mods:      int64(enums.ModMirror),
			wantQuery: "AND (s.mods & ?) != 0 ",
			wantArgs:  []any{int64(enums.ModMirror)},
		},
		{
			name:      "legacy unsupported modifier remains available",
			alias:     "s",
			mods:      int64(enums.ModNoFail),
			wantQuery: "AND (s.mods & ?) != 0 ",
			wantArgs:  []any{int64(enums.ModNoFail)},
		},
		{
			name:      "lookup single supported modifier",
			alias:     "s",
			mods:      int64(enums.ModMirror),
			ready:     true,
			wantQuery: "AND s.mirror = 1 ",
			wantArgs:  []any{},
		},
		{
			name:      "lookup multiple supported modifiers",
			alias:     "scores",
			mods:      int64(enums.ModMirror | enums.ModNoMiss),
			ready:     true,
			wantQuery: "AND scores.mirror = 1 AND scores.no_miss = 1 ",
			wantArgs:  []any{},
		},
		{
			name:    "lookup unsupported modifier errors",
			alias:   "s",
			mods:    int64(enums.ModNoFail),
			ready:   true,
			wantErr: ErrUnsupportedScoreModifierFilter,
		},
		{
			name:    "lookup mixed supported and unsupported errors",
			alias:   "s",
			mods:    int64(enums.ModMirror | enums.ModNoFail),
			ready:   true,
			wantErr: ErrUnsupportedScoreModifierFilter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setScoreModColumnsReadyForTest(t, tt.ready)
			gotQuery, gotArgs, err := getModifierScoreFilter(tt.alias, tt.mods)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if gotQuery != tt.wantQuery {
				t.Fatalf("query = %q, want %q", gotQuery, tt.wantQuery)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Fatalf("args = %#v, want %#v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestGetRateScoreFilter(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		mods      int64
		ready     bool
		wantQuery string
		wantArgs  []any
		wantErr   error
	}{
		{
			name:      "legacy none",
			alias:     "s",
			mods:      0,
			wantQuery: "AND (s.mods = 0 OR s.mods = ?) ",
			wantArgs:  []any{int64(enums.ModMirror)},
		},
		{
			name:      "legacy speed modifier",
			alias:     "scores",
			mods:      int64(enums.ModSpeed15X),
			wantQuery: "AND (scores.mods & ?) != 0 ",
			wantArgs:  []any{int64(enums.ModSpeed15X)},
		},
		{
			name:      "legacy non-speed modifier remains available",
			alias:     "s",
			mods:      int64(enums.ModMirror),
			wantQuery: "AND (s.mods & ?) != 0 ",
			wantArgs:  []any{int64(enums.ModMirror)},
		},
		{
			name:      "lookup none",
			alias:     "s",
			mods:      0,
			ready:     true,
			wantQuery: "AND s.speed_rate = ? ",
			wantArgs:  []any{100},
		},
		{
			name:      "lookup speed modifier",
			alias:     "scores",
			mods:      int64(enums.ModSpeed15X),
			ready:     true,
			wantQuery: "AND scores.speed_rate = ? ",
			wantArgs:  []any{150},
		},
		{
			name:    "lookup non-speed errors",
			alias:   "s",
			mods:    int64(enums.ModMirror),
			ready:   true,
			wantErr: ErrUnsupportedScoreModifierFilter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setScoreModColumnsReadyForTest(t, tt.ready)
			gotQuery, gotArgs, err := getRateScoreFilter(tt.alias, tt.mods)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if gotQuery != tt.wantQuery {
				t.Fatalf("query = %q, want %q", gotQuery, tt.wantQuery)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Fatalf("args = %#v, want %#v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestScoreboardRedisKeySeparatesModifierFilterModes(t *testing.T) {
	setScoreModColumnsReadyForTest(t, false)
	legacyKey := scoreboardRedisKey("map", scoreboardMods, int64(enums.ModMirror))

	config.Instance.ScoreModColumnsReady = true
	columnsKey := scoreboardRedisKey("map", scoreboardMods, int64(enums.ModMirror))

	if legacyKey == columnsKey {
		t.Fatalf("modifier scoreboard cache key did not change with filter mode: %q", legacyKey)
	}
}

func TestScoreAfterFindPopulatesLegacyModifierFields(t *testing.T) {
	timestamp := int64(1_721_234_567_890)
	score := &Score{
		Timestamp: timestamp,
		Modifiers: int64(enums.ModSpeed15X | enums.ModMirror | enums.ModNoMiss),
	}

	if err := score.AfterFind(nil); err != nil {
		t.Fatalf("AfterFind() error = %v", err)
	}

	if want := time.UnixMilli(timestamp); !score.TimestampJSON.Equal(want) {
		t.Fatalf("TimestampJSON = %v, want %v", score.TimestampJSON, want)
	}

	if score.SpeedRate != 150 {
		t.Fatalf("SpeedRate = %d, want 150", score.SpeedRate)
	}

	if !score.Mirror || !score.NoMiss {
		t.Fatalf("legacy modifier fields were not populated: mirror=%v no_miss=%v", score.Mirror, score.NoMiss)
	}
}

func TestScoreAfterFindPreservesMigratedModifierFields(t *testing.T) {
	score := &Score{
		Modifiers:    int64(enums.ModSpeed15X | enums.ModMirror),
		SpeedRate:    125,
		Mirror:       false,
		ModsMigrated: true,
	}

	if err := score.AfterFind(nil); err != nil {
		t.Fatalf("AfterFind() error = %v", err)
	}

	if score.SpeedRate != 125 || score.Mirror {
		t.Fatalf("migrated modifier fields changed: speed_rate=%d mirror=%v", score.SpeedRate, score.Mirror)
	}
}
