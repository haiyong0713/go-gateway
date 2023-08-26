package dao

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/fawkes/job/internal/model"
	"go-gateway/app/app-svr/fawkes/job/internal/model/pack"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
)

// QueryPackList 查询pack信息
func (d *dao) QueryPackList(c context.Context, oc *model.OutCfg, tStart, tEnd int64, pkgTypes []int64, appKey string) (res []*cimdl.BuildPack, err error) {
	var ip = metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	if tStart != 0 {
		params.Set("start_time", strconv.FormatInt(tStart, 10))
	}
	if tEnd != 0 {
		params.Set("end_time", strconv.FormatInt(tEnd, 10))
	}
	for _, v := range pkgTypes {
		params.Add("pkg_type", strconv.FormatInt(v, 10))
	}

	if appKey != "" {
		params.Set("app_key", appKey)
	}
	var re struct {
		Msg  string             `json:"message"`
		Code int                `json:"code"`
		Data []*cimdl.BuildPack `json:"data"`
	}

	if err = d.client.Get(c, oc.FAWKES+oc.LIST, ip, params, &re); err != nil {
		log.Error("err:%+v", err)
		return
	}
	return re.Data, nil
}

// DeleteExpiredPack 删除过期的构建包
func (d *dao) DeleteExpiredPack(c context.Context, oc *model.OutCfg, keys []*pack.BuildKey) (res *pack.DeleteResp, err error) {
	var ip = metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	bytes, _ := json.Marshal(keys)
	params.Set("delete_keys", string(bytes))
	var re struct {
		Msg  string           `json:"message"`
		Code int              `json:"code"`
		Data *pack.DeleteResp `json:"data"`
	}
	if err = d.client.Post(c, oc.FAWKES+oc.DELETE, ip, params, &re); err != nil {
		log.Error("err:%+v", err)
		return
	}
	return re.Data, err
}
