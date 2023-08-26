package taishan

import (
	"context"
	"fmt"

	"go-common/library/database/taishan"
	"go-common/library/log"

	"go-gateway/app/app-svr/playurl/service/conf"
	tmdl "go-gateway/app/app-svr/playurl/service/model/taishan"
)

type Dao struct {
	c       *conf.Config
	taishan taishan.TaishanProxyClient
}

func playConfKey(buvid string) []byte {
	// 缩短key前缀
	return []byte(fmt.Sprintf("p_%s", buvid))
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.taishan, err = taishan.NewClient(c.TaiShanClient); err != nil {
		panic(fmt.Sprintf("taishan NewClient error(%v)", err))
	}
	return
}

// PlayConfGet .
// nolint:gomnd
func (d *Dao) PlayConfGet(c context.Context, buvid string) (rly *tmdl.PlayConfs, err error) {
	var (
		resp *taishan.GetResp
		ukey = playConfKey(buvid)
	)
	rly = &tmdl.PlayConfs{}
	cReq := &taishan.GetReq{
		Table:  d.c.TaiShanConf.PlayConfTable,
		Auth:   &taishan.Auth{Token: d.c.TaiShanConf.PlayConfToken},
		Record: &taishan.Record{Key: ukey},
	}
	if resp, err = d.taishan.Get(c, cReq); err != nil {
		log.Error("PlayConfGet d.taishan.Get (key:%s) err(%v)", string(ukey), err)
		return
	}
	if resp.GetRecord().GetStatus().GetErrNo() == 404 {
		err = nil
		return
	}
	if resp.GetRecord().GetStatus().GetErrNo() != 0 {
		err = fmt.Errorf("PlayConfGet error code %d", resp.GetRecord().GetStatus().GetErrNo())
		log.Error("PlayConfGet d.taishan.Get (key:%s) err(%v)", string(ukey), resp.Record.Status)
		return
	}
	// 没有查到对应的数据
	if len(resp.Record.Columns) == 0 || resp.Record.Columns[0] == nil {
		log.Error("PlayConfGet d.taishan.Get (key:%s) colums is nil ", string(ukey))
		return
	}
	if err = rly.Unmarshal(resp.Record.Columns[0].Value); err != nil {
		log.Error("PlayConfGet Unmarshal(key:%s) err(%v)", string(ukey), err)
	}
	return
}

// PlayConfSet .
func (d *Dao) PlayConfSet(c context.Context, arg *tmdl.PlayConfs, buvid string) (err error) {
	var (
		ukey  = playConfKey(buvid)
		value []byte
		resp  *taishan.PutResp
		req   = &taishan.PutReq{
			Table: d.c.TaiShanConf.PlayConfTable,
			Auth: &taishan.Auth{
				Token: d.c.TaiShanConf.PlayConfToken,
			},
		}
	)
	if value, err = arg.Marshal(); err != nil {
		log.Error("PlayConfSet data(%+v) err(%v)", arg, err)
		return
	}
	req.Record = &taishan.Record{Key: ukey, Columns: []*taishan.Column{{Value: value}}}
	if resp, err = d.taishan.Put(c, req); err != nil {
		log.Error("PlayConfSet (key:%s, data:%+v) err(%v)", ukey, arg, err)
		return
	}
	if resp.GetStatus().GetErrNo() != 0 {
		err = fmt.Errorf("PlayConfSet error code(%d)", resp.GetStatus().GetErrNo())
		log.Error("PlayConfSet status (key:%s, data:%+v) err(%v)", ukey, arg, resp.Status)
	}
	return
}
