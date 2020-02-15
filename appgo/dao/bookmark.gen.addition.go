package dao

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"time"
)

//书签列表
type bookmarkList struct {
	Id           int       `json:"id,omitempty"`
	Title        string    `json:"title"`
	Identify     string    `json:"identify"`
	BookId       int       `json:"book_id"`
	Uid          int       `json:"uid"`
	DocId        int       `json:"doc_id"`
	CreateAt     int       `json:"-"`
	CreateAtTime time.Time `json:"created_at"`
}

//添加或移除书签（如果书签不存在，则添加书签，如果书签存在，则移除书签）
func (m *bookmark) InsertOrDelete(uid, docId int) (insert bool, err error) {
	if uid*docId == 0 {
		err = errors.New("用户id和文档id均不能为空")
		return
	}

	var (
		bookmark mysql.Bookmark
		doc      = mysql.Document{}
	)

	mus.Db.Where("uid = ? and doc_id = ?", uid, docId).Find(&bookmark)
	if bookmark.Id > 0 { //删除书签
		err = mus.Db.Where("id = ?", bookmark.Id).Delete(&bookmark).Error
		return
	}

	//新增书签
	//查询文档id是属于哪个文档项目
	mus.Db.Where("document_id = ?", docId).Find(&doc)

	bookmark.BookId = doc.BookId
	bookmark.CreateAt = int(time.Now().Unix())
	bookmark.Uid = uid
	bookmark.DocId = docId
	err = mus.Db.Create(&bookmark).Error
	insert = true
	return
}

//查询书签是否存在
func (m *bookmark) Exist(uid, docId int) (exist bool) {
	if uid*docId > 0 {
		var bk mysql.Bookmark
		mus.Db.Where("uid = ? and doc_id = ?", uid, docId).Find(&bk)
		return bk.Id > 0
	}
	return
}

//删除书签
//1、只有 bookId > 0，则删除bookId所有书签【用于文档项目被删除的情况】
//2、bookId>0 && uid > 0 ，删除用户的书籍书签【用户用户清空书签的情况】
//3、uid > 0 && docId>0 ，删除指定书签【用于删除某条书签】
//4、其余情况不做处理
func (m *bookmark) DeleteXX(uid, bookId, docId int) (err error) {
	q := mus.Db

	if bookId > 0 {
		err = q.Where("book_id = ?", bookId).Delete(&mysql.Bookmark{}).Error
	} else if bookId > 0 && uid > 0 {
		err = q.Where("book_id = ? and uid = ?", bookId, uid).Delete(&mysql.Bookmark{}).Error
	} else if uid > 0 && docId > 0 {
		err = q.Where("doc_id = ? and uid = ?", docId, uid).Delete(&mysql.Bookmark{}).Error
	}
	return
}

//查询书签列表
func (m *bookmark) ListXX(uid, bookId int) (bl []bookmarkList, rows int64, err error) {
	o := orm.NewOrm()
	fields := "b.id,d.document_name title,d.identify,b.book_id,b.uid,b.doc_id,b.create_at"
	sql := "select %v from " + mysql.Bookmark{}.TableName() + " b left join " + mysql.Document{}.TableName() + " d on b.doc_id=d.document_id where b.uid=? and b.book_id=? order by b.id desc limit 1000"
	sql = fmt.Sprintf(sql, fields)
	rows, err = o.Raw(sql, uid, bookId).QueryRows(&bl)
	return
}
