package dao

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"

	"github.com/pkg/errors"
)

const (
	_audios = "/x/internal/v1/audio/menus/batch"
)

type audioConfig struct {
	Host string
}

type audioDao struct {
	client *bm.Client
	cfg    audioConfig
}

func (d *audioDao) Audios(ctx context.Context, ids []int64) (map[int64]*audio.Audio, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("ids", xstr.JoinInts(ids))
	var res struct {
		Code int                    `json:"code"`
		Data map[int64]*audio.Audio `json:"data"`
	}
	if err := d.client.Get(ctx, d.cfg.Host+_audios, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.cfg.Host+_audios+"?"+params.Encode())
	}
	return res.Data, nil
}
