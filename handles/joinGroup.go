package handles

import (
	"sync"

	"github.com/labstack/echo/v4"
)

type Group struct {
	Inclient   []string // 这个群的所有客户端
	Group_id   string   // 群id
	Group_name string   // 群名称
}

var groups = make(map[string]*Group) // 群组列表
var groupMu sync.Mutex

func InitGroup() {
	groups["g100"] = &Group{
		Inclient: make([]string, 0),
		Group_id: "g100",
	}
}

// 加入一个群组
func JoinGroup(ctx echo.Context) error {
	type mess struct {
		OriginUsername string `json:"originUsername"` //请求的用户名
		OriginUserID   string `json:"originUserID"`   //请求的用户id
		GroupID        string `json:"groupID"`        //请求的body
	}
	rec := mess{}
	if err := ctx.Bind(&rec); err != nil {
		ctx.JSON(400, map[string]string{"message": "请求参数错误"})
	}
	groupMu.Lock()
	defer groupMu.Unlock()
	if _, ok := groups[rec.GroupID]; !ok {
		ctx.JSON(400, map[string]string{"message": "群组不存在"})
		return nil
	}
	groups[rec.GroupID].Inclient = append(groups[rec.GroupID].Inclient, rec.OriginUserID)
	ctx.JSON(200, map[string]string{
		"message": "加入群组成功"})
	return nil
}
