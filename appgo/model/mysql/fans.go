package mysql

import (
	"fmt"

	"github.com/i2eco/ecology/appgo/pkg/mus"
)

type FansResult struct {
	Uid      int
	Nickname string
	Avatar   string
	Account  string
}

//粉丝表
type Fans struct {
	Id     int `gorm:"not null;"json:"id"`          //自增主键
	Uid    int `gorm:"not null;index"json:"uid"`    //被关注的用户id
	FansId int `gorm:"not null;index"json:"fansId"` //粉丝id
}

func (Fans) TableName() string {
	return "fans"
}

// 多字段唯一键
func (this *Fans) TableUnique() [][]string {
	return [][]string{
		[]string{"Uid", "FansId"},
	}
}

//关注和取消关注
func (this *Fans) FollowOrCancel(uid, fansId int) (cancel bool, err error) {
	var fans Fans
	mus.Db.Where("uid = ? and fans_id = ?", uid, fansId).Where(&fans)

	if fans.Id > 0 { //已关注，则取消关注
		err = mus.Db.Delete(&fans).Error
		cancel = true
	} else { //未关注，则新增关注
		fans.Uid = uid
		fans.FansId = fansId
		err = mus.Db.Create(&fans).Error
	}
	return
}

//查询是否已经关注了用户
func (this *Fans) Relation(uid, fansId interface{}) (ok bool) {
	var fans Fans
	mus.Db.Where("uid = ? and fans_id = ?", uid, fansId).Where(&fans)
	return fans.Id != 0
}

//查询用户的粉丝（用户id作为被关注对象）
func (this *Fans) GetFansList(uid, page, pageSize int) (fans []FansResult, total int, err error) {
	err = mus.Db.Model(Fans{}).Where("uid = ?", uid).Count(&total).Error
	if err != nil {
		return
	}

	if total > 0 {
		sql := fmt.Sprintf(
			"select m.member_id uid,m.avatar,m.account,m.nickname from "+Member{}.TableName()+" m left join "+Fans{}.TableName()+" f on m.member_id=f.fans_id where f.uid=?  order by f.id desc limit %v offset %v",
			pageSize, (page-1)*pageSize,
		)
		err = mus.Db.Raw(sql, uid).Scan(&fans).Error
	}
	return
}

//查询用户的关注（用户id作为fans_id）
func (this *Fans) GetFollowList(fansId, page, pageSize int) (fans []FansResult, total int, err error) {
	err = mus.Db.Model(Fans{}).Where("fans_id = ?", fansId).Count(&total).Error
	if total > 0 {
		sql := fmt.Sprintf(
			"select m.member_id uid,m.avatar,m.account,m.nickname from md_members m left join md_fans f on m.member_id=f.uid where f.fans_id=?  order by f.id desc limit %v offset %v",
			pageSize, (page-1)*pageSize,
		)
		err = mus.Db.Raw(sql, fansId).Scan(&fans).Error
	}
	return
}
