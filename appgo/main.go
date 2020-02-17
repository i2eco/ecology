package main

import (
	"github.com/goecology/ecology/appgo/command"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/register"
	"github.com/goecology/ecology/appgo/router"
	"github.com/goecology/ecology/appgo/service"
	"github.com/goecology/muses"
	"github.com/goecology/muses/pkg/cache/mixcache"
	"github.com/goecology/muses/pkg/cmd"
	"github.com/goecology/muses/pkg/database/mysql"
	"github.com/goecology/muses/pkg/oss"
	musgin "github.com/goecology/muses/pkg/server/gin"
	"github.com/goecology/muses/pkg/server/stat"
	"github.com/goecology/muses/pkg/session/ginsession"
	"github.com/goecology/muses/pkg/tpl/tplbeego"
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
	app.SetPreRun(register.Init)
	app.SetPostRun(conf.Init, register.Init, mus.Init, service.Init)
	err := app.Run()
	if err != nil {
		panic(err)
	}
}
