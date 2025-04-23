package handles

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"kkj123/database"
	"kkj123/models"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
func GenerateSHA256UserID(email string) string {
	normalized := normalizeEmail(email)
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:]) // 64位十六进制
}

func Login(ctx echo.Context) error {
	loginData := struct {
		Usernmae string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := ctx.Bind(&loginData); err != nil {
		return ctx.JSON(400, map[string]string{"error": "请求格式错误"})
	}
	if loginData.Email == "" || loginData.Password == "" {
		return ctx.JSON(400, map[string]string{"error": "邮箱和密码不能为空"})
	}
	// 在数据库中查找这个邮箱地址的用户
	// 通过邮箱查询
	var user models.User
	if err := database.DB.Where("email = ? AND password = ?", loginData.Email, loginData.Password).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(401, map[string]string{"error": "邮箱和用户名或密码错误"})
		}
	}
	return ctx.JSON(200, map[string]string{"message": "登录成功"})
}
