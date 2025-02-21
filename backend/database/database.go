package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	db, err := gorm.Open(sqlite.Open("library.db"), &gorm.Config{})
	if err != nil {
		panic("Database connection failed")
	}
	db.AutoMigrate(&Users{}, &Book{}, &IssueRegistery{}, &Library{}, &ReaderRequestEvents{})
	DB = db
}
