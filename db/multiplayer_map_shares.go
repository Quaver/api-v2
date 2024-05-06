package db

import "encoding/json"

type MultiplayerMapShare struct {
	Id         int    `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId     int    `gorm:"column:user_id" json:"uploader_id"`
	GameId     int    `gorm:"column:game_id" json:"game_id"`
	MapMD5     string `gorm:"column:map_md5" json:"map_md5"`
	PackageMD5 string `gorm:"column:package_md5" json:"package_md5"`
	Timestamp  int64  `gorm:"timestamp" json:"timestamp"`
}

func (*MultiplayerMapShare) TableName() string {
	return "multiplayer_map_shares"
}

func (m *MultiplayerMapShare) Insert() error {
	if err := SQL.Create(&m).Error; err != nil {
		return err
	}

	return nil
}

func (m *MultiplayerMapShare) PublishToRedis() error {
	if err := Redis.Publish(RedisCtx, "quaver:multiplayer_map_shares", m).Err(); err != nil {
		return err
	}

	return nil
}

func (m *MultiplayerMapShare) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}
func (m *MultiplayerMapShare) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &m)
}
