package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/popular"
	"go-gateway/app/app-svr/app-car/interface/model/search"
)

func mediaPopularWeb(c *bm.Context) {
	param := &popular.MediaPopularParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(showSvc.MediaPopularWeb(c, param))
}

func mediaSearchWeb(c *bm.Context) {
	param := &search.MediaSearchParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(showSvc.MediaSearchWeb(c, param))
}

func mediaPGCWeb(c *bm.Context) {
	param := &bangumi.MediaPGCParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(showSvc.MediaPGCWeb(c, param))
}

func mediaRegion(c *bm.Context) {
	param := new(struct {
		Pn     int64  `form:"pn" default:"1" validate:"min=1"`
		Ps     int64  `form:"ps" default:"1" validate:"min=1,max=50"`
		Source string `form:"source" validate:"required"`
	})
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, url, err := showSvc.MediaRegionV2(c, param.Pn, param.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(struct {
		Item    []*card.MediaItem `json:"items"`
		MoreURL string            `json:"more_url"`
	}{Item: data,
		MoreURL: url,
	}, nil)
}

func mediaSearch(c *bm.Context) {
	param := new(struct {
		Pn      int    `form:"pn" default:"1" validate:"min=1"`
		Ps      int    `form:"ps" default:"1" validate:"min=1,max=50"`
		Keyword string `form:"keyword" validate:"required"`
		Source  string `form:"source" validate:"required"`
	})
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, url, err := showSvc.MediaRegionSearchV2(c, param.Pn, param.Ps, param.Keyword)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(struct {
		Item    []*card.MediaItem `json:"items"`
		MoreURL string            `json:"more_url"`
	}{Item: data,
		MoreURL: url,
	}, nil)
}
