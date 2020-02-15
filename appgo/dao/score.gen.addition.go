package dao

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/jinzhu/gorm"
)

//添加评分
//score的值只能是1-5，然后需要对scorex10，50则表示5.0分
func (this *score) AddScore(c *gin.Context, uid, bookId, score int) (err error) {
	//查询评分是否已存在
	var one mysql.Score
	err = mus.Db.Where("uid = ? and book_id = ?", uid, bookId).Find(&one).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	// 创建评分
	if err != gorm.ErrRecordNotFound {
		err = errors.New("您已给当前文档打过分了")
		return
	}

	transdb := mus.Db.Begin()
	score = score * 10
	createScore := mysql.Score{
		BookId:     bookId,
		Uid:        uid,
		Score:      score,
		TimeCreate: time.Now(),
	}
	err = transdb.Create(&createScore).Error
	if err != nil {
		transdb.Rollback()
		return
	}
	var oneBook mysql.Book
	mus.Db.Where("book_id = ?", bookId).Find(&oneBook)
	// todo 操作不原子
	if oneBook.CntScore == 0 {
		oneBook.CntScore = 1
		oneBook.Score = 0
	} else {
		oneBook.CntScore = oneBook.CntScore + 1
	}
	oneBook.Score = (oneBook.Score*(oneBook.CntScore-1) + score) / oneBook.CntScore
	err = Book.Update(c, transdb, bookId, mysql.Ups{
		"cnt_score": oneBook.CntScore,
		"score":     oneBook.Score,
	})
	if err != nil {
		transdb.Rollback()
		return
	}
	transdb.Commit()

	return
}

//查询用户对文档的评分
func (this *score) BookScoreByUid(uid, bookId interface{}) int {
	var score mysql.Score
	mus.Db.Select("score").Where("uid = ? and book_id = ?", uid, bookId).Find(&score)
	return score.Score
}
