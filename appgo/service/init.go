package service

import "github.com/goecology/ecology/appgo/dao"

func Init() error {
	dao.Global = dao.NewGlobal()
	dao.DocumentSearchResult = dao.NewDocumentSearchResult()
	InitGen()
	dao.ReadRecord.UpdateReadingRule()
	dao.GithubApi = dao.NewGithubApi()
	return nil
}
