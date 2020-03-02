package tool

import (
	"fmt"
	"github.com/i2eco/ecology/appgo/pkg/tool/jsontostruct"
	"github.com/i2eco/ecology/appgo/router/core"
)

// json转struct
func JsonToStruct(c *core.Context) {
	b := c.PostForm("json")
	fmt.Println("b------>", b)
	f, err := jsontostruct.ParseJson([]byte(b))
	if err != nil {
		c.JSONErrTips("解析失败", err)
		return
	}
	jsontostruct.WriteGo(c.Writer, f, "Example")
}
