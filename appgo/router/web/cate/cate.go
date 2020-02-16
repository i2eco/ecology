package cate

import (
	"fmt"
	"strconv"

	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/router/core"
)

func Index(c *core.Context) {
	cid, _ := strconv.Atoi(c.Param("cid"))
	if cid > 0 {
		c.Redirect(302, "/"+c.Context.Request.RequestURI)
	}
	List(c)
}

//分类
func List(c *core.Context) {
	if cates, err := dao.Category.GetCates(c.Context, -1, 1); err == nil {
		fmt.Println("cates------>", cates)
		c.Tpl().Data["Cates"] = cates
	} else {
		mus.Logger.Error(err.Error())
	}

	c.GetSeoByPage("cate", map[string]string{
		"title":       "书籍分类",
		"keywords":    "文档托管,在线创作,文档在线管理,在线知识管理,文档托管平台,在线写书,文档在线转换,在线编辑,在线阅读,开发手册,api手册,文档在线学习,技术文档,在线编辑",
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Tpl().Data["IsCate"] = true
	c.Tpl().Data["Recommends"], _, _ = dao.Book.HomeData(1, 12, mysql.OrderLatestRecommend, "", 0)
	c.Html("cates/list")
}
