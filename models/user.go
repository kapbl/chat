package models

import (
	"gorm.io/gorm"
)

type Channel struct {
	gorm.Model
	ChannelID string `gorm:"unique"`
	Name      string
}

type User struct {
	gorm.Model
	UserID   string // unique and generate by server
	Email    string // unique
	Username string // unique
	Password string
	Channels []Channel `gorm:"many2many:user_channels;"`
}
