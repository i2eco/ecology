package dao

import (
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/jinzhu/gorm"
)

func (m *memberToken) FindByFieldFirst(field string, value interface{}) (resp *mysql.MemberToken, err error) {
	err = mus.Db.Where("token = ?", value).Order("token_id desc").Find(resp).Error
	return
}

//插入或者更新
func (this *memberToken) InsertOrUpdate(db *gorm.DB, data *mysql.MemberToken) (err error) {
	if data.TokenId > 0 {
		err = db.Save(data).Error
		return
	}
	err = db.Create(data).Error
	return
}
