package web

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	webmdl "go-gateway/app/web-svr/web-goblin/interface/model/web"
)

const (
	_del = "del"
)

// UgcFull search all ugc .
func (s *Service) UgcFull(ctx context.Context, pn, ps int64, source string) (res []*webmdl.Mi, err error) {
	if res, err = s.FullShort(ctx, pn, ps, source); err != nil {
		log.Error("UgcFull error (%v)", err)
		return
	}
	if len(res) > 0 {
		for idx := range res {
			res[idx].UgcFullDeal()
		}
	}
	return
}

// UgcIncre search ugc after a certain time .
func (s *Service) UgcIncre(ctx context.Context, pn, ps int, start, end int64, source string) (res []*webmdl.Mi, err error) {
	var (
		aids    []*webmdl.SearchAids
		opmap   map[int64]string
		delaids []int64
		tmpAids []int64
		ip      = metadata.String(ctx, metadata.RemoteIP)
	)
	if aids, err = s.dao.UgcIncre(ctx, pn, ps, start, end); err != nil {
		log.Error("s.dao.UgcIncre error (%v)", err)
		return
	}
	opmap = make(map[int64]string, len(aids))
	for _, v := range aids {
		opmap[v.Aid] = v.Action
		if v.Action == _del {
			delaids = append(delaids, v.Aid)
		} else {
			tmpAids = append(tmpAids, v.Aid)
		}
	}
	if res, err = s.archiveWithTag(ctx, tmpAids, ip, opmap, source); err != nil {
		log.Warn("s.archiveWithTag ip(%s) aids(%s) error(%v)", err, ip, xstr.JoinInts(tmpAids))
	}
	for _, v := range delaids {
		m := &webmdl.Mi{}
		m.Op = _del
		m.ID = v
		res = append(res, m)
	}
	if len(res) > 0 {
		for idx := range res {
			res[idx].UgcIncreDeal()
		}
	}
	return
}

// RankingReg .
func (s *Service) RankingReg(ctx context.Context, rid, day int, source string) (res []*webmdl.Mi, err error) {
	var (
		aids []int64
		rs   []*webmdl.NewArchive
		op   = make(map[int64]string)
		ip   = metadata.String(ctx, metadata.RemoteIP)
	)
	res = make([]*webmdl.Mi, 0)
	if rs, err = s.dao.Ranking(ctx, rid, day); err != nil {
		log.Error("[RankingReg] s.dao.RankingRegion rid(%d) error(%v)", rid, err)
		return
	}
	if len(rs) == 0 {
		log.Info("big data return is nil")
		return
	}
	for _, v := range rs {
		aids = append(aids, v.Aid)
	}
	if res, err = s.archiveWithTag(ctx, aids, ip, op, source); err != nil {
		log.Error("[RankingReg]s.archiveWithTag error(%v)", err)
	}
	return
}
