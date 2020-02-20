MUSES_SYSTEM:=github.com/i2eco/muses/pkg/system
APPNAME:=ecology
APPPATH:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
APPOUT:=$(APPPATH)/appgo/$(APPNAME)

# 执行wechat
wechat:
	@cd $(APPPATH)/appuni && npm run dev:mp-weixin


ant:
	@cd $(APPPATH)/adminant && npm start

# 执行go指令
go:
	@cd $(APPPATH)/appgo && $(APPPATH)/tool/build.sh $(APPNAME) $(APPOUT) $(MUSES_SYSTEM)
	@cd $(APPPATH)/appgo && $(APPOUT) start --conf=conf/conf.toml


install:
	@cd $(APPPATH)/appgo && $(APPPATH)/tool/build.sh $(APPNAME) $(APPOUT) $(MUSES_SYSTEM)
	@cd $(APPPATH)/appgo && $(APPOUT) install --conf=conf/conf.toml

install.create:
	@cd $(APPPATH)/appgo && $(APPPATH)/tool/build.sh $(APPNAME) $(APPOUT) $(MUSES_SYSTEM)
	@cd $(APPPATH)/appgo && $(APPOUT) install --conf=conf/conf.toml --mode=create

install.clear:
	@cd $(APPPATH)/appgo && $(APPPATH)/tool/build.sh $(APPNAME) $(APPOUT) $(MUSES_SYSTEM)
	@cd $(APPPATH)/appgo && $(APPOUT) install --conf=conf/conf.toml --clear=true

all:
	@cd $(APPPATH)/appgo && $(APPPATH)/tool/build.sh $(APPNAME) $(APPOUT) $(MUSES_SYSTEM)
	@tar zxvf build.tar.gz build

generator:
	@cd $(GOPATH)/src/github.com/i2eco/generator && make build
	@cd $(APPPATH)/tool/gencode && go build && $(APPPATH)/tool/gencode/gencode -g $(GOPATH)/src/github.com/i2eco/generator -m "root:root@tcp(localhost:3306)" --model "github.com/i2eco/ecology/appgo/model" --dao "github.com/i2eco/ecology/appgo/dao" --outdao "$(APPPATH)/appgo/dao" --app "$(APPPATH)/appgo" --module "github.com/i2eco/ecology/appgo"
