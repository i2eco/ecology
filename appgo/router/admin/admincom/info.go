package admincom

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
	conds := mysql.Conds{}

	// 搜索名称
	if req.Name != "" {
		conds["name"] = mysql.Cond{
			"like",
			req.Name,
		}
	}

	// 搜索繁育人
	if req.Author != "" {
		conds["author"] = mysql.Cond{
			"like",
			req.Author,
		}
	}

	// 搜索地址
	if req.Author != "" {
		conds["address"] = mysql.Cond{
			"like",
			req.Address,
		}
	}

	req.Sort = "id desc"
	total, list := dao.Com.ListPage(c.Context, conds, &req.ReqPage)

	//处理封面图片
	for idx, comInfo := range list {
		comInfo.Cover = mus.Oss.ShowImg(comInfo.Cover, "x1")
		list[idx] = comInfo
	}

	c.JSONList(list, req.ReqPage.Current, req.ReqPage.PageSize, total)
}
