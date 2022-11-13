package database

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	conn, err := gorm.Open(sqlite.Open("./database.sqlite"))
	if err != nil {
		panic("error")
	}

	conn.Migrator().AutoMigrate(User{})

	return conn
}

type User struct {
	Id        uint   `gorm:"primaryKey"`
	Username  string `gorm:"primaryKey"`
	Hash      string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
