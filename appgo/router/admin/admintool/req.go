package admintool

import "github.com/i2eco/ecology/appgo/model/trans"

type ReqList struct {
	Name string `form:"name"`
	trans.ReqPage
}

type ReqInfo struct {
	Id int `form:"id"`
}

type ReqCreate struct {
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Identify string `json:"identify"`
	Cover    string `json:"cover"`
}

type ReqUpdate struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Identify string `json:"identify"`
	Cover    string `json:"cover"`
}
