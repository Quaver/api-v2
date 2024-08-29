package db

type Metrics struct {
	FailedScores int64 `gorm:"column:failed_scores" json:"failed_scores"`
}

func (*Metrics) TableName() string {
	return "metrics"
}

func IncrementFailedScoresMetric() error {
	return SQL.Exec("UPDATE metrics SET failed_scores = failed_scores + 1").Error
}

func GetMetrics() (*Metrics, error) {
	var metrics Metrics

	result := SQL.First(&metrics)

	if result.Error != nil {
		return nil, result.Error
	}

	return &metrics, nil
}
