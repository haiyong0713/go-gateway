package http

import (
	"go-common/library/log"
	"strconv"
	"strings"

	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/space/ecode"
	"go-gateway/app/web-svr/space/interface/conf"
	"go-gateway/app/web-svr/space/interface/model"
)

func channel(c *bm.Context) {
	var (
		vmid, mid, cid int64
		isGuest        bool
		err            error
	)
	params := c.Request.Form
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	vmidStr := params.Get("mid")
	cidStr := params.Get("cid")
	guestStr := params.Get("guest")
	if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil || (vmid <= 0 && mid <= 0) {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if cid, err = strconv.ParseInt(cidStr, 10, 64); err != nil || cid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if guestStr != "" {
		if isGuest, err = strconv.ParseBool(guestStr); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if !isGuest && vmid > 0 && mid != vmid {
		mid = vmid
	}
	if mid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(spcSvc.Channel(c, mid, cid))
}

func channelIndex(c *bm.Context) {
	var (
		vmid, mid int64
		isGuest   bool
		err       error
	)
	params := c.Request.Form
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	vmidStr := params.Get("mid")
	guestStr := params.Get("guest")
	if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil || (vmid <= 0 && mid <= 0) {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if guestStr != "" {
		if isGuest, err = strconv.ParseBool(guestStr); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if !isGuest && vmid > 0 && mid != vmid {
		isGuest = true
		mid = vmid
	}
	if mid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(spcSvc.ChannelIndex(c, mid, isGuest))
}

func channelList(c *bm.Context) {
	var (
		vmid, mid int64
		channels  []*model.Channel
		isGuest   bool
		err       error
	)
	params := c.Request.Form
	vmidStr := params.Get("mid")
	guestStr := params.Get("guest")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil || (vmid <= 0 && mid <= 0) {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if guestStr != "" {
		if isGuest, err = strconv.ParseBool(guestStr); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if !isGuest && vmid > 0 && mid != vmid {
		isGuest = true
		mid = vmid
	}
	if mid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if channels, err = spcSvc.ChannelList(c, mid, isGuest); err != nil {
		c.JSON(nil, err)
		return
	}
	data := map[string]interface{}{}
	data["count"] = len(channels)
	data["list"] = channels
	c.JSON(data, nil)
}

func addChannel(c *bm.Context) {
	var (
		mid, cid int64
		err      error
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	name := params.Get("name")
	intro := params.Get("intro")
	if name == "" || len([]rune(name)) > conf.Conf.Rule.MaxChNameLen {
		c.JSON(nil, ecode.ChNameToLong)
		return
	}
	if intro != "" && len([]rune(intro)) > conf.Conf.Rule.MaxChIntroLen {
		c.JSON(nil, ecode.ChIntroToLong)
		return
	}
	if cid, err = spcSvc.AddChannel(c, mid, name, intro); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(struct {
		Cid int64 `json:"cid"`
	}{Cid: cid}, nil)
}

func editChannel(c *bm.Context) {
	var (
		mid, cid int64
		err      error
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	cidStr := params.Get("cid")
	name := params.Get("name")
	intro := params.Get("intro")
	if cid, err = strconv.ParseInt(cidStr, 10, 64); err != nil || cid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if name == "" {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if len([]rune(name)) > conf.Conf.Rule.MaxChNameLen {
		c.JSON(nil, ecode.ChNameToLong)
		return
	}
	if intro != "" && len([]rune(intro)) > conf.Conf.Rule.MaxChIntroLen {
		c.JSON(nil, ecode.ChIntroToLong)
		return
	}
	c.JSON(nil, spcSvc.EditChannel(c, mid, cid, name, intro))
}

func delChannel(c *bm.Context) {
	var (
		mid, cid int64
		err      error
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	cidStr := params.Get("cid")
	if cid, err = strconv.ParseInt(cidStr, 10, 64); err != nil || cid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, spcSvc.DelChannel(c, mid, cid))
}

func channelAids(ctx *bm.Context) {
	v := new(struct {
		ChannelID int64  `form:"channel_id" validate:"min=1"` // 频道ID，值的算法为cid * 10 + vmid % 10
		Sort      string `form:"sort"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	aids, err := spcSvc.ChannelAids(ctx, v.ChannelID, v.Sort)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["aids"] = aids
	ctx.JSON(res, nil)
}

func channelDetail(ctx *bm.Context) {
	v := new(struct {
		ChannelID int64 `form:"channel_id" validate:"min=1"` // 频道ID，值的算法为cid * 10 + vmid % 10
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	detail, err := spcSvc.ChannelDetail(ctx, v.ChannelID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["detail"] = detail
	ctx.JSON(res, nil)
}

func channelVideo(c *bm.Context) {
	var (
		vmid, mid, cid int64
		pn, ps         int
		isGuest, order bool
		err            error
	)
	params := c.Request.Form
	vmidStr := params.Get("mid")
	cidStr := params.Get("cid")
	pnStr := params.Get("pn")
	psStr := params.Get("ps")
	guestStr := params.Get("guest")
	orderStr := params.Get("order")
	ctypeStr := params.Get("ctype")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil || (vmid <= 0 && mid <= 0) {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if cid, err = strconv.ParseInt(cidStr, 10, 64); err != nil || cid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(pnStr); err != nil || pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(psStr); err != nil || ps < 1 || ps > conf.Conf.Rule.MaxChArcsPs {
		ps = conf.Conf.Rule.MaxChArcsPs
	}
	if guestStr != "" {
		if isGuest, err = strconv.ParseBool(guestStr); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if !isGuest && vmid > 0 && mid != vmid {
		isGuest = true
		mid = vmid
	}
	if mid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if orderStr != "" {
		if order, err = strconv.ParseBool(orderStr); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	ctype, _ := strconv.ParseInt(ctypeStr, 10, 64)
	channelDetail, button, err := spcSvc.ChannelVideos(c, mid, cid, pn, ps, isGuest, order, ctype)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   pn,
		"size":  ps,
		"count": channelDetail.Count,
	}
	data["page"] = page
	data["list"] = channelDetail
	if button != nil {
		data["episodic_button"] = button
	}
	c.JSON(data, nil)
}

func addChannelVideo(c *bm.Context) {
	var (
		aid  int64
		aids []int64
		err  error
	)
	v := new(struct {
		Cid   int64  `form:"cid" validate:"min=1"`
		Aids  string `form:"aids"`
		Bvids string `form:"bvids"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	if v.Aids == "" && v.Bvids == "" {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if v.Bvids != "" {
		bvids := strings.Split(v.Bvids, ",")
		if len(bvids) == 0 || len(bvids) > conf.Conf.Rule.MaxChArcAddLimit {
			c.JSON(nil, xecode.RequestErr)
			return
		}
		for _, bvid := range bvids {
			if aid, err = bvArgCheck(0, bvid); err != nil {
				c.JSON(nil, err)
				return
			}
			aids = append(aids, aid)
		}
	} else {
		if aids, err = xstr.SplitInts(v.Aids); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if l := len(aids); l == 0 || l > conf.Conf.Rule.MaxChArcAddLimit {
		log.Warn("len(aids)(%d) warn", l)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	aidMap := make(map[int64]int64, len(aids))
	for _, aid := range aids {
		aidMap[aid] = aid
	}
	if len(aidMap) < len(aids) {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(spcSvc.AddChannelArc(c, mid, v.Cid, aids))
}

func delChannelVideo(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
		Cid  int64  `form:"cid" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.DelChannelArc(c, mid, v.Cid, aid))
}

func sortChannelVideo(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
		Cid  int64  `form:"cid" validate:"min=1"`
		To   int    `form:"to" validate:"min=0"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.SortChannelArc(c, mid, v.Cid, aid, v.To))
}

func checkChannelVideo(c *bm.Context) {
	var (
		mid, cid int64
		err      error
	)
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	cidStr := c.Request.Form.Get("cid")
	if cid, err = strconv.ParseInt(cidStr, 10, 64); err != nil || cid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, spcSvc.CheckChannelVideo(c, mid, cid))
}
