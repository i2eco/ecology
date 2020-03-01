package adminawesome

import (
	"github.com/i2eco/ecology/appgo/model/trans"
)

type ReqList struct {
	Name string `form:"name"`
	trans.ReqPage
}

type ReqInfo struct {
	Id int `form:"id"`
}

type ReqCreate struct {
	Name string `json:"name"` // 唯一标识
	Desc string `json:"desc"`
}

type ReqUpdate struct {
	Id   int    `json:"id"`
	Name string `json:"name"` // 唯一标识
	Desc string `json:"desc"`
}

type ReqDelete struct {
	Id int `form:"id"`
}
