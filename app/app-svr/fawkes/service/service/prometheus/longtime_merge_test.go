package prometheus

import (
	"flag"
	"fmt"
	"os"
	"testing"

	bm "go-common/library/net/http/blademaster"

	"github.com/BurntSushi/toml"
	"github.com/robfig/cron"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
)

var c *conf.Config

func init() {
	var confPath string
	strings := flag.CommandLine.Args()
	fmt.Printf("%s", strings)
	args := os.Args[1:]
	for i, v := range args {
		if v == "-conf" {
			args = args[i:]
			break
		}
	}
	tf := flag.NewFlagSet("test", flag.ContinueOnError)
	tf.StringVar(&confPath, "conf", "", "config")
	if err := tf.Parse(args); err != nil {
		return
	}
	_, _ = toml.DecodeFile(confPath, &c)
}

func TestService_longProcessMerge(t *testing.T) {
	type fields struct {
		c          *conf.Config
		fkDao      *fkdao.Dao
		httpClient *bm.Client
		cron       *cron.Cron
		cronSwitch string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "CASE1",
			fields: fields{
				c:          c,
				fkDao:      fkdao.New(c),
				cronSwitch: "ON",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				c:          tt.fields.c,
				fkDao:      tt.fields.fkDao,
				httpClient: tt.fields.httpClient,
				cron:       tt.fields.cron,
				cronSwitch: tt.fields.cronSwitch,
			}
			s.longProcessMerge()
		})
	}
}
