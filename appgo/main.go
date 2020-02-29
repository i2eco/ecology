package main

import (
	"github.com/i2eco/ecology/appgo/command"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/pkg/register"
	"github.com/i2eco/ecology/appgo/router"
	"github.com/i2eco/ecology/appgo/service"
	"github.com/i2eco/muses"
	"github.com/i2eco/muses/pkg/cache/mixcache"
	"github.com/i2eco/muses/pkg/cmd"
	"github.com/i2eco/muses/pkg/database/mysql"
	"github.com/i2eco/muses/pkg/oss"
	musgin "github.com/i2eco/muses/pkg/server/gin"
	"github.com/i2eco/muses/pkg/server/stat"
	"github.com/i2eco/muses/pkg/session/ginsession"
	"github.com/i2eco/muses/pkg/tpl/tplbeego"
	"github.com/spf13/cobra"
)

func main() {
	app := muses.Container(
		cmd.Register,
		stat.Register,
		mixcache.Register,
		mysql.Register,
		musgin.Register,
		tplbeego.Register,
		oss.Register,
		ginsession.Register,
	)
	app.SetRootCommand(func(cobraCommand *cobra.Command) {
		cobraCommand.AddCommand(command.InstallCmd)
		cobraCommand.AddCommand(command.AwesomeCmd)
	})
	app.SetStartCommand(func(cobraCommand *cobra.Command) {
		cobraCommand.PersistentFlags().StringVar(&command.Mode, "mode", "all", "设置启动模式")
	})
	app.SetGinRouter(router.InitRouter)
	app.SetPreRun(register.Init, register.InitDictionary)
	app.SetPostRun(conf.Init, register.Init, mus.Init, service.Init)
	err := app.Run()
	if err != nil {
		panic(err)
	}
}
