package awesome

import (
	"math"

	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/model/trans"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/ecology/appgo/router/core"
	"github.com/jinzhu/gorm"
)

func Index(c *core.Context) {
	req := ReqAwesome{}
	err := c.Bind(&req)
	if err != nil {
		// 首页就用默认值
	}
	pageIndex := req.Page
	if pageIndex == 0 {
		pageIndex = 1
	}

	//每页显示24个，为了兼容Pad、mobile、PC
	pageSize := 24
	reqPage := &trans.ReqPage{
		Current:  pageIndex,
		PageSize: 24,
		Sort:     "git_updated_at desc",
	}

	totalCount, lists := dao.Awesome.ListPage(c.Context, mysql.Conds{
		"git_name": mysql.Cond{"!=", ""},
	}, reqPage)

	//books, totalCount, err := dao.Book.HomeData(pageIndex, pageSize, mysql.BookOrder(req.Tab), bookType, cid)
	if err != nil {
		c.Html404()
		return
	}
	if totalCount > 0 {
		html := utils.NewPaginations(conf.RollPage, totalCount, pageSize, pageIndex, "/awesome", "")
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}
	c.Tpl().Data["Type"] = "latest"
	c.Tpl().Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))

	c.Tpl().Data["IsAwesome"] = true

	c.Tpl().Data["Lists"] = lists

	c.Html("awesome/index")
}

func Click(c *core.Context) {
	q1 := c.Query("q1") // github
	q2 := c.Query("q2") // web home url

	if q1 != "" {
		go func() {
			mus.Db.Model(mysql.Awesome{}).Where("html_url = ?", q1).Updates(mysql.Ups{
				"read_count": gorm.Expr("read_count+?", 1),
			})
		}()
		c.Redirect(302, q1)
		return
	}

	if q2 != "" {
		go func() {
			mus.Db.Model(mysql.Awesome{}).Where("home_page = ?", q2).Updates(mysql.Ups{
				"read_count": gorm.Expr("read_count+?", 1),
			})
		}()
		c.Redirect(302, q2)
		return
	}

}
