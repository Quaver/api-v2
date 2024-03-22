package db

type CrashLog struct {
	Id        int    `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId    int    `gorm:"column:user_id" json:"user_id"`
	Timestamp int64  `gorm:"column:timestamp"`
	Runtime   string `gorm:"column:runtime_log" json:"runtime_log"`
	Network   string `gorm:"column:network_log" json:"network_log"`
}

func (*CrashLog) TableName() string {
	return "crash_logs"
}

// InsertCrashLog Inserts a crash log into the database
func InsertCrashLog(cl *CrashLog) error {
	if result := SQL.Create(cl); result.Error != nil {
		return result.Error
	}

	return nil
}
