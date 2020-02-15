package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/model/trans"
	"github.com/goecology/muses/pkg/logger"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type documentStore struct {
	logger *logger.Client
	db     *gorm.DB
}

func InitDocumentStore(logger *logger.Client, db *gorm.DB) *documentStore {
	return &documentStore{
		logger: logger,
		db:     db,
	}
}

// Create 新增一条记
func (g *documentStore) Create(c *gin.Context, db *gorm.DB, data *mysql.DocumentStore) (err error) {

	if err = db.Create(data).Error; err != nil {
		g.logger.Error("create documentStore create error", zap.Error(err))
		return
	}
	return nil
}

// Update 根据主键更新一条记录
func (g *documentStore) Update(c *gin.Context, db *gorm.DB, paramDocumentId int, ups mysql.Ups) (err error) {
	var sql = "`document_id`=?"
	var binds = []interface{}{paramDocumentId}

	if err = db.Table("document_store").Where(sql, binds...).Updates(ups).Error; err != nil {
		g.logger.Error("document_store update error", zap.Error(err))
		return
	}
	return
}

// UpdateX Update的扩展方法，根据Cond更新一条或多条记录
func (g *documentStore) UpdateX(c *gin.Context, db *gorm.DB, conds mysql.Conds, ups mysql.Ups) (err error) {

	sql, binds := mysql.BuildQuery(conds)
	if err = db.Table("document_store").Where(sql, binds...).Updates(ups).Error; err != nil {
		g.logger.Error("document_store update error", zap.Error(err))
		return
	}
	return
}

// Delete 根据主键删除一条记录。如果有delete_time则软删除，否则硬删除。
func (g *documentStore) Delete(c *gin.Context, db *gorm.DB, paramDocumentId int) (err error) {
	var sql = "`document_id`=?"
	var binds = []interface{}{paramDocumentId}

	if err = db.Table("document_store").Where(sql, binds...).Delete(&mysql.DocumentStore{}).Error; err != nil {
		g.logger.Error("document_store delete error", zap.Error(err))
		return
	}

	return
}

// DeleteX Delete的扩展方法，根据Cond删除一条或多条记录。如果有delete_time则软删除，否则硬删除。
func (g *documentStore) DeleteX(c *gin.Context, db *gorm.DB, conds mysql.Conds) (err error) {
	sql, binds := mysql.BuildQuery(conds)

	if err = db.Table("document_store").Where(sql, binds...).Delete(&mysql.DocumentStore{}).Error; err != nil {
		g.logger.Error("document_store delete error", zap.Error(err))
		return
	}

	return
}

// Info 根据PRI查询单条记录
func (g *documentStore) Info(c *gin.Context, paramDocumentId int) (resp mysql.DocumentStore, err error) {
	var sql = "`document_id`=?"
	var binds = []interface{}{paramDocumentId}

	if err = g.db.Table("document_store").Where(sql, binds...).First(&resp).Error; err != nil {
		g.logger.Error("document_store info error", zap.Error(err))
		return
	}
	return
}

// InfoX Info的扩展方法，根据Cond查询单条记录
func (g *documentStore) InfoX(c *gin.Context, conds mysql.Conds) (resp mysql.DocumentStore, err error) {
	sql, binds := mysql.BuildQuery(conds)

	if err = g.db.Table("document_store").Where(sql, binds...).First(&resp).Error; err != nil {
		g.logger.Error("document_store info error", zap.Error(err))
		return
	}
	return
}

// List 查询list，extra[0]为sorts
func (g *documentStore) List(c *gin.Context, conds mysql.Conds, extra ...string) (resp []mysql.DocumentStore, err error) {
	sql, binds := mysql.BuildQuery(conds)

	sorts := ""
	if len(extra) >= 1 {
		sorts = extra[0]
	}
	if err = g.db.Table("document_store").Where(sql, binds...).Order(sorts).Find(&resp).Error; err != nil {
		g.logger.Error("document_store info error", zap.Error(err))
		return
	}
	return
}

// ListMap 查询map，map遍历的时候是无序的，所以指定sorts参数没有意义
func (g *documentStore) ListMap(c *gin.Context, conds mysql.Conds) (resp map[int]mysql.DocumentStore, err error) {
	sql, binds := mysql.BuildQuery(conds)

	mysqlSlice := make([]mysql.DocumentStore, 0)
	resp = make(map[int]mysql.DocumentStore, 0)
	if err = g.db.Table("document_store").Where(sql, binds...).Find(&mysqlSlice).Error; err != nil {
		g.logger.Error("document_store info error", zap.Error(err))
		return
	}
	for _, value := range mysqlSlice {
		resp[value.DocumentId] = value
	}
	return
}

// ListPage 根据分页条件查询list
func (g *documentStore) ListPage(c *gin.Context, conds mysql.Conds, reqList *trans.ReqPage) (total int, respList []mysql.DocumentStore) {
	if reqList.PageSize == 0 {
		reqList.PageSize = 10
	}
	if reqList.Current == 0 {
		reqList.Current = 1
	}
	sql, binds := mysql.BuildQuery(conds)

	db := g.db.Table("document_store").Where(sql, binds...)
	respList = make([]mysql.DocumentStore, 0)
	db.Count(&total)
	db.Order(reqList.Sort).Offset((reqList.Current - 1) * reqList.PageSize).Limit(reqList.PageSize).Find(&respList)
	return
}
