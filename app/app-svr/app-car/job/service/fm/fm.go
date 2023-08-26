package fm

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-car/job/conf"
	"go-gateway/app/app-svr/app-car/job/model/fm"

	"github.com/pkg/errors"
)

const _maxOidCount = 300 // 合集中最大稿件数

func (s *Service) initFmSeasonInputer(c *conf.Config) {
	inputer := railgun.NewKafkaInputer(c.FmSeasonGun.KafkaCfg)
	processor := railgun.NewSingleProcessor(c.FmSeasonGun.SingleConfig, s.unpackFmSeason, s.upsertSeason)
	s.fmSeasonGun = railgun.NewRailGun("FmSeasonKafkaInputer", c.FmSeasonGun.Cfg, inputer, processor)
	s.fmSeasonGun.Start()
	log.Warn("initFmSeasonInputer success!")
}

func (s *Service) unpackFmSeason(msg railgun.Message) (single *railgun.SingleUnpackMsg, err error) {
	log.Warn("unpackFmSeason receive msg: %+v", string(msg.Payload()))
	tmp := &struct {
		Scene fm.Scene `json:"scene"`
	}{}
	err = json.Unmarshal(msg.Payload(), tmp)
	if err != nil {
		log.Error("unpackFmSeason fail, scene unmarshal error, str(%s) err(%+v)", string(msg.Payload()), err)
		return nil, err
	}
	if tmp.Scene == "" {
		tmp.Scene = fm.SceneFm // 兼容旧的schema
	}

	var (
		season  = &fm.CommonSeason{Scene: tmp.Scene}
		groupId int64
	)
	switch tmp.Scene {
	case fm.SceneFm:
		fmSeason := new(fm.FmSeason)
		err = json.Unmarshal(msg.Payload(), fmSeason)
		if err != nil {
			log.Error("unpackFmSeason fail, fmSeason unmarshal error, str(%s) err(%+v)", string(msg.Payload()), err)
			return nil, err
		}
		if len(fmSeason.FmListStr) > 0 && fmSeason.FmListStr != "null" {
			fmSeason.FmList = make([]*fm.Item, 0)
			err = json.Unmarshal([]byte(fmSeason.FmListStr), &fmSeason.FmList)
			if err != nil {
				log.Error("unpackFmSeason fail, fmListStr unmarshal error, str(%s) err(%+v)", fmSeason.FmListStr, err)
				return nil, err
			}
			if len(fmSeason.FmList) >= _maxOidCount {
				log.Error("unpackFmSeason fail, too many oids, fm_type:%s, fm_id:%d, count:%d", fmSeason.FmType, fmSeason.FmId, len(fmSeason.FmList))
				return nil, ecode.RequestErr
			}
			if fmSeason.Scene == "" {
				fmSeason.Scene = fm.SceneFm
			}
		}
		groupId = fmSeason.FmId
		season.Fm = *fmSeason
	case fm.SceneVideo:
		videoSeason := new(fm.VideoSeason)
		err = json.Unmarshal(msg.Payload(), videoSeason)
		if err != nil {
			log.Error("unpackFmSeason fail, videoSeason unmarshal error, str(%s) err(%+v)", string(msg.Payload()), err)
			return nil, err
		}
		if len(videoSeason.SeasonListStr) > 0 && videoSeason.SeasonListStr != "null" {
			videoSeason.SeasonList = make([]*fm.Item, 0)
			err = json.Unmarshal([]byte(videoSeason.SeasonListStr), &videoSeason.SeasonList)
			if err != nil {
				log.Error("unpackFmSeason fail, seasonListStr unmarshal error, str(%s) err(%+v)", videoSeason.SeasonListStr, err)
				return nil, err
			}
			if len(videoSeason.SeasonList) >= _maxOidCount {
				log.Error("unpackFmSeason fail, too many oids, season_id:%d, count:%d", videoSeason.SeasonId, len(videoSeason.SeasonList))
				return nil, ecode.RequestErr
			}
		}
		groupId = videoSeason.SeasonId
		season.Video = *videoSeason
	default:
		log.Error("unpackFmSeason unknown scene:%s", tmp.Scene)
		return nil, errors.Wrapf(ecode.RequestErr, "unknown scene:%s", tmp.Scene)
	}
	return &railgun.SingleUnpackMsg{
		Group: groupId,
		Item:  season,
	}, nil
}

// upsertSeason 插入或更新FM合集
func (s *Service) upsertSeason(ctx context.Context, item interface{}) railgun.MsgPolicy {
	var (
		season *fm.CommonSeason
		policy railgun.MsgPolicy
		ok     bool
		err    error
	)
	season, ok = item.(*fm.CommonSeason)
	if !ok {
		return railgun.MsgPolicyIgnore
	}
	// 1. 生成合集封面
	policy, ok = s.generateCover(ctx, season)
	if !ok {
		return policy
	}
	// 2. 导入db
	if season.Scene == fm.SceneFm {
		policy, err = s.fmDao.UpsertSeasonWithLock(ctx, season, s.upsertFmSeasonDb)
	} else if season.Scene == fm.SceneVideo {
		policy, err = s.fmDao.UpsertSeasonWithLock(ctx, season, s.upsertVideoSeasonDb)
	}

	if err != nil {
		return policy
	}
	// 3. 删除redis缓存
	for i := 0; i < 3; i++ {
		if season.Scene == fm.SceneFm {
			err = s.fmDao.DelSeasonCache(ctx, season.Scene, season.Fm.FmType, season.Fm.FmId)
		} else if season.Scene == fm.SceneVideo {
			err = s.fmDao.DelSeasonCache(ctx, season.Scene, "", season.Video.SeasonId)
		}
		if err == nil {
			break
		}
		log.Error("upsertSeason fail, s.fmDao.DelSeasonCache err:%+v, season:%+v", err, season)
		time.Sleep(50 * time.Millisecond)
	}
	return railgun.MsgPolicyNormal
}

// generateCover 生成合集封面（若算法侧未指定）
func (s *Service) generateCover(ctx context.Context, season *fm.CommonSeason) (railgun.MsgPolicy, bool) {
	switch season.Scene {
	case fm.SceneFm:
		if len(season.Fm.Cover) == 0 {
			if season.Fm.FmType == fm.AudioSeason {
				if len(season.Fm.FmList) == 0 || season.Fm.FmList[0] == nil {
					log.Error("generateCover AudioSeason fail, fmList empty, fmId:%d", season.Fm.FmId)
					return railgun.MsgPolicyIgnore, false
				}
				arc, err := s.fmDao.Arc(ctx, season.Fm.FmList[0].Aid)
				if err != nil {
					log.Error("generateCover AudioSeason fail, s.fmDao.Arc err:%+v, aid:%d", err, season.Fm.FmList[0].Aid)
					return railgun.MsgPolicyAttempts, false
				}
				season.Fm.Cover = arc.Pic
			} else if season.Fm.FmType == fm.AudioSeasonUp {
				account, err := s.fmDao.Profile3(ctx, season.Fm.FmId/100) // up主合集id规则：up主id * 100 + 编号0~99
				if err != nil {
					log.Error("generateCover AudioSeasonUp fail, s.fmDao.Profile3 err:%+v, fmId:%d, upMid:%d", err, season.Fm.FmId, season.Fm.FmId/100)
					return railgun.MsgPolicyAttempts, false
				}
				season.Fm.Cover = account.Face
			} else {
				log.Error("generateCover fail, unknown fmType:%s, season:%+v", season.Fm.FmType, season.Fm)
				return railgun.MsgPolicyIgnore, false
			}
		}
	case fm.SceneVideo:
		if len(season.Video.Cover) == 0 {
			if len(season.Video.SeasonList) == 0 || season.Video.SeasonList[0] == nil {
				log.Error("generateCover VideoSeason fail, seasonList empty, seasonId:%d", season.Video.SeasonId)
				return railgun.MsgPolicyIgnore, false
			}
			arc, err := s.fmDao.Arc(ctx, season.Video.SeasonList[0].Aid)
			if err != nil {
				log.Error("generateCover VideoSeason fail, s.fmDao.Arc err:%+v, aid:%d", err, season.Video.SeasonList[0].Aid)
				return railgun.MsgPolicyAttempts, false
			}
			season.Video.Cover = arc.Pic
		}
	default:
		log.Error("generateCover unknown scene:%s", season.Scene)
		return railgun.MsgPolicyIgnore, false
	}
	return railgun.MsgPolicyNormal, true
}

// upsertFmSeasonDb FM合集导入DB
func (s *Service) upsertFmSeasonDb(ctx context.Context, season *fm.CommonSeason) (railgun.MsgPolicy, error) {
	// 1. 查询数据库中是否有当前合集
	var (
		exist    = true  // 是否有当前合集
		same     = false // 新合集信息与库中信息是否相同
		err      error
		infoResp *fm.SeasonInfoResp
		oids     []*fm.FmSeasonOidPo
	)
	infoResp, err = s.fmDao.QuerySeasonInfo(ctx, fm.SeasonInfoReq{Scene: season.Scene, FmType: season.Fm.FmType, SeasonId: season.Fm.FmId})
	if err != nil {
		if err == ecode.NothingFound {
			exist = false
		} else {
			log.Error("upsertFmSeasonDb fail, s.fmDao.QuerySeasonInfo err:%+v, scene:%s, fm:%+v, video:%+v", err, season.Scene, season.Fm, season.Video)
			return railgun.MsgPolicyAttempts, err
		}
	}
	if infoResp == nil || (infoResp.Scene == fm.SceneFm && infoResp.Fm == nil) {
		exist = false
	}
	if exist {
		// 2.1 存在原合集，则进行比对（比对结果相同，无需改动）
		oids, _, err = s.fmDao.QuerySeasonOid(ctx, season.Scene, season.Fm.FmType, season.Fm.FmId)
		if err == nil && len(oids) > 0 {
			same = s.compareFmSeason(infoResp.Fm, oids, &season.Fm)
		}
		if !same {
			originOidsBytes, _ := json.Marshal(oids)
			seasonBytes, _ := json.Marshal(season.Fm.FmList)
			log.Warn("upsertFmSeasonDb s.compareFmSeason not same, origin info:%+v, origin oids:%s, latest season:%s", infoResp.Fm, string(originOidsBytes), string(seasonBytes))
			// 3.2 比对结果不同，删除原合集后插入
			err = s.fmDao.ModifySeasonWithTx(ctx, season, fm.TypeUpdate)
		}
	} else {
		// 2.2 不存在原合集，则直接插入
		err = s.fmDao.ModifySeasonWithTx(ctx, season, fm.TypeInsert)
	}
	if err != nil {
		log.Error("upsertFmSeasonDb fail, s.fmDao.ModifySeasonWithTx err:%+v, isUpdate:%t, season:%+v", err, exist, season)
		return railgun.MsgPolicyAttempts, err
	}
	return railgun.MsgPolicyNormal, nil
}

// upsertVideoSeasonDb 视频合集导入DB
func (s *Service) upsertVideoSeasonDb(ctx context.Context, season *fm.CommonSeason) (railgun.MsgPolicy, error) {
	// 1. 查询数据库中是否有当前合集
	var (
		exist    = true  // 是否有当前合集
		same     = false // 新合集信息与库中信息是否相同
		err      error
		infoResp *fm.SeasonInfoResp
		oids     []*fm.VideoSeasonOidPo
	)
	infoResp, err = s.fmDao.QuerySeasonInfo(ctx, fm.SeasonInfoReq{Scene: season.Scene, SeasonId: season.Video.SeasonId})
	if err != nil {
		if err == ecode.NothingFound {
			exist = false
		} else {
			log.Error("upsertVideoSeasonDb fail, s.fmDao.QuerySeasonInfo err:%+v, scene:%s, fm:%+v, video:%+v", err, season.Scene, season.Fm, season.Video)
			return railgun.MsgPolicyAttempts, err
		}
	}
	if infoResp == nil || (infoResp.Scene == fm.SceneVideo && infoResp.Video == nil) {
		exist = false
	}
	if exist {
		// 2.1 存在原合集，则进行比对（比对结果相同，无需改动）
		_, oids, err = s.fmDao.QuerySeasonOid(ctx, season.Scene, "", season.Video.SeasonId)
		if err == nil && len(oids) > 0 {
			same = s.compareVideoSeason(infoResp.Video, oids, &season.Video)
		}
		if !same {
			originOidsBytes, _ := json.Marshal(oids)
			seasonBytes, _ := json.Marshal(season.Video.SeasonList)
			log.Warn("upsertVideoSeasonDb s.compareVideoSeason not same, origin info:%+v, origin oids:%s, latest season:%s", infoResp.Video, string(originOidsBytes), string(seasonBytes))
			// 3.2 比对结果不同，删除原合集后插入
			err = s.fmDao.ModifySeasonWithTx(ctx, season, fm.TypeUpdate)
		}
	} else {
		// 2.2 不存在原合集，则直接插入
		err = s.fmDao.ModifySeasonWithTx(ctx, season, fm.TypeInsert)
	}
	if err != nil {
		log.Error("upsertVideoSeasonDb fail, s.fmDao.ModifySeasonWithTx err:%+v, isUpdate:%t, season:%+v", err, exist, season)
		return railgun.MsgPolicyAttempts, err
	}
	return railgun.MsgPolicyNormal, nil
}

// compareFmSeason 新老FM合集比对（老合集：info/oids 新合集：latest）
// 比对内容包含：合集标题、合集封面、稿件池oid、稿件排序
func (s *Service) compareFmSeason(info *fm.FmSeasonInfoPo, oids []*fm.FmSeasonOidPo, latest *fm.FmSeason) bool {
	if info.FmId != latest.FmId || info.FmType != string(latest.FmType) || info.Title != latest.Title || info.Cover != latest.Cover {
		return false
	}
	if len(oids) != len(latest.FmList) {
		return false
	}
	oidsDeepCopy := make([]*fm.FmSeasonOidPo, 0)
	for _, v := range oids {
		oidsDeepCopy = append(oidsDeepCopy, &fm.FmSeasonOidPo{
			Oid: v.Oid,
			Seq: v.Seq,
		})
	}
	sort.SliceStable(oidsDeepCopy, func(i, j int) bool {
		return oidsDeepCopy[i].Seq < oidsDeepCopy[j].Seq
	})
	for i, v := range oidsDeepCopy {
		if latest.FmList[i] == nil || latest.FmList[i].Aid != v.Oid {
			return false
		}
	}
	return true
}

// compareFmSeason 新老FM合集比对（老合集：info/oids 新合集：latest）
// 比对内容包含：合集标题、合集封面、稿件池oid、稿件排序
func (s *Service) compareVideoSeason(info *fm.VideoSeasonInfoPo, oids []*fm.VideoSeasonOidPo, latest *fm.VideoSeason) bool {
	if info.SeasonId != latest.SeasonId || info.Title != latest.Title || info.Cover != latest.Cover {
		return false
	}
	if len(oids) != len(latest.SeasonList) {
		return false
	}
	oidsDeepCopy := make([]*fm.VideoSeasonOidPo, 0)
	for _, v := range oids {
		oidsDeepCopy = append(oidsDeepCopy, &fm.VideoSeasonOidPo{
			Oid: v.Oid,
			Seq: v.Seq,
		})
	}
	sort.SliceStable(oidsDeepCopy, func(i, j int) bool {
		return oidsDeepCopy[i].Seq < oidsDeepCopy[j].Seq
	})
	for i, v := range oidsDeepCopy {
		if latest.SeasonList[i] == nil || latest.SeasonList[i].Aid != v.Oid {
			return false
		}
	}
	return true
}
