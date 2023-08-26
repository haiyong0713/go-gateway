package dao

import (
	"context"
	"testing"

	pb "go-gateway/app/app-svr/app-gw/management/api"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestDaosnapshotKey(t *testing.T) {
	Convey("snapshotKey", t, func() {
		var (
			node    = "main.web-svr"
			gateway = "playlist-gateway"
			uuid    = "sdfafdasasq21"
		)
		Convey("When everything goes positive", func() {
			p1 := snapshotKey(uuid, node, gateway)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaosnapshotBreakerAPIKey(t *testing.T) {
	Convey("snapshotBreakerAPIKey", t, func() {
		var (
			node    = "main.web-svr"
			gateway = "playlist-gateway"
			uuid    = "sdfafdasasq21"
			api     = "/test_11"
		)
		Convey("When everything goes positive", func() {
			p1 := snapshotBreakerAPIKey(uuid, node, gateway, api)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaosnapshotDynPathKey(t *testing.T) {
	Convey("snapshotDynPathKey", t, func() {
		var (
			node    = "main.web-svr"
			gateway = "playlist-gateway"
			uuid    = "sdfafdasasq21"
			pattern = ""
		)
		Convey("When everything goes positive", func() {
			p1 := snapshotDynPathKey(uuid, node, gateway, pattern)
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
			node    = "main.web-svr"
			gateway = "playlist-gateway"
			uuid    = "sdfafdasasq21"
		)
		Convey("When everything goes positive", func() {
			p1, err := d.CreateSnapshotDao().ListBreakerAPI(ctx, node, gateway, uuid)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetBreakerAPI(t *testing.T) {
	Convey("SetBreakerAPI", t, func() {
		var (
			ctx  = context.Background()
			req  = &pb.SetBreakerAPIReq{}
			uuid = ""
		)
		Convey("When everything goes positive", func() {
			err := d.CreateSnapshotDao().SetBreakerAPI(ctx, req, uuid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoEnableBreakerAPI(t *testing.T) {
	Convey("EnableBreakerAPI", t, func() {
		var (
			ctx  = context.Background()
			req  = &pb.EnableBreakerAPIReq{}
			uuid = ""
		)
		Convey("When everything goes positive", func() {
			err := d.CreateSnapshotDao().EnableBreakerAPI(ctx, req, uuid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoDeleteBreakerAPI(t *testing.T) {
	Convey("DeleteBreakerAPI", t, func() {
		var (
			ctx  = context.Background()
			req  = &pb.DeleteBreakerAPIReq{}
			uuid = ""
		)
		Convey("When everything goes positive", func() {
			err := d.CreateSnapshotDao().DeleteBreakerAPI(ctx, req, uuid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoListDynPath(t *testing.T) {
	Convey("ListDynPath", t, func() {
		var (
			ctx     = context.Background()
			node    = "main.web-svr"
			gateway = "playlist-gateway"
			uuid    = "sdfafdasasq21"
		)
		Convey("When everything goes positive", func() {
			p1, err := d.CreateSnapshotDao().ListDynPath(ctx, node, gateway, uuid)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetDynPath(t *testing.T) {
	Convey("SetDynPath", t, func() {
		var (
			ctx = context.Background()
			req = &pb.SetDynPathReq{
				Node:    "main.web-svr",
				Gateway: "playlist-gateway",
				Pattern: "/test_11",
			}
			uuid = ""
		)
		Convey("When everything goes positive", func() {
			err := d.CreateSnapshotDao().SetDynPath(ctx, req, uuid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoDeleteDynPath(t *testing.T) {
	Convey("DeleteDynPath", t, func() {
		var (
			ctx = context.Background()
			req = &pb.DeleteDynPathReq{
				Node:    "main.web-svr",
				Gateway: "playlist-gateway",
				Pattern: "/test_11",
			}
			uuid = ""
		)
		Convey("When everything goes positive", func() {
			err := d.CreateSnapshotDao().DeleteDynPath(ctx, req, uuid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoEnableDynPath(t *testing.T) {
	Convey("EnableDynPath", t, func() {
		var (
			ctx = context.Background()
			req = &pb.EnableDynPathReq{
				Node:    "main.web-svr",
				Gateway: "playlist-gateway",
				Pattern: "/test_11",
			}
			uuid = "sdfafdasasq21"
		)
		Convey("When everything goes positive", func() {
			err := d.CreateSnapshotDao().EnableDynPath(ctx, req, uuid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoGetSnapshotMeta(t *testing.T) {
	Convey("GetSnapshotMeta", t, func() {
		var (
			ctx     = context.Background()
			node    = "main.web-svr"
			gateway = "playlist-gateway"
			uuid    = "sdfafdasasq21"
		)
		Convey("When everything goes positive", func() {
			p1, err := d.CreateSnapshotDao().GetSnapshotMeta(ctx, node, gateway, uuid)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddSnapshot(t *testing.T) {
	Convey("AddSnapshot", t, func() {
		var (
			ctx = context.Background()
			req = &pb.AddSnapshotReq{
				Node:    "main.web-svr",
				Gateway: "playlist-gateway",
			}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.CreateSnapshotDao().AddSnapshot(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoBuildPlan(t *testing.T) {
	Convey("BuildPlan", t, func() {
		var (
			ctx     = context.Background()
			node    = "main.web-svr"
			gateway = "playlist-gateway"
			uuid    = "sdfafdasasq21"
		)
		Convey("When everything goes positive", func() {
			p1, err := d.CreateSnapshotDao().BuildPlan(ctx, node, gateway, uuid)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoRunPlan(t *testing.T) {
	Convey("RunPlan", t, func() {
		var (
			ctx = context.Background()
			req = &pb.SnapshotRunPlan{}
		)
		Convey("When everything goes positive", func() {
			err := d.CreateSnapshotDao().RunPlan(ctx, req)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestConstructBatchReq(t *testing.T) {
	ba := []*pb.BreakerAPI{
		{
			Api:     "/x/test_1",
			Ratio:   100,
			Node:    "main.web-svr",
			Gateway: "playlist-gateway",
		},
		{
			Api:     "/x/test_2",
			Ratio:   80,
			Node:    "main.web-svr",
			Gateway: "playlist-gateway",
		},
		{
			Api:     "/x/test_3",
			Ratio:   50,
			Node:    "main.web-svr",
			Gateway: "playlist-gateway",
		},
	}
	ssba := []*pb.BreakerAPI{
		{
			Api:     "/x/test_1",
			Ratio:   20,
			Node:    "main.web-svr",
			Gateway: "playlist-gateway",
		},
		{
			Api:     "/x/test_2",
			Ratio:   80,
			Node:    "main.web-svr",
			Gateway: "playlist-gateway",
		},
	}
	dp := []*pb.DynPath{
		{
			Node:    "main.web-svr",
			Gateway: "playlist-gateway",
			Pattern: "~ ^/test",
		},
		{
			Node:    "main.web-svr",
			Gateway: "playlist-gateway",
			Pattern: "~ ^/test",
		},
	}
	ssdp := []*pb.DynPath{
		{
			Node:    "main.web-svr",
			Gateway: "playlist-gateway",
			Pattern: "~ ^/test",
		},
		{
			Node:    "main.web-svr",
			Gateway: "playlist-gateway",
			Pattern: "~ ^/test",
		},
	}
	res := constructBatchReq(ssba, ba, ssdp, dp)
	assert.Equal(t, len(res.DelDynReq), 0)
	assert.Equal(t, len(res.DelBreakerReq), 1)
	assert.Equal(t, len(res.SetDynReq), 0)
	assert.Equal(t, len(res.SetBreakerReq), 1)
}
