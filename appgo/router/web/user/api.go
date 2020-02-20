package user

import (
	"strconv"

	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/code"
	"github.com/i2eco/ecology/appgo/router/core"
)

//关注或取消关注
func SetFollow(c *core.Context) {
	var cancel bool
	if c.Member() == nil || c.Member().MemberId == 0 {
		c.JSONCode(code.MsgErr)
		return
	}
	uid, _ := strconv.Atoi(c.Param("uid"))
	if uid == c.Member().MemberId {
		c.JSONCode(code.MsgErr)
		return
	}
	cancel, _ = new(mysql.Fans).FollowOrCancel(uid, c.Member().MemberId)
	if cancel {
		//this.JsonResult(0, "您已经成功取消了关注")
		c.JSONOK()
	}
	c.JSONOK()

	//this.JsonResult(0, "您已经成功关注了Ta")
}

func SignToday(c *core.Context) {
	if c.Member() == nil || c.Member().MemberId == 0 {
		c.JSONCode(code.MsgErr)
		return
	}
	reward, err := mysql.NewSign().Sign(c.Member().MemberId, false)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	c.JSONOK(reward)
	//this.JsonResult(0, fmt.Sprintf("恭喜您，签到成功,奖励阅读时长 %v 秒", reward))
}
