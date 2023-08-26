package ci

import (
	"context"
	"flag"
	"os"
	"reflect"
	"sync"
	"testing"

	bm "go-common/library/net/http/blademaster"

	"github.com/BurntSushi/toml"
	"github.com/robfig/cron"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	cdSvr "go-gateway/app/app-svr/fawkes/service/service/cd"
	gitSvr "go-gateway/app/app-svr/fawkes/service/service/gitlab"
)

var c conf.Config
var d *fawkes.Dao

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
	_, _ = toml.DecodeFile(confPath, &c)
	d = fawkes.New(&c)
}

func TestService_DeleteCINas(t *testing.T) {
	dk := cimdl.CISpecifyTimeDelete{
		AppKey:  "iphone",
		BuildId: 11111,
	}
	var deleteKeys []*cimdl.CISpecifyTimeDelete
	deleteKeys = append(deleteKeys, &dk)

	type fields struct {
		c             *conf.Config
		fkDao         *fawkes.Dao
		httpClient    *bm.Client
		ciChan        chan func()
		crontabCIProc map[int64]bool
		gitSvr        *gitSvr.Service
		cdSvr         *cdSvr.Service
		cron          *cron.Cron
		cronSwitch    string
	}
	type args struct {
		c   context.Context
		req *cimdl.CISpecifyTimeDeleteReq
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes *cimdl.CISpecifyTimeDeleteResp
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				c:             &c,
				fkDao:         d,
				httpClient:    nil,
				ciChan:        nil,
				crontabCIProc: nil,
				gitSvr:        nil,
				cdSvr:         nil,
				cron:          nil,
				cronSwitch:    "",
			},
			args: args{
				c: context.TODO(),
				req: &cimdl.CISpecifyTimeDeleteReq{
					DeleteKeys: deleteKeys,
				},
			},
			wantRes: nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				c:          tt.fields.c,
				fkDao:      tt.fields.fkDao,
				httpClient: tt.fields.httpClient,
				ciChan:     tt.fields.ciChan,
				gitSvr:     tt.fields.gitSvr,
				cdSvr:      tt.fields.cdSvr,
				cron:       tt.fields.cron,
				cronSwitch: tt.fields.cronSwitch,
			}
			gotRes, err := s.DeleteCINas(tt.args.c, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCINas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("DeleteCINas() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestService_saveMaven(t *testing.T) {
	type fields struct {
		c             *conf.Config
		fkDao         *fawkes.Dao
		httpClient    *bm.Client
		ciChan        chan func()
		crontabCIProc sync.Locker
		gitSvr        *gitSvr.Service
		cdSvr         *cdSvr.Service
		cron          *cron.Cron
		cronSwitch    string
	}
	type args struct {
		src        string
		dst        string
		AppKey     string
		bundleName string
		jobId      int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "maven上传",
			fields: fields{
				c:             &c,
				fkDao:         d,
				httpClient:    nil,
				ciChan:        nil,
				crontabCIProc: nil,
				gitSvr:        nil,
				cdSvr:         nil,
				cron:          nil,
				cronSwitch:    "",
			},
			args: args{
				src:        "/Users/wdlu/Downloads/main.bbr",
				dst:        "/Users/wdlu/Downloads/maven/",
				AppKey:     "TestApp",
				bundleName: "host",
				jobId:      10222,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				c:          tt.fields.c,
				fkDao:      tt.fields.fkDao,
				httpClient: tt.fields.httpClient,
				ciChan:     tt.fields.ciChan,
				gitSvr:     tt.fields.gitSvr,
				cdSvr:      tt.fields.cdSvr,
				cron:       tt.fields.cron,
				cronSwitch: tt.fields.cronSwitch,
			}
			if err := s.saveMaven(tt.args.src, tt.args.dst, tt.args.AppKey, tt.args.bundleName, tt.args.jobId); (err != nil) != tt.wantErr {
				t.Errorf("saveMaven() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getApiFromBbr(t *testing.T) {
	Convey("test tribe api command", t, func() {
		execTribeAPI(context.Background(), "/Users/wdlu/Downloads/main.bbr", "")
	})
}

func Test_readbbr(t *testing.T) {
	Convey("test tribe api command", t, func() {

		meta, _ := getMetaJson("/Users/wdlu/Downloads/main.bbr")
		vers := meta["compatibleVersions"].([]interface{})
		v := vers[len(vers)-1]
		v = v.(float64)
		v = int(v.(float64))
		So(v, ShouldBeGreaterThan, 0)
	})
}
