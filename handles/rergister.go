package handles

import (
	"errors"
	"kkj123/database"
	"kkj123/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// 注册路由的处理函数
func Register(ctx echo.Context) error {
	registerData := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := ctx.Bind(&registerData); err != nil {
		ctx.JSON(400, map[string]string{"message": "请求参数错误"})
		return err
	}
	// 注册数据格式通过
	// 检查数据库中是否有重复的用户名和电子邮箱
	if registerData.Username == "" || registerData.Password == "" {
		ctx.JSON(400, map[string]string{"message": "用户名或密码不能为空"})
		return nil
	}
	// 检查用户名和邮箱是否已存在
	// 通过用户名和邮箱查询
	// 这里的查询条件是 OR 的关系
	// 也就是只要有一个条件满足就可以
	var registerUser models.User
	if err := database.DB.Where("username = ? OR email = ?", registerData.Username, registerData.Email).First(&registerUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户名和邮箱都不存在，可以注册
			// 创建用户
			registerUser = models.User{
				Username: registerData.Username,
				Email:    registerData.Email,
				Password: registerData.Password,
				UserID:   GenerateSHA256UserID(registerData.Email),
			}
			// 判断是否插入成功
			if err := database.DB.Create(&registerUser).Error; err != nil {
				ctx.JSON(500, map[string]string{"message": "注册失败"})
				return err
			} else {
				ctx.JSON(200, map[string]string{"message": "注册成功"})
				return nil
			}
		} else {
			ctx.JSON(500, map[string]string{"message": "数据库查询错误"})
		}
	} else {
		if registerUser.Username == registerData.Username {
			ctx.JSON(409, map[string]string{"message": "用户名已存在"})
			return nil
		}
		if registerUser.Email == registerData.Email {
			ctx.JSON(409, map[string]string{"message": "邮箱已存在"})
			return nil
		}
	}
	return nil
}
