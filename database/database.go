package database

import (
	"kkj123/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabse() {
	dsn := "root:0220059cyCY@tcp(127.0.0.1:3306)/chat?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println(err)
	}
	db.AutoMigrate(&models.User{})
	DB = db
}

func GetDatabase() *gorm.DB {
	return DB
}
