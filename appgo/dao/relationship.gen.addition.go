package dao

import (
	"errors"

	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/jinzhu/gorm"
)

func (m *relationship) Find(id int) (resp *mysql.Relationship, err error) {
	err = mus.Db.Where("relationship_id = ?", id).Find(resp).Error
	return
}

//查询指定项目的创始人.
func (m *relationship) FindFounder(bookId int) (resp *mysql.Relationship, err error) {
	err = mus.Db.Where("book_id = ? and role_id = ?", bookId, 0).Find(resp).Error
	return
}

func (m *relationship) UpdateRoleId(bookId, memberId, roleId int) (resp *mysql.Relationship, err error) {

	oneBook := mysql.Book{}
	err = mus.Db.Where("book_id =?", bookId).Find(&oneBook).Error
	if err != nil {
		return
	}

	oneRelationShip := mysql.Relationship{}
	err = mus.Db.Where("member_id = ? and book_id = ?", memberId, bookId).Find(&oneRelationShip).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}

	if err == gorm.ErrRecordNotFound {
		oneRelationShip.BookId = bookId
		oneRelationShip.MemberId = memberId
		oneRelationShip.RoleId = roleId
	}

	if oneRelationShip.RoleId == conf.Conf.Info.BookFounder {
		err = errors.New("不能变更创始人的权限")
		return
	}

	oneRelationShip.RoleId = roleId

	if oneRelationShip.RelationshipId > 0 {
		err = mus.Db.UpdateColumns(oneRelationShip).Error
		return
	}
	err = mus.Db.Create(&oneRelationShip).Error
	return

}

func (m *relationship) FindForRoleId(bookId, memberId int) (roleId int, err error) {
	var one mysql.Relationship
	err = mus.Db.Select("role_id").Where("book_id = ? AND member_id = ?", bookId, memberId).Find(&one).Error
	if err != nil {
		roleId = 0
		return
	}
	roleId = one.RoleId
	return
}

func (m *relationship) FindByBookIdAndMemberId(bookId, memberId int) (resp *mysql.Relationship, err error) {
	resp = &mysql.Relationship{}
	err = mus.Db.Where("book_id = ? and member_id = ?", bookId, memberId).Find(resp).Error
	return
}

func (m *relationship) Insert(db *gorm.DB, create *mysql.Relationship) error {
	return db.Create(create).Error
}

func (m *relationship) Update(db *gorm.DB, update *mysql.Relationship) error {
	return db.UpdateColumns(update).Error
}

func (m *relationship) DeleteByBookIdAndMemberId(bookId, memberId int) (err error) {

	var one mysql.Relationship
	err = mus.Db.Where("book_id = ? and member_id = ?", bookId, memberId).Find(&one).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}

	if err == gorm.ErrRecordNotFound {
		return errors.New("用户未参与该项目")
	}
	if one.RoleId == conf.BookFounder {
		return errors.New("不能删除创始人")
	}
	err = mus.Db.Delete(&one).Error
	if err != nil {
		return
	}
	return

}

func (m *relationship) Transfer(bookId, founderId, receiveId int) (err error) {

	var founder mysql.Relationship
	err = mus.Db.Where("book_id = ? and member_id = ?", bookId, founderId).Find(&founder).Error

	if err != nil {
		return
	}
	if founder.RoleId != conf.BookFounder {
		return errors.New("转让者不是创始人")
	}

	var receiver mysql.Relationship
	err = mus.Db.Where("book_id = ? and member_id = ?", bookId, founderId).Find(&receiver).Error

	if err != gorm.ErrRecordNotFound && err != nil {
		return
	}

	transdb := mus.Db.Begin()

	founder.RoleId = conf.BookAdmin

	receiver.MemberId = receiveId
	receiver.RoleId = conf.BookFounder
	receiver.BookId = bookId

	if err := Relationship.Update(transdb, &founder); err != nil {
		transdb.Rollback()
		return err
	}

	if receiver.RelationshipId > 0 {
		if err := Relationship.Update(transdb, &receiver); err != nil {
			transdb.Rollback()
			return err
		}
	}

	if err := Relationship.Insert(transdb, &receiver); err != nil {
		transdb.Rollback()
		return err
	}
	transdb.Commit()
	return
}
