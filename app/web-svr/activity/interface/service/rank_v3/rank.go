package rank

import (
	"context"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank_v3"
	"go-gateway/pkg/idsafe/bvid"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"time"
)

func (s *Service) checkRank(ctx context.Context, base *rankmdl.Base, rule *rankmdl.Rule) (err error) {
	if base == nil || rule == nil {
		return ecode.RankConfigError
	}
	if rule.BaseID != base.ID {
		return ecode.RankConfigError
	}
	return nil
}

// GetRankByID 根据id获取排行榜
func (s *Service) GetRankByID(ctx context.Context, rankID int64, pn, ps int) (res *rankmdl.ResultList, err error) {
	res = &rankmdl.ResultList{}
	rule, err := s.dao.RankRule(ctx, rankID)
	if err != nil {
		log.Errorc(ctx, "s.dao.RankRule err(%v)", err)
		return
	}
	base, err := s.dao.RankBase(ctx, rule.BaseID)
	if err != nil {
		log.Errorc(ctx, "s.dao.RankBase err(%v)", err)
		return
	}
	err = s.checkRank(ctx, base, rule)
	if err != nil {
		log.Errorc(ctx, "s.checkRank baseID(%d) rankID(%d)", base.ID, rankID)
		return
	}

	result, err := s.getRankResult(ctx, base, rule)
	if err != nil {
		log.Errorc(ctx, "s.getRankResult err(%v)", err)
		return
	}
	res, err = s.getRankDetail(ctx, result, pn, ps)
	res.ShowBatchTime = rule.ShowBatchTime
	res.StatisticsType = rule.StatisticsType
	return
}

// getRankDetail ...
func (s *Service) getRankDetail(ctx context.Context, result *rankmdl.ResultRank, pn, ps int) (res *rankmdl.ResultList, err error) {
	if result == nil || result.List == nil {
		return
	}
	res = &rankmdl.ResultList{}
	res.Page = &rankmdl.Page{Num: pn, Size: ps}
	res.List = make([]*rankmdl.ResultRes, 0)
	res.Page.Total = len(result.List)
	start := (pn - 1) * ps
	if start >= len(result.List) {
		return
	}
	newRes := result.List[start:]
	end := start + ps

	if end < len(result.List) {
		newRes = result.List[start:end]
	}
	list := make([]*rankmdl.ResultRes, 0)
	if result.StatisticsType == rankmdl.StatisticsTypeArchive || result.StatisticsType == rankmdl.StatisticsTypeDistinctArchive {
		list, err = s.getRankArchive(ctx, newRes)
		if err != nil {
			log.Errorc(ctx, "s.getRankArchive err(%v)", err)
		}
	}
	if result.StatisticsType == rankmdl.StatisticsTypeUp {
		list, err = s.getRankUp(ctx, newRes)
		if err != nil {
			log.Errorc(ctx, "s.getRankUp err(%v)", err)
		}
	}
	if result.StatisticsType == rankmdl.StatisticsTypeTag {
		list, err = s.getRankTag(ctx, newRes)
		if err != nil {
			log.Errorc(ctx, "s.getRankTag err(%v)", err)
		}
	}
	res.List = list
	return
}

// getRankArchive ...
func (s *Service) getRankArchive(ctx context.Context, list []*rankmdl.ResultRankList) (res []*rankmdl.ResultRes, err error) {
	res = make([]*rankmdl.ResultRes, 0)
	aids := make([]int64, 0)

	for _, v := range list {
		aids = append(aids, v.OID)

	}
	var archive map[int64]*api.Arc
	if len(aids) > 0 {
		archive, err = s.archive.AllArchiveInfo(ctx, aids)
		if err != nil {
			log.Errorc(ctx, "s.archive.AllArchiveInfo err(%v)", err)
			return
		}
	}

	for _, v := range list {
		arcList := make([]*rankmdl.Archive, 0)
		account := &rankmdl.Account{}
		if arc, ok := archive[v.OID]; ok {
			arcInfo := s.archiveToArchive(arc, v.Score)
			account = s.archiveAccountToAccount(arc)

			if arcInfo == nil || account == nil {
				continue
			}
			arcList = append(arcList, arcInfo)
			res = append(res, &rankmdl.ResultRes{
				Archive:    arcList,
				Account:    account,
				ObjectType: rankmdl.ObjectTypeArchive,
				Score:      0,
				ShowScore:  v.ShowScore,
			})
		}

	}
	return res, nil

}

func (s *Service) getRankUp(ctx context.Context, list []*rankmdl.ResultRankList) (res []*rankmdl.ResultRes, err error) {
	res = make([]*rankmdl.ResultRes, 0)
	aids := make([]int64, 0)
	mids := make([]int64, 0)
	for _, v := range list {
		mids = append(mids, v.OID)
		if v.AID != nil && len(v.AID) > 0 {
			for _, aid := range v.AID {
				aids = append(aids, aid.AID)
			}
		}
	}
	var archiveAll map[int64]*api.Arc
	var accountAll map[int64]*accountapi.Card
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if len(aids) > 0 {
			archiveAll, err = s.archive.AllArchiveInfo(ctx, aids)
			if err != nil {
				log.Errorc(ctx, "s.archive.AllArchiveInfo err(%v)", err)
				return
			}
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if len(mids) > 0 {
			accountAll, err = s.account.CardsInfo(ctx, mids)
			if err != nil {
				log.Errorc(ctx, "s.account.MemberInfo err(%v)", err)
				return
			}
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return
	}

	for _, v := range list {
		arcList := make([]*rankmdl.Archive, 0)
		account := &rankmdl.Account{}
		if len(v.AID) > 0 {

			for _, a := range v.AID {

				if arcInfo, ok := archiveAll[a.AID]; ok {
					arc := s.archiveToArchive(arcInfo, a.Score)
					if arc == nil {
						continue
					}
					arcList = append(arcList, arc)
				}

			}
		}
		if acc, ok := accountAll[v.OID]; ok {
			account = s.accountToAccount(acc)
			if account == nil {
				continue
			}
		}
		res = append(res, &rankmdl.ResultRes{
			Archive:    arcList,
			Account:    account,
			ObjectType: rankmdl.ObjectTypeUp,
			Score:      0,
			ShowScore:  v.ShowScore,
		})
	}
	return res, nil

}

func (s *Service) getRankTag(ctx context.Context, list []*rankmdl.ResultRankList) (res []*rankmdl.ResultRes, err error) {
	res = make([]*rankmdl.ResultRes, 0)
	aids := make([]int64, 0)
	tagIDs := make([]int64, 0)
	for _, v := range list {
		tagIDs = append(tagIDs, v.OID)
		if v.AID != nil && len(v.AID) > 0 {
			for _, aid := range v.AID {
				aids = append(aids, aid.AID)
			}
		}

	}
	var archiveAll map[int64]*api.Arc
	var tagAll map[int64]*tagrpc.Tag
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if len(aids) > 0 {
			archiveAll, err = s.archive.AllArchiveInfo(ctx, aids)
			if err != nil {
				log.Errorc(ctx, "s.archive.ArchiveInfo err(%v)", err)
				return
			}
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if len(tagIDs) > 0 {
			tagAll, err = s.tag.TagInfo(ctx, tagIDs)
			if err != nil {
				log.Errorc(ctx, "s.account.MemberInfo err(%v)", err)
				return
			}
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return
	}

	for _, v := range list {
		arcList := make([]*rankmdl.Archive, 0)
		tag := &rankmdl.Tag{}
		if v.AID == nil {
			continue
		}
		for _, a := range v.AID {

			if arcInfo, ok := archiveAll[a.AID]; ok {
				arc := s.archiveToArchive(arcInfo, a.Score)
				if arc == nil {
					continue
				}
				arcList = append(arcList, arc)
			}

		}
		if len(arcList) == 0 {
			continue
		}
		if t, ok := tagAll[v.OID]; ok {
			tag = s.tagToTag(t)
			if tag == nil {
				continue
			}
		}
		res = append(res, &rankmdl.ResultRes{
			Archive:    arcList,
			ObjectType: rankmdl.ObjectTypeTag,
			Score:      0,
			ShowScore:  v.ShowScore,
			Tag:        tag,
		})
	}
	return res, nil

}

func (s *Service) accountToAccount(acc *accountapi.Card) *rankmdl.Account {
	if acc == nil {
		return nil
	}
	res := &rankmdl.Account{
		Mid:      acc.Mid,
		Name:     acc.Name,
		Face:     acc.Face,
		Official: acc.Official,
		VipInfo:  acc.Vip,
	}
	return res
}

func (s *Service) tagToTag(tag *tagrpc.Tag) *rankmdl.Tag {
	if tag == nil {
		return nil
	}
	return &rankmdl.Tag{TID: tag.Id, Name: tag.Name}
}
func (s *Service) archiveAccountToAccount(arc *api.Arc) *rankmdl.Account {

	res := &rankmdl.Account{
		Mid:  arc.Author.Mid,
		Name: arc.Author.Name,
		Face: arc.Author.Face,
	}
	return res
}

func (s *Service) archiveToArchive(arc *api.Arc, score int64) *rankmdl.Archive {
	var bvidStr string
	var err error
	if bvidStr, err = bvid.AvToBv(arc.Aid); err != nil || bvidStr == "" {
		return nil
	}
	if err != nil {
		return nil
	}
	account := s.archiveAccountToAccount(arc)
	res := &rankmdl.Archive{
		Bvid:     bvidStr,
		TypeName: arc.TypeName,
		Title:    arc.Title,
		Desc:     arc.Desc,
		Duration: arc.Duration,
		Pic:      arc.Pic,
		View:     arc.Stat.View,
		Like:     arc.Stat.Like,
		Danmaku:  arc.Stat.Danmaku,
		Reply:    arc.Stat.Reply,
		Fav:      arc.Stat.Fav,
		Coin:     arc.Stat.Coin,
		Share:    arc.Stat.Share,
		Score:    0,
		PubDate:  arc.PubDate,
		ShowLink: arc.ShortLinkV2,
		Account:  account,
	}
	return res
}

func (s *Service) getRankResult(ctx context.Context, base *rankmdl.Base, rule *rankmdl.Rule) (result *rankmdl.ResultRank, err error) {
	result, ok := s.getRankResultByRule(base.ID, rule.ID)
	if ok && result.Batch == rule.ShowBatch {
		return result, nil
	}

	// 从db获取
	result, err = s.getRankResultFromDB(ctx, base, rule)
	if err != nil {
		log.Errorc(ctx, "s.getRankResultFromDB err(%v)", err)
		return
	}
	s.setRankResultByRule(base.ID, rule.ID, result)
	return result, nil

}

// getRankResultFromDB ...
func (s *Service) getRankResultFromDB(ctx context.Context, base *rankmdl.Base, rule *rankmdl.Rule) (result *rankmdl.ResultRank, err error) {
	eg := errgroup.WithContext(ctx)
	var (
		oidResult []*rankmdl.ResultOid
		aidResult []*rankmdl.ResultOidArchive
	)
	eg.Go(func(ctx context.Context) (err error) {
		oidResult, err = s.dao.GetRankOid(ctx, base.ID, rule.ID, rule.ShowBatch)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetRankOid err(%v)", err)
			return
		}
		return nil
	})
	if rule.StatisticsType == rankmdl.StatisticsTypeTag || rule.StatisticsType == rankmdl.StatisticsTypeUp {
		eg.Go(func(ctx context.Context) (err error) {
			aidResult, err = s.dao.GetRankArchive(ctx, base.ID, rule.ID, rule.ShowBatch)
			if err != nil {
				log.Errorc(ctx, "s.dao.GetRankOid err(%v)", err)
				return
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return
	}

	mapOidArchive := make(map[int64][]*rankmdl.ResultOidArchive)
	if aidResult != nil && len(aidResult) > 0 {
		for _, v := range aidResult {
			if _, ok := mapOidArchive[v.OID]; ok {
				mapOidArchive[v.OID] = append(mapOidArchive[v.OID], v)
				continue
			}
			mapOidArchive[v.OID] = make([]*rankmdl.ResultOidArchive, 0)
			mapOidArchive[v.OID] = append(mapOidArchive[v.OID], v)
		}
	}
	result = &rankmdl.ResultRank{
		BaseID:         base.ID,
		RankID:         rule.ID,
		Batch:          rule.ShowBatch,
		StatisticsType: rule.StatisticsType,
		AddTime:        time.Now().Unix(),
	}
	list := make([]*rankmdl.ResultRankList, 0)
	if oidResult != nil && len(oidResult) > 0 {
		for _, v := range oidResult {
			aid := make([]*rankmdl.ResultOidScore, 0)
			if archive, ok := mapOidArchive[v.OID]; ok {
				for _, v := range archive {
					aid = append(aid, &rankmdl.ResultOidScore{
						AID:       v.AID,
						Score:     v.Score,
						ShowScore: v.ShowScore,
					})
				}
			}
			if rule.StatisticsType == rankmdl.StatisticsTypeTag {
				if len(aid) > 0 {
					list = append(list, &rankmdl.ResultRankList{
						OID:       v.OID,
						Score:     v.Score,
						ShowScore: v.ShowScore,
						Rank:      v.Rank,
						AID:       aid,
					})
				}
			}
			if rule.StatisticsType == rankmdl.StatisticsTypeUp {
				list = append(list, &rankmdl.ResultRankList{
					OID:       v.OID,
					Score:     v.Score,
					ShowScore: v.ShowScore,
					Rank:      v.Rank,
					AID:       aid,
				})
			}
			if rule.StatisticsType == rankmdl.StatisticsTypeArchive || rule.StatisticsType == rankmdl.StatisticsTypeDistinctArchive {
				list = append(list, &rankmdl.ResultRankList{
					OID:       v.OID,
					Score:     v.Score,
					ShowScore: v.ShowScore,
					Rank:      v.Rank,
					AID:       aid,
				})
			}

		}
	}
	result.List = list
	return result, nil
}
