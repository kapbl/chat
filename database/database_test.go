package database

import (
	"kkj123/models"
	"testing"
)

func TestInsertData(t *testing.T) {
	user1 := models.User{
		Username: "caoyan",
		Password: "123456",
	}
	InitDatabse()
	if DB == nil {
		t.Error("DB is nil")
	}
	// 检查数据是否已经存在
	var count int64
	DB.Model(&models.User{}).Where("username = ?", user1.Username).Count(&count)
	if count > 0 {
		t.Error("User already exists in the database")
		return
	}
	// 如果不存在，则继续执行插入语句
	tx := DB.Create(&user1)
	if tx.Error != nil {
		t.Error("Insert data failed:", tx.Error)
	} else {
		t.Log("Insert data success")
	}
}
