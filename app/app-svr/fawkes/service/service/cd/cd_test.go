package cd

import (
	"context"
	"testing"

	"github.com/robfig/cron"
	"github.com/xanzy/go-gitlab"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	ossdao "go-gateway/app/app-svr/fawkes/service/dao/oss"
	gitSvr "go-gateway/app/app-svr/fawkes/service/service/gitlab"
	mdlSvr "go-gateway/app/app-svr/fawkes/service/service/modules"
	"go-gateway/app/app-svr/fawkes/service/tools/appstoreconnect"
)

func TestService_AppCDGenerateAddGit(t *testing.T) {
	conf.Conf = C
	type fields struct {
		c              *conf.Config
		fkDao          *fkdao.Dao
		ossDao         *ossdao.Dao
		appstoreClient *appstoreconnect.Client
		gitlabClient   *gitlab.Client
		gitSvr         *gitSvr.Service
		mdlSvr         *mdlSvr.Service
		hanlderChan    []chan func()
		cron           *cron.Cron
	}
	type args struct {
		ctx      context.Context
		appKey   string
		channels string
		userName string
		buildID  int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			fields: fields{
				c:              C,
				fkDao:          D,
				ossDao:         nil,
				appstoreClient: nil,
				gitlabClient:   nil,
				gitSvr:         gitSvr.New(C),
				mdlSvr:         nil,
				hanlderChan:    nil,
				cron:           nil,
			},
			args: args{
				ctx:      context.Background(),
				appKey:   "w19e",
				channels: "[{\"channel_id\":8599,\"channel\":\"932_xxl_jrtt_lm_1\"},{\"channel_id\":8208,\"channel\":\"954_xxl_wx_yf_4\"},{\"channel_id\":8207,\"channel\":\"951_xxl_gdt_dw_5\"},{\"channel_id\":8206,\"channel\":\"951_xxl_gdt_dw_4\"},{\"channel_id\":8205,\"channel\":\"954_xxl_wx_yf_3\"},{\"channel_id\":8204,\"channel\":\"951_xxl_gdt_dw_3\"},{\"channel_id\":8203,\"channel\":\"951_xxl_gdt_dw_2\"},{\"channel_id\":8202,\"channel\":\"951_xxl_gdt_dw_1\"},{\"channel_id\":8201,\"channel\":\"954_xxl_wx_yf_2\"},{\"channel_id\":8200,\"channel\":\"954_xxl_wx_yf_1\"},{\"channel_id\":2724,\"channel\":\"wm_wdj_15395\"},{\"channel_id\":2708,\"channel\":\"APP_sub_15379\"},{\"channel_id\":2709,\"channel\":\"APP_sub_15380\"},{\"channel_id\":2710,\"channel\":\"APP_sub_15381\"},{\"channel_id\":2711,\"channel\":\"APP_sub_15382\"},{\"channel_id\":2712,\"channel\":\"APP_sub_15383\"},{\"channel_id\":46,\"channel\":\"chuizi\"},{\"channel_id\":66,\"channel\":\"sanxing\"},{\"channel_id\":13,\"channel\":\"yingyongbao\"},{\"channel_id\":12,\"channel\":\"yingyonghui\"},{\"channel_id\":11,\"channel\":\"360os\"},{\"channel_id\":8,\"channel\":\"baidu\"},{\"channel_id\":98,\"channel\":\"test\"},{\"channel_id\":5,\"channel\":\"master\"},{\"channel_id\":64,\"channel\":\"jinli\"},{\"channel_id\":856,\"channel\":\"test-app\"},{\"channel_id\":2485,\"channel\":\"xxl_jrtt_249\"},{\"channel_id\":2486,\"channel\":\"xxl_jrtt_250\"},{\"channel_id\":2487,\"channel\":\"xxl_jrtt_251\"},{\"channel_id\":2488,\"channel\":\"xxl_jrtt_252\"},{\"channel_id\":2489,\"channel\":\"xxl_APPfx_253\"},{\"channel_id\":14,\"channel\":\"alifenfa\"},{\"channel_id\":16,\"channel\":\"oppo\"},{\"channel_id\":48,\"channel\":\"lenovo\"},{\"channel_id\":45,\"channel\":\"vivo\"},{\"channel_id\":51,\"channel\":\"zhongxing\"},{\"channel_id\":41,\"channel\":\"sougou2\"},{\"channel_id\":53,\"channel\":\"meizu\"},{\"channel_id\":34,\"channel\":\"liantong\"},{\"channel_id\":30,\"channel\":\"nby\"},{\"channel_id\":20,\"channel\":\"nokia\"},{\"channel_id\":58,\"channel\":\"pairui01\"},{\"channel_id\":59,\"channel\":\"sougou\"},{\"channel_id\":60,\"channel\":\"anzhi\"},{\"channel_id\":61,\"channel\":\"360\"},{\"channel_id\":62,\"channel\":\"xiaomi\"},{\"channel_id\":18,\"channel\":\"huawei\"},{\"channel_id\":17,\"channel\":\"baidusem1\"},{\"channel_id\":2490,\"channel\":\"xxl_jrtt_254\"},{\"channel_id\":2491,\"channel\":\"xxl_jrtt_256\"}]",
				userName: "luweidan",
				buildID:  2276252,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				c:              tt.fields.c,
				fkDao:          tt.fields.fkDao,
				ossDao:         tt.fields.ossDao,
				appstoreClient: tt.fields.appstoreClient,
				gitlabClient:   tt.fields.gitlabClient,
				gitSvr:         tt.fields.gitSvr,
				mdlSvr:         tt.fields.mdlSvr,
				hanlderChan:    tt.fields.hanlderChan,
				cron:           tt.fields.cron,
			}
			if err := s.AppCDGenerateAddGit(tt.args.ctx, tt.args.appKey, tt.args.channels, tt.args.userName, tt.args.buildID); (err != nil) != tt.wantErr {
				t.Errorf("AppCDGenerateAddGit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkService_AppCDGenerateAddGit(b *testing.B) {
	conf.Conf = C
	type fields struct {
		c              *conf.Config
		fkDao          *fkdao.Dao
		ossDao         *ossdao.Dao
		appstoreClient *appstoreconnect.Client
		gitlabClient   *gitlab.Client
		gitSvr         *gitSvr.Service
		mdlSvr         *mdlSvr.Service
		hanlderChan    []chan func()
		cron           *cron.Cron
	}
	type args struct {
		ctx      context.Context
		appKey   string
		channels string
		userName string
		buildID  int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			fields: fields{
				c:              C,
				fkDao:          D,
				ossDao:         nil,
				appstoreClient: nil,
				gitlabClient:   nil,
				gitSvr:         gitSvr.New(C),
				mdlSvr:         nil,
				hanlderChan:    nil,
				cron:           nil,
			},
			args: args{
				ctx:      context.Background(),
				appKey:   "w19e",
				channels: "[{\"channel_id\":8599,\"channel\":\"932_xxl_jrtt_lm_1\"},{\"channel_id\":8208,\"channel\":\"954_xxl_wx_yf_4\"},{\"channel_id\":8207,\"channel\":\"951_xxl_gdt_dw_5\"},{\"channel_id\":8206,\"channel\":\"951_xxl_gdt_dw_4\"},{\"channel_id\":8205,\"channel\":\"954_xxl_wx_yf_3\"},{\"channel_id\":8204,\"channel\":\"951_xxl_gdt_dw_3\"},{\"channel_id\":8203,\"channel\":\"951_xxl_gdt_dw_2\"},{\"channel_id\":8202,\"channel\":\"951_xxl_gdt_dw_1\"},{\"channel_id\":8201,\"channel\":\"954_xxl_wx_yf_2\"},{\"channel_id\":8200,\"channel\":\"954_xxl_wx_yf_1\"},{\"channel_id\":2724,\"channel\":\"wm_wdj_15395\"},{\"channel_id\":2708,\"channel\":\"APP_sub_15379\"},{\"channel_id\":2709,\"channel\":\"APP_sub_15380\"},{\"channel_id\":2710,\"channel\":\"APP_sub_15381\"},{\"channel_id\":2711,\"channel\":\"APP_sub_15382\"},{\"channel_id\":2712,\"channel\":\"APP_sub_15383\"},{\"channel_id\":46,\"channel\":\"chuizi\"},{\"channel_id\":66,\"channel\":\"sanxing\"},{\"channel_id\":13,\"channel\":\"yingyongbao\"},{\"channel_id\":12,\"channel\":\"yingyonghui\"},{\"channel_id\":11,\"channel\":\"360os\"},{\"channel_id\":8,\"channel\":\"baidu\"},{\"channel_id\":98,\"channel\":\"test\"},{\"channel_id\":5,\"channel\":\"master\"},{\"channel_id\":64,\"channel\":\"jinli\"},{\"channel_id\":856,\"channel\":\"test-app\"},{\"channel_id\":2485,\"channel\":\"xxl_jrtt_249\"},{\"channel_id\":2486,\"channel\":\"xxl_jrtt_250\"},{\"channel_id\":2487,\"channel\":\"xxl_jrtt_251\"},{\"channel_id\":2488,\"channel\":\"xxl_jrtt_252\"},{\"channel_id\":2489,\"channel\":\"xxl_APPfx_253\"},{\"channel_id\":14,\"channel\":\"alifenfa\"},{\"channel_id\":16,\"channel\":\"oppo\"},{\"channel_id\":48,\"channel\":\"lenovo\"},{\"channel_id\":45,\"channel\":\"vivo\"},{\"channel_id\":51,\"channel\":\"zhongxing\"},{\"channel_id\":41,\"channel\":\"sougou2\"},{\"channel_id\":53,\"channel\":\"meizu\"},{\"channel_id\":34,\"channel\":\"liantong\"},{\"channel_id\":30,\"channel\":\"nby\"},{\"channel_id\":20,\"channel\":\"nokia\"},{\"channel_id\":58,\"channel\":\"pairui01\"},{\"channel_id\":59,\"channel\":\"sougou\"},{\"channel_id\":60,\"channel\":\"anzhi\"},{\"channel_id\":61,\"channel\":\"360\"},{\"channel_id\":62,\"channel\":\"xiaomi\"},{\"channel_id\":18,\"channel\":\"huawei\"},{\"channel_id\":17,\"channel\":\"baidusem1\"},{\"channel_id\":2490,\"channel\":\"xxl_jrtt_254\"},{\"channel_id\":2491,\"channel\":\"xxl_jrtt_256\"}]",
				userName: "luweidan",
				buildID:  2276252,
			},
			wantErr: false,
		},
	}
	ts := tests[0]
	s := &Service{
		c:              ts.fields.c,
		fkDao:          ts.fields.fkDao,
		ossDao:         ts.fields.ossDao,
		appstoreClient: ts.fields.appstoreClient,
		gitlabClient:   ts.fields.gitlabClient,
		gitSvr:         ts.fields.gitSvr,
		mdlSvr:         ts.fields.mdlSvr,
		hanlderChan:    ts.fields.hanlderChan,
		cron:           ts.fields.cron,
	}
	for i := 0; i < b.N; i++ {
		err := s.AppCDGenerateAddGit(ts.args.ctx, ts.args.appKey, ts.args.channels, ts.args.userName, ts.args.buildID)
		if err != nil {
			return
		}
	}

}

func TestService_GeneratesUpdate(t *testing.T) {
	conf.Conf = C
	type fields struct {
		c              *conf.Config
		fkDao          *fkdao.Dao
		ossDao         *ossdao.Dao
		appstoreClient *appstoreconnect.Client
		gitlabClient   *gitlab.Client
		gitSvr         *gitSvr.Service
		mdlSvr         *mdlSvr.Service
		hanlderChan    []chan func()
		cron           *cron.Cron
	}
	type args struct {
		c                  context.Context
		appKey             string
		channelFileInfoStr string
		jobID              int64
		channelStatus      int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "GeneratesUpdate_HappyPass",
			fields: fields{
				c:              C,
				fkDao:          D,
				ossDao:         ossdao.New(C),
				appstoreClient: appstoreconnect.NewClient(C),
				gitlabClient:   gitlab.NewClient(nil, C.Gitlab.Token),
				gitSvr:         gitSvr.New(C),
				mdlSvr:         mdlSvr.New(C),
				hanlderChan:    make([]chan func(), 10),
				cron:           cron.New(),
			},
			args: args{
				c:                  context.Background(),
				appKey:             "w19e",
				channelFileInfoStr: "[{\"id\":251752,\"md5\":\"4c563d09c51a80d1a7d217720f059c14\",\"path\":\"/mnt/build-archive/archive/fawkes/pack/w19e/5432620/channel/iBiliPlayer-apinkRelease-6.25.0-b5432620-951_xxl_gdt_dw_1.apk\",\"size\":70878642}]",
				jobID:              251752,
				channelStatus:      0,
			},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				c:              tt.fields.c,
				fkDao:          tt.fields.fkDao,
				ossDao:         tt.fields.ossDao,
				appstoreClient: tt.fields.appstoreClient,
				gitlabClient:   tt.fields.gitlabClient,
				gitSvr:         tt.fields.gitSvr,
				mdlSvr:         tt.fields.mdlSvr,
				hanlderChan:    tt.fields.hanlderChan,
				cron:           tt.fields.cron,
			}
			if err := s.GeneratesUpdate(tt.args.c, tt.args.appKey, tt.args.channelFileInfoStr, tt.args.jobID, tt.args.channelStatus); (err != nil) != tt.wantErr {
				t.Errorf("GeneratesUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
