package adminarea

import (
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/router/core"
)

func List(c *core.Context) {
	req := ReqList{}
	if err := c.Bind(&req); err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	total, list := dao.Area.ListPage(c.Context, mysql.Conds{}, &req.ReqPage)
	c.JSONList(list, req.ReqPage.Current, req.ReqPage.PageSize, total)
}
