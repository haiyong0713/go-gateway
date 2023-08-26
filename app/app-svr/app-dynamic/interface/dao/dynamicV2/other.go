package dynamicV2

import (
	"context"
	"net/url"
	"strings"

	"go-common/library/ecode"

	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"

	dynactivitygrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/activity"
	dyntopicextgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic-ext"
	"github.com/pkg/errors"
)

const (
	_emojiURL = "/x/internal/emote/by/text"
)

func (d *Dao) GetEmoji(ctx context.Context, emojis []string) (map[string]*mdlv2.EmojiItem, error) {
	params := url.Values{}
	params.Set("texts", strings.Join(emojis, ","))
	params.Set("business", "dynamic")
	emojiURL := d.c.Hosts.ApiCo + _emojiURL
	var ret struct {
		Code int          `json:"code"`
		Msg  string       `json:"msg"`
		Data *mdlv2.Emoji `json:"data"`
	}
	if err := d.client.Get(ctx, emojiURL, "", params, &ret); err != nil {
		xmetric.DynamicBackfillAPI.Inc(emojiURL, "request_error")
		return nil, errors.WithStack(err)
	}
	if ret.Code != 0 || ret.Data == nil {
		xmetric.DynamicBackfillAPI.Inc(emojiURL, "reply_error")
		return nil, errors.Wrapf(ecode.Int(ret.Code), "getEmoji url: %v, code: %v msg: %v or data nil", emojiURL, ret.Code, ret.Msg)
	}
	return ret.Data.Emote, nil
}

func (d *Dao) DynamicAttachedPromo(c context.Context, args []*dynactivitygrpc.DynamicAttachedPromoInfo) (map[int64]*dynactivitygrpc.DynamicAttachedPromoInfo, error) {
	resTmp, err := d.dynamicActivityGRPC.DynamicAttachedPromo(c, &dynactivitygrpc.DynamicAttachedPromoReq{Dynamics: args})
	if err != nil {
		return nil, err
	}
	var res = make(map[int64]*dynactivitygrpc.DynamicAttachedPromoInfo)
	if resTmp != nil {
		for _, re := range resTmp.AttachedPromos {
			if re == nil {
				continue
			}
			res[re.DynamicId] = re
		}
	}
	return res, nil
}

func (d *Dao) ListTopicAdditiveCards(c context.Context, args []int64) (map[int64]*dyntopicextgrpc.TopicAdditiveCard, error) {
	resTmp, err := d.dynamicTopicExtGRPC.ListTopicAdditiveCards(c, &dyntopicextgrpc.ListTopicAdditiveCardReq{TopicIds: args})
	if err != nil {
		return nil, err
	}
	var res = make(map[int64]*dyntopicextgrpc.TopicAdditiveCard)
	if resTmp != nil {
		for _, re := range resTmp.TopicAdditiveCards {
			if re == nil {
				continue
			}
			res[re.TopicId] = re
		}
	}
	return res, nil
}
