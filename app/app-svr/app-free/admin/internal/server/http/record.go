package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-free/admin/internal/model"
)

var (
	_ispm = map[model.ISP]struct{}{
		model.ISPMobile: {},
		model.ISPTelcom: {},
		model.ISPUnicom: {},
	}
)

func addRecord(ctx *bm.Context) {
	param := new(struct {
		Records string `form:"records" validate:"required"`
	})
	if err := ctx.Bind(param); err != nil {
		return
	}
	var (
		res = map[string]interface{}{}
		rs  []*model.FreeRecord
	)
	if err := json.Unmarshal([]byte(param.Records), &rs); err != nil {
		res["message"] = fmt.Sprintf("数据解析错误 %v", err)
		ctx.JSONMap(res, ecode.RequestErr)
		return
	}
	if len(rs) == 0 {
		res["message"] = fmt.Sprintf("数据错误 %s", param.Records)
		ctx.JSON(res, ecode.RequestErr)
		return
	}
	for _, r := range rs {
		if !model.CheckIP(r.IPStart, r.IPEnd) {
			res["message"] = fmt.Sprintf("起始IP或结束IP不是IPV4,或者起始IP大于结束IP %+v", r)
			ctx.JSONMap(res, ecode.RequestErr)
			return
		}
		if _, ok := _ispm[r.ISP]; !ok {
			res["message"] = fmt.Sprintf("数据isp错误 %+v", r)
			ctx.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if err := svc.InsertFreeRecords(ctx, rs); err != nil {
		res["message"] = fmt.Sprintf("记录添加错误 %v", err)
		ctx.JSONMap(res, ecode.RequestErr)
		return
	}
	ctx.JSONMap(res, nil)
}

func Record(ctx *bm.Context) {
	param := new(struct {
		IP string `form:"ip"`
	})
	if err := ctx.Bind(param); err != nil {
		return
	}
	var ips []string
	if param.IP != "" {
		ips = strings.Split(param.IP, ",")
	}
	ctx.JSON(svc.Records(ctx, ips))
}
