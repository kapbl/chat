package main

import (
	"kkj123/database"
	"kkj123/router"
)

func main() {
	// 启动数据库
	database.InitDatabse()
	// 启动服务器
	router.InitRouter()
	// todo ss
}
