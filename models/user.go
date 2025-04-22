package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID   string
	Email    string
	Username string
	Password string
}
