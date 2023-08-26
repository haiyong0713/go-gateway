package initialize

import (
	"context"
	"flag"
	"fmt"
	"go-common/library/conf/flagvar"
	"go-common/library/conf/paladin"
	"go-common/library/log"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	options flagvar.StringVars
	reg, _  = regexp.Compile("\\([^\\)]+\\)\\.")
	ctx     = context.Background()
	config  = &struct {
		Base    string
		Options struct {
			Switch     map[string]bool
			CloseAfter map[string]time.Time
		}
	}{}
)

func init() {
	flag.CommandLine.Var(&options, "init.option", "usage: -init.option=databus.lottery=1 -grpc.target=service.like=0")
}

func Init() {
	// 加载配置文件选项
	if err := paladin.Get("options.toml").UnmarshalTOML(&config); err != nil {
		panic(err)
	}
	// 初始化过期配置
	now := time.Now()
	for option, expire := range config.Options.CloseAfter {
		if now.After(expire) {
			config.Options.Switch[option] = false
		}
	}
	// 加载命令行参数选项
	for _, option := range options {
		if strings.Contains(option, "=") {
			p := strings.Split(option, "=")
			config.Options.Switch[p[0]] = p[1] == "1" || p[1] == "true"
		} else {
			config.Options.Switch[option] = true
		}
	}
}
func IsOpen(f interface{}) (string, bool) {
	fName, ok := f.(string)
	if !ok {
		fName = runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		fName = reg.ReplaceAllString(fName, "")
		if strings.HasPrefix(fName, config.Base) {
			fName = fName[len(config.Base):]
		}
		if strings.HasPrefix(fName, "git.bilibili.co/bapis/bapis-go/") {
			fName = strings.ReplaceAll(fName, "git.bilibili.co/bapis/bapis-go/", "bapis/")
			if strings.HasSuffix(fName, ".NewClient") {
				fName = fName[0 : len(fName)-10]
			}
		} else {
			if idx := strings.LastIndex(fName, "-"); idx > 0 {
				if !strings.Contains(fName[idx:], "/") && !strings.Contains(fName[idx:], ".") {
					fName = fName[0:idx]
				}
			}
		}
	}
	if open, ok := config.Options.Switch[fName]; !ok || open {
		return fName, true
	}
	return fName, false
}

func do(f interface{}, a func() error) {
	if action, open := IsOpen(f); open {
		if err := a(); err != nil {
			log.Errorv(ctx, log.KV("action", action), log.KV("err", err))
			panic(fmt.Errorf("initialize %s with err: %v", action, err))
		}
	}
}

func Call(f func() error) {
	do(f, f)
}

func CallC(f func(ctx context.Context) error) {
	do(f, func() error {
		return f(ctx)
	})
}

func NewC(f interface{}, a func(ctx context.Context) error) {
	do(f, func() error {
		return a(ctx)
	})
}

func NewE(f interface{}, a func() error) {
	do(f, a)
}

func New(f interface{}, a func()) {
	do(f, func() error {
		a()
		return nil
	})
}
