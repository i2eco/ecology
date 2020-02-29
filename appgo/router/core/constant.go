package core

import (
	"encoding/gob"

	"github.com/i2eco/ecology/appgo/model/mysql"
)

const AdminSessionKey = "ecology/session/admin"
const AdminContextKey = "ecology/context/admin"
const FrontSessionKey = "ecology/session/front"
const FrontContextKey = "ecology/context/front"
const TPL = "MDW_TPL"
const Options = "MDW_OPTIONS"

func init() {
	gob.Register(&mysql.Member{})
	gob.Register(&mysql.User{})
}

type RespList struct {
	List       interface{} `json:"list"`
	Pagination struct {
		Current  int `json:"current"`
		PageSize int `json:"pageSize"`
		Total    int `json:"total"`
	} `json:"pagination"`
}

type WechatRespList struct {
	List  interface{} `json:"list"`
	Total int         `json:"total"`
}
