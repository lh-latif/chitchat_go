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
	Id        uint      `gorm:"primaryKey;autoIncrement:true" json:"id"`
	Username  string    `json:"username"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserContact struct {
	Id          uint      `gorm:"primaryKey" json:"id"`
	UserId      uint      `json:"user_id"`
	ContactId   uint      `json:"contact_id"`
	ContactName string    `json:"contact_name"`
	CreatedAt   time.Time `json:"created_at"`
}
