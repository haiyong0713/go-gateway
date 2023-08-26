package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"
)

func noteAdd(c *bm.Context) {
	req := new(note.NoteAddReq)
	err := c.Bind(req)
	if err != nil {
		return
	}
	// TODO 废弃aid
	if req.Aid > 0 {
		req.Oid = req.Aid
	}
	oid, err := bvArgCheck(req.Oid, req.Bvid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	req.Oid = oid
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(noteSvr.NoteAdd(c, req, mid.(int64)))
}

func noteInfo(c *bm.Context) {
	req := new(note.NoteInfoReq)
	err := c.Bind(req)
	if err != nil {
		return
	}
	// TODO 废弃aid
	if req.Aid > 0 {
		req.Oid = req.Aid
	}
	oid, err := bvArgCheck(req.Oid, req.Bvid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	req.Oid = oid
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(noteSvr.NoteInfo(c, req, mid.(int64)))
}

func noteListArc(c *bm.Context) {
	req := new(struct {
		Bvid    string `form:"bvid"`
		Aid     int64  `form:"aid"`
		Oid     int64  `form:"oid"`
		OidType int    `form:"oid_type" validate:"lt=2"`
	})
	err := c.Bind(req)
	if err != nil {
		return
	}
	// TODO 废弃aid
	if req.Aid > 0 {
		req.Oid = req.Aid
	}
	oid, err := bvArgCheck(req.Oid, req.Bvid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	req.Oid = oid
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(noteSvr.NoteListArc(c, req.Oid, mid.(int64), req.OidType))
}

func noteList(c *bm.Context) {
	req := new(struct {
		Pn int64 `form:"pn" validate:"required"`
		Ps int64 `form:"ps" validate:"required"`
	})
	err := c.Bind(req)
	if err != nil {
		return
	}
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(noteSvr.NoteList(c, mid.(int64), req.Pn, req.Ps))
}

func noteDel(c *bm.Context) {
	req := new(note.NoteDelReq)
	err := c.Bind(req)
	if err != nil {
		return
	}
	if len(req.NoteIds) == 0 && req.NoteId == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if req.NoteId > 0 {
		req.NoteIds = []int64{req.NoteId}
	}
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, noteSvr.NoteDel(c, req, mid.(int64)))
}

func isGray(c *bm.Context) {
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(noteSvr.IsGray(c, mid.(int64)))
}

func links(c *bm.Context) {
	c.JSON(noteSvr.Links(c))
}

func noteCount(c *bm.Context) {
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(noteSvr.NoteCount(c, mid.(int64)))
}

func isForbid(c *bm.Context) {
	req := new(struct {
		Aid int64 `form:"aid" validate:"required"`
	})
	err := c.Bind(req)
	if err != nil {
		return
	}
	c.JSON(noteSvr.IsForbid(c, req.Aid))
}
