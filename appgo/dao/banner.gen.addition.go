package dao

import (
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/pkg/utils"
)

func (m *banner) Lists(t string) (banners []mysql.Banner, err error) {
	err = mus.Db.Where("type = ? and status = ?", t, true).Order("sort desc,id desc").Find(&banners).Error
	return
}

func (m *banner) All() (banners []mysql.Banner, err error) {
	err = mus.Db.Order("sort desc,status desc").Find(&banners).Error
	return
}

func (m *banner) UpdateXX(id int, field string, value interface{}) (err error) {
	err = mus.Db.Model(mysql.Banner{}).Where("id = ?", id).Updates(mysql.Ups{field: value}).Error
	return
}

func (m *banner) DeleteXX(id int) (err error) {
	var banner mysql.Banner
	mus.Db.Where("id= ?", id).Find(&banner)
	if banner.Id > 0 {
		err = mus.Db.Delete(&banner).Error
		if err == nil {
			utils.DeleteFile(banner.Image)
		}
	}
	return
}
