package dao

import (
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
)

//分页查询指定文档的历史.
func (m *documentHistory) FindToPager(docId, pageIndex, pageSize int) (docs []*mysql.DocumentHistorySimpleResult, totalCount int, err error) {
	offset := (pageIndex - 1) * pageSize

	totalCount = 0

	sql := `SELECT history.*,m1.account,m1.nickname,m2.account as modify_name
FROM ` + mysql.DocumentHistory{}.TableName() + ` AS history
LEFT JOIN ` + mysql.Member{}.TableName() + ` AS m1 ON history.member_id = m1.member_id
LEFT JOIN ` + mysql.Member{}.TableName() + ` AS m2 ON history.modify_at = m2.member_id
WHERE history.document_id = ? ORDER BY history.history_id DESC LIMIT ?,?;`

	err = mus.Db.Raw(sql, docId, offset, pageSize).Scan(&docs).Error

	if err != nil {
		return
	}
	err = mus.Db.Model(mysql.DocumentHistory{}).Where("document_id = ?", docId).Count(&totalCount).Error

	if err != nil {
		return
	}
	return
}

func (m *documentHistory) InsertOrUpdate(info *mysql.DocumentHistory) (err error) {
	if info.HistoryId > 0 {
		err = mus.Db.Model(mysql.DocumentHistory{}).Where("history_id = ?", info.HistoryId).UpdateColumns(info).Error
	} else {
		err = mus.Db.Create(info).Error
	}
	return
}

// 根据文档id删除
func (history *documentHistory) DeleteByLimit(docId, limit int) (err error) {
	if limit <= 0 {
		return
	}

	filter := mus.Db.Model(mysql.DocumentHistory{}).Where("document_id = ?", docId)
	var cnt int
	err = filter.Count(&cnt).Error
	if err != nil {
		return
	}

	if cnt > limit {

		var histories []mysql.DocumentHistory
		var historyIds []interface{}

		filter2 := filter.Order("history_id desc").Limit(cnt - limit)
		filter2.Find(&histories)

		for _, item := range histories {
			ver := NewVersionControl(item.DocumentId, item.Version)
			err = ver.DeleteVersion()
			if err != nil {
				// todo log
			}
			historyIds = append(historyIds, item.HistoryId)
		}
		err = filter.Where("history_id in (?)", historyIds).Delete(&mysql.DocumentHistory{}).Error
	}
	return
}
