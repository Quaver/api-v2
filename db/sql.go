package db

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// The SQL database to use throughout the application
var SQL *gorm.DB

const testConfigPath string = "../config.json"

// ConnectMySQL Connects to a MySQL database
func ConnectMySQL() {
	cfg := config.Instance.SQL

	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", cfg.Username, cfg.Password, cfg.Host, cfg.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		logrus.Panic(err)
	}

	SQL = db
	logrus.Infof("Connected to MySQL database: %v/%v", cfg.Host, cfg.Database)
}

// CloseMySQL Closes the MySQL database connection
func CloseMySQL() {
	db, err := SQL.DB()

	if err != nil {
		logrus.Panic(err)
	}

	err = db.Close()

	if err != nil {
		logrus.Panic(err)
	}
}
