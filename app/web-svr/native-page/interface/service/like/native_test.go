package like

import (
	"context"
	"encoding/json"
	"go-gateway/app/web-svr/native-page/interface/api"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_AutoDispense(t *testing.T) {
	Convey("AutoDispense", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg = ""
		)
		err := svf.AutoDispense(c, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_PageUpdate(t *testing.T) {
	Convey("PageUpdate", t, WithService(func(svf *Service) {
		var (
			c        = context.Background()
			msg, old json.RawMessage
		)
		err := svf.PageUpdate(c, msg, old, "del")
		So(err, ShouldBeNil)
	}))
}

func TestService_LikePageDel(t *testing.T) {
	Convey("LikePageDel", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.PageDel(c, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_ModuleUpdate(t *testing.T) {
	Convey("LikePageDel", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.ModuleUpdate(c, msg, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_ModuleDel(t *testing.T) {
	Convey("ModuleDel", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.ModuleUpdate(c, msg, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_NatClickUpdate(t *testing.T) {
	Convey("NatClickUpdate", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.NatClickUpdate(c, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_NatClickDel(t *testing.T) {
	Convey("NatClickDel", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.NatClickDel(c, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_NatActUpdate(t *testing.T) {
	Convey("NatActUpdate", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.NatActUpdate(c, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_NatActDel(t *testing.T) {
	Convey("NatActDel", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.NatActDel(c, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_NatDynamicUpdate(t *testing.T) {
	Convey("NatDynamicUpdate", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.NatDynamicUpdate(c, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_NatDynamicDel(t *testing.T) {
	Convey("NatDynamicDel", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.NatDynamicDel(c, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_VideoUpdate(t *testing.T) {
	Convey("NatVideoUpdate", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg json.RawMessage
		)
		err := svf.NatVideoUpdate(c, msg)
		So(err, ShouldBeNil)
	}))
}

func TestService_NatConfig(t *testing.T) {
	Convey("NatConfig", t, WithService(func(svf *Service) {
		var (
			c = context.Background()
		)
		_, err := svf.NatConfig(c, &api.NatConfigReq{Pid: 15})
		So(err, ShouldBeNil)
	}))
}

func TestService_NatMixtureUpdate(t *testing.T) {
	Convey("NatMixtureUpdate", t, WithService(func(svf *Service) {
		var (
			c   = context.Background()
			msg = json.RawMessage(`{"id":1,"module_id":1,"m_type":1}`)
		)
		err := svf.NatMixtureUpdate(c, msg)
		So(err, ShouldBeNil)
	}))
}
