package handles

import (
	"kkj123/database"
	"kkj123/models"

	"github.com/labstack/echo/v4"
)

func Login(ctx echo.Context) error {
	loginData := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := ctx.Bind(&loginData); err != nil {
		return ctx.JSON(400, map[string]string{"error": "请求格式错误"})
	}
	var user models.User
	if err := database.DB.Where("email = ? AND password = ?", loginData.Email, loginData.Password).First(&user).Error; err != nil {
		return ctx.JSON(401, map[string]string{"error": "邮箱或密码错误"})
	}

	return ctx.JSON(200, map[string]string{"message": "登录成功"})
}
