package commands

import "testing"

func TestDatabaseScoresModsBackfillValidation(t *testing.T) {
	originalBatchSize := scoreModsBackfillBatchSize
	originalSleepMs := scoreModsBackfillSleepMs
	originalMaxBatches := scoreModsBackfillMaxBatches

	t.Cleanup(func() {
		scoreModsBackfillBatchSize = originalBatchSize
		scoreModsBackfillSleepMs = originalSleepMs
		scoreModsBackfillMaxBatches = originalMaxBatches
	})

	tests := []struct {
		name       string
		batchSize  int
		sleepMs    int
		maxBatches int64
		wantError  string
	}{
		{
			name:      "zero batch size",
			batchSize: 0,
			wantError: "batch-size must be greater than 0",
		},
		{
			name:      "negative sleep",
			batchSize: 1,
			sleepMs:   -1,
			wantError: "sleep-ms cannot be negative",
		},
		{
			name:       "negative max batches",
			batchSize:  1,
			maxBatches: -1,
			wantError:  "max-batches cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoreModsBackfillBatchSize = tt.batchSize
			scoreModsBackfillSleepMs = tt.sleepMs
			scoreModsBackfillMaxBatches = tt.maxBatches
			err := DatabaseScoresModsBackfill.RunE(DatabaseScoresModsBackfill, nil)

			if err == nil || err.Error() != tt.wantError {
				t.Fatalf("RunE() error = %v, want %q", err, tt.wantError)
			}
		})
	}
}
