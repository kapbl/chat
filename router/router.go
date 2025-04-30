package router

import (
	"kkj123/handles"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func InitRouter() {
	server := echo.New()
	server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://127.0.0.1:8081", "http://192.168.10.68:8081", "http://localhost:8081"}, // 前端地址
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
	BindRouter(server)
	handles.IninDafaultChannel()
	handles.InitChannels()
	server.Start("0.0.0.0:8080")
}

func BindRouter(server *echo.Echo) {
	// 设置登录和注册的路由组
	auth := server.Group("/auth")
	auth.POST("/login", handles.Login)
	auth.POST("/register", handles.Register)
	auth.POST("/createUser", handles.CreateUser) // 创建用户的路由

	server.GET("/ws", handles.HandleWebSocker)
	// 设置群组的路由组
	api := server.Group("/api")
	api.Use(handles.JWTValidator([]byte("my_secret")))
	api.GET("/SearchChannels", handles.ChannelSearch)         // 获取频道列表
	api.POST("/JoinChannel", handles.JoinChannel)             // 加入频道
	api.POST("/CreateChannel", handles.CreateChannel)         // 加入频道
	api.GET("/InitJoinedChannel", handles.InitJoinedChannels) // 加入频道
	api.PUT("/ModifyChannel", handles.ModifyChannel)
	api.DELETE("/DeleteChannel", handles.DeleteChannel)
}
