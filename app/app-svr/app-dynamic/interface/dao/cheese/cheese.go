package cheese

import (
	"context"
	"net/http"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	cheesemdl "go-gateway/app/app-svr/app-dynamic/interface/model/cheese"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"

	dynCheeseGrpc "git.bilibili.co/bapis/bapis-go/cheese/service/dynamic"

	"github.com/pkg/errors"
)

func (d *Dao) MyPaid(ctx context.Context, mid int64) (*dynCheeseGrpc.MyPaidReply, error) {
	req := &dynCheeseGrpc.MyPaidReq{
		Mid: mid,
	}
	ret, err := d.cheeseGrpc.MyPaid(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, err
}

func (d *Dao) AdditionalCheese(c context.Context, ssids []int64) (map[int64]*cheesemdl.Cheese, error) {
	params := url.Values{}
	params.Set("season_ids", xstr.JoinInts(ssids))
	// new request
	req, err := d.client.NewRequest(http.MethodGet, d.attachCard, metadata.String(c, metadata.RemoteIP), params)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	var resTmp struct {
		Code int                 `json:"code"`
		Msg  string              `json:"message"`
		Data []*cheesemdl.Cheese `json:"data"`
	}
	if err = d.client.Do(c, req, &resTmp); err != nil {
		xmetric.DyanmicItemAPI.Inc(d.attachCard, "request_error")

		return nil, err
	}
	if resTmp.Code != ecode.OK.Code() {
		xmetric.DyanmicItemAPI.Inc(d.attachCard, "reply_code_error")
		err = errors.Wrap(ecode.Int(resTmp.Code), d.attachCard+"?"+params.Encode())
		return nil, err
	}
	var res = make(map[int64]*cheesemdl.Cheese)
	for _, cheese := range resTmp.Data {
		if cheese == nil || cheese.ID == 0 {
			continue
		}
		res[cheese.ID] = cheese
	}
	return res, nil
}
