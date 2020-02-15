package mysql

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
)

//评论表
type Comments struct {
	Id         int       `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"`
	Uid        int       `gorm:"not null;index"json:"uid"`   //用户id
	BookId     int       `gorm:"not null;"json:"bookId"`     //文档项目id
	Content    string    `gorm:"not null;"json:"content"`    //评论内容
	TimeCreate time.Time `gorm:"not null;"json:"timeCreate"` //评论时间
	Status     int8      `gorm:"not null;"json:"status"`     //  审核状态; 0，待审核，1 通过，-1 不通过
}

func (Comments) TableName() string {
	return "comments"
}

//评分表
type Score struct {
	Id         int       `gorm:"not null;"json:"id"`
	BookId     int       `gorm:"not null;"json:"bookId"`
	Uid        int       `gorm:"not null;"json:"uid"`
	Score      int       `gorm:"not null;"json:"score"` //评分
	TimeCreate time.Time `gorm:"not null;"json:"timeCreate"`
}

func (Score) TableName() string {
	return "score"
}

// 多字段唯一键
func (this *Score) TableUnique() [][]string {
	return [][]string{
		[]string{"Uid", "BookId"},
	}
}

//评论内容
type BookCommentsResult struct {
	Id         int       `json:"id"`
	Uid        int       `json:"uid"`
	Score      int       `json:"score"`
	Avatar     string    `json:"avatar"`
	Account    string    `json:"account"`
	Nickname   string    `json:"nickname"`
	BookId     int       `json:"book_id"`
	BookName   string    `json:"book_name"`
	Identify   string    `json:"identify"`
	Content    string    `json:"content"`
	Status     int8      `json:"status"`
	TimeCreate time.Time `json:"created_at"` //评论时间
}

func NewComments() *Comments {
	return &Comments{}
}

func (this *Comments) SetCommentStatus(id, status int) (err error) {
	_, err = orm.NewOrm().QueryTable(this).Filter("id", id).Update(orm.Params{"status": status})
	return
}

type CommentCount struct {
	Id     int
	BookId int
	Cnt    int
}

func (this *Comments) Count(bookId int, status ...int) (int64, error) {
	query := orm.NewOrm().QueryTable(this)
	if bookId > 0 {
		query = query.Filter("book_id", bookId)
	}
	if len(status) > 0 {
		query = query.Filter("status", status[0])
	}
	return query.Count()
}

//评分内容
type BookScoresResult struct {
	Avatar     string    `json:"avatar"`
	Nickname   string    `json:"nickname"`
	Score      string    `json:"score"`
	TimeCreate time.Time `json:"time_create"` //评论时间
}

//获取评分内容
func (this *Score) BookScores(p, listRows, bookId int) (scores []BookScoresResult, err error) {
	sql := `select s.score,s.time_create,m.avatar,m.nickname from md_score s left join md_members m on m.member_id=s.uid where s.book_id=? order by s.id desc limit %v offset %v`
	sql = fmt.Sprintf(sql, listRows, (p-1)*listRows)
	_, err = orm.NewOrm().Raw(sql, bookId).QueryRows(&scores)
	return
}
