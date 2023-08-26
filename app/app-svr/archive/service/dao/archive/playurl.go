package archive

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/archive/service/model/archive"

	"github.com/pkg/errors"
)

// PGCPlayerInfos cid with pgc player info
func (d *Dao) PGCPlayerInfos(c context.Context, aids []int64, platform, ip, session string, fnval, fnver int64) (pgcm map[int64]*archive.PlayerInfo, err error) {
	params := url.Values{}
	params.Set("aids", xstr.JoinInts(aids))
	params.Set("mobi_app", platform)
	params.Set("ip", ip)
	params.Set("fnver", strconv.FormatInt(fnver, 10))
	params.Set("fnval", strconv.FormatInt(fnval, 10))
	params.Set("session", session)
	res := struct {
		Code   int                          `json:"code"`
		Result map[int64]*archive.PGCPlayer `json:"result"`
	}{}
	if err = d.playerClient.Get(c, d.c.PGCPlayerAPI, ip, params, &res); err != nil {
		return
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.c.PGCPlayerAPI+params.Encode())
		return
	}
	pgcm = make(map[int64]*archive.PlayerInfo)
	for _, v := range res.Result {
		if v.PlayerInfo != nil {
			pgcm[v.PlayerInfo.Cid] = v.PlayerInfo
		}
	}
	return
}

// PGCPlayURLs cid with pgc player info
func (d *Dao) PGCPlayURLs(c context.Context, aids []int64, platform, ip, session string, fnval, fnver int64) (pgcm map[int64]*archive.PGCPlayurl, err error) {
	params := url.Values{}
	params.Set("aids", xstr.JoinInts(aids))
	params.Set("mobi_app", platform)
	params.Set("ip", ip)
	params.Set("fnver", strconv.FormatInt(fnver, 10))
	params.Set("fnval", strconv.FormatInt(fnval, 10))
	params.Set("session", session)
	res := struct {
		Code   int                           `json:"code"`
		Result map[int64]*archive.PGCPlayurl `json:"result"`
	}{}
	if err = d.playerClient.Get(c, d.c.PGCPlayerV2API, ip, params, &res); err != nil {
		return
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.c.PGCPlayerV2API+params.Encode())
		return
	}
	pgcm = make(map[int64]*archive.PGCPlayurl)
	for _, v := range res.Result {
		if v.PlayerInfo != nil {
			pgcm[int64(v.PlayerInfo.Cid)] = v
		}
	}
	return
}
