package dao

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/spf13/viper"
)

//添加评论
func (this *comments) AddComments(uid, book_id int, content string) (err error) {
	//查询该用户现有的评论
	second := viper.GetInt("app.commentInterval")
	var comment mysql.Comments
	err = mus.Db.Where("uid = ? and time_create > ?", uid, time.Now().Add(-time.Duration(second)*time.Second)).Order("id desc").Find(&comment).Error
	if err != nil {
		return
	}

	if comment.Id > 0 {
		err = errors.New(fmt.Sprintf("您距离上次发表评论时间小于 %v 秒，请歇会儿再发。", second))
		return
	}
	//项目被评论是量+1
	var comments = mysql.Comments{
		Uid:        uid,
		BookId:     book_id,
		Content:    content,
		TimeCreate: time.Now(),
	}
	err = mus.Db.Create(&comments).Error
	if err != nil {
		err = errors.New("发表评论失败")
		return
	}
	err = SetIncreAndDecre(mysql.Book{}.TableName(), "cnt_comment", fmt.Sprintf("book_id=%v", book_id), true)
	return
}

// 获取可显示的评论内容
func (this *comments) Comments(p, listRows, bookId int, status ...int) (comments []mysql.BookCommentsResult, err error) {
	sql := `select c.id,c.content,s.score,c.uid,c.status,c.time_create,m.avatar,m.nickname,m.account,b.book_id,b.book_name,b.identify from ` + mysql.Comments{}.TableName() + ` c left join ` + mysql.Member{}.TableName() + ` m on m.member_id=c.uid left join ` + mysql.Score{}.TableName() + ` s on s.uid=c.uid and s.book_id=c.book_id left join ` + mysql.Book{}.TableName() + ` b on b.book_id = c.book_id %v order by c.id desc limit %v offset %v`
	whereStr := ""
	whereSlice := []string{"true"}
	if bookId > 0 {
		whereSlice = append(whereSlice, "c.book_id = "+strconv.Itoa(bookId))
	}
	if len(status) > 0 {
		whereSlice = append(whereSlice, "c.status = "+strconv.Itoa(status[0]))
	}

	if len(whereSlice) > 0 {
		whereStr = " where " + strings.Join(whereSlice, " and ")
	}

	sql = fmt.Sprintf(sql, whereStr, listRows, (p-1)*listRows)
	err = mus.Db.Raw(sql).Scan(&comments).Error
	return
}

func (this *comments) Count(bookId int, status ...int) (cnt int64, err error) {
	qs := mus.Db.Model(mysql.Comments{})

	if bookId > 0 {
		qs = qs.Where("book_id = ?", bookId)
	}
	if len(status) > 0 {
		qs = qs.Where("status = ?", status[0])
	}
	err = qs.Count(&cnt).Error
	return
}

// 清空评论
func (this *comments) ClearComments(uid int) {
	var (
		comments  []mysql.CommentCount
		cid       []interface{}
		bookId    []interface{}
		bookIdMap = make(map[int]int)
	)
	sql := "select count(id) cnt,id,book_id from md_comments where uid = ? group by book_id"
	mus.Db.Raw(sql, uid).Scan(&comments)
	if len(comments) > 0 {
		for _, comment := range comments {
			cid = append(cid, comment.Id)
			bookId = append(bookId, comment.BookId)
			bookIdMap[comment.BookId] = comment.Cnt
		}
		var err error
		mus.Db.Begin()
		defer func() {
			if err != nil {
				mus.Db.Rollback()
			} else {
				mus.Db.Commit()
			}
		}()
		err = mus.Db.Where("book_id in (?) ", bookId).Delete(&mysql.Comments{}).Error
		if err != nil {
			return
		}
		sqlUpdate := "update md_books set cnt_comment = cnt_comment - ? where cnt_comment> ? and book_id = ?"
		for bid, cnt := range bookIdMap {
			if err = mus.Db.Exec(sqlUpdate, cnt, cnt-1, bid).Error; err != nil {
				return
			}
		}
	}
}

// 删除评论
func (this *comments) DeleteComment(id int) {
	m := &mysql.Comments{}
	mus.Db.Where("id = ?", id).Find(m)
	if m.BookId > 0 {
		mus.Db.Where("id = ?", id).Delete(m)
		sql := "update md_books set cnt_comment = cnt_comment - 1 where cnt_comment > 0 and book_id = ?"
		err := mus.Db.Exec(sql, m.BookId).Error
		if err != nil {
			mus.Logger.Error(err.Error())
		}
	}
}

func (this *comments) SetCommentStatus(id, status int) (err error) {
	err = mus.Db.Model(mysql.Comments{}).Where("id = ?", id).Updates(mysql.Ups{"status": status}).Error
	return
}
