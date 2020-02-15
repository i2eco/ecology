package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/model/trans"
	"github.com/goecology/muses/pkg/logger"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type bookmark struct {
	logger *logger.Client
	db     *gorm.DB
}

func InitBookmark(logger *logger.Client, db *gorm.DB) *bookmark {
	return &bookmark{
		logger: logger,
		db:     db,
	}
}

// Create 新增一条记
func (g *bookmark) Create(c *gin.Context, db *gorm.DB, data *mysql.Bookmark) (err error) {

	if err = db.Create(data).Error; err != nil {
		g.logger.Error("create bookmark create error", zap.Error(err))
		return
	}
	return nil
}

// Update 根据主键更新一条记录
func (g *bookmark) Update(c *gin.Context, db *gorm.DB, paramId int, ups mysql.Ups) (err error) {
	var sql = "`id`=?"
	var binds = []interface{}{paramId}

	if err = db.Table("bookmark").Where(sql, binds...).Updates(ups).Error; err != nil {
		g.logger.Error("bookmark update error", zap.Error(err))
		return
	}
	return
}

// UpdateX Update的扩展方法，根据Cond更新一条或多条记录
func (g *bookmark) UpdateX(c *gin.Context, db *gorm.DB, conds mysql.Conds, ups mysql.Ups) (err error) {

	sql, binds := mysql.BuildQuery(conds)
	if err = db.Table("bookmark").Where(sql, binds...).Updates(ups).Error; err != nil {
		g.logger.Error("bookmark update error", zap.Error(err))
		return
	}
	return
}

// Delete 根据主键删除一条记录。如果有delete_time则软删除，否则硬删除。
func (g *bookmark) Delete(c *gin.Context, db *gorm.DB, paramId int) (err error) {
	var sql = "`id`=?"
	var binds = []interface{}{paramId}

	if err = db.Table("bookmark").Where(sql, binds...).Delete(&mysql.Bookmark{}).Error; err != nil {
		g.logger.Error("bookmark delete error", zap.Error(err))
		return
	}

	return
}

// DeleteX Delete的扩展方法，根据Cond删除一条或多条记录。如果有delete_time则软删除，否则硬删除。
func (g *bookmark) DeleteX(c *gin.Context, db *gorm.DB, conds mysql.Conds) (err error) {
	sql, binds := mysql.BuildQuery(conds)

	if err = db.Table("bookmark").Where(sql, binds...).Delete(&mysql.Bookmark{}).Error; err != nil {
		g.logger.Error("bookmark delete error", zap.Error(err))
		return
	}

	return
}

// Info 根据PRI查询单条记录
func (g *bookmark) Info(c *gin.Context, paramId int) (resp mysql.Bookmark, err error) {
	var sql = "`id`=?"
	var binds = []interface{}{paramId}

	if err = g.db.Table("bookmark").Where(sql, binds...).First(&resp).Error; err != nil {
		g.logger.Error("bookmark info error", zap.Error(err))
		return
	}
	return
}

// InfoX Info的扩展方法，根据Cond查询单条记录
func (g *bookmark) InfoX(c *gin.Context, conds mysql.Conds) (resp mysql.Bookmark, err error) {
	sql, binds := mysql.BuildQuery(conds)

	if err = g.db.Table("bookmark").Where(sql, binds...).First(&resp).Error; err != nil {
		g.logger.Error("bookmark info error", zap.Error(err))
		return
	}
	return
}

// List 查询list，extra[0]为sorts
func (g *bookmark) List(c *gin.Context, conds mysql.Conds, extra ...string) (resp []mysql.Bookmark, err error) {
	sql, binds := mysql.BuildQuery(conds)

	sorts := ""
	if len(extra) >= 1 {
		sorts = extra[0]
	}
	if err = g.db.Table("bookmark").Where(sql, binds...).Order(sorts).Find(&resp).Error; err != nil {
		g.logger.Error("bookmark info error", zap.Error(err))
		return
	}
	return
}

// ListMap 查询map，map遍历的时候是无序的，所以指定sorts参数没有意义
func (g *bookmark) ListMap(c *gin.Context, conds mysql.Conds) (resp map[int]mysql.Bookmark, err error) {
	sql, binds := mysql.BuildQuery(conds)

	mysqlSlice := make([]mysql.Bookmark, 0)
	resp = make(map[int]mysql.Bookmark, 0)
	if err = g.db.Table("bookmark").Where(sql, binds...).Find(&mysqlSlice).Error; err != nil {
		g.logger.Error("bookmark info error", zap.Error(err))
		return
	}
	for _, value := range mysqlSlice {
		resp[value.Id] = value
	}
	return
}

// ListPage 根据分页条件查询list
func (g *bookmark) ListPage(c *gin.Context, conds mysql.Conds, reqList *trans.ReqPage) (total int, respList []mysql.Bookmark) {
	if reqList.PageSize == 0 {
		reqList.PageSize = 10
	}
	if reqList.Current == 0 {
		reqList.Current = 1
	}
	sql, binds := mysql.BuildQuery(conds)

	db := g.db.Table("bookmark").Where(sql, binds...)
	respList = make([]mysql.Bookmark, 0)
	db.Count(&total)
	db.Order(reqList.Sort).Offset((reqList.Current - 1) * reqList.PageSize).Limit(reqList.PageSize).Find(&respList)
	return
}
