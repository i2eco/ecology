package record

import (
	"github.com/astaxie/beego"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/router/core"
	"time"
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
	bookId, _ := this.GetInt(":book_id")
	if bookId > 0 {
		m := new(mysql.ReadRecord)
		if rl, count, err = m.List(c.Member().MemberId, bookId); err == nil && len(rl) > 0 {
			rp, _ = m.Progress(c.Member().MemberId, bookId)
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
		errCode = 1
		message = "您当前没有阅读记录"
	}
	c.JSONOK(map[string]interface{}{
		"lists":    lists,
		"count":    count,
		"progress": rp,
		"clear":    beego.URLFor("RecordController.Clear", ":book_id", bookId),
	})
}

//重置阅读进度(清空阅读历史)
func Clear() {
	bookId, _ := this.GetInt(":book_id")
	if bookId > 0 {
		m := new(mysql.ReadRecord)
		if err := m.Clear(c.Member().MemberId, bookId); err != nil {
			mus.Logger.Error(err)
		}
	}
	//不管删除是否成功，均返回成功
	c.JSONErrStr(0, "重置阅读进度成功")
}

//删除单条阅读历史
func Delete() {
	docId, _ := this.GetInt(":doc_id")
	if docId > 0 {
		if err := new(mysql.ReadRecord).Delete(c.Member().MemberId, docId); err != nil {
			mus.Logger.Error(err)
		}
	}
	c.JSONErrStr(0, "删除成功")
}
