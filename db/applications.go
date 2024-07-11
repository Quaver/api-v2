package db

import (
	"gorm.io/gorm"
	"time"
)

type Application struct {
	Id            int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId        int       `gorm:"column:user_id" json:"user_id"`
	Name          string    `gorm:"column:name" json:"name"`
	RedirectURL   string    `gorm:"column:redirect_url" json:"redirect_url"`
	ClientId      string    `gorm:"column:client_id" json:"client_id"`
	ClientSecret  string    `gorm:"column:client_secret" json:"client_secret,omitempty"`
	Timestamp     int64     `gorm:"column:timestamp" json:"-"`
	TimestampJSON time.Time `gorm:":-all" json:"timestamp"`
	Active        bool      `gorm:"column:active" json:"active"`
}

func (*Application) TableName() string {
	return "applications"
}

func (app *Application) AfterFind(db *gorm.DB) error {
	app.TimestampJSON = time.UnixMilli(app.Timestamp)
	return nil
}

// GetUserActiveApplications Retrieves a user's active applications
func GetUserActiveApplications(userId int) ([]*Application, error) {
	var applications = make([]*Application, 0)

	result := SQL.
		Where("user_id = ? AND active = 1", userId).
		Order("id DESC").
		Find(&applications)

	if result.Error != nil {
		return nil, result.Error
	}

	return applications, nil
}

// GetApplicationById Retrieves an application by id
func GetApplicationById(id int) (*Application, error) {
	var application *Application

	result := SQL.
		Where("id = ? AND active = 1", id).
		First(&application)

	if result.Error != nil {
		return nil, result.Error
	}

	return application, nil
}

// SetActiveStatus Sets an applications active status
func (app *Application) SetActiveStatus(active bool) error {
	app.Active = active

	result := SQL.Model(&Application{}).
		Where("id = ?", app.Id).
		Update("active", app.Active)

	return result.Error
}

// SetClientSecret Sets a new secret for the application
func (app *Application) SetClientSecret(secret string) error {
	app.ClientSecret = secret

	result := SQL.Model(&Application{}).
		Where("id = ?", app.Id).
		Update("client_secret", app.ClientSecret)

	return result.Error
}
