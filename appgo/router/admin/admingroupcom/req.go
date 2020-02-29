package admingroupcom

import "github.com/i2eco/ecology/appgo/model/trans"

type ReqList struct {
	Name string `form:"name"`
	trans.ReqPage
}
