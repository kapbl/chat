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
	UserID   string `gorm:"unique"`
	Email    string
	Username string
	Password string
	Channels []Channel `gorm:"many2many:user_channels;"`
}

type UserChannel struct {
	UserID    uint    `gorm:"primaryKey"`
	ChannelID uint    `gorm:"primaryKey"`
	User      User    `gorm:"foreignKey:UserID; constraint:OnDelete:CASCADE"`
	Channel   Channel `gorm:"foreignKey:ChannelID; constraint:OnDelete:RESTRICT;"`
}
