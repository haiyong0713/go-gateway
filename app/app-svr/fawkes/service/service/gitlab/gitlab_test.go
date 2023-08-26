package gitlab

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/robfig/cron"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/xanzy/go-gitlab"
	ggl "github.com/xanzy/go-gitlab"

	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
)

var (
	srv *Service
	D   *fawkes.Dao
)

func init() {
	dir, _ := filepath.Abs("../../fawkes_admin.toml")
	flag.Set("conf", dir)
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if srv == nil {
		srv = New(conf.Conf)
	}
	D = fawkes.New(conf.Conf)
	srv = New(conf.Conf)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(srv)
	}
}

func TestService_TriggerPipeline(t *testing.T) {

	var varMap = map[string]string{
		"APP_KEY":        "android",
		"TASK":           "CHANNEL",
		"BUILD_ID":       "2276252",
		"ORIGIN_APK_URL": "https://dl.hdslb.com/mobile/pack/w19e/2276252/iBiliPlayer-release-5.46.0-b356045.apk",
		"LOCAL_APK_URL":  "https://macross-jks.bilibili.co/archive/fawkes/pack/w19e/2276252/iBiliPlayer-release-5.46.0-b356045.apk",
		"CHANNELS":       "[{\"channel\":\"932_xxl_jrtt_lm_1\",\"id\":161519},{\"channel\":\"954_xxl_wx_yf_4\",\"id\":161520},{\"channel\":\"951_xxl_gdt_dw_5\",\"id\":161521},{\"channel\":\"951_xxl_gdt_dw_4\",\"id\":161522},{\"channel\":\"954_xxl_wx_yf_3\",\"id\":161523},{\"channel\":\"951_xxl_gdt_dw_3\",\"id\":161524},{\"channel\":\"951_xxl_gdt_dw_2\",\"id\":161525},{\"channel\":\"951_xxl_gdt_dw_1\",\"id\":161526},{\"channel\":\"954_xxl_wx_yf_2\",\"id\":161527},{\"channel\":\"954_xxl_wx_yf_1\",\"id\":161528},{\"channel\":\"wm_wdj_15395\",\"id\":161529},{\"channel\":\"APP_sub_15379\",\"id\":161530},{\"channel\":\"APP_sub_15380\",\"id\":161531},{\"channel\":\"APP_sub_15381\",\"id\":161532},{\"channel\":\"APP_sub_15382\",\"id\":161533},{\"channel\":\"APP_sub_15383\",\"id\":161534},{\"channel\":\"chuizi\",\"id\":161535},{\"channel\":\"sanxing\",\"id\":161536},{\"channel\":\"yingyongbao\",\"id\":161537},{\"channel\":\"yingyonghui\",\"id\":161538},{\"channel\":\"360os\",\"id\":161539},{\"channel\":\"baidu\",\"id\":161540},{\"channel\":\"test\",\"id\":161541},{\"channel\":\"master\",\"id\":161542},{\"channel\":\"jinli\",\"id\":161543},{\"channel\":\"test-app\",\"id\":161544},{\"channel\":\"xxl_jrtt_249\",\"id\":161545},{\"channel\":\"xxl_jrtt_250\",\"id\":161546},{\"channel\":\"xxl_jrtt_251\",\"id\":161547},{\"channel\":\"xxl_jrtt_252\",\"id\":161548},{\"channel\":\"xxl_APPfx_253\",\"id\":161549},{\"channel\":\"alifenfa\",\"id\":161550},{\"channel\":\"oppo\",\"id\":161551},{\"channel\":\"lenovo\",\"id\":161552},{\"channel\":\"vivo\",\"id\":161553},{\"channel\":\"zhongxing\",\"id\":161554},{\"channel\":\"sougou2\",\"id\":161555},{\"channel\":\"meizu\",\"id\":161556},{\"channel\":\"liantong\",\"id\":161557},{\"channel\":\"nby\",\"id\":161558},{\"channel\":\"nokia\",\"id\":161559},{\"channel\":\"pairui01\",\"id\":161560},{\"channel\":\"sougou\",\"id\":161561},{\"channel\":\"anzhi\",\"id\":161562},{\"channel\":\"360\",\"id\":161563},{\"channel\":\"xiaomi\",\"id\":161564},{\"channel\":\"huawei\",\"id\":161565},{\"channel\":\"baidusem1\",\"id\":161566},{\"channel\":\"xxl_jrtt_254\",\"id\":161567},{\"channel\":\"xxl_jrtt_256\",\"id\":161568}]",
	}

	type fields struct {
		c            *conf.Config
		fkDao        *fkdao.Dao
		httpClient   *bm.Client
		gitlabClient *gitlab.Client
		cron         *cron.Cron
	}
	type args struct {
		ctx       context.Context
		appKey    string
		gitType   int
		gitName   string
		variables map[string]string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantPipeline *ggl.Pipeline
		wantErr      bool
	}{
		{
			name: "case1",
			fields: fields{
				c:            conf.Conf,
				fkDao:        D,
				httpClient:   bm.NewClient(conf.Conf.HTTPClient),
				gitlabClient: gitlab.NewClient(nil, conf.Conf.Gitlab.Token),
				cron:         cron.New(),
			},
			args: args{
				ctx:       context.Background(),
				appKey:    "android",
				gitType:   0,
				gitName:   "keep/build_apk_channels",
				variables: varMap,
			},
			wantPipeline: nil,
			wantErr:      false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				c:            tt.fields.c,
				fkDao:        tt.fields.fkDao,
				httpClient:   tt.fields.httpClient,
				gitlabClient: tt.fields.gitlabClient,
				cron:         tt.fields.cron,
			}
			_, err := s.TriggerPipeline(tt.args.ctx, tt.args.appKey, tt.args.gitType, tt.args.gitName, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("TriggerPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestService_TriggerPipeline1(t *testing.T) {
	Convey("trigger", t, WithService(func(s *Service) {
		_, err := s.TriggerPipeline(context.Background(), "w19e", 0, "", nil)
		So(err, ShouldBeNil)
	}))
}
