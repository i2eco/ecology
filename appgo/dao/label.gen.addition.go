package dao

import (
	"strings"

	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
)

func (m *label) FindFirst(field string, value interface{}) (resp *mysql.Label, err error) {
	err = mus.Db.Where(field+"=?", value).Find(&resp).Error
	return
}

//插入或更新标签.
func (m *label) InsertOrUpdate(labelName string) (err error) {
	var (
		cnt int
		one mysql.Label
	)
	err = mus.Db.Model(mysql.Label{}).Where("label_name=?", labelName).Count(&cnt).Error
	if err != nil {
		return
	}

	one.BookNumber = int(cnt) + 1
	one.LabelName = labelName
	if cnt == 0 {
		err = mus.Db.Create(&one).Error
		return
	}
	err = mus.Db.Model(mysql.Label{}).Where("label_name=?", labelName).UpdateColumns(one).Error
	return
}

//批量插入或更新标签.
func (m *label) InsertOrUpdateMulti(labels string) {
	if labels != "" {
		labelArray := strings.Split(labels, ",")
		for _, label := range labelArray {
			if label != "" {
				err := m.InsertOrUpdate(strings.TrimSpace(label))
				if err != nil {
					// todo log
				}
			}
		}
	}
}

//分页查找标签.
func (m *label) FindToPager(pageIndex, pageSize int, word ...string) (labels []*mysql.Label, totalCount int, err error) {
	var count int64

	qs := mus.Db
	if len(word) > 0 {
		qs.Where("label_name like %?%", word[0])
	}
	err = qs.Model(mysql.Label{}).Count(&count).Error
	totalCount = int(count)

	offset := (pageIndex - 1) * pageSize
	err = qs.Offset(offset).Limit(pageSize).Order("book_number desc").Find(&labels).Error
	return
}
