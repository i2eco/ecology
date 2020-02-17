package dao

import (
	"fmt"
	"strconv"
	"time"

	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
)

//收藏或者取消收藏
//@param            uid         用户id
//@param            bid         书籍id
//@return           cancel      是否是取消收藏，只是标记属于取消还是收藏操作，err才表示执行操作成功与否
func (m *star) Star(uid, bid int) (cancel bool, err error) {
	var star mysql.Star
	mus.Db.Where("uid = ? and bid = ?", uid, bid).Find(&star)
	if star.Id > 0 { //取消收藏
		err = mus.Db.Table("star").Where("id = ?", star.Id).Delete(&mysql.Star{}).Error
		if err != nil {
			return
		}
		// todo sql 注入, transaction
		err = SetIncreAndDecre("book", "star", fmt.Sprintf("book_id=%v and star>0", bid), false, 1)
		if err != nil {
			return
		}
		cancel = true
	} else { //添加收藏
		cancel = false
		star = mysql.Star{
			Id:  0,
			Uid: uid,
			Bid: bid,
		}
		err = mus.Db.Create(&star).Error
		if err != nil {
			//收藏计数+1
			return
		}
		err = SetIncreAndDecre("book", "star", "book_id="+strconv.Itoa(bid), true, 1)
		if err != nil {
			return
		}

	}
	return
}

//是否收藏了文档
func (this *star) DoesStar(uid, bid interface{}) bool {
	var star mysql.Star
	mus.Db.Where("uid = ? and bid = ?", uid, bid).Find(&star)
	if star.Id > 0 {
		return true
	}
	return false
}

//获取收藏列表，查询项目信息
func (this *star) ListXX(uid, p, listRows int, order ...string) (cnt int64, books []mysql.StarResult, err error) {
	//根据用户id查询用户的收藏，先从收藏表中查询book_id
	sqlCount := `select count(s.bid) cnt from ` + mysql.Book{}.TableName() + ` b left join ` + mysql.Star{}.TableName() + ` s on s.bid=b.book_id where s.uid=? and b.privately_owned=0`
	var count mysql.DataCount

	mus.Db.Raw(sqlCount, uid).Scan(&count)

	//这里先暂时每次都统计一次用户的收藏数量。合理的做法是在用户表字段中增加一个收藏计数
	orderBy := "last_read desc,id desc"
	if len(order) > 0 && order[0] == "new" {
		orderBy = "id desc"
	}
	if cnt = count.Cnt; cnt > 0 {
		sql := `select b.*,m.nickname from ` + mysql.Book{}.TableName() + ` b left join ` + mysql.Star{}.TableName() + ` s on s.bid=b.book_id left join  ` + mysql.Member{}.TableName() + ` m on m.member_id=b.member_id where s.uid=? and b.privately_owned=0 order by %v limit %v offset %v`
		sql = fmt.Sprintf(sql, orderBy, listRows, (p-1)*listRows)
		err = mus.Db.Raw(sql, uid).Scan(&books).Error
	}
	return
}

func (this *star) SetLastReadTime(uid, bid int) {
	mus.Db.Model(mysql.Star{}).Where("uid = ? and bid = ?", uid, bid).Updates(mysql.Ups{
		"last_read": time.Now().Unix(),
	})
}
