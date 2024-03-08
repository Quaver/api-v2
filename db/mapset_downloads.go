package db

type MapsetDownload struct {
	Id        int    `gorm:"column:id; PRIMARY_KEY"`
	UserId    int    `gorm:"column:user_id"`
	MapsetId  int    `gorm:"column:mapset_id"`
	Timestamp int64  `gorm:"column:timestamp"`
	Method    string `gorm:"column:method"`
}

type MapsetDownloadMethod int8

const (
	DownloadMethodWeb MapsetDownloadMethod = iota
	DownloadMethodInGame
)

func (*MapsetDownload) TableName() string {
	return "mapset_downloads"
}

// InsertMapsetDownload Inserts a mapset download into the database
func InsertMapsetDownload(download *MapsetDownload) error {
	download.Method = "Web"

	if err := SQL.Create(&download).Error; err != nil {
		return err
	}

	return nil
}
