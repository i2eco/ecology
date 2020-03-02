package admintool

import (
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/router/core"
)

func List(c *core.Context) {
	req := ReqList{}
	if err := c.Bind(&req); err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	total, list := dao.Tool.ListPage(c.Context, mysql.Conds{}, &req.ReqPage)
	c.JSONList(list, req.ReqPage.Current, req.ReqPage.PageSize, total)
}

func Info(c *core.Context) {
	req := ReqInfo{}
	err := c.Bind(&req)
	if err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	resp, err := dao.Seo.Info(c.Context, req.Id)
	if err != nil {
		c.JSONErrTips("获取信息失败", err)
		return
	}
	c.JSONOK(resp)
}

func Create(c *core.Context) {
	req := ReqCreate{}
	err := c.Bind(&req)
	if err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	createInfo := &mysql.Tool{
		Name:     req.Name,
		Desc:     req.Desc,
		Identify: req.Identify,
		Cover:    req.Cover,
		Uid:      c.AdminUid(),
	}
	err = dao.Tool.Create(c.Context, mus.Db, createInfo)
	if err != nil {
		c.JSONErrTips("创建失败1", err)
		return
	}
	c.JSONOK()

}

func Update(c *core.Context) {
	req := ReqUpdate{}
	err := c.Bind(&req)
	if err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	err = mus.Db.Model(mysql.Tool{}).Where("id = ?", req.Id).UpdateColumns(&mysql.Tool{
		Name:     req.Name,
		Desc:     req.Desc,
		Identify: req.Identify,
		Cover:    req.Cover,
		Uid:      c.AdminUid(),
	}).Error
	if err != nil {
		c.JSONErrTips("更新失败", err)
		return
	}
	c.JSONOK()
}
