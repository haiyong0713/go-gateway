package cd

import (
	"bytes"
	"flag"
	"os"
	"testing"

	"github.com/BurntSushi/toml"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
)

var C *conf.Config
var D *fawkes.Dao

func init() {
	var confPath string
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
	_, _ = toml.DecodeFile(confPath, &C)
	D = fawkes.New(C)
}

func Test_resultReport(t *testing.T) {
	type args struct {
		success     bool
		brief       bytes.Buffer
		appKey      string
		logPath     string
		gitlabJobId int64
		d           *fawkes.Dao
		conf        *conf.Config
	}

	tests := []struct {
		name string
		args args
	}{
		{
			"正确示例",
			args{success: true,
				brief:       bytes.Buffer{},
				appKey:      "test",
				logPath:     "/mnt/build-archive/archive/fawkes/pack/iphone/5634318/store/upload_log.txt",
				gitlabJobId: 1111,
				conf:        C,
				d:           D,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.args
			resultReport(a.success, a.brief, a.appKey, a.logPath, a.gitlabJobId, a.d, a.conf)
		})
	}
}
