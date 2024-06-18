package db

import (
	"gorm.io/gorm"
	"time"
)

type MultiplayerGame struct {
	Id              int                       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UniqueId        string                    `gorm:"column:unique_id" json:"unique_id"`
	Name            string                    `gorm:"column:name" json:"name"`
	Type            int8                      `gorm:"column:type" json:"-"`
	TimeCreated     int64                     `gorm:"time_created" json:"-"`
	TimeCreatedJSON time.Time                 `gorm:"-:all" json:"time_created"`
	Matches         []*MultiplayerGameMatches `gorm:"foreignKey:GameId" json:"matches"`
}

func (*MultiplayerGame) TableName() string {
	return "multiplayer_games"
}

func (mg *MultiplayerGame) AfterFind(*gorm.DB) (err error) {
	mg.TimeCreatedJSON = time.UnixMilli(mg.TimeCreated)
	return nil
}

// GetRecentMultiplayerGames Retrieves recently played multiplayer games from the DB
func GetRecentMultiplayerGames(limit int, page int) ([]*MultiplayerGame, error) {
	var games []*MultiplayerGame

	result := SQL.
		Preload("Matches").
		Preload("Matches.Map").
		Order("id DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&games)

	if result.Error != nil {
		return nil, result.Error
	}

	return games, nil
}

// GetMultiplayerGame Gets an individual multiplayer game from the database
func GetMultiplayerGame(id int) (*MultiplayerGame, error) {
	var game *MultiplayerGame

	result := SQL.
		Preload("Matches").
		Preload("Matches.Map").
		Preload("Matches.Scores").
		Preload("Matches.Scores.User").
		Where("id = ?", id).
		First(&game)

	if result.Error != nil {
		return nil, result.Error
	}

	return game, nil
}