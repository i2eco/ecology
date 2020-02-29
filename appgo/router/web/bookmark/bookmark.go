package bookmark

import (
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/router/core"
)

//添加或者移除书签
func Bookmark(c *core.Context) {
	docId, _ := strconv.Atoi(c.Param("id"))
	if docId <= 0 {
		c.JSONErrStr(1, "收藏失败，文档id参数错误")
	}

	insert, err := dao.Bookmark.InsertOrDelete(c.Member().MemberId, docId)
	if err != nil {
		mus.Logger.Error(err.Error())
		if insert {
			c.JSONErrStr(1, "添加书签失败")
			return
		}
		c.JSONErrStr(1, "移除书签失败")
		return
	}

	if insert {
		c.JSONErrStr(0, "添加书签成功")
		return
	}
	c.JSONOK()
}

//获取书签列表
func List(c *core.Context) {
	bookId, _ := strconv.Atoi(c.Param("book_id"))
	if bookId <= 0 {
		c.JSONErrStr(1, "获取书签列表失败：参数错误")
		return
	}

	bl, rows, err := dao.Bookmark.ListXX(c.Member().MemberId, bookId)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.JSONErrStr(1, "获取书签列表失败")
		return
	}

	var (
		book  mysql.Book
		lists []map[string]interface{}
	)

	mus.Db.Select("identify").Where("book_id = ?", bookId).Find(&book)

	for _, item := range bl {
		var list = make(map[string]interface{})
		list["url"] = "/read/" + book.Identify + "/" + item.Identify
		list["title"] = item.Title
		list["doc_id"] = item.DocId
		list["del"] = beego.URLFor("BookmarkController.Bookmark", ":id", item.DocId)
		list["time"] = time.Unix(int64(item.CreateAt), 0).Format("01-02 15:04")
		lists = append(lists, list)
	}

	c.JSONOK(map[string]interface{}{
		"count":   rows,
		"book_id": bookId,
		"list":    lists,
	})
}
