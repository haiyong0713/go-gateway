package http

import (
	_ "embed"
	"fmt"
	"io/ioutil"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"
)

func botVerify(ctx *bm.Context) {
	req := new(model.BotVerifyReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	rst, err := svc.BotVerify(ctx, req)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.Bytes(200, "text/plain", []byte(rst))
}

func botCallback(ctx *bm.Context) {
	req := new(model.BotCallbackReq)
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer ctx.Request.Body.Close()
	if err = ctx.Bind(req); err != nil {
		return
	}
	rst, err := svc.BotCallback(ctx, req, data)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.Bytes(200, "text/plain", []byte(rst))
}

//go:embed static/deploy.html
var deploy string

func deployment(ctx *bm.Context) {
	req := new(struct {
		Jwt string `form:"jwt"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}
	session, err := ctx.Request.Cookie("_AJSESSIONID")
	if err != nil {
		ctx.Redirect(302, "https://dashboard-mng.biliapi.net/api/v4/user/dashboard_login?caller=gateway-dev-mgt&path=%2Fx%2Fadmin%2Fgateway-dev-management%2Fbot%2Fdeploy"+fmt.Sprintf("?jwt=%v", req.Jwt))
		return
	}
	if _, err = svc.DashboardVerify(ctx, session.Value); err != nil {
		ctx.Redirect(302, "https://dashboard-mng.biliapi.net/api/v4/user/dashboard_login?caller=gateway-dev-mgt&path=%2Fx%2Fadmin%2Fgateway-dev-management%2Fbot%2Fdeploy"+fmt.Sprintf("?jwt=%v", req.Jwt))
		return
	}
	ctx.Bytes(200, "text/html", []byte(deploy))
}

func getDeploy(ctx *bm.Context) {
	req := new(struct {
		Jwt string `form:"jwt"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	reply, err := svc.GetDeploy(ctx, req.Jwt, cookie)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.JSON(reply, nil)
}

func startDeploy(ctx *bm.Context) {
	req := new(struct {
		DeployId int64 `form:"deployId"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	reply, err := svc.StartDeploy(ctx, req.DeployId, cookie)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.JSON(reply, nil)
}

func resumeDeploy(ctx *bm.Context) {
	req := new(struct {
		DeployId int64 `form:"deployId"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	reply, err := svc.ResumeDeploy(ctx, req.DeployId, cookie)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.JSON(reply, nil)
}

func doneDeploy(ctx *bm.Context) {
	req := new(struct {
		DeployId int64 `form:"deployId"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	reply, err := svc.DoneDeploy(ctx, req.DeployId, cookie)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.JSON(reply, nil)
}

func rollbackDeploy(ctx *bm.Context) {
	req := new(struct {
		DeployId int64 `form:"deployId"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	reply, err := svc.RollbackDeploy(ctx, req.DeployId, cookie)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.JSON(reply, nil)
}

//go:embed static/script.html
var script string

func scripts(ctx *bm.Context) {
	session, err := ctx.Request.Cookie("_AJSESSIONID")
	if err != nil {
		ctx.Redirect(302, "https://dashboard-mng.biliapi.net/api/v4/user/dashboard_login?caller=gateway-dev-mgt&path=%2Fx%2Fadmin%2Fgateway-dev-management%2Fbot%2Fscript")
		return
	}
	if _, err = svc.DashboardVerify(ctx, session.Value); err != nil {
		ctx.Redirect(302, "https://dashboard-mng.biliapi.net/api/v4/user/dashboard_login?caller=gateway-dev-mgt&path=%2Fx%2Fadmin%2Fgateway-dev-management%2Fbot%2Fscript")
		return
	}
	ctx.Bytes(200, "text/html", []byte(script))
}

func newScript(ctx *bm.Context) {
	req := &model.NewScriptReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	err := svc.NewScript(ctx, req)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.JSON("", nil)
}

func getScript(ctx *bm.Context) {
	session, err := ctx.Request.Cookie("username")
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	reply, err := svc.GetScript(ctx, session.Value)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.JSON(reply, nil)
}
func doScript(ctx *bm.Context) {
	req := new(struct {
		Id string `form:"id"`
	})
	defer ctx.Request.Body.Close()
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	rst, err := svc.DoScript(ctx, req.Id, cookie)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.JSON(rst, nil)
}
