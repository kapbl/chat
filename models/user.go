package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID   string // unique and generate by server
	Email    string // unique
	Username string // unique
	Password string
}
