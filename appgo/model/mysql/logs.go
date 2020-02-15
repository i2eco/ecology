package mysql

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goecology/ecology/appgo/pkg/conf"
)

var loggerQueue = &logQueue{channel: make(chan *Logs, 100), isRunning: 0}

type logQueue struct {
	channel   chan *Logs
	isRunning int32
}

// Logs struct .
type Logs struct {
	LoggerId int64 `gorm:"not null;primary_key;AUTO_INCREMENT"json:"loggerId"`
	MemberId int   `gorm:"not null;"json:"memberId"`
	// 日志类别：operate 操作日志/ system 系统日志/ exception 异常日志 / document 文档操作日志
	Category     string    `gorm:"not null;"json:"category"`
	Content      string    `gorm:"not null;"json:"content"`
	OriginalData string    `gorm:"not null;"json:"originalData"`
	PresentData  string    `gorm:"not null;"json:"presentData"`
	CreateTime   time.Time `gorm:""json:"createTime"`
	UserAgent    string    `gorm:"not null;"json:"userAgent"`
	IPAddress    string    `gorm:"not null;"json:"ipAddress"`
}

// TableName 获取对应数据库表名.
func (m *Logs) TableName() string {
	return "logs"
}

// TableEngine 获取数据使用的引擎.
func (m *Logs) TableEngine() string {
	return "INNODB"
}
func (m *Logs) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

func NewLogger() *Logs {
	return &Logs{}
}

func (m *Logs) Add() error {
	if m.MemberId <= 0 {
		return errors.New("用户ID不能为空")
	}
	if m.Category == "" {
		m.Category = "system"
	}
	if m.Content == "" {
		return errors.New("日志内容不能为空")
	}
	loggerQueue.channel <- m
	if atomic.LoadInt32(&(loggerQueue.isRunning)) <= 0 {
		atomic.AddInt32(&(loggerQueue.isRunning), 1)
		go addLoggerAsync()
	}
	return nil
}

func addLoggerAsync() {
	defer atomic.AddInt32(&(loggerQueue.isRunning), -1)
	o := orm.NewOrm()

	for {
		logger := <-loggerQueue.channel

		o.Insert(logger)
	}
}
