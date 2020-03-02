package tool

import (
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/router/core"
)

func Index(c *core.Context) {
	lists, _ := dao.Tool.List(c.Context, mysql.Conds{}, "updated_at desc")
	c.Tpl().Data["Lists"] = lists
	c.Tpl().Data["IsTool"] = true
	c.Html("tool/index")
}

func Info(c *core.Context) {
	name := c.Param("toolName")
	if name != "jsontostruct" && name != "timestamp" {
		c.Html404()
		return
	}
	c.Tpl().Data["IsTool"] = true
	c.Html("tool/" + name)
}
