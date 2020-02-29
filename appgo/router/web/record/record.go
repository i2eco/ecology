package record

import (
	"strconv"
	"time"

	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/pkg/code"

	"github.com/astaxie/beego"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/router/core"
)

//获取阅读记录列表
func List(c *core.Context) {
	var (
		lists []map[string]interface{}
		err   error
		rl    []mysql.RecordList
		rp    mysql.ReadProgress
		//errCode int
		//message string = "数据查询成功"
		count int64
	)
	bookId, _ := strconv.Atoi(c.Param("bookId"))
	if bookId > 0 {
		m := new(mysql.ReadRecord)
		if rl, count, err = m.List(c.Member().MemberId, bookId); err == nil && len(rl) > 0 {
			rp, _ = dao.ReadRecord.Progress(c.Member().MemberId, bookId)
			for _, item := range rl {
				var list = make(map[string]interface{})
				list["title"] = item.Title
				list["url"] = "/read/" + rp.BookIdentify + "/" + item.Identify
				list["time"] = time.Unix(int64(item.CreateAt), 0).Format("01-02 15:04")
				list["del"] = beego.URLFor("RecordController.Delete", ":doc_id", item.DocId)
				lists = append(lists, list)
			}
		}
	}
	if len(lists) == 0 {
		// "您当前没有阅读记录"
		c.JSONErr(code.MsgErr, nil)
	}
	c.JSONOK(map[string]interface{}{
		"lists":    lists,
		"count":    count,
		"progress": rp,
		"clear":    "/record/clear/" + strconv.Itoa(bookId),
	})
}

//重置阅读进度(清空阅读历史)
func Clear(c *core.Context) {
	bookId, _ := strconv.Atoi(c.Param("bookId"))
	if bookId > 0 {
		if err := dao.ReadRecord.Clear(c.Member().MemberId, bookId); err != nil {
			mus.Logger.Error(err.Error())
		}
	}
	//不管删除是否成功，均返回成功
	c.JSONOK()
}

//删除单条阅读历史
func Delete(c *core.Context) {
	docId, _ := strconv.Atoi(c.Param("docId"))
	if docId > 0 {
		if err := dao.ReadRecord.DeleteXX(c.Member().MemberId, docId); err != nil {
			mus.Logger.Error(err.Error())
			c.JSONErr(code.MsgErr, err)
			return
		}
	}
	c.JSONOK()
}
