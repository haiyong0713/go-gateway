package http

import (
	"fmt"
	"time"

	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/ecode"
	"go-gateway/app/web-svr/esports/interface/model"
	favEcode "go-main/app/community/favorite/service/ecode"
)

func addFav(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	v := new(struct {
		Cid int64 `form:"cid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, switchCode(eSvc.AddFav(c, mid, v.Cid)))
}

func batchAddFav(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}

	v := new(struct {
		IDList string `form:"id_list" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}

	d, err := eSvc.BatchAddFav(c, mid, v.IDList)
	if err != nil {
		c.JSON(nil, switchCode(err))
	} else {
		c.JSON(d, nil)
	}
}

func batchQueryFav(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}

	v := new(struct {
		IDList string `form:"id_list" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		fmt.Println("batchQueryFav >>> bind", err)
		return
	}

	d, err := eSvc.BatchQueryFav(c, mid, v.IDList)
	if err != nil {
		c.JSON(nil, switchCode(err))
	} else {
		c.JSON(d, nil)
	}
}

func delFav(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	v := new(struct {
		Cid int64 `form:"cid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, switchCode(eSvc.DelFav(c, mid, v.Cid)))
}

func listFav(c *bm.Context) {
	var (
		mid     int64
		total   int
		contest []*model.Contest
		err     error
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	v := new(struct {
		VMID int64 `form:"vmid"`
		Pn   int   `form:"pn" default:"1" validate:"min=1"`
		Ps   int   `form:"ps" default:"5" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if contest, total, err = eSvc.ListFav(c, mid, v.VMID, v.Pn, v.Ps); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = contest
	c.JSON(data, nil)
}
func appListFav(c *bm.Context) {
	var (
		mid     int64
		total   int
		contest []*model.Contest
		err     error
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	v := new(model.ParamFav)
	if err = c.Bind(v); err != nil {
		return
	}
	if mid == 0 && v.VMID == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if v.Stime != "" {
		if _, err = time.Parse("2006-01-02", v.Stime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if v.Etime != "" {
		if _, err = time.Parse("2006-01-02", v.Etime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if contest, total, err = eSvc.ListAppFav(c, mid, v); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = contest
	c.JSON(data, nil)
}

func seasonFav(c *bm.Context) {
	var (
		mid     int64
		total   int
		seasons []*model.Season
		err     error
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	v := new(model.ParamSeason)
	if err = c.Bind(v); err != nil {
		return
	}
	if mid == 0 && v.VMID == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if seasons, total, err = eSvc.SeasonFav(c, mid, v); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = seasons
	c.JSON(data, nil)
}

func stimeFav(c *bm.Context) {
	var (
		mid    int64
		total  int
		stimes []string
		err    error
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	v := new(model.ParamSeason)
	if err = c.Bind(v); err != nil {
		return
	}
	if mid == 0 && v.VMID == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if stimes, total, err = eSvc.StimeFav(c, mid, v); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = stimes
	c.JSON(data, nil)
}

func switchCode(err error) error {
	if err == nil {
		return err
	}
	switch xecode.Cause(err) {
	case favEcode.FavResourceOverflow:
		err = ecode.EsportsContestMaxCount
	case favEcode.FavResourceAlreadyDel:
		err = ecode.EsportsContestFavDel
	case favEcode.FavResourceExist:
		err = ecode.EsportsContestFavExist
	case favEcode.FavFolderNotExist:
		err = ecode.EsportsContestNotExist
	}
	return err
}
