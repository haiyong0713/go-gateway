package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/story/internal/model"

	"github.com/pkg/errors"
)

const _storyCart = "/bce/api/story/cart"

func (d *dao) StoryCart(ctx context.Context, param *model.StoryCartParam) (*model.StoryCartReply, error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(param.Mid, 10))
	params.Set("buvid", param.Buvid)
	params.Set("ip", param.IP)
	params.Set("mobi_app", param.MobiApp)
	params.Set("ad_extra", param.AdExtra)
	params.Set("build", strconv.Itoa(param.Build))
	params.Set("aid", strconv.FormatInt(param.Aid, 10))
	params.Set("cid", strconv.FormatInt(param.Cid, 10))
	params.Set("av_rid", strconv.FormatInt(param.AvRid, 10))
	params.Set("av_up_id", strconv.FormatInt(param.AvUpId, 10))
	params.Set("ua", param.Ua)
	params.Set("resource", param.Resource)
	params.Set("country", param.Country)
	params.Set("province", param.Province)
	params.Set("city", param.City)
	var res struct {
		Code int                   `json:"code"`
		Data *model.StoryCartReply `json:"data"`
	}
	if err := d.adClient.Get(ctx, d.storyCart, param.IP, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.storyCart+"?"+params.Encode())
	}
	return res.Data, nil
}
