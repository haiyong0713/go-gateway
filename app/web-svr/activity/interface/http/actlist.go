package http

import (
	bm "go-common/library/net/http/blademaster"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/service"
)

func listDomain(c *bm.Context) {
	args := new(struct {
		PageNo   int `json:"page_no" form:"page_no" default:"1"`
		PageSize int `json:"page_size" form:"page_size" default:"50"`
	})

	if err := c.Bind(args); err != nil {
		return
	}
	list, err := service.LikeSvc.ListDomain(c, args.PageNo, args.PageSize)
	if err != nil {
		return
	}
	if list == nil {
		list = []*lmdl.Record{}
	}
	c.JSON(struct {
		PageNo   int            `json:"page_no"`
		PageSize int            `json:"page_size"`
		List     []*lmdl.Record `json:"list"`
	}{
		PageNo:   args.PageNo,
		PageSize: args.PageSize,
		List:     list,
	}, err)
}

func searchDomain(c *bm.Context) {
	args := new(struct {
		ActDomain string `json:"act_domain" form:"act_domain" validate:"required"`
	})

	if err := c.Bind(args); err != nil {
		return
	}
	record, err := service.LikeSvc.SearchDomain(c, args.ActDomain)
	if err != nil {
		return
	}

	c.JSON(struct {
		DomainInfo *lmdl.Record `json:"domain_info"`
	}{
		DomainInfo: record,
	}, err)
}
