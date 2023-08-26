package up

import (
	"context"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
	"go-gateway/app/app-svr/app-car/interface/model/medialist"

	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

const _mediaListSuffix = "/x/v2/medialist/resource/list"

// Dao is coin dao
type Dao struct {
	c            *conf.Config
	upClient     upgrpc.UpArchiveClient
	client       *bm.Client
	mediaListUrl string
}

// New initial coin dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:            c,
		client:       bm.NewClient(c.HTTPClient),
		mediaListUrl: c.Host.APICom + _mediaListSuffix,
	}
	var err error
	if d.upClient, err = upgrpc.NewClient(c.UpGRPC); err != nil {
		panic(err)
	}
	return
}

// UpArcs get upper archives
func (d *Dao) UpArcs(c context.Context, mid, pn, ps int64) ([]*upgrpc.Arc, int64, error) {
	reply, err := d.upClient.ArcPassed(c, &upgrpc.ArcPassedReq{Mid: mid, Pn: pn, Ps: ps})
	if err != nil {
		log.Error("%v", err)
		return nil, 0, err
	}
	return reply.Archives, reply.Total, nil
}

// UpArcsWithUpward 支持向上翻页的up主稿件接口
func (d *Dao) UpArcsWithUpward(c context.Context, req fm_v2.OidsWithUpwardReq) (aids []int64, hasMore bool, err error) {
	if req.FmType != fm_v2.AudioUp {
		log.Warn("UpArcsWithUpward unknown fmType, req:%+v", req)
		return make([]int64, 0), false, nil
	}
	httpReq := &medialist.MediaListReq{
		Type:        _typeUpArcs,
		BizId:       req.FmId,
		OType:       _oTypeUgc,
		Oid:         req.Cursor,
		Desc:        true,
		Direction:   req.Upward,
		WithCurrent: req.WithCurrent,
		Ps:          req.Ps,
		MobiApp:     req.MobiApp,
	}
	res, err := d.MediaList(c, httpReq)
	if err != nil {
		return nil, false, err
	}
	hasMore = res.Data.HasMore
	for _, v := range res.Data.MediaList {
		if v.Type != _oTypeUgc {
			continue
		}
		aids = append(aids, v.ID)
	}
	return aids, hasMore, nil
}
