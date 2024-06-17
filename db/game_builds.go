package db

import "time"

type GameBuild struct {
	Id                    int    `gorm:"column:id; PRIMARY_KEY"`
	Version               string `gorm:"column:version"`
	QuaverDll             string `gorm:"column:quaver_dll"`
	QuaverApiDll          string `gorm:"column:quaver_api_dll"`
	QuaverServerClientDll string `gorm:"column:quaver_server_client_dll"`
	QuaverServerCommonDll string `gorm:"column:quaver_server_common_dll"`
	QuaverSharedDll       string `gorm:"column:quaver_shared_dll"`
	Allowed               bool   `gorm:"column:allowed"`
	Timestamp             int64  `gorm:"column:timestamp"`
}

func (*GameBuild) TableName() string {
	return "game_builds"
}

// Insert Inserts a new game build into the database
func (g *GameBuild) Insert() error {
	g.Allowed = true
	g.Timestamp = time.Now().UnixMilli()

	if err := SQL.Create(&g).Error; err != nil {
		return err
	}

	return nil
}
