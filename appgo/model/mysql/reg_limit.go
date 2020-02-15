package mysql

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"time"
)

type RegLimit struct {
	Id          int
	Ip          string    `orm:"size(15);index"`
	CreatedAt   time.Time `orm:"index"`
	DailyRegNum int       `orm:"-"`
	HourRegNum  int       `orm:"-"`
	RealIPField string    `orm:"-"`
}

func (rl *RegLimit) CheckIPIsAllowed(ip string) (allowHour, allowDaily bool) {
	now := time.Now()
	o := orm.NewOrm()
	if rl.HourRegNum > 0 {
		hourBefore := now.Add(-1 * time.Hour)
		cnt, _ := o.QueryTable(rl).Filter("ip", ip).Filter("created_at__gt", hourBefore).Filter("created_at__lt", now).Count()
		if int(cnt) >= rl.HourRegNum {
			return false, true
		}
	}

	DayBefore := now.Add(-24 * time.Hour)
	if rl.DailyRegNum > 0 {
		cnt, _ := o.QueryTable(rl).Filter("ip", ip).Filter("created_at__gt", DayBefore).Filter("created_at__lt", now).Count()
		if int(cnt) >= rl.DailyRegNum {
			return true, false
		}
	}
	o.QueryTable(rl).Filter("created_at__lt", DayBefore).Delete()
	return true, true
}

func (rl *RegLimit) Insert(ip string) (err error) {
	rl.Ip = ip
	rl.CreatedAt = time.Now()
	if _, err = orm.NewOrm().Insert(rl); err != nil {
		beego.Error(err)
	}
	return
}
