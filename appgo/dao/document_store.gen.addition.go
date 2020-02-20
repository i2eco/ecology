package dao

import (
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/jinzhu/gorm"
)

//插入或者更新
func (*documentStore) InsertOrUpdate(db *gorm.DB, ds *mysql.DocumentStore) (err error) {
	var one mysql.DocumentStore
	db.Where("document_id = ?", ds.DocumentId).Find(&one)
	if one.DocumentId > 0 {
		err = db.Model(mysql.DocumentStore{}).Where("document_id = ?", ds.DocumentId).UpdateColumns(ds).Error
	} else {
		err = db.Create(ds).Error
	}
	return
}

//查询markdown内容或者content内容
func (this *documentStore) GetFiledById(docId interface{}, field string) string {
	var ds = mysql.DocumentStore{}
	if field != "markdown" {
		field = "content"
	}

	mus.Db.Select(field).Where("document_id = ?", docId).Find(&ds)
	if field == "content" {
		return ds.Content
	}
	return ds.Markdown
}

//查询markdown内容或者content内容
func (this *documentStore) GetById(docId interface{}) (ds mysql.DocumentStore, err error) {
	err = mus.Db.Where("document_id = ?", docId).Find(&ds).Error
	return
}

//查询markdown内容或者content内容
func (this *documentStore) DeleteById(docId ...interface{}) {
	if len(docId) > 0 {
		mus.Db.Where("document_id in (?)", docId).Delete(mysql.DocumentStore{})
	}
}
