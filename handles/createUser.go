package handles

import "github.com/labstack/echo/v4"

// 在用户登录或者注册后，创建一个用户
func CreateUser(ctx echo.Context) error {
	type mess struct {
		Username string `json:"username"` // 请求的用户名
		Password string `json:"password"` // 请求的密码
	}
	rec := mess{}
	if err := ctx.Bind(&rec); err != nil {
		ctx.JSON(400, map[string]string{"message": "请求参数错误"})
		return nil
	}
	if rec.Username == "" || rec.Password == "" {
		ctx.JSON(400, map[string]string{"message": "用户名或密码不能为空"})
		return nil
	}
	ctx.JSON(200, map[string]string{"message": "创建用户成功"})
	return nil
}
