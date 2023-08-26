package dao

import (
	"context"
	"testing"

	gwconfig "go-gateway/app/app-svr/app-gw/management-job/internal/model/gateway-config"
	pb "go-gateway/app/app-svr/app-gw/management/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaobreakerAPIKey(t *testing.T) {
	Convey("breakerAPIKey", t, func() {
		var (
			node    = ""
			gateway = ""
			api     = ""
		)
		Convey("When everything goes positive", func() {
			p1 := breakerAPIKey(node, gateway, api)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaofullRange(t *testing.T) {
	Convey("fullRange", t, func() {
		var (
			prefix = ""
		)
		Convey("When everything goes positive", func() {
			p1, p2 := fullRange(prefix)
			Convey("Then p1,p2 should not be nil.", func() {
				So(p2, ShouldNotBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaofilterBreakderAPI(t *testing.T) {
	Convey("filterBreakderAPI", t, func() {
		var (
			in     = []*pb.BreakerAPI{}
			filter func(*pb.BreakerAPI) bool
		)
		Convey("When everything goes positive", func() {
			p1 := filterBreakderAPI(in, filter)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoListBreakerAPI(t *testing.T) {
	Convey("ListBreakerAPI", t, func() {
		var (
			ctx     = context.Background()
			node    = ""
			gateway = ""
		)
		Convey("When everything goes positive", func() {
			p1, err := d.ListBreakerAPI(ctx, node, gateway)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoPushConfigs(t *testing.T) {
	Convey("PushConfigs", t, func() {
		var (
			ctx = context.Background()
			req = &gwconfig.PushConfigReq{}
		)
		Convey("When everything goes positive", func() {
			err := d.PushConfigs(ctx, req)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoRawConfigs(t *testing.T) {
	Convey("RawConfigs", t, func() {
		var (
			ctx = context.Background()
			req = &gwconfig.RawConfigReq{AppID: "main.web-svr.web-gateway", TreeID: 207634, ConfigMeta: &pb.ConfigMeta{Token: "b3f5524a5bcd95822c7517052e3f3cda", Env: "uat", Zone: "sh001", BuildName: "docker-1", Filename: "proxy-config.toml"}}
		)
		Convey("When everything goes positive", func() {
			res, err := d.RawConfigs(ctx, req)
			Convey("Then err should be nil.", func() {
				Print(res)
				So(err, ShouldBeNil)
			})
		})
	})
}
