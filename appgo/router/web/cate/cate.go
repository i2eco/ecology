package cate

import (
	"fmt"
	"strconv"

	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/router/core"
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

	c.Tpl().Data["IsCate"] = true
	c.Tpl().Data["Recommends"], _, _ = dao.Book.HomeData(1, 12, mysql.OrderLatestRecommend, "", 0)
	c.Html("cates/list")
}
