package dao

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/model/trans"
	"github.com/goecology/muses/pkg/logger"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type member struct {
	logger *logger.Client
	db     *gorm.DB
}

func InitMember(logger *logger.Client, db *gorm.DB) *member {
	return &member{
		logger: logger,
		db:     db,
	}
}

// Create 新增一条记
func (g *member) Create(c *gin.Context, db *gorm.DB, data *mysql.Member) (err error) {
	data.CreateTime = time.Now()
	if err = db.Create(data).Error; err != nil {
		g.logger.Error("create member create error", zap.Error(err))
		return
	}
	return nil
}

// UpdateX Update的扩展方法，根据Cond更新一条或多条记录
func (g *member) UpdateX(c *gin.Context, db *gorm.DB, conds mysql.Conds, ups mysql.Ups) (err error) {

	sql, binds := mysql.BuildQuery(conds)
	if err = db.Table("member").Where(sql, binds...).Updates(ups).Error; err != nil {
		g.logger.Error("member update error", zap.Error(err))
		return
	}
	return
}

// DeleteX Delete的扩展方法，根据Cond删除一条或多条记录。如果有delete_time则软删除，否则硬删除。
func (g *member) DeleteX(c *gin.Context, db *gorm.DB, conds mysql.Conds) (err error) {
	sql, binds := mysql.BuildQuery(conds)

	if err = db.Table("member").Where(sql, binds...).Delete(&mysql.Member{}).Error; err != nil {
		g.logger.Error("member delete error", zap.Error(err))
		return
	}

	return
}

// InfoX Info的扩展方法，根据Cond查询单条记录
func (g *member) InfoX(c *gin.Context, conds mysql.Conds) (resp mysql.Member, err error) {
	sql, binds := mysql.BuildQuery(conds)

	if err = g.db.Table("member").Where(sql, binds...).First(&resp).Error; err != nil {
		g.logger.Error("member info error", zap.Error(err))
		return
	}
	return
}

// List 查询list，extra[0]为sorts
func (g *member) List(c *gin.Context, conds mysql.Conds, extra ...string) (resp []mysql.Member, err error) {
	sql, binds := mysql.BuildQuery(conds)

	sorts := ""
	if len(extra) >= 1 {
		sorts = extra[0]
	}
	if err = g.db.Table("member").Where(sql, binds...).Order(sorts).Find(&resp).Error; err != nil {
		g.logger.Error("member info error", zap.Error(err))
		return
	}
	return
}

// ListPage 根据分页条件查询list
func (g *member) ListPage(c *gin.Context, conds mysql.Conds, reqList *trans.ReqPage) (total int, respList []mysql.Member) {
	if reqList.PageSize == 0 {
		reqList.PageSize = 10
	}
	if reqList.Current == 0 {
		reqList.Current = 1
	}
	sql, binds := mysql.BuildQuery(conds)

	db := g.db.Table("member").Where(sql, binds...)
	respList = make([]mysql.Member, 0)
	db.Count(&total)
	db.Order(reqList.Sort).Offset((reqList.Current - 1) * reqList.PageSize).Limit(reqList.PageSize).Find(&respList)
	return
}
