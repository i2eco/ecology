package adminseo

import (
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/constx"
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
	total, list := dao.Seo.ListPage(c.Context, mysql.Conds{}, &req.ReqPage)
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
	createInfo := &mysql.Seo{
		Page:        req.Page,
		Statement:   req.Statement,
		Title:       req.Title,
		Keywords:    req.Keywords,
		Description: req.Description,
	}
	err = dao.Seo.Create(c.Context, mus.Db, createInfo)
	if err != nil {
		c.JSONErrTips("创建失败1", err)
		return
	}
	_, err = mus.Mixcache.Set(constx.SEO_REDIS_PREFIX_KEY+req.Page, createInfo, 0)
	if err != nil {
		c.JSONErrTips("创建失败2", err)
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
	err = mus.Db.Model(mysql.Seo{}).Where("id = ?", req.Id).UpdateColumns(&mysql.Seo{
		Page:        req.Page,
		Statement:   req.Statement,
		Title:       req.Title,
		Keywords:    req.Keywords,
		Description: req.Description,
	}).Error
	if err != nil {
		c.JSONErrTips("更新失败", err)
		return
	}
	resp, err := dao.Seo.Info(c.Context, req.Id)
	if err != nil {
		c.JSONErrTips("获取信息失败", err)
		return
	}
	_, err = mus.Mixcache.Set(constx.SEO_REDIS_PREFIX_KEY+resp.Page, resp, 0)
	if err != nil {
		c.JSONErrTips("更新失败2", err)
		return
	}
	c.JSONOK()

}
