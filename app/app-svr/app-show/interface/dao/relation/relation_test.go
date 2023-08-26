package relation

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-show")
		flag.Set("conf_token", "Pae4IDOeht4cHXCdOkay7sKeQwHxKOLA")
		flag.Set("tree_id", "2687")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-show-test.toml")

	}
	flag.Parse()
	cfg, err := confInit()
	if err != nil {
		panic(err)
	}
	d = New(cfg)
	os.Exit(m.Run())
}

func confInit() (*conf.Config, error) {
	err := paladin.Init()
	if err != nil {
		return nil, err
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("app-show.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func ctx() context.Context {
	return context.Background()
}

func TestRelations(t *testing.T) {
	Convey("get Relations all", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      map[int64]*relationgrpc.FollowingReply
			err      error
			mid      = int64(1581872)
			fids     = []int64{1581871, 1581873}
		)
		defer mockCtrl.Finish()
		mockArc := relationgrpc.NewMockRelationClient(mockCtrl)
		d.relGRPC = mockArc
		mockArc.EXPECT().Relations(context.TODO(), gomock.Any()).Return(res, nil)
		res, err = d.RelationsGRPC(ctx(), mid, fids)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

// AddFollowing
func TestAddFollowing(t *testing.T) {
	Convey("get Relations all", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			err      error
			spmid    string
			mid      = int64(1581872)
			fid      = int64(1581872)
		)
		defer mockCtrl.Finish()
		mockArc := relationgrpc.NewMockRelationClient(mockCtrl)
		d.relGRPC = mockArc
		mockArc.EXPECT().AddFollowing(context.TODO(), gomock.Any()).Return(nil, nil)
		err = d.AddFollowing(ctx(), fid, mid, spmid)
		So(err, ShouldBeNil)
	})
}

func TestDelFollowing(t *testing.T) {
	Convey("TestDelFollowing", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			err      error
			spmid    string
			mid      = int64(1581872)
			fid      = int64(1581872)
		)
		defer mockCtrl.Finish()
		mockArc := relationgrpc.NewMockRelationClient(mockCtrl)
		d.relGRPC = mockArc
		mockArc.EXPECT().DelFollowing(context.TODO(), gomock.Any()).Return(nil, nil)
		err = d.DelFollowing(ctx(), fid, mid, spmid)
		So(err, ShouldBeNil)
	})
}

func TestStatsGRPC(t *testing.T) {
	Convey("get StatsGRPC all", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      map[int64]*relationgrpc.StatReply
			err      error
			mids     = []int64{1581872}
		)
		defer mockCtrl.Finish()
		mockArc := relationgrpc.NewMockRelationClient(mockCtrl)
		d.relGRPC = mockArc
		mockArc.EXPECT().Stats(context.TODO(), gomock.Any()).Return(res, nil)
		res, err = d.StatsGRPC(ctx(), mids)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestAttentions(t *testing.T) {
	Convey("Attentions", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      []*relationgrpc.FollowingReply
			err      error
			mid      = int64(1581872)
		)
		defer mockCtrl.Finish()
		mockArc := relationgrpc.NewMockRelationClient(mockCtrl)
		d.relGRPC = mockArc
		mockArc.EXPECT().Attentions(context.TODO(), gomock.Any()).Return(res, nil)
		res, err = d.Attentions(ctx(), mid)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
