package dao

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/jinzhu/gorm"
)

//查询所有分类
//@param            pid         -1表示不限（即查询全部），否则表示查询指定pid的分类
//@param            status      -1表示不限状态(即查询所有状态的分类)，0表示关闭状态，1表示启用状态
func (m *category) GetCates(c *gin.Context, pid int, status int) (resp []mysql.Category, err error) {
	conds := mysql.Conds{}

	if pid > -1 {
		conds["pid"] = pid
	}

	if status == 0 || status == 1 {
		conds["status"] = status
	}
	resp, err = m.List(c, conds, "status desc,sort asc,title asc")

	return
}

//新增分类
func (this *category) AddCates(pid int, cates string) (err error) {
	slice := strings.Split(cates, "\n")
	if len(slice) == 0 {
		return
	}

	for _, item := range slice {
		if item = strings.TrimSpace(item); item != "" {
			var cate = mysql.Category{
				Pid:    pid,
				Title:  item,
				Status: true,
			}
			var one mysql.Category
			err = mus.Db.Where("title = ?", cate.Title).Find(&one).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				continue
			}
			err = mus.Db.Create(&cate).Error

		}
	}
	return
}

//根据字段更新内容
func (this *category) UpdateByField(id int, field, val string) (err error) {
	err = mus.Db.Model(mysql.Category{}).Where("id = ?", id).Updates(mysql.Ups{field: val}).Error
	return
}
