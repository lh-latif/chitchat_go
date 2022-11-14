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

	conn.Migrator().AutoMigrate(User{}, UserContact{})

	return conn
}

type User struct {
	Id        uint `gorm:"primaryKey"`
	Username  string
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserContact struct {
	Id          uint `gorm:"primaryKey"`
	UserId      uint
	ContactId   uint
	ContactName string
	CreatedAt   time.Time
}
