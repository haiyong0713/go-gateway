package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"
)

func modifyPage(c *bm.Context) {
	var (
		err error
		arg = &natmdl.ModifyParam{}
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, natSrv.ModifyPage(c, arg))
}

func reOnline(c *bm.Context) {
	var (
		err error
		arg = &natmdl.OnlineParam{}
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, natSrv.ReOnline(c, arg))
}

func addPage(c *bm.Context) {
	var (
		err error
		arg = &natmdl.AddPageParam{}
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(natSrv.PageSave(c, arg))
}

func delPage(c *bm.Context) {
	var (
		err error
		arg = new(struct {
			ID        int64  `form:"id" validate:"required"`
			UserName  string `form:"user_name" validate:"required"`
			OffReason string `form:"off_reason"`
		})
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, natSrv.DelPage(c, arg.ID, arg.UserName, arg.OffReason))
}

func editPage(c *bm.Context) {
	arg := &natmdl.EditParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	pType := 0
	if arg.SkipUrl == "" {
		c.JSON(nil, natSrv.EditPage(c, arg, pType))
	} else {
		c.JSON(nil, natSrv.PageSkipUrl(c, arg, pType))
	}
}

func searchPage(c *bm.Context) {
	arg := &natmdl.SearchParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(natSrv.SearchPage(c, arg))
}

func upPage(c *bm.Context) {
	arg := &natmdl.UpParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(natSrv.UpPage(c, arg))
}

func findPage(c *bm.Context) {
	var (
		err error
		arg = new(struct {
			ID    int64  `form:"id"`
			Title string `form:"title"`
		})
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	if arg.Title == "" && arg.ID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.FindPage(c, arg.Title, arg.ID))
}

func searchModule(c *bm.Context) {
	arg := &natmdl.SearchModule{}
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.ID == 0 && arg.ModuleID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.SearchModule(c, arg))
}

func saveModule(c *bm.Context) {
	arg := &natmdl.ModuleParam{}
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.Data == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.Module(c, arg))
}

func saveTab(c *bm.Context) {
	var (
		err error
		req = &natmdl.SaveTabReq{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}
	loginUser, exist := c.Get("username")
	if !exist {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(natSrv.SaveTab(c, req, loginUser.(string)))
}

func editTab(c *bm.Context) {
	var (
		err error
		req = new(struct {
			ID    int32 `form:"id" validate:"required"`
			Stime int64 `form:"stime" validate:"min=0"`
			Etime int64 `form:"etime" validate:"min=0"`
		})
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, natSrv.EditTab(c, req.ID, req.Stime, req.Etime))
}

func tabList(c *bm.Context) {
	var (
		err error
		req = &natmdl.SearchTabReq{}
	)
	if err = c.Bind(req); err != nil {
		return
	}
	c.JSON(natSrv.SearchTab(c, req))
}

func tsOnline(c *bm.Context) {
	req := &natmdl.TsOnlineReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(nil, natSrv.TsOnline(c, req))
}

func pageTab(c *bm.Context) {
	var (
		err error
		req = new(struct {
			PId int32 `form:"pid" validate:"required"`
		})
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.GetTabOfPage(c, req.PId))
}

func findCounters(c *bm.Context) {
	var (
		err error
		req = new(struct {
			Activity string `form:"activity" validate:"required"`
		})
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.FindCounters(c, req.Activity))
}

func spaceOffline(c *bm.Context) {
	req := &natmdl.SpaceOfflineReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(nil, natSrv.SpaceOffline(c, req))
}

func topicUpgrade(c *bm.Context) {
	c.JSON(nil, nil)
}

func gameDetail(c *bm.Context) {
	var (
		err error
		req = new(struct {
			GameID int64 `form:"game_id"  validate:"min=1" `
		})
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.GameDetail(c, req.GameID))
}

func cartoonDetail(c *bm.Context) {
	var (
		err error
		req = new(struct {
			ID int64 `form:"id"  validate:"min=1" `
		})
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.CartoonDetail(c, req.ID))
}

func channelDetail(c *bm.Context) {
	var (
		err error
		req = new(struct {
			ID int64 `form:"id"  validate:"min=1"`
			Ps int32 `form:"ps" default:"15" validate:"min=1"`
		})
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var buvid string
	if res, err := c.Request.Cookie("Buvid"); err == nil {
		buvid = res.Value
	} else {
		if res, err := c.Request.Cookie("buvid3"); err == nil {
			buvid = res.Value
		}
	}
	c.JSON(natSrv.ChannelDetail(c, req.ID, req.Ps, buvid))
}

func reserveDetail(c *bm.Context) {
	var (
		err error
		req = new(struct {
			SID int64 `form:"sid"  validate:"min=1" `
		})
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.ReserveDetail(c, req.SID))
}

func tsPage(c *bm.Context) {
	req := &natmdl.TsPageReq{}
	if err := c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.TsPage(c, req))
}

func upVote(c *bm.Context) {
	req := new(struct {
		VoteID int64 `form:"vote_id" validate:"min=1"`
	})
	if err := c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(natSrv.UpVote(c, req.VoteID))
}

func addNewact(c *bm.Context) {
	req := &natmdl.AddNewactReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(natSrv.AddNewact(c, req))
}
