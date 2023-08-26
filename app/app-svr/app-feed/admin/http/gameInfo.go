package http

import "go-common/library/ecode"

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/game"
)

func gameInfoApp(c *bm.Context) {
	var (
		err  error
		res  = map[string]interface{}{}
		data *game.Info
	)
	type GameReq struct {
		ID       int64 `form:"id" validate:"required"`
		PlatFrom int   `form:"platfrom"`
	}
	card := &GameReq{}
	if err = c.Bind(card); err != nil {
		return
	}
	if data, err = commonSvc.Game(c, card.ID, card.PlatFrom); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(data, nil)
}
