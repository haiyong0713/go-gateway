package http

import (
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

func upperPassed(c *bm.Context) {
	params := c.Request.Form
	midStr := params.Get("mid")
	pnStr := params.Get("pn")
	psStr := params.Get("ps")
	// check params
	mid, err := strconv.ParseInt(midStr, 10, 64)
	if err != nil {
		log.Error("strconv.ParseInt(%s) error(%v)", midStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// deal
	pn, err := strconv.Atoi(pnStr)
	if err != nil || pn < 1 {
		pn = 1
	}
	ps, err := strconv.Atoi(psStr)
	if err != nil || ps < 1 || ps > 100 {
		ps = 20
	}
	as, err := arcSvc.UpperPassed3(c, mid, pn, ps)
	if err != nil {
		if ec := ecode.Cause(err); ec != ecode.NothingFound {
			log.Error("arcSvc.UpperPassed(%d) error(%d)", mid, err)
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(as, err)
}

// upperCount write the count of archives of Up.
func upperCount(c *bm.Context) {
	params := c.Request.Form
	midStr := params.Get("mid")
	// check params
	mid, err := strconv.ParseInt(midStr, 10, 64)
	if err != nil {
		log.Error("strconv.ParseInt(%s) error(%v)", midStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	count, err := arcSvc.UpperCount(c, mid)
	if err != nil {
		c.JSON(nil, err)
		log.Error("arcSvc.UpperCount(%d) error(%d)", mid, err)
		return
	}
	var res struct {
		Count int `json:"count"`
	}
	res.Count = count
	c.JSON(res, nil)
}
