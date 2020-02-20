package awesome

import (
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/router/core"
)

func Info(c *core.Context) {
	q := c.Query("q")
	output, err := dao.GithubApi.Info(q)
	if err != nil {
		c.JSONOK(err.Error())
		return
	}
	c.JSONOK(output)
}

func Gen(c *core.Context) {
	q := c.Query("q")
	output, err := dao.GithubApi.Info(q)
	if err != nil {
		c.JSONOK(err.Error())
		return
	}
	err = dao.GithubApi.Update(output, 0)
	if err != nil {
		c.JSONOK(err.Error())
		return
	}

	c.JSONOK(output)
}

func All(c *core.Context) {
	err := dao.GithubApi.All()
	if err != nil {
		c.JSONOK(err.Error())
		return
	}

	c.JSONOK()
}
