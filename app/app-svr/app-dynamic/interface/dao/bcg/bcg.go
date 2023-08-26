package bcg

import (
	"context"
	"net/url"
	"strconv"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	bcgmdl "go-gateway/app/app-svr/app-dynamic/interface/model/bcg"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"

	bcgvo "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/vo"
	"github.com/pkg/errors"
)

const (
	_goodsDetails = "/dwp/api/openApi/v1/window/get"
	_creative     = "/mgk/api/open_api/v1/creative/%d"
)

func (d *Dao) DynamicAdInfo(c context.Context, mid int64, CreativeIds []int64, build, buvid, mobiAp, AdExtra, requestId, device, status, from string) (map[int64]*bcgvo.DynamicAdDto, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	resTmp, err := d.bcggrpc.DynamicAdInfo(c, &bcgvo.DynamicAdRequestForInfo{
		Mid:         mid,
		Buvid:       buvid,
		Ip:          ip,
		MobiApp:     mobiAp,
		AdExtra:     AdExtra,
		Build:       build,
		CreativeIds: CreativeIds,
		RequestId:   requestId,
		From:        from,
		Device:      device,
		CardStatus:  status,
	})
	if err != nil {
		log.Error("DynamicAdInfo mid %v, CreativeIds %v err %v", mid, CreativeIds, err)
		return nil, err
	}
	var res = make(map[int64]*bcgvo.DynamicAdDto)
	for _, ad := range resTmp.GetDynamicAds() {
		if ad == nil {
			continue
		}
		res[ad.CreativeId] = ad
	}
	return res, nil
}

func (d *Dao) GoodsDetails(c context.Context, req *bcgmdl.GoodsParams) (map[int]map[string]*bcgmdl.GoodsItem, error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(req.Uid, 10))
	params.Set("up_uid", strconv.FormatInt(req.UpUid, 10))
	params.Set("dynamic_id", strconv.FormatInt(req.DynamicID, 10))
	params.Set("ctx", req.Ctx)
	params.Set("input_extend", req.InputExtend)
	goodsDetailsURL := d.c.Hosts.CmCom + _goodsDetails
	var ret struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data *bcgmdl.GoodsRes `json:"data"`
	}
	if err := d.client.Get(c, goodsDetailsURL, "", params, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(goodsDetailsURL, "request_error")
		return nil, err
	}
	if ret.Code != 0 {
		xmetric.DyanmicItemAPI.Inc(goodsDetailsURL, "reply_code_error")
		return nil, errors.Wrap(ecode.Int(ret.Code), goodsDetailsURL+"?"+params.Encode())
	}
	if ret.Data == nil || ret.Data.OutputExtend.List == nil {
		xmetric.DyanmicItemAPI.Inc(goodsDetailsURL, "reply_date_error")
		return nil, errors.New("GoodsDetails ret.data is empty.")
	}
	rsp := make(map[int]map[string]*bcgmdl.GoodsItem)
	for _, item := range ret.Data.OutputExtend.List {
		resMap, ok := rsp[item.Type]
		if !ok {
			resMap = make(map[string]*bcgmdl.GoodsItem)
			rsp[item.Type] = resMap
		}
		resMap[strconv.FormatInt(item.ItemsID, 10)] = item
	}
	return rsp, nil
}

func (d *Dao) Creative(c context.Context, creativeID int64) (int64, error) {
	params := url.Values{}
	params.Set("dynamic_id", strconv.FormatInt(creativeID, 10))
	url := d.c.Hosts.CmCom + _creative
	var ret struct {
		Code int `json:"code"`
		Data struct {
			BizID int64 `json:"biz_id"`
		} `json:"data"`
	}
	if err := d.client.Get(c, url, "", params, &ret); err != nil {
		return 0, err
	}
	if ret.Code != 0 {
		return 0, errors.Wrap(ecode.Int(ret.Code), url+"?"+params.Encode())
	}
	return ret.Data.BizID, nil
}

func (d *Dao) Creatives(c context.Context, creativeIDs []int64) (map[int64]int64, error) {
	g := errgroup.WithContext(c)
	res := map[int64]int64{}
	mu := sync.Mutex{}
	for _, id := range creativeIDs {
		creativeID := id
		g.Go(func(ctx context.Context) error {
			bizID, err := d.Creative(ctx, creativeID)
			if err != nil {
				return err
			}
			mu.Lock()
			res[creativeID] = bizID
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}
