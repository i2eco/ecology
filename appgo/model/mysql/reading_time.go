package mysql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/goecology/ecology/appgo/pkg/mus"
	"go.uber.org/zap"
)

// 阅读时长
type ReadingTime struct {
	Id       int `gorm:"not null;"json:"id"`
	Uid      int `gorm:"not null;"json:"uid"`
	Day      int `gorm:"not null;"json:"day"`      // 日期，如 20191212
	Duration int `gorm:"not null;"json:"duration"` // 每天的阅读时长
}

func (ReadingTime) TableName() string {
	return "reading_time"
}

type sum struct {
	SumVal int
}

type ReadingSortedUser struct {
	Uid      int
	Account  string
	Nickname string
	Avatar   string
	SumTime  int
}

const (
	readingTimeCacheDir = "cache/rank/reading-time"
	readingTimeCacheFmt = "cache/rank/reading-time/%v-%v.json"
)

func init() {
	if _, err := os.Stat(readingTimeCacheDir); err != nil {
		err = os.MkdirAll(readingTimeCacheDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func NewReadingTime() *ReadingTime {
	return &ReadingTime{}
}

func (*ReadingTime) TableUnique() [][]string {
	return [][]string{[]string{"uid", "day"}}
}

func (r *ReadingTime) GetReadingTime(uid int, prd period) int {
	sum := &sum{}
	sqlSum := "select sum(duration) sum_val from " + ReadingTime{}.TableName() + " where uid = ? and day>=? and day<=? limit 1"
	now := time.Now()
	if prd == PeriodAll {
		var member Member
		mus.Db.Where("member_id = ?", uid).Find(&member)
		return member.TotalReadingTime
	}
	start, end := getTimeRange(now, prd)

	mus.Db.Raw(sqlSum, uid, start, end).Scan(sum)
	return sum.SumVal
}

func (r *ReadingTime) Sort(prd period, limit int, withCache ...bool) (users []ReadingSortedUser) {
	var b []byte
	cache := false
	if len(withCache) > 0 {
		cache = withCache[0]
	}
	file := fmt.Sprintf(readingTimeCacheFmt, prd, limit)
	if cache {
		if info, err := os.Stat(file); err == nil && time.Now().Sub(info.ModTime()).Seconds() <= cacheTime {
			// 文件存在，且在缓存时间内
			if b, err = ioutil.ReadFile(file); err == nil {
				err = json.Unmarshal(b, &users)
				if err != nil {
					mus.Logger.Error("read time sort error", zap.Error(err))
					return
				}
				if len(users) > 0 {
					return
				}
			}
		}
	}

	sqlSort := "SELECT t.uid,sum(t.duration) sum_time,m.account,m.avatar,m.nickname FROM `" + ReadingTime{}.TableName() + "` t left JOIN " + Member{}.TableName() + " m on t.uid=m.member_id WHERE t.day>=? and t.day<=? GROUP BY t.uid ORDER BY sum_time desc limit ?"
	start, end := getTimeRange(time.Now(), prd)
	mus.Db.Raw(sqlSort, start, end, limit).Scan(&users)

	if cache && len(users) > 0 {
		b, _ = json.Marshal(users)
		ioutil.WriteFile(file, b, os.ModePerm)
	}
	return
}
