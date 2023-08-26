package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	plugin2 "go-gateway/app/app-svr/app-resource/interface/model/plugin"
)

func plugin(c *bm.Context) {
	var (
		params   = c.Request.Form
		build    int
		baseCode int
		seed     int
		err      error
	)
	name := params.Get("name")
	buildStr := params.Get("build")
	baseCodeStr := params.Get("base_code")
	seedStr := params.Get("seed")
	if build, err = strconv.Atoi(buildStr); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if seed, err = strconv.Atoi(seedStr); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if baseCode, err = strconv.Atoi(baseCodeStr); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pg := pgSvr.Plugin(build, baseCode, seed, name); pg != nil {
		c.JSON(pg, nil)
	} else {
		c.JSON(struct{}{}, nil)
	}

}

func serviceDependencies(ctx *bm.Context) {
	req := &plugin2.TraceParam{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Cookie = ctx.Request.Header.Get("Cookie")
	ctx.JSON(pgSvr.Dependencies(ctx, req))
}
