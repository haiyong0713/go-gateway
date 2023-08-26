package service

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-card/interface/model/card/game"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	"github.com/pkg/errors"
)

func (s *Service) MultiGameInfos(ctx context.Context, ids []int64) (map[int64]*game.Game, error) {
	params := url.Values{}
	params.Set("game_base_ids", xstr.JoinInts(ids))
	params.Set("sdk_type", castSDKType(ctx))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	var res struct {
		Code int          `json:"code"`
		Data []*game.Game `json:"data"`
	}
	if err := s.httpGameCo.Get(ctx, s.gameMultiInfos, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), s.gameMultiInfos+"?"+params.Encode())
	}
	out := make(map[int64]*game.Game, len(res.Data))
	for _, v := range res.Data {
		if !v.IsOnline {
			continue
		}
		out[v.GameBaseID] = v
	}
	return out, nil
}

func castSDKType(ctx context.Context) string {
	if pd.WithContext(ctx).IsIOSAll().MustFinish() {
		return "2"
	}
	return "1"
}
