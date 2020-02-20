package command

import (
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/muses"
	mmysql "github.com/i2eco/muses/pkg/database/mysql"
	"github.com/i2eco/muses/pkg/oss"
	"github.com/spf13/cobra"
)

var AwesomeCmd = &cobra.Command{
	Use:  "awesome",
	Long: `Show awesome information`,
	Run:  awesomeCmd,
}

type GithubItem struct {
	Name string `csv:"name"`
	Desc string `csv:"desc"`
}

type CateItem struct {
	Name  int    `csv:"name"`
	Name2 string `csv:"name2"`
}

var AwesomeConfigPath string
var AwesomeCsvPath string
var AwesomeMode string

func init() {
	AwesomeCmd.PersistentFlags().StringVar(&AwesomeConfigPath, "conf", "conf/conf.toml", "conf path")
	AwesomeCmd.PersistentFlags().StringVar(&AwesomeCsvPath, "path", "", "csv path")
	AwesomeCmd.PersistentFlags().StringVar(&AwesomeMode, "mode", "github", "mode")
}

func awesomeCmd(cmd *cobra.Command, args []string) {
	app := muses.Container(
		mmysql.Register,
		oss.Register,
	)
	app.SetCfg(AwesomeConfigPath)
	app.SetPostRun(func() error {
		if AwesomeMode == "github" {
			f, err := os.OpenFile(AwesomeCsvPath, os.O_RDWR|os.O_CREATE, os.ModePerm) // 此处假设当前目录下已存在test目录
			if err != nil {
				return err
			}

			defer f.Close()
			var out []*GithubItem
			if err = gocsv.UnmarshalFile(f, &out); err != nil {
				return err
			}
			for _, value := range out {
				if value.Name == "" {
					continue
				}
				db := mmysql.Caller("ecology")
				var info mysql.Awesome
				db.Where("name = ?", value.Name).Find(&info)
				if info.Id > 0 {
					db.Model(mysql.Awesome{}).Where("id=?", info.Id).Updates(mysql.Ups{"desc": value.Desc})
					continue
				}

				db.Create(&mysql.Awesome{
					Name:           value.Name,
					GitName:        "",
					OwnerAvatarUrl: "",
					HtmlUrl:        "",
					GitDescription: "",
					GitCreatedAt:   time.Now(),
					GitUpdatedAt:   time.Now(),
					GitUrl:         "",
					SshUrl:         "",
					CloneUrl:       "",
					HomePage:       "",
					StarCount:      0,
					WatcherCount:   0,
					Language:       "",
					ForkCount:      0,
					LicenseUrl:     "",
					Desc:           value.Desc,
					LongDesc:       "",
					Version:        0,
				})
			}
		}
		if AwesomeMode == "cate" {

		}
		return nil
	})
	err := app.Run()
	if err != nil {
		panic(err)
	}
}
