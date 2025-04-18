package handles

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestJoinGroup(t *testing.T) {
	echo := echo.New()
	echo.POST("/joinGroup", JoinGroup)

	// 模拟一个请求
	type mess struct {
		originUsername string //请求的用户名
		groupID        string //请求的body
	}
	testCase := mess{
		originUsername: "testUser",
		groupID:        "testGroup",
	}
	req := httptest.NewRequest("POST", "/joinGroup", strings.NewReader(`{"originUsername":"`+testCase.originUsername+`","groupID":"`+testCase.groupID+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	echo.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}
