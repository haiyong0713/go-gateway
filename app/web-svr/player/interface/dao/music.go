package dao

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/player/interface/model"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

const (
	_bgmURL    = "/x/copyright-music-publicity/bgm/entrance"
	_bgmClient = "2"
)

// bgm tags
func (d *Dao) BgmEntrance(c context.Context, aid, cid int64) (res *model.BgmEntranceReply, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("aid", strconv.FormatInt(aid, 10))
	params.Set("cid", strconv.FormatInt(cid, 10))
	params.Set("client", _bgmClient)
	var rs struct {
		Code int                     `json:"code"`
		Data *model.BgmEntranceReply `json:"data"`
	}
	if err = d.client.Get(c, d.c.Host.MusicAPI+_bgmURL, ip, params, &rs); err != nil {
		log.Error("BgmEntrance d.client.Get(%s, %s, %v) error(%v)", d.c.Host.MusicAPI+_bgmURL, ip, params, err)
		return
	}
	if rs.Code != 0 {
		return nil, errors.Wrap(ecode.Int(rs.Code), d.c.Host.MusicAPI+_bgmURL+"?"+params.Encode())
	}
	return rs.Data, nil
}
