package book

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"

	"github.com/i2eco/ecology/appgo/pkg/mus"

	"github.com/i2eco/ecology/appgo/model/mysql"

	"github.com/astaxie/beego"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/ecology/appgo/router/core"
)

func Index(c *core.Context) {
	var req ReqIndex
	err := c.Bind(&req)
	if err != nil {
		req.Page = 1
		req.Private = 1
	}

	if req.Page == 0 {
		req.Page = 1
	}

	c.Tpl().Data["SettingBook"] = true
	c.Tpl().Data["Private"] = req.Private //是否是私有文档
	pageIndex := req.Page
	books, totalCount, err := dao.Book.FindToPager(pageIndex, conf.Conf.Info.PageSize, c.Member().MemberId, req.Private)
	fmt.Println("books------>", books)
	if err != nil {
		c.Html404()
		return
	}
	if totalCount > 0 {
		//c.Tpl().Data["PageHtml"] = utils.GetPagerHtml(c.Context.Request.RequestURI, pageIndex, conf.PageSize, totalCount)
		c.Tpl().Data["PageHtml"] = utils.NewPaginations(conf.RollPage, totalCount, conf.PageSize, pageIndex, beego.URLFor("BookController.Index"), fmt.Sprintf("&private=%v", req.Private))
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}
	//处理封面图片
	for idx, book := range books {
		book.Cover = utils.ShowImg(book.Cover, "cover")
		books[idx] = book
	}
	b, err := json.Marshal(books)
	if err != nil || len(books) <= 0 {
		c.Tpl().Data["Result"] = template.JS("[]")
	} else {
		c.Tpl().Data["Result"] = template.JS(string(b))
	}
	c.Html("book/index")
}

// Dashboard 项目概要 .
func Dashboard(c *core.Context) {
	key := c.Param("key")
	if key == "" {
		c.Html404()
		return
	}

	book, err := dao.Book.ResultFindByIdentify(key, c.Member().MemberId)
	if err != nil {
		if err == dao.ErrPermissionDenied {
			c.Html404()
			return
		}
		c.Html404()
		return
	}

	c.Tpl().Data["Model"] = *book
	c.Html("book/dashboard")
}

// Setting 项目设置 .
func Setting(c *core.Context) {
	key := c.Param("key")
	if key == "" {
		c.Html404()
		return
	}

	member := c.Member()

	book, err := dao.Book.ResultFindByIdentify(key, member.MemberId)
	if err != nil {
		c.Html404()
		return
	}
	//如果不是创始人也不是管理员则不能操作
	if book.RoleId != conf.BookFounder && book.RoleId != conf.BookAdmin {
		c.Html404()
		return
	}

	if book.PrivateToken != "" {
		tipsFmt := "访问链接：%v  访问密码：%v"
		book.PrivateToken = fmt.Sprintf(tipsFmt, "/books/"+book.Identify+"?token="+book.PrivateToken)
	}

	//查询当前书籍的分类id
	if selectedCates, _ := dao.BookCategory.GetByBookId(book.BookId); len(selectedCates) > 0 {
		var maps = make(map[int]bool)
		for _, cate := range selectedCates {
			maps[cate.Id] = true
		}
		c.Tpl().Data["Maps"] = maps
	}

	c.Tpl().Data["Cates"], _ = dao.Category.GetCates(c.Context, -1, 1)
	book.DealCover()
	c.Tpl().Data["Model"] = book
	c.Html("book/setting")
}

// Users 用户列表.
func Users(c *core.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("page"))
	if pageIndex == 0 {
		pageIndex = 1
	}
	key := c.Param("key")
	if key == "" {
		c.Html404()
		return
	}

	member := c.Member()
	book, err := dao.Book.ResultFindByIdentify(key, member.MemberId)
	if err != nil {
		c.Html404()
		return
	}

	c.Tpl().Data["Model"] = *book
	pageSize := 10
	members, totalCount, _ := mysql.NewMemberRelationshipResult().FindForUsersByBookId(book.BookId, pageIndex, pageSize)

	for idx, member := range members {
		member.Avatar = mus.Oss.ShowImg(member.Avatar, "")
		members[idx] = member
	}

	if totalCount > 0 {
		html := utils.GetPagerHtml(c.Context.Request.RequestURI, pageIndex, pageSize, totalCount)
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}

	b, err := json.Marshal(members)
	if err != nil {
		c.Tpl().Data["Result"] = template.JS("[]")
	} else {
		c.Tpl().Data["Result"] = template.JS(string(b))
	}
	c.Html("book/users")
}
