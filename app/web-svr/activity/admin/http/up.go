package http

import bm "go-common/library/net/http/blademaster"

func upActlist(c *bm.Context) {
	v := new(struct {
		Uid   int64 `form:"uid"`
		State int64 `form:"state" default:"0"`
		Pn    int64 `form:"pn" default:"1" validate:"min=1"`
		Ps    int64 `form:"ps" default:"20" validate:"min=1,max=50"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, count, err := actSrv.UpActList(c, v.Uid, v.State, v.Pn, v.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": count,
	}
	data["list"] = list
	data["page"] = page
	c.JSON(data, nil)
}

func upActEdit(c *bm.Context) {
	v := new(struct {
		ID    int64 `form:"id"`
		Uid   int64 `form:"uid"`
		State int64 `form:"state" validate:"required"`
		IsBig int   `form:"is_big"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, actSrv.UpActEdit(c, v.ID, v.Uid, v.State, v.IsBig))
}

func upActOffline(c *bm.Context) {
	v := new(struct {
		ID      int64 `form:"id"`
		Offline int64 `form:"offline" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, actSrv.UpActOffline(c, v.ID, v.Offline))
}
