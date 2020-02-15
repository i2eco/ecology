package mysql

import (
	"github.com/astaxie/beego/orm"
)

//友链数据表
type FriendLink struct {
	Id     int    `gorm:"not null;"json:"id"`     //自增主键
	Sort   int    `gorm:"not null;"json:"sort"`   //排序
	Link   string `gorm:"not null;"json:"link"`   //链接地址
	Title  string `gorm:"not null;"json:"title"`  //链接名称
	Remark string `gorm:"not null;"json:"remark"` // 备注
	Status bool   `gorm:"not null;"json:"status"`
}

func (FriendLink) TableName() string {
	return "friend_link"
}

//添加友情链接
func (this *FriendLink) Add(title, link string) (err error) {
	var fl = FriendLink{
		Title:  title,
		Link:   link,
		Sort:   0,
		Status: true,
	}
	_, err = orm.NewOrm().Insert(&fl)
	return
}

//根据字段更新友链
func (this *FriendLink) Update(id int, field string, value interface{}) (err error) {
	_, err = orm.NewOrm().QueryTable(this).Filter("id", id).Update(orm.Params{field: value})
	return
}

//删除友情链接
func (this *FriendLink) Del(id int) (err error) {
	var link = FriendLink{Id: id}
	_, err = orm.NewOrm().Delete(&link)
	return
}

//查询友链列表
//all表示是否查询全部，当为false时，只查询启用状态的友链，否则查询全部
func (this *FriendLink) GetList(all bool) (links []FriendLink) {
	qs := orm.NewOrm().QueryTable("friend_link")
	if !all {
		qs = qs.Filter("status", 1)
	}
	qs.OrderBy("-status").OrderBy("sort").All(&links)
	return
}
