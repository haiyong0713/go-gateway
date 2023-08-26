package http

import (
	"encoding/json"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func saveGraph(c *bm.Context) {
	v := new(struct {
		Preview int    `form:"preview"`
		Data    string `form:"data" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	param := new(model.SaveGraphParam)
	if err := json.Unmarshal([]byte(v.Data), &param); err != nil || param.Graph == nil {
		log.Warn("saveGraph json.Unmarshal(%s) error(%v)", v.Data, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	if v.Preview != model.GraphIsPreview {
		v.Preview = 0
	}
	if graphID, errInfo, err := svc.SaveGraph(c, mid, v.Preview, param); err != nil {
		c.JSONMap(map[string]interface{}{"err_type": errInfo.ErrType, "err_id": errInfo.ErrId}, err)
	} else {
		c.JSON(map[string]interface{}{"graph_id": graphID}, nil)
	}
}

func latestGraphList(c *bm.Context) {
	v := new(struct {
		Aid int64 `form:"aid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(svc.LatestGraphList(c, mid, v.Aid))
}

func msgCheck(c *bm.Context) {
	v := new(struct {
		Msg string `form:"msg" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, svc.MsgCheck(c, v.Msg))
}

func playurl(c *bm.Context) {
	v := new(model.PlayurlParam)
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(svc.Playurl(c, mid, v))
}

func graphShow(c *bm.Context) {
	v := new(struct {
		Aid     int64 `form:"aid" validate:"min=1"`
		GraphID int64 `form:"graph_id"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(svc.GraphShow(c, mid, v.Aid, v.GraphID))
}

func graphCheck(c *bm.Context) {
	v := new(struct {
		Aid int64 `form:"aid" validate:"min=1"`
		Cid int64 `form:"cid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(svc.GraphCheck(c, v.Aid, v.Cid))
}

func managerGraph(c *bm.Context) {
	v := new(struct {
		Aid int64 `form:"aid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(svc.ManagerList(c, v.Aid))
}

func videoInfo(c *bm.Context) {
	v := new(model.VideoInfoParam)
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(svc.VideoInfo(c, v, mid))
}

func recentArcs(c *bm.Context) {
	v := new(model.RecentArcReq)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(svc.RecentArcs(c, v))
}

func skinList(c *bm.Context) {
	c.JSON(svc.SkinList(c), nil)

}
