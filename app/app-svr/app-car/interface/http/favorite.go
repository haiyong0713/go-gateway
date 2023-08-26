package http

import (
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/favorite"
	"go-gateway/app/app-svr/app-car/interface/model/show"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func mediaFavorite(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.FavoriteParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.Favorite(c, plat, mid, param)
	c.JSON(struct {
		Title string       `json:"title"`
		Item  []*show.Item `json:"items"`
	}{Title: "我的收藏夹", Item: data}, err)
}

func mediaList(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.MediaListParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.MediaList(c, plat, mid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}

func toview(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.ToViewParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.ToView(c, plat, mid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}

func userFolders(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.UserFolderParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := showSvc.UserFolders(c, mid, param)
	c.JSON(struct {
		Item []*favorite.UserFolder `json:"items"`
	}{Item: data}, err)
}

func favAddOrDelFolders(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.FavAddOrDelFolders{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, showSvc.FavAddOrDelFolders(c, mid, param))
}

func addFolder(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.AddFolder{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	fid, err := showSvc.AddFolder(c, mid, param)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(struct {
		Fid int64 `json:"fid"`
	}{Fid: fid}, nil)
}

func favAddOrDelFoldersWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.FavAddOrDelFolders{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, showSvc.FavAddOrDelFoldersWeb(c, mid, param))
}

func addFolderWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.AddFolder{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	fid, err := showSvc.AddFolderWeb(c, mid, param)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(struct {
		Fid int64 `json:"fid"`
	}{Fid: fid}, nil)
}

func favoriteWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.FavoriteParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := showSvc.FavoriteWeb(c, mid, param)
	c.JSON(struct {
		Title string          `json:"title"`
		Item  []*show.ItemWeb `json:"items"`
	}{Title: "我的收藏夹", Item: data}, err)

}

func mediaListWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.MediaListParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, page, err := showSvc.MediaListWeb(c, model.PlatH5, mid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data, Page: page}, err)
}

func toviewWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &favorite.ToViewParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := showSvc.ToViewWeb(c, mid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}

func favoriteV2(c *bm.Context) {
	var (
		req = new(commonmdl.FavoriteReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	referer := c.Request.Referer()
	cookie := c.Request.Header.Get(_headerCookie)
	c.JSON(commonSvc.Favorite(c, req, mid, buvid, cookie, referer))
}

func favoriteVideoV2(c *bm.Context) {
	var (
		req = new(commonmdl.FavoriteVideoReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	c.JSON(commonSvc.FavoriteVideo(c, req, mid, buvid))
}

func favoriteBangumiV2(c *bm.Context) {
	var (
		req = new(commonmdl.FavoriteBangumiReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	c.JSON(commonSvc.FavoriteBangumi(c, req, mid, buvid))
}

func favoriteCinemaV2(c *bm.Context) {
	var (
		req = new(commonmdl.FavoriteCinemaReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	c.JSON(commonSvc.FavoriteCinema(c, req, mid, buvid))
}

func favoriteToView(c *bm.Context) {
	var (
		req = new(commonmdl.FavoriteToViewReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if mid <= 0 {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	c.JSON(commonSvc.FavoriteToView(c, req, mid, buvid))
}
