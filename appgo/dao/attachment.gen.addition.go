package dao

import (
	"os"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/pkg/utils"
)

func (m *attachment) FindListByDocumentId(docId int) (attaches []*mysql.Attachment, err error) {
	err = mus.Db.Where("document_id = ?", docId).Order("attachment_id asc").Find(&attaches).Error
	return attaches, err
}

func (m *attachment) Find(id int) (attach *mysql.Attachment, err error) {
	if id <= 0 {
		return nil, ErrInvalidParameter
	}
	err = mus.Db.Where("attachment_id = ?", id).Find(attach).Error
	return
}

//分页查询附件
func (m *attachment) FindToPager(pageIndex, pageSize int) (attachList []*mysql.AttachmentResult, totalCount int64, err error) {
	err = mus.Db.Model(mysql.Attachment{}).Count(&totalCount).Error
	if err != nil {
		return
	}
	offset := (pageIndex - 1) * pageSize

	var list []*mysql.Attachment

	err = mus.Db.Order("attachment_id desc").Offset(offset).Limit(pageSize).Find(&list).Error
	if err != nil {
		return
	}

	for _, item := range list {
		attach := &mysql.AttachmentResult{}
		attach.Attachment = *item
		attach.FileShortSize = utils.FormatBytes(int64(attach.FileSize))

		var book mysql.Book
		if e := mus.Db.Where("book_id = ?", item.BookId).Find(&book).Error; e == nil {
			attach.BookName = book.BookName
		} else {
			attach.BookName = "[不存在]"
		}

		var doc mysql.Document
		if e := mus.Db.Where("document_id = ?", item.DocumentId).Find(&doc).Error; e == nil {
			attach.DocumentName = doc.DocumentName
		} else {
			attach.DocumentName = "[不存在]"
		}
		attach.LocalHttpPath = strings.Replace(item.FilePath, "\\", "/", -1)
		attachList = append(attachList, attach)
	}

	return
}

func (a *attachment) DeleteFilePath(param *mysql.Attachment) (err error) {
	o := orm.NewOrm()

	if _, err = o.Delete(param); err != nil {
		return err
	}

	return os.Remove(param.FilePath)
}

func (m *attachment) ResultFind(id int) (*mysql.AttachmentResult, error) {
	attachResult := &mysql.AttachmentResult{}

	attach := &mysql.Attachment{}
	err := mus.Db.Where("attachment_id = ?", id).Find(attach).Error

	if err != nil {
		return attachResult, err
	}

	attachResult.Attachment = *attach

	book := mysql.NewBook()

	err = mus.Db.Where("book_id = ?", attach.BookId).Find(book).Error
	if err != nil {
		attachResult.BookName = "[不存在]"
	} else {
		attachResult.BookName = book.BookName
	}

	doc := mysql.NewDocument()

	err = mus.Db.Where("document_id = ?", attach.DocumentId).Find(doc).Error
	if err != nil {
		attachResult.DocumentName = "[不存在]"
	} else {
		attachResult.DocumentName = doc.DocumentName
	}

	if attach.CreateAt > 0 {
		member := mysql.NewMember()
		err = mus.Db.Where("member_id = ?", attach.CreateAt).Find(member).Error
		if err == nil {
			attachResult.Account = member.Account
		}
	}
	attachResult.FileShortSize = utils.FormatBytes(int64(attach.FileSize))
	attachResult.LocalHttpPath = strings.Replace(attachResult.FilePath, "\\", "/", -1)

	return attachResult, nil
}
