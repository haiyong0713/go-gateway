package http

import (
	"strconv"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favEcode "git.bilibili.co/bapis/bapis-go/community/service/favorite/ecode"
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/playlist/ecode"
	"go-gateway/app/web-svr/playlist/interface/conf"
	"go-gateway/app/web-svr/playlist/interface/model"
)

func whiteList(c *bm.Context) {
	var (
		err       error
		vmid, mid int64
	)
	params := c.Request.Form
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	vmidStr := params.Get("vmid")
	if vmidStr != "" {
		if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if mid == 0 && vmid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if vmid > 0 {
		mid = vmid
	}
	c.JSON(plSvc.White(c, mid))
}

func add(c *bm.Context) {
	var (
		err                      error
		mid, pid, public         int64
		name, description, cover string
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	publicStr := params.Get("public")
	if publicStr != "" {
		if public, err = strconv.ParseInt(publicStr, 10, 64); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	name = params.Get("name")
	if name == "" || len([]rune(name)) > conf.Conf.Rule.MaxNameLimit {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	description = params.Get("description")
	if description != "" {
		if len([]rune(description)) > conf.Conf.Rule.MaxPlDescLimit {
			c.JSON(nil, ecode.PlDescTooLong)
			return
		}
	}
	cover = params.Get("cover")
	if pid, err = plSvc.Add(c, mid, int32(public), name, description, cover, c.Request.Header.Get("Cookie"), params.Get("access_key")); err != nil {
		c.JSON(nil, switchCode(err, favmdl.TypePlayVideo))
		return
	}
	data := make(map[string]interface{}, 1)
	data["pid"] = pid
	c.JSON(data, nil)
}

func del(c *bm.Context) {
	var (
		err      error
		mid, pid int64
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	pidStr := params.Get("pid")
	if pid, err = strconv.ParseInt(pidStr, 10, 64); err != nil || pid < 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, switchCode(plSvc.Del(c, mid, pid), favmdl.TypePlayVideo))
}

func update(c *bm.Context) {
	var (
		err                      error
		mid, pid, public         int64
		name, description, cover string
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	publicStr := params.Get("public")
	if publicStr != "" {
		if public, err = strconv.ParseInt(publicStr, 10, 64); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	pidStr := params.Get("pid")
	if pid, err = strconv.ParseInt(pidStr, 10, 64); err != nil || pid < 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	name = params.Get("name")
	if name == "" || len([]rune(name)) > conf.Conf.Rule.MaxNameLimit {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	description = params.Get("description")
	if description != "" {
		if len([]rune(description)) > conf.Conf.Rule.MaxPlDescLimit {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	cover = params.Get("cover")
	c.JSON(nil, switchCode(plSvc.Update(c, mid, pid, int32(public), name, description, cover, c.Request.Header.Get("Cookie"), params.Get("access_key")), favmdl.TypePlayVideo))
}

func info(c *bm.Context) {
	var (
		err      error
		pid, mid int64
		list     *model.Playlist
	)
	params := c.Request.Form
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	pidStr := params.Get("pid")
	if pid, err = strconv.ParseInt(pidStr, 10, 64); err != nil || pid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if list, err = plSvc.Info(c, mid, pid); err != nil {
		c.JSON(nil, switchCode(err, favmdl.TypePlayVideo))
		return
	}
	c.JSON(list, nil)
}

func report(c *bm.Context) {
	var (
		err      error
		pid, aid int64
	)
	params := c.Request.Form
	pidStr := params.Get("pid")
	if pid, err = strconv.ParseInt(pidStr, 10, 64); err != nil || pid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	aidStr := params.Get("aid")
	if aidStr != "" {
		if aid, err = strconv.ParseInt(aidStr, 10, 64); err != nil || aid <= 0 {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	c.JSON(true, plSvc.PubView(c, pid, aid))
}

func reportShare(c *bm.Context) {
	var (
		err      error
		pid, aid int64
	)
	params := c.Request.Form
	pidStr := params.Get("pid")
	if pid, err = strconv.ParseInt(pidStr, 10, 64); err != nil || pid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	aidStr := params.Get("aid")
	if aid, err = strconv.ParseInt(aidStr, 10, 64); err != nil || aid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(true, plSvc.PubShare(c, pid, aid))
}

func list(c *bm.Context) {
	var (
		err                 error
		vmid, mid           int64
		pn, ps, sort, total int
		list                []*model.Playlist
	)
	params := c.Request.Form
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	vmidStr := params.Get("vmid")
	if vmidStr != "" {
		if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if mid == 0 && vmid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if vmid > 0 {
		mid = vmid
	}
	pnStr := params.Get("pn")
	if pn, err = strconv.Atoi(pnStr); err != nil || pn < 1 {
		pn = 1
	}
	psStr := params.Get("ps")
	if ps, err = strconv.Atoi(psStr); err != nil || ps < 1 || ps > conf.Conf.Rule.MaxPlsPageSize {
		ps = conf.Conf.Rule.MaxPlsPageSize
	}
	sortStr := params.Get("sort")
	if sortStr != "" {
		if sort, err = strconv.Atoi(sortStr); err != nil || sort < 0 {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if list, total, err = plSvc.List(c, mid, pn, ps, sort); err != nil {
		c.JSON(nil, switchCode(err, favmdl.TypePlayVideo))
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   pn,
		"size":  ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func addFavorite(c *bm.Context) {
	var (
		err      error
		mid, pid int64
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	pidStr := params.Get("pid")
	if pid, err = strconv.ParseInt(pidStr, 10, 64); err != nil || pid < 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, switchCode(plSvc.AddFavorite(c, mid, pid), favmdl.TypePlayList))
}

func delFavorite(c *bm.Context) {
	var (
		err      error
		mid, pid int64
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	pidStr := params.Get("pid")
	if pid, err = strconv.ParseInt(pidStr, 10, 64); err != nil || pid < 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, switchCode(plSvc.DelFavorite(c, mid, pid), favmdl.TypePlayList))
}

func listFavorite(c *bm.Context) {
	var (
		err                 error
		mid, vmid           int64
		pn, ps, sort, total int
		list                []*model.Playlist
	)
	params := c.Request.Form
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	vmidStr := params.Get("vmid")
	if vmidStr != "" {
		if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	pnStr := params.Get("pn")
	if pn, err = strconv.Atoi(pnStr); err != nil || pn < 1 {
		pn = 1
	}
	psStr := params.Get("ps")
	if ps, err = strconv.Atoi(psStr); err != nil || ps < 1 || ps > conf.Conf.Rule.MaxPlsPageSize {
		ps = conf.Conf.Rule.MaxPlsPageSize
	}
	sortStr := params.Get("sort")
	if sortStr != "" {
		if sort, err = strconv.Atoi(sortStr); err != nil || sort < 0 {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if list, total, err = plSvc.ListFavorite(c, mid, vmid, pn, ps, sort); err != nil {
		c.JSON(nil, switchCode(err, favmdl.TypePlayList))
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   pn,
		"size":  ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func switchCode(err error, tp int32) error {
	if err == nil {
		return err
	}
	switch xecode.Cause(err) {
	case favEcode.FavNameTooLong:
		err = ecode.PlNameTooLong
	case favEcode.FavFolderExist:
		err = ecode.PlExist
	case favEcode.FavMaxFolderCount:
		err = ecode.PlMaxCount
	case favEcode.FavCanNotDelDefault:
		err = ecode.PlCanNotDelDefault
	case favEcode.FavFloderAlreadyDel:
		err = ecode.PlAlreadyDel
	case favEcode.FavResourceOverflow:
		err = ecode.PlVideoOverflow
	case favEcode.FavResourceAlreadyDel:
		if tp == favmdl.TypePlayVideo {
			err = ecode.PlVideoAlreadyDel
		} else if tp == favmdl.TypePlayList {
			err = ecode.PlFavAlreadyDel
		}
	case favEcode.FavResourceExist:
		if tp == favmdl.TypePlayVideo {
			err = ecode.PlVideoExist
		} else if tp == favmdl.TypePlayList {
			err = ecode.PlFavExist
		}
	case favEcode.FavFolderNotExist:
		err = ecode.PlNotExist
	}
	return err
}
