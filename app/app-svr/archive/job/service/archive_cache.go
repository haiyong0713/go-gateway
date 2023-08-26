package service

import (
	"context"
	"encoding/json"
	ugcmdl "go-gateway/app/app-svr/ugc-season/service/api"
	"math/rand"
	"strconv"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive/service/api"
	arcmdl "go-gateway/app/app-svr/archive/service/model"
	"go-gateway/app/app-svr/archive/service/model/archive"
)

const (
	_prefixStat = "stpr_"
)

func statKey(aid int64) (key string) {
	return _prefixStat + strconv.FormatInt(aid, 10)
}

func (s *Service) loadTypes() (err error) {
	types, err := s.resultDao.RawTypes(context.Background())
	if err != nil {
		log.Error("s.dao.Types error(%+v)", err)
		return err
	}
	s.tNames = types
	log.Info("loadTypes success allTypesCount(%d)", len(types))
	return nil
}

func (s *Service) setArcCache(c context.Context, arc *api.Arc) error {
	arc.Fill()
	if arc.AttrVal(api.AttrBitIsCooperation) == api.AttrYes {
		staff, err := s.resultDao.RawStaff(c, arc.Aid)
		if err != nil {
			return err
		}
		arc.StaffInfo = staff
	}
	if arc.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes {
		p, err := s.getArcPremiereInfo(c, arc.Aid)
		if err != nil {
			return err
		}
		arc.Premiere = p
	}
	desc, descV2, subType, err := s.resultDao.RawAdditV2(c, arc.Aid)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	//设置付费属性
	if arc.AttrValV2(api.AttrBitV2Pay) == api.AttrYes {
		if subType == 0 {
			log.Error("日志告警 付费稿件 付费属性为0 arc(%+v)", arc)
		}
		arc.Pay = &api.PayInfo{
			PayAttr: subType,
		}
		err := s.setArcPayInfo(c, arc)
		if err != nil {
			return err
		}
	}
	if desc == "" {
		desc = arc.Desc
	}
	arcBs, err := arc.Marshal()
	if err != nil {
		log.Error("setArcCache Marshal aid(%d) arc(%+v) error(%+v)", arc.Aid, arc, err)
		return nil
	}
	// a3p_{aid}缓存设置过期时间，默认10小时+48小时内随机数
	rand.Seed(time.Now().UnixNano())
	exp := int64(36000)
	arcExp := exp + rand.Int63n(172800)
	if err = s.setCacheWithExp(c, "SET", arcmdl.ArcKey(arc.Aid), arcBs, arcExp); err != nil {
		return err
	}
	if err = s.setTaishan(c, []byte(arcmdl.ArcKey(arc.Aid)), arcBs); err != nil {
		return err
	}
	resultAddit := &archive.Addit{
		Aid:         arc.Aid,
		Description: desc,
		DescV2:      descV2,
	}
	resultAdditBs, err := json.Marshal(resultAddit)
	if err != nil {
		log.Error("setArcCache resultAddit Marshal aid(%d) addit(%+v) error(%+v)", arc.Aid, resultAddit, err)
		return nil
	}
	if err = s.setTaishan(c, []byte(arcmdl.DescKeyV2(arc.Aid)), resultAdditBs); err != nil {
		return err
	}
	return nil
}

func (s *Service) setVideosPageCache(c context.Context, aid int64, vs []*api.Page) error {
	v := &api.AidVideos{Aid: aid, Pages: vs}
	bs, err := v.Marshal()
	if err != nil {
		log.Error("setVideosPageCache Marshal aid(%d) v(%+v) err(%+v) ", aid, v, err)
		return nil
	}
	// psb_{aid}缓存设置过期时间，默认10小时+48小时内随机数
	rand.Seed(time.Now().UnixNano())
	exp := int64(36000)
	psbExp := exp + rand.Int63n(172800)
	if err = s.setCacheWithExp(c, "SET", arcmdl.PageKey(aid), bs, psbExp); err != nil {
		return err
	}
	if err = s.setTaishan(c, []byte(arcmdl.PageKey(aid)), bs); err != nil {
		return err
	}
	return nil
}

func (s *Service) setVideoCache(c context.Context, aid, cid int64, video *api.Page) error {
	bs, err := video.Marshal()
	if err != nil {
		log.Error("setVideoCache Marshal aid(%d) cid(%d) video(%+v) err(%+v) ", aid, cid, video, err)
		return nil
	}

	// psb_#{aid}_#{cid}缓存设置过期时间，默认10小时+48小时内随机数
	rand.Seed(time.Now().UnixNano())
	exp := int64(36000) + rand.Int63n(172800)

	if err = s.setCacheWithExp(c, "SET", arcmdl.VideoKey(aid, cid), bs, exp); err != nil {
		return err
	}
	if err = s.setTaishan(c, []byte(arcmdl.VideoKey(aid, cid)), bs); err != nil {
		return err
	}
	return nil
}

func (s *Service) initStatCache(c context.Context, aid int64) error {
	_, err := s.resultDao.RawStat(c, aid)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}
	st := &api.Stat{Aid: aid}
	bs, err := st.Marshal()
	if err != nil {
		return err
	}
	if err = s.setCache(c, "SETNX", statKey(aid), bs); err != nil {
		return err
	}
	return nil
}

func (s *Service) setCache(c context.Context, command string, key string, val []byte) (err error) {
	for _, rds := range s.arcRedises {
		if err = func() (err error) {
			conn := rds.Get(c)
			defer conn.Close()
			if _, err = conn.Do(command, key, val); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) setCacheWithExp(c context.Context, command string, key string, val []byte, exp int64) (err error) {
	for _, rds := range s.arcRedises {
		if err = func() (err error) {
			conn := rds.Get(c)
			defer conn.Close()
			if _, err = conn.Do(command, key, val, "EX", exp); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) setSimpleArcCache(c context.Context, arc *api.Arc, pages []*api.Page) error {
	var cids []int64
	for _, p := range pages {
		cids = append(cids, p.Cid)
	}
	sa := &api.SimpleArc{
		Aid:         arc.Aid,
		Cids:        cids,
		TypeId:      arc.TypeID,
		Copyright:   arc.Copyright,
		State:       arc.State,
		Access:      arc.Access,
		Attribute:   arc.Attribute,
		Duration:    arc.Duration,
		RedirectUrl: arc.RedirectURL,
		Mid:         arc.Author.Mid,
		SeasonId:    arc.SeasonID,
		AttributeV2: arc.AttributeV2,
		Pubdate:     int64(arc.PubDate),
		Rights: &api.SimpleRights{
			ArcPay: arc.AttrValV2(api.AttrBitV2Pay),
		},
	}
	if arc.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes {
		p, err := s.getArcPremiereInfo(c, arc.Aid)
		if err != nil {
			return err
		}
		sa.Premiere = p
	}
	//设置付费属性
	if arc.AttrValV2(api.AttrBitV2Pay) == api.AttrYes {
		err := s.setSimpleArcPayInfo(c, sa)
		if err != nil {
			return err
		}
	}
	bs, err := sa.Marshal()
	if err != nil {
		log.Error("setSimpleArcCache Marshal aid(%d) sa(%+v) err(%+v) ", arc.Aid, sa, err)
		return nil
	}
	for k, rds := range s.sArcRds {
		if err = func() error {
			conn := rds.Get(c)
			defer conn.Close()
			_, e := conn.Do("SET", arcmdl.SimpleArcKey(arc.Aid), bs)
			return e
		}(); err != nil {
			log.Error("setSimpleArcCache conn.Do key(%s) k(%d) err(%+v)", arcmdl.SimpleArcKey(arc.Aid), k, err)
			return err
		}
	}
	if err = s.setTaishan(c, []byte(arcmdl.SimpleArcKey(arc.Aid)), bs); err != nil {
		return err
	}
	return nil
}

func (s *Service) getArcPremiereInfo(c context.Context, aid int64) (*api.Premiere, error) {
	expand, err := s.resultDao.RawArchiveExpand(c, aid)
	if err != nil {
		return nil, err
	}
	if expand == nil {
		return nil, nil
	}
	return &api.Premiere{
		StartTime: expand.PremiereTime.Unix(),
		RoomId:    expand.RoomId,
	}, nil
}

func (s *Service) setArcPayInfo(c context.Context, sa *api.Arc) error {
	if sa.SeasonID != 0 && sa.Pay.AttrVal(api.PaySubTypeAttrBitSeason) == api.AttrYes {
		episode, err := s.resultDao.RawSeasonEpisode(c, sa.SeasonID, sa.Aid)
		if err != nil {
			return err
		}
		if episode == nil {
			log.Error("日志告警 付费稿件 未查询到episode信息 sa(%v)", sa)
		} else {
			sa.Rights.ArcPayFreeWatch = episode.AttrVal(ugcmdl.EpisodeAttrSnFreeWatch)
		}
	}
	return nil
}

func (s *Service) setSimpleArcPayInfo(c context.Context, sa *api.SimpleArc) error {
	_, _, subType, err := s.resultDao.RawAdditV2(c, sa.Aid)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	sa.Pay = &api.PayInfo{
		PayAttr: subType,
	}
	if sa.Pay.AttrVal(api.PaySubTypeAttrBitSeason) == api.AttrYes {
		episode, err := s.resultDao.RawSeasonEpisode(c, sa.SeasonId, sa.Aid)
		if err != nil {
			return err
		}
		if episode != nil {
			sa.Rights.ArcPayFreeWatch = episode.AttrVal(ugcmdl.EpisodeAttrSnFreeWatch)
		}
	}
	return nil
}
