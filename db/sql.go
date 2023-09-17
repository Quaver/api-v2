package db

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// The SQL database to use throughout the application
var SQL *gorm.DB

// ConnectMySQL Connects to a MySQL database
func ConnectMySQL() {
	cfg := config.Instance.SQL

	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", cfg.Username, cfg.Password, cfg.Host, cfg.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		logrus.Panic(err)
	}

	SQL = db
	logrus.Infof("Connected to MySQL database: %v/%v", cfg.Host, cfg.Database)
}
