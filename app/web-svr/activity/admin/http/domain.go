package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/model/domain"
	"strconv"
	"time"
)

func addDomain(c *bm.Context) {
	var (
		err  error
		args = &domain.AddDomainParam{}
	)

	if err = c.Bind(args); err != nil {
		return
	}

	if args.Stime > args.Etime || args.Etime < (xtime.Time)(time.Now().Unix()) {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "参数错误"))
		return
	}
	rows, err := domainSrv.AddDomain(c, args)

	c.JSON(domain.UpdateRespone{Rows: rows}, err)
}

func editDomain(c *bm.Context) {
	var (
		err  error
		args = &domain.Record{}
	)

	if err = c.Bind(args); err != nil {
		return
	}

	rows, err := domainSrv.EditDomain(c, args)

	c.JSON(domain.UpdateRespone{Rows: rows}, err)
}

func stopDomain(c *bm.Context) {
	var (
		err error
		aid int64
	)

	if aid, err = strconv.ParseInt(c.Request.Form.Get("id"), 10, 64); err != nil || aid <= 0 {
		log.Errorc(c, "stopDomain aid:%v , strconv.ParseInt() failed. error(%v)", aid, err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "参数错误"))
		return
	}
	rows, err := domainSrv.StopDomain(c, aid)
	c.JSON(domain.UpdateRespone{Rows: rows}, err)
}

func searchDomain(c *bm.Context) {
	var (
		err  error
		args = &domain.Search{}
	)

	if err = c.Bind(args); err != nil {
		return
	}

	records, total, err := domainSrv.SearchDomain(c, args)

	c.JSON(struct {
		PageNo   int              `json:"page_no"`
		PageSize int              `json:"page_size"`
		Total    int              `json:"total"`
		List     []*domain.Record `json:"list"`
	}{
		PageNo:   args.PageNo,
		PageSize: args.PageSize,
		Total:    total,
		List:     records,
	}, err)
}

func syncCacheDomain(c *bm.Context) {
	args := new(struct {
		SynNum int `form:"sync_num" json:"sync_num" validate:"min=1,max=20"`
	})
	if err := c.Bind(args); err != nil {
		return
	}
	rows, err := domainSrv.SyncScript(c, args.SynNum)
	c.JSON(domain.UpdateRespone{Rows: rows}, err)
}
