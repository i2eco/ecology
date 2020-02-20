package dao

import (
	"errors"
	"sync"

	"github.com/i2eco/ecology/appgo/model/constx"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/robfig/cron"
)

// Option struct .
type global struct {
	l          sync.RWMutex
	c          *cron.Cron
	data       map[string]string
	dataOption map[string]*mysql.Option
}

func NewGlobal() *global {
	obj := &global{
		c: cron.New(),
	}
	obj.Run()
	//obj.c.Schedule(cron.Every(5*time.Second), obj)
	//obj.c.Start()
	return obj
}

func (t *global) Run() {
	options, err := t.innerAll()
	// 获取失败，不更新
	if err != nil {
		return
	}
	tmpData := make(map[string]string, len(options))
	tmpDataOption := make(map[string]*mysql.Option, len(options))
	for _, item := range options {
		tmpData[item.OptionName] = item.OptionValue
		tmpDataOption[item.OptionName] = item
	}

	t.l.Lock()
	t.data = tmpData
	t.dataOption = tmpDataOption
	t.l.Unlock()
}

func (t *global) AllOptions() (resp map[string]string) {
	resp = make(map[string]string)

	t.l.RLock()
	for key, value := range t.data {
		resp[key] = value
	}
	t.l.RUnlock()
	return
}

func (t *global) All() (resp []*mysql.Option) {
	resp = make([]*mysql.Option, 0)

	t.l.RLock()
	for _, value := range t.dataOption {
		resp = append(resp, value)
	}
	t.l.RUnlock()
	return
}

func (t *global) FindByKey(key string) (*mysql.Option, error) {
	t.l.RLock()
	defer t.l.RUnlock()
	option, flag := t.dataOption[key]
	if !flag {
		return nil, errors.New("option empty")
	}
	return option, nil

}

func (p *global) innerAll() ([]*mysql.Option, error) {
	//o := orm.NewOrm()
	var options []*mysql.Option

	mus.Db.Find(&options)
	//_, err := o.QueryTable(bootstrap.Conf.App.DbPrefix + "options").All(&options)
	//if err != nil {
	//	return options, err
	//}
	return options, nil
}

func (p *global) GetOptionValue(key, def string) string {
	if option, err := p.FindByKey(key); err == nil {
		return option.OptionValue
	}
	return def
}

func (p *global) Get(key string) string {
	if option, err := p.FindByKey(key); err == nil {
		return option.OptionValue
	}
	return ""
}

func (g *global) GetSiteName() string {
	return g.Get(constx.SITE_NAME)
}

func (g *global) IsEnableAnonymous() bool {
	return g.Get("ENABLE_ANONYMOUS") == "true"
}

func (p *global) InsertOrUpdate() (err error) {

	//o := orm.NewOrm()
	//
	//var err error
	//
	//if p.OptionId > 0 || o.QueryTable(p.TableNameWithPrefix()).Filter("option_name", p.OptionName).Exist() {
	//	_, err = o.Update(p)
	//} else {
	//	_, err = o.Insert(p)
	//}
	return
}
