package dao

import (
	"context"
	"testing"

	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDaodeploymentMetaKey(t *testing.T) {
	Convey("deploymentMetaKey", t, func() {
		var (
			node    = "main.web-svr"
			gateway = "playlist-gateway"
			id      = "9223372035264712524/1590063283"
		)
		Convey("When everything goes positive", func() {
			p1 := deploymentMetaKey("http", node, gateway, id)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaodeploymentConfirmKey(t *testing.T) {
	Convey("deploymentConfirmKey", t, func() {
		var (
			node    = "main.web-svr"
			gateway = "playlist-gateway"
			id      = "9223372035264712524/1590063283"
		)
		Convey("When everything goes positive", func() {
			p1 := deploymentConfirmKey(node, gateway, id)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaodeploymentActionLogKey(t *testing.T) {
	Convey("deploymentActionLogKey", t, func() {
		var (
			node      = "main.web-svr"
			gateway   = "playlist-gateway"
			id        = "9223372035264712524/1590063283"
			createdAt = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := deploymentActionLogKey(node, gateway, id, createdAt)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoCreateDeploymentMeta(t *testing.T) {
	Convey("CreateDeploymentMeta", t, func() {
		var (
			ctx  = context.Background()
			meta = &pb.DeploymentMeta{}
		)
		Convey("When everything goes positive", func() {
			err := d.CreateDeploymentMeta(ctx, meta)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoSetDeploymentMeta(t *testing.T) {
	Convey("SetDeploymentMeta", t, func() {
		var (
			ctx  = context.Background()
			meta = &pb.DeploymentMeta{}
		)
		Convey("When everything goes positive", func() {
			err := d.SetDeploymentMeta(ctx, meta)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoUpdateDeploymentState(t *testing.T) {
	Convey("UpdateDeploymentState", t, func() {
		var (
			ctx = context.Background()
			src = &pb.DeploymentMeta{}
			dst = &pb.DeploymentMeta{}
		)
		Convey("When everything goes positive", func() {
			err := d.UpdateDeploymentState(ctx, src, dst)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoSetDeploymentConfirm(t *testing.T) {
	Convey("SetDeploymentConfirm", t, func() {
		var (
			ctx = context.Background()
			req = &pb.DeploymentReq{
				Node:         "main.web-svr",
				Gateway:      "playlist-gateway",
				DeploymentId: "9223372035264712524/1590063283",
			}
			confirm = &pb.DeploymentConfirm{
				Sponsor: "test",
			}
		)
		Convey("When everything goes positive", func() {
			err := d.SetDeploymentConfirm(ctx, req, confirm)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoGetDeploymentMeta(t *testing.T) {
	Convey("GetDeploymentMeta", t, func() {
		var (
			ctx = context.Background()
			req = &pb.DeploymentReq{
				Node:         "main.web-svr",
				Gateway:      "playlist-gateway",
				DeploymentId: "9223372035264712524/1590063283",
			}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.GetDeploymentMeta(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoGetDeploymentConfirm(t *testing.T) {
	Convey("GetDeploymentConfirm", t, func() {
		var (
			ctx = context.Background()
			req = &pb.DeploymentReq{
				Node:         "main.web-svr",
				Gateway:      "playlist-gateway",
				DeploymentId: "9223372035264712524/1590063283",
			}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.GetDeploymentConfirm(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoDeploymentIsConfirmed(t *testing.T) {
	Convey("DeploymentIsConfirmed", t, func() {
		var (
			ctx = context.Background()
			req = &pb.DeploymentReq{
				Node:         "main.web-svr",
				Gateway:      "playlist-gateway",
				DeploymentId: "9223372035264712524/1590063283",
			}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.DeploymentIsConfirmed(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaocheckErr(t *testing.T) {
	Convey("checkErr", t, func() {
		var (
			err = errors.New("test error")
		)
		Convey("When everything goes positive", func() {
			p1, err := checkErr(err)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoGetDeploymentActionLog(t *testing.T) {
	Convey("GetDeploymentActionLog", t, func() {
		var (
			ctx = context.Background()
			req = &pb.DeploymentReq{
				Node:         "main.web-svr",
				Gateway:      "playlist-gateway",
				DeploymentId: "9223372035264712524/1590063283",
			}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.GetDeploymentActionLog(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoReloadConfig(t *testing.T) {
	Convey("ReloadConfig", t, func() {
		var (
			ctx = context.Background()
			req = &model.ReloadConfigReq{}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.ReloadConfig(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddActionLog(t *testing.T) {
	Convey("AddActionLog", t, func() {
		var (
			ctx = context.Background()
			req = &pb.AddActionLogReq{
				Node:         "main.web-svr",
				Gateway:      "playlist-gateway",
				DeploymentId: "9223372035264712524/1590063283",
				ActionLog: pb.ActionLog{
					Instance: "test",
					Action:   "test",
					Level:    "INFO",
				},
			}
		)
		Convey("When everything goes positive", func() {
			d.AddActionLog(ctx, req)
			Convey("No return values", func() {
			})
		})
	})
}

func TestDaoListDeployment(t *testing.T) {
	Convey("ListDeployment", t, func() {
		var (
			ctx = context.Background()
			req = &pb.ListDeploymentReq{
				Node:    "main.web-svr",
				Gateway: "playlist-gateway",
			}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.ListDeployment(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}
