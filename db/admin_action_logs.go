package db

type AdminActionLog struct {
	Id             int                `gorm:"column:id; PRIMARY_KEY"`
	AuthorId       int                `gorm:"column:author_id"`
	AuthorUsername string             `gorm:"column:author_username"`
	TargetId       int                `gorm:"column:target_id"`
	TargetUsername string             `gorm:"column:target_username"`
	Action         AdminActionLogType `gorm:"column:action"`
	Notes          string             `gorm:"column:notes"`
	Timestamp      int64              `gorm:"column:timestamp"`
}

type AdminActionLogType string

const (
	AdminActionBanned  AdminActionLogType = "Banned"
	AdminActionKicked  AdminActionLogType = "Kicked"
	AdminActionUnmuted AdminActionLogType = "Unmuted"
	AdminActionUpdated AdminActionLogType = "Updated"
)

func (*AdminActionLog) TableName() string {
	return "admin_action_logs"
}

// Insert Inserts an admin action log to the database
func (log *AdminActionLog) Insert() error {
	if err := SQL.Create(&log).Error; err != nil {
		return err
	}

	return nil
}
