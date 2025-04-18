package main

import (
	"kkj123/database"
	"kkj123/handles"

	"github.com/labstack/echo/v4"
)

func main() {
	// 启动数据库
	database.InitDatabse()
	// 启动服务器
	server := echo.New()
	// 设置websocket的路由
	server.GET("/ws", handles.HandleWebSocker)
	// 设置登录和注册的路由组
	api := server.Group("/api")
	api.POST("/login", handles.Login)
	api.POST("/register", handles.Register)
	api.POST("/createUser", handles.CreateUser) // 创建用户的路由
	// 设置群组的路由组
	api.POST("/joinGroup", handles.JoinGroup)
	handles.InitGroup()
	server.Start(":8080")
}
