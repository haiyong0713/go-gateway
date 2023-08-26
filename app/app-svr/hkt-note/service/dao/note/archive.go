package note

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/hkt-note/service/model/note"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	cepgrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/episode"
	cssngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	"github.com/pkg/errors"
)

const (
	_needAll         = 1
	_ps              = 50
	_feaContListPath = "/x/internal/feature/content/list"
)

func (d *Dao) ViewPage(c context.Context, aid int64) (map[int64]*note.PageCore, error) {
	var arg = &arcapi.PageRequest{Aid: aid}
	reply, err := d.arcClient.Page(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "View aid(%d)", aid)
	}
	page := make(map[int64]*note.PageCore)
	if reply == nil || len(reply.Pages) == 0 { // 失效稿件
		return page, nil
	}
	for _, p := range reply.Pages {
		page[p.Cid] = note.ToArcPage(p)
	}
	return page, nil
}

func (d *Dao) Arcs(c context.Context, aids []int64) (map[int64]*arcapi.Arc, error) {
	var arg = &arcapi.ArcsRequest{Aids: aids}
	reply, err := d.arcClient.Arcs(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "Arcs aids(%v)", aids)
	}
	if reply == nil || len(reply.Arcs) == 0 {
		log.Warnc(c, "noteInfo Arcs aids(%v) res nil", aids)
		return make(map[int64]*arcapi.Arc), nil
	}
	return reply.Arcs, nil
}

func (d *Dao) SimpleArcs(c context.Context, aids []int64) (map[int64]*arcapi.SimpleArc, error) {
	var arg = &arcapi.SimpleArcsRequest{Aids: aids}
	reply, err := d.arcClient.SimpleArcs(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "SimpleArcs aids(%v)", aids)
	}
	if reply == nil || len(reply.Arcs) == 0 {
		log.Warn("noteInfo SimpleArcs aids(%v) res nil", aids)
		return make(map[int64]*arcapi.SimpleArc), nil
	}
	return reply.Arcs, nil
}

func (d *Dao) CheeseSeasons(c context.Context, sids []int32) (map[int32]*cssngrpc.SeasonCard, error) {
	arg := &cssngrpc.SeasonCardsReq{Ids: sids, NeedAll: _needAll}
	reply, err := d.chSsnClient.Cards(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "CheeseSeasons sids(%v)", sids)
	}
	if reply == nil || reply.Cards == nil {
		log.Warn("noteInfo CheeseSeasons sids(%v) res nil", sids)
		return make(map[int32]*cssngrpc.SeasonCard), nil
	}
	return reply.Cards, nil
}

func (d *Dao) seasonEp(c context.Context, sid, pn int32) (res []*cepgrpc.EpisodeModel, total int32, err error) {
	var arg = &cepgrpc.SeasonEpReq{SeasonId: sid, Pn: pn, Ps: _ps}
	reply, err := d.chEpClient.SeasonEp(c, arg)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "SeasonEps arg(%+v)", arg)
	}
	if reply == nil {
		return nil, 0, errors.Wrapf(ecode.NothingFound, "SeasonEps arg(%+v)", arg)
	}
	return reply.Items, reply.Total, nil
}

func (d *Dao) SeasonEps(c context.Context, sid int32) (map[int64]*note.PageCore, error) {
	items, total, err := d.seasonEp(c, sid, 1)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]*note.PageCore)
	for _, p := range items {
		if p != nil {
			res[int64(p.Id)] = note.ToEpPage(p)
		}
	}
	if total <= _ps {
		return res, nil
	}
	// 数据不止一页时
	var (
		eg    = errgroup.WithContext(c)
		mutex sync.Mutex
	)
	for i := 2; i <= int(total/_ps)+1; i++ {
		curPn := int32(i)
		eg.Go(func(c context.Context) error {
			items, _, err := d.seasonEp(c, sid, curPn)
			if err != nil {
				return err
			}
			mutex.Lock()
			for _, p := range items {
				if p != nil {
					res[int64(p.Id)] = note.ToEpPage(p)
				}
			}
			mutex.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) FeatureContList(c context.Context) ([]*note.FeatureCont, error) {
	var (
		params = url.Values{}
		cfg    = d.c.NoteCfg.ForbidCfg
	)
	params.Set("group_id", cfg.PoliticsGroupId)
	params.Set("type", cfg.PoliticsType)
	listUrl := fmt.Sprintf("%s%s?%s", cfg.FeaHost, _feaContListPath, params.Encode())
	req, err := http.NewRequest("GET", listUrl, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "FeatureContList path(%s)", listUrl)
	}
	var res struct {
		Code    int                 `json:"code"`
		Message string              `json:"message"`
		List    []*note.FeatureCont `json:"list"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, errors.Wrapf(err, "FeatureContList path(%s)", listUrl)
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrapf(ecode.Int(res.Code), "FeatureContList path(%s)", listUrl)

	}
	return res.List, nil
}
