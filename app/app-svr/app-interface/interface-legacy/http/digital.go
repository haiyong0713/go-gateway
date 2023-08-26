package http

import (
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

func digitalInfo(c *bm.Context) {
	var (
		mid   int64
		vmid  int64
		nftID string
		err   error
	)
	params := c.Request.Form
	if nftID = params.Get("nft_id"); nftID == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if vmid, err = strconv.ParseInt(params.Get("vmid"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spaceSvr.DigitalInfo(c, mid, vmid, nftID))
}

func digitalBind(c *bm.Context) {
	param := &struct {
		Mid    int64  `form:"mid"`
		ItemID int64  `form:"item_id"`
		NftID  string `form:"nft_id"`
	}{}
	if err := c.Bind(param); err != nil {
		log.Error("digitalBind request err:%+v", err)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	err := spaceSvr.DigitalBind(c, param.Mid, param.ItemID, param.NftID)
	c.JSON(nil, err)
}

func digitalUnbind(c *bm.Context) {
	param := &struct {
		Mid int64 `form:"mid"`
	}{}
	if err := c.Bind(param); err != nil {
		log.Error("digitalUnbind request err:%+v", err)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	err := spaceSvr.DigitalUnbind(c, param.Mid)
	c.JSON(nil, err)
}

func digitalExtraInfo(c *bm.Context) {
	var (
		mid   int64
		nftID string
	)
	params := c.Request.Form
	if nftID = params.Get("nft_id"); nftID == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spaceSvr.DigitalExtraInfo(c, mid, nftID))
}
