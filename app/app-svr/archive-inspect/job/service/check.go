package service

import (
	"bytes"
	"context"
	"fmt"
	"go-common/library/log"
	ugcmdl "go-gateway/app/app-svr/ugc-season/service/api"

	"go-gateway/app/app-svr/archive/job/model/archive"
	"go-gateway/app/app-svr/archive/service/api"
	arcmdl "go-gateway/app/app-svr/archive/service/model"
)

const _processCnt = 50

func (s *Service) checkModifyAids() {
	//select last 10 minutes modified aids
	avs, err := s.resultDao.LastMinuteAvs(context.Background())
	if err != nil {
		log.Error("checkEqual s.resultDao.LastMinuteAvs err(%+v)", err)
		return
	}
	log.Info("checkEqual len(%d)", len(avs))
	for i := 0; i < len(avs); i += _processCnt {
		var partAvs []int64
		if i+_processCnt > len(avs) {
			partAvs = avs[i:]
		} else {
			partAvs = avs[i : i+_processCnt]
		}
		errMsg := s.MultiCheck(partAvs)
		if errMsg != "" {
			log.Error("checkEqual 日志告警：(%s)", errMsg)
		}
	}
}

func (s *Service) MultiCheck(partAvs []int64) string {
	var errMsg string
	//实时查询，减少因时间延误导致的误判
	avRly, err := s.resultDao.ArcInfoAvs(context.Background(), partAvs)
	if err != nil {
		return fmt.Sprintf("%s s.resultDao.ArcInfoAvs err(%+v) aids(%+v)", errMsg, err, partAvs)
	}
	log.Info("checkEqual in(%d) out(%d)", len(partAvs), len(avRly))
	for _, val := range avRly {
		if val == nil || val.Arc == nil {
			continue
		}
		av := val.Arc
		c := context.Background()
		log.Info("checkEqual aid(%d)", av.Aid)
		s.fillArc(c, av, val.Ip)
		originAv, err := s.archiveDao.GrpcRawArchive(c, av.Aid)
		if err != nil {
			errMsg = fmt.Sprintf("%s s.archiveDao.GrpcRawArchive err(%+v) aid(%d)", errMsg, err, av.Aid)
			continue
		}
		originVs, err := s.archiveDao.GrpcRawVideos(c, av.Aid)
		if err != nil {
			errMsg = fmt.Sprintf("%s s.archiveDao.GrpcRawVideos err(%+v) aid(%d)", errMsg, err, av.Aid)
			continue
		}
		vs, err := s.resultDao.RawVideos(c, av.Aid)
		if err != nil {
			errMsg = fmt.Sprintf("%s s.resultDao.RawVideos err(%+v) aid(%d)", errMsg, err, av.Aid)
			continue
		}
		// check DB
		if errStr := s.AvDBEqual(av, originAv, vs, originVs); errStr != "" {
			errMsg = fmt.Sprintf("%s s.AvDBEqual %s aid(%d)", errMsg, errStr, av.Aid)
			continue
		}
		// check redis & taishan
		if errStr := s.AvCacheEqual(c, av, vs); errStr != "" {
			errMsg = fmt.Sprintf("%s s.AvCacheEqual %s aid(%d)", errMsg, errStr, av.Aid)
			continue
		}
	}
	return errMsg
}

func (s *Service) AvCacheEqual(c context.Context, av *api.Arc, vs []*api.Page) string {
	avBs, err := av.Marshal()
	if err != nil {
		return fmt.Sprintf("av.Marshal aid(%d) err(%+v)", av.Aid, err)
	}
	view := &api.AidVideos{Aid: av.Aid, Pages: vs}
	vsBs, err := view.Marshal()
	if err != nil {
		return fmt.Sprintf("view.Marshal aid(%d) err(%+v)", av.Aid, err)
	}
	// 检查redis数据
	avRds, viewRds, err := s.getCache(c, av.Aid)
	if err != nil {
		return fmt.Sprintf("s.getCache aid(%d) key(%s) err(%+v)", av.Aid, arcmdl.ArcKey(av.Aid), err)
	}
	for k, rdsBs := range avRds {
		if bytes.Equal(rdsBs, avBs) {
			continue
		}
		var rs = &api.Arc{}
		err := rs.Unmarshal(rdsBs)
		return fmt.Sprintf("arc in db and redis is not equal k(%d) aid(%d) db(%+v) redis(%+v) err(%+v)", k, av.Aid, av, rs, err)
	}
	for k, rdsBs := range viewRds {
		if bytes.Equal(rdsBs, vsBs) {
			continue
		}
		var rs = &api.AidVideos{}
		err = rs.Unmarshal(rdsBs)
		return fmt.Sprintf("view in db and redis is not equal k(%d) aid(%d) db(%+v) redis(%+v) err(%+v)", k, av.Aid, view, rs, err)
	}
	// 检查taishan数据
	avTs, err := s.getFromTaishan(c, arcmdl.ArcKey(av.Aid))
	if err != nil {
		return fmt.Sprintf("s.getFromTaishan aid(%d) key(%s) err(%+v)", av.Aid, arcmdl.ArcKey(av.Aid), err)
	}
	if !bytes.Equal(avTs, avBs) {
		var ts = &api.Arc{}
		err = ts.Unmarshal(avTs)
		return fmt.Sprintf("arc in db and taishan is not equal aid(%d) db(%+v) taishan(%+v) err(%+v)", av.Aid, av, ts, err)
	}
	viewTs, err := s.getFromTaishan(c, arcmdl.PageKey(av.Aid))
	if err != nil {
		return fmt.Sprintf("s.getFromTaishan aid(%d) key(%s) err(%+v)", av.Aid, arcmdl.PageKey(av.Aid), err)
	}
	if !bytes.Equal(viewTs, vsBs) {
		var ts = &api.AidVideos{}
		err = ts.Unmarshal(viewTs)
		return fmt.Sprintf("view in db and taishan is not equal aid(%d) db(%+v) taishan(%+v) err(%+v)", av.Aid, vs, ts, err)
	}
	return ""
}

func (s *Service) AvDBEqual(av, originAv *api.Arc, vs []*api.Page, originVs []*archive.Video) string {
	var errMsg string
	// check av
	if originAv.State == api.StateForbidFixed || originAv.State == api.StateForbidSubmit {
		return errMsg
	}
	if av.State != originAv.State {
		errMsg = fmt.Sprintf("%s 稿件状态异常(r:%d,o:%d)", errMsg, av.State, originAv.State)
	}
	if av.Title != originAv.Title {
		errMsg = fmt.Sprintf("%s 稿件标题异常(r:%s,o:%s)", errMsg, av.Title, originAv.Title)
	}
	if av.Author.Mid != originAv.Author.Mid {
		errMsg = fmt.Sprintf("%s 稿件up主异常(r:%d,o:%d)", errMsg, av.Author.Mid, originAv.Author.Mid)
	}
	// 年报、动态视频、愚人节等先发后审稿件会有回查，高清H5标志位可能会不一致
	if av.Attribute != originAv.Attribute && (av.Attribute+256 != originAv.Attribute) {
		errMsg = fmt.Sprintf("%s 稿件attribute异常(r:%d,o:%d)", errMsg, av.Attribute, originAv.Attribute)
	}
	if av.TypeID != originAv.TypeID {
		errMsg = fmt.Sprintf("%s 稿件分区异常(r:%d,o:%d)", errMsg, av.TypeID, originAv.TypeID)
	}
	//稿件不可见不校验分p
	if av.State < 0 {
		return errMsg
	}
	// check vs
	validVs := make(map[int64]*archive.Video)
	for _, v := range originVs {
		if v.Cid > 0 && (v.Status == archive.VideoStatusAccess || v.Status == archive.VideoStatusOpen) && v.State == archive.VideoRelationBind {
			validVs[v.Cid] = v
		}
	}
	if len(validVs) != len(vs) {
		errMsg = fmt.Sprintf("%s 稿件分p异常(len(r):%d,len(o):%d)", errMsg, len(vs), len(validVs))
	}
	for _, v := range vs {
		vv, ok := validVs[v.Cid]
		if !ok {
			errMsg = fmt.Sprintf("%s 稿件分p异常 cid=%d is not exist", errMsg, v.Cid)
			continue
		}
		if v.Part != vv.Title {
			errMsg = fmt.Sprintf("%s 稿件分p标题异常(r:%s,o:%s) cid=%d", errMsg, v.Part, vv.Title, v.Cid)
		}
		if v.Duration != vv.Duration {
			errMsg = fmt.Sprintf("%s 稿件分p时长异常(r:%d,o:%d) cid=%d", errMsg, v.Duration, vv.Duration, v.Cid)
		}
	}
	return errMsg
}

// fillArc is
func (s *Service) fillArc(c context.Context, a *api.Arc, ip string) {
	a.Fill() // set attribute and rights
	func() { // set 联合投稿
		if a.AttrVal(api.AttrBitIsCooperation) != api.AttrYes {
			return
		}
		staffs, err := s.resultDao.RawStaff(c, a.Aid)
		if err != nil {
			log.Error("日志报警 fillArc aid(%d) error(%+v)", a.Aid, err)
			return
		}
		a.StaffInfo = staffs
	}()
	func() { //set PubLocation
		if ip != "" {
			ipstr, err := s.locDao.Info2Special(c, ip)
			if err != nil {
				log.Error("日志报警 fillArc aid(%d) Info2Special error(%+v)", a.Aid, err)
				return
			}
			a.PubLocation = ipstr
		}
	}()
	func() {
		if a.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes {
			p, err := s.resultDao.RawArchiveExpand(c, a.Aid)
			if err != nil {
				log.Error("日志报警 fillArc aid(%d) RawArchiveExpand error(%+v)", a.Aid, err)
				return
			}
			if p == nil {
				return
			}
			a.Premiere = &api.Premiere{
				StartTime: p.PremiereTime.Unix(),
				RoomId:    p.RoomId,
			}
		}
	}()
	func() {
		//设置付费属性
		if a.AttrValV2(api.AttrBitV2Pay) != api.AttrYes {
			return
		}
		_, _, subType, err := s.resultDao.RawAdditV2(c, a.Aid)
		if err != nil {
			return
		}
		if a.Pay == nil {
			a.Pay = &api.PayInfo{
				PayAttr: subType,
			}
		} else {
			a.Pay.PayAttr = subType
		}
		if a.SeasonID != 0 && a.Pay.AttrVal(api.PaySubTypeAttrBitSeason) == api.AttrYes {
			episode, err := s.resultDao.RawSeasonEpisode(c, a.SeasonID, a.Aid)
			if err != nil {
				return
			}
			if episode != nil {
				a.Rights.ArcPayFreeWatch = episode.AttrVal(ugcmdl.EpisodeAttrSnFreeWatch)
			}
		}
	}()
}
