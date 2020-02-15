package service

import (
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/pkg/mus"
)

func InitGen() {
	dao.ReadCount = dao.InitReadCount(mus.Logger, mus.Db)
	dao.Banner = dao.InitBanner(mus.Logger, mus.Db)
	dao.DocumentHistory = dao.InitDocumentHistory(mus.Logger, mus.Db)
	dao.Star = dao.InitStar(mus.Logger, mus.Db)
	dao.Score = dao.InitScore(mus.Logger, mus.Db)
	dao.Member = dao.InitMember(mus.Logger, mus.Db)
	dao.Wechat = dao.InitWechat(mus.Logger, mus.Db)
	dao.ReadRecord = dao.InitReadRecord(mus.Logger, mus.Db)
	dao.GithubUser = dao.InitGithubUser(mus.Logger, mus.Db)
	dao.Github = dao.InitGithub(mus.Logger, mus.Db)
	dao.BookCounter = dao.InitBookCounter(mus.Logger, mus.Db)
	dao.Category = dao.InitCategory(mus.Logger, mus.Db)
	dao.Attachment = dao.InitAttachment(mus.Logger, mus.Db)
	dao.MemberToken = dao.InitMemberToken(mus.Logger, mus.Db)
	dao.Document = dao.InitDocument(mus.Logger, mus.Db)
	dao.Relationship = dao.InitRelationship(mus.Logger, mus.Db)
	dao.BookCategory = dao.InitBookCategory(mus.Logger, mus.Db)
	dao.Comments = dao.InitComments(mus.Logger, mus.Db)
	dao.Logs = dao.InitLogs(mus.Logger, mus.Db)
	dao.Bookmark = dao.InitBookmark(mus.Logger, mus.Db)
	dao.Label = dao.InitLabel(mus.Logger, mus.Db)
	dao.Book = dao.InitBook(mus.Logger, mus.Db)
	dao.DocumentStore = dao.InitDocumentStore(mus.Logger, mus.Db)
	dao.Option = dao.InitOption(mus.Logger, mus.Db)

}
