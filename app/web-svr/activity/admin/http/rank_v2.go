package http

import (
	"context"
	"encoding/csv"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank_v3"
	"io/ioutil"
	"strconv"
	"strings"
)

func rankV2Create(c *bm.Context) {
	arg := new(rankmdl.CreateReq)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(rankv2Srv.Create(c, arg, userName))
}

func rankV2Update(c *bm.Context) {
	arg := new(rankmdl.UpdateReq)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankv2Srv.Update(c, arg, userName))
}

func rankV2List(c *bm.Context) {
	arg := new(rankmdl.ListReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(rankv2Srv.List(c, arg))
}

func rankV2UpdateRule(c *bm.Context) {
	arg := new(rankmdl.RuleUpdateReq)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankv2Srv.UpdateRules(c, arg, userName))
}

func UpdateRulesShowInfo(c *bm.Context) {
	arg := new(rankmdl.RuleUpdateShowReq)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankv2Srv.UpdateRulesShowInfo(c, arg, userName))
}

func rankV2Rank(c *bm.Context) {
	arg := new(rankmdl.BaseReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(rankv2Srv.GetRank(c, arg.ID))
}

func rankV2BlackWhite(c *bm.Context) {
	arg := new(rankmdl.BlackWhiteReq)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankv2Srv.UpdateBlackWhite(c, arg))
}

func rankV2UpdateAdjust(c *bm.Context) {
	arg := new(rankmdl.UpdateAdjustObject)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankv2Srv.UpdateAdjust(c, arg, userName))
}

func rankV2Detail(c *bm.Context) {
	arg := new(rankmdl.ResultReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(rankv2Srv.GetRankResult(c, arg))
}

func rankV2Publish(c *bm.Context) {
	arg := new(rankmdl.PublishReq)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankv2Srv.Publish(c, arg, userName))
}

func rankV2RuleOffline(c *bm.Context) {
	arg := new(rankmdl.RuleOfflineReq)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankv2Srv.RankRuleOffline(c, arg, userName))
}

func rankV2Export(c *bm.Context) {
	arg := new(rankmdl.ExportReq)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankv2Srv.Export(c, arg, userName))
}

func rankV2ExportResult(c *bm.Context) {
	arg := new(rankmdl.ExportRankResultReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankv2Srv.ExportRankResult(c, arg, userName))
}

func rankV2Source(c *bm.Context) {
	arg := new(rankmdl.SourceListReq)
	if err := c.BindWith(arg, binding.JSON, binding.Form); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(rankv2Srv.GetSourceList(c, arg))

}

func uploadSource(c *bm.Context) {
	v := new(struct {
		BaseID int64 `form:"base_id" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	var (
		err  error
		data []byte
	)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorc(c, "importDetailCSV upload err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Errorc(c, "importDetailCSV ioutil.ReadAll err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Errorc(c, "importDetailCSV r.ReadAll() err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var args []int64
Loop:
	for i, row := range records {
		if i == 0 {
			continue
		}
		// import csv state online
		var arg int64
		for field, value := range row {
			value = strings.TrimSpace(value)
			switch field {
			case 0:
				if value == "" {
					log.Warn("importDetailCSV name provinceID(%s)", value)
					continue Loop
				}
				mid, err := strconv.ParseInt(value, 10, 64)

				if err != nil {
					// continue Loop
				}
				arg = mid
			}

		}
		if arg != 0 {
			args = append(args, arg)
		}
	}
	if len(args) == 0 {
		log.Errorc(c, "importDetailCSV args no after filter")
		c.JSON(nil, ecode.RequestErr)
		return
	}
	go func() {
		ctx := context.Background()
		rankv2Srv.UploadSource(ctx, v.BaseID, args)
	}()
	c.JSON(nil, nil)

}
