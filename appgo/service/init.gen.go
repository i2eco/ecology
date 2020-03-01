package service

import (
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/pkg/mus"
)

func InitGen() {
	dao.MemberToken = dao.InitMemberToken(mus.Logger, mus.Db)
	dao.Document = dao.InitDocument(mus.Logger, mus.Db)
	dao.ReadRecord = dao.InitReadRecord(mus.Logger, mus.Db)
	dao.FriendLink = dao.InitFriendLink(mus.Logger, mus.Db)
	dao.Sign = dao.InitSign(mus.Logger, mus.Db)
	dao.Bookmark = dao.InitBookmark(mus.Logger, mus.Db)
	dao.Score = dao.InitScore(mus.Logger, mus.Db)
	dao.DocumentHistory = dao.InitDocumentHistory(mus.Logger, mus.Db)
	dao.Awesome = dao.InitAwesome(mus.Logger, mus.Db)
	dao.ReadCount = dao.InitReadCount(mus.Logger, mus.Db)
	dao.Banner = dao.InitBanner(mus.Logger, mus.Db)
	dao.Logs = dao.InitLogs(mus.Logger, mus.Db)
	dao.Label = dao.InitLabel(mus.Logger, mus.Db)
	dao.GithubUser = dao.InitGithubUser(mus.Logger, mus.Db)
	dao.Member = dao.InitMember(mus.Logger, mus.Db)
	dao.Relationship = dao.InitRelationship(mus.Logger, mus.Db)
	dao.Category = dao.InitCategory(mus.Logger, mus.Db)
	dao.User = dao.InitUser(mus.Logger, mus.Db)
	dao.AwesomeCate = dao.InitAwesomeCate(mus.Logger, mus.Db)
	dao.Star = dao.InitStar(mus.Logger, mus.Db)
	dao.AdsCont = dao.InitAdsCont(mus.Logger, mus.Db)
	dao.Book = dao.InitBook(mus.Logger, mus.Db)
	dao.Option = dao.InitOption(mus.Logger, mus.Db)
	dao.Tool = dao.InitTool(mus.Logger, mus.Db)
	dao.Comments = dao.InitComments(mus.Logger, mus.Db)
	dao.ReadingTime = dao.InitReadingTime(mus.Logger, mus.Db)
	dao.DocumentStore = dao.InitDocumentStore(mus.Logger, mus.Db)
	dao.Github = dao.InitGithub(mus.Logger, mus.Db)
	dao.BookCounter = dao.InitBookCounter(mus.Logger, mus.Db)
	dao.Seo = dao.InitSeo(mus.Logger, mus.Db)
	dao.Fans = dao.InitFans(mus.Logger, mus.Db)
	dao.Attachment = dao.InitAttachment(mus.Logger, mus.Db)
	dao.Wechat = dao.InitWechat(mus.Logger, mus.Db)
	dao.BookCategory = dao.InitBookCategory(mus.Logger, mus.Db)

}
