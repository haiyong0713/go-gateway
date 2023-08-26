package dao

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	tusm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tus"

	"github.com/pkg/errors"
)

func (d *dao) SaveTusConfigs(ctx context.Context, details []*tusm.Detail) error {
	taishanBatchPutKey := make(map[string][]byte, len(details))
	for _, v := range details {
		tkf := tusm.TaishanKeyInfos{
			TusValue: v.TusValue,
			Result:   v.Result,
		}
		key := tkf.BuildKeyByResult()
		taishanBatchPutKey[key] = v.Config
	}
	req := d.kv.NewBatchPutReq(ctx, taishanBatchPutKey)
	resp, err := d.kv.BatchPut(ctx, req)
	if err != nil {
		return err
	}
	if !resp.AllSucceed {
		return errors.Errorf("Failed to put all config to taishan")
	}
	return nil
}

func (d *dao) BatchFetchTusInfos(ctx context.Context, tusValue []string) ([]*tusm.Info, error) {
	params := url.Values{}
	for _, v := range tusValue {
		params.Add("ids", v)
	}
	uri := fmt.Sprintf("%s%s", d.tusHost, "/titan/foreign/crowd/query/detail")
	res := struct {
		Code int64 `json:"code"`
		Data []struct {
			Owner    string `json:"owner"`
			IsValid  int64  `json:"isValid"`
			Ctime    int64  `json:"ctime"`
			VaildDay int64  `json:"validDay"`
			Name     string `json:"crowdName"`
			Count    int64  `json:"crowdCount"`
			CrowdId  int64  `json:"crowdId"`
		} `json:"data"`
	}{}
	if err := d.bmClient.Get(ctx, uri, "", params, &res); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch tus infos uri(%s)", uri)
	}
	if res.Code != http.StatusOK {
		return nil, errors.Errorf("failed to fetch tus infos code(%d), uri(%s)", res.Code, uri)
	}
	var tusInfos []*tusm.Info
	for _, v := range res.Data {
		tusInfo := &tusm.Info{
			TusValue:   strconv.FormatInt(v.CrowdId, 10),
			Name:       v.Name,
			Creator:    v.Owner,
			Status:     v.IsValid,
			ValidDay:   v.VaildDay,
			CrowdCount: v.Count,
			Ctime:      v.Ctime,
		}
		tusInfos = append(tusInfos, tusInfo)
	}
	return tusInfos, nil
}
