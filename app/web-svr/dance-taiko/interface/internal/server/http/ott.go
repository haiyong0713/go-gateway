package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	xecode "go-gateway/app/web-svr/dance-taiko/interface/ecode"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

const (
	_preRank  = 10
	_dancebiz = "2"
)

func loadRank(c *bm.Context) {
	req := new(struct {
		Cid       int64  `form:"cid" validate:"required"`
		AccessKey string `form:"access_key"`
		Pn        int    `form:"pn"`
		Ps        int    `form:"ps"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	mid, _ := midFromCtx(c)
	var ps = req.Ps
	if req.Ps == 0 {
		ps = _preRank
	}
	c.JSON(svc.GameRanks(c, req.Cid, mid, req.Pn, ps))
}

func midFromCtx(ctx *bm.Context) (int64, error) {
	midIface, ok := ctx.Get("mid")
	if !ok {
		return 0, ecode.NoLogin
	}
	mid, ok := midIface.(int64)
	if !ok {
		return 0, ecode.NoLogin
	}
	return mid, nil
}

func loadKeyFrames(c *bm.Context) {
	req := new(struct {
		Aid  int64  `form:"aid" validate:"required"`
		Cid  int64  `form:"cid" validate:"required"`
		Plat string `form:"plat" validate:"required"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.KeyFrames(c, req.Cid, req.Plat))
}

func gameCreate(c *bm.Context) {
	req := new(struct {
		Aid   int64  `form:"aid" validate:"required"`
		Cid   int64  `form:"cid" validate:"required"`
		Buvid string `form:"buvid"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.GameCreate(c, req.Aid, req.Cid))
}

func gameStart(c *bm.Context) {
	req := new(struct {
		GameId int64 `form:"game_id" validate:"required"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.GameStart(c, req.GameId))
}

func gameJoin(c *bm.Context) {
	req := new(struct {
		GameId    int64  `form:"game_id" validate:"required"`
		AccessKey string `form:"access_key" validate:"required"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	mid, _ := midFromCtx(c)
	c.JSON(svc.GameJoin(c, req.GameId, mid))
}

func gameFinish(c *bm.Context) {
	req := new(struct {
		GameId int64 `form:"game_id" validate:"required"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(nil, svc.GameFinish(c, req.GameId))
}

func gameStatus(c *bm.Context) {
	req := new(struct {
		GameId      int64 `form:"game_id" validate:"required"`
		PlayTime    int64 `form:"play_time"`
		NaturalTime int64 `form:"natural_time"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.GameStatus(c, req.GameId, req.PlayTime, req.NaturalTime))
}

func ottGameStat(c *bm.Context) {
	v := new(struct {
		GameID int64  `form:"game_id" validate:"required"`
		Stats  string `form:"stats" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}

	header := c.Request.Header
	buvid := header.Get("Buvid")

	sts := make([]*model.Stat, 0)
	if err := json.Unmarshal([]byte(v.Stats), &sts); err != nil {
		c.JSON(nil, xecode.JsonFormatErr)
		return
	}
	if len(sts) == 0 {
		log.Error("GameStat GameID %d Stats Empty", v.GameID)
		c.JSON(nil, xecode.JsonFormatErr)
		return
	}

	mid, _ := c.Get("mid")
	c.JSON(nil, svc.OttGameStat(c, buvid, v.GameID, mid.(int64), sts))
}

func pkgUpload(c *bm.Context) {
	var (
		fileType string
		body     []byte
		file     multipart.File
		err      error
	)
	if c.Request.FormValue("biz_id") != _dancebiz {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if file, _, err = c.Request.FormFile("file"); err != nil {
		c.JSON(nil, ecode.RequestErr)
		log.Error("c.Request.FormFile(\"file\") error(%v)", err)
		return
	}
	defer file.Close()
	if body, err = ioutil.ReadAll(file); err != nil {
		c.JSON(nil, ecode.RequestErr)
		log.Error("ioutil.ReadAll(c.Request.Body) error(%v)", err)
		return
	}
	fileType = http.DetectContentType(body)
	c.JSON(nil, svc.GamePkgUpload(c, fileType, bytes.NewReader(body)))
}

func loadQRCode(c *bm.Context) {
	req := new(struct {
		GameId int64 `form:"game_id" validate:"required"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.LoadQRCode(c, req.GameId))
}

func gameRestart(c *bm.Context) {
	req := new(struct {
		GameId int64 `form:"game_id" validate:"required"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.GameRestart(c, req.GameId))
}
