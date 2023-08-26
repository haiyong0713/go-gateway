package rank

import (
	"context"
	"encoding/csv"
	"encoding/json"

	"bytes"
	"errors"
	"fmt"
	"go-common/library/ecode"
	"go-gateway/app/web-svr/activity/admin/component/boss"
	"go-gateway/app/web-svr/activity/admin/service/exporttask"
	"go-gateway/pkg/idsafe/bvid"
	"strconv"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"

	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank"
)

func (s *Service) hasAuthority(c context.Context, author string) bool {
	if s.c.Rank.Reviewers != nil {
		for _, v := range s.c.Rank.Reviewers {
			if author == v {
				return true
			}
		}
	}
	return false
}

func (s *Service) getRank(c context.Context, sid int64, attributeType int) (rank *rankmdl.Rank, res []*rankmdl.MidRank, err error) {
	rank, err = s.Detail(c, sid, rankmdl.AidSource)
	if err != nil || rank == nil {
		return nil, nil, err
	}
	rankName := fmt.Sprintf("%d_%d", rank.ID, attributeType)
	res, err = s.dao.GetRank(c, rankName)
	return
}

// GetRankDetail 排行榜结果
func (s *Service) GetRankDetail(c context.Context, sid int64, attributeType int, ps, pn int) (*rankmdl.ResultReply, error) {
	res := rankmdl.ResultReply{}
	replyRankBatch := make([]*rankmdl.Result, 0)
	var rank []*rankmdl.MidRank
	var config *rankmdl.Rank
	var (
		start = ((pn - 1) * ps)
		end   = start + ps
	)
	config, rank, err := s.getRank(c, sid, attributeType)
	if err != nil {
		log.Errorc(c, "s.getRank(%d) error(%v)", sid, err)
		return nil, err
	}
	page := &rankmdl.Page{}
	page.Total = len(rank)
	page.Num = pn
	page.Size = ps
	res.Page = page
	if end > len(rank) {
		end = len(rank)
	}
	rank = rank[start:end]
	var (
		memberInfo map[int64]*accountapi.Info
		archive    map[int64]*api.Arc
	)
	mids := make([]int64, 0)
	aids := make([]int64, 0)
	mapAidScore := make(map[int64]*rankmdl.AidScore)
	for _, v := range rank {
		mids = append(mids, v.Mid)
		if v.Aids != nil {
			for _, v := range v.Aids {
				aids = append(aids, v.Aid)
				mapAidScore[v.Aid] = v
			}

		}
	}
	newAids := aids
	// newAids, err := s.filterList(c, aids, subject)
	eg2 := errgroup.WithContext(c)
	eg2.Go(func(ctx context.Context) (err error) {

		memberInfo, err = s.account.MemberInfo(c, mids)
		return err
	})
	eg2.Go(func(ctx context.Context) (err error) {
		archive, err = s.archive.ArchiveInfo(c, newAids)
		return err
	})
	if err := eg2.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, err
	}
	var count int64
	for _, v := range rank {
		if count >= config.Top {
			break
		}
		replyRank := &rankmdl.Result{}
		if account, ok := memberInfo[v.Mid]; ok {
			video := []*rankmdl.Video{}
			if v.Aids != nil && len(v.Aids) > 0 {
				for _, v := range v.Aids {
					if arc, ok := archive[v.Aid]; ok {
						if score, ok1 := mapAidScore[v.Aid]; ok1 {
							if arc.IsNormal() {

								video = append(video, s.archiveInfoToRankVideoInt(arc, score.Score))
							}

						}
					}
				}
				if len(video) > 0 {
					replyRank.Videos = video
					replyRank.Account = s.accountToRankAccount(c, account)

					replyRank.Score = v.Score
					count++
					replyRankBatch = append(replyRankBatch, replyRank)

				}
			}
		}
	}
	res.Rank = replyRankBatch
	return &res, nil
}

func (s *Service) archiveInfoToRankVideoInt(arc *api.Arc, score int64) *rankmdl.Video {
	var bvidStr string
	var err error
	if bvidStr, err = bvid.AvToBv(arc.Aid); err != nil || bvidStr == "" {
		return nil
	}
	return &rankmdl.Video{
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
		Score:    score,
		PubDate:  arc.PubDate,
		Mid:      arc.Author.Mid,
	}
}

// Create rank
func (s *Service) Create(c context.Context, sid int64, sidSource int, stime, etime xtime.Time, userName string) (id int64, err error) {
	if !s.hasAuthority(c, userName) {
		return 0, ecode.Error(ecode.RequestErr, "请联系贾花朵进行配置")
	}
	var (
		tx *xsql.Tx
	)
	if tx, err = s.dao.BeginTran(c); err != nil {
		log.Errorc(c, "s.dao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() (id int64, err error) {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			err = errors.New("err")
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
				err = err1
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
		return
	}()
	if id, err = s.dao.Create(c, tx, sid, sidSource, stime, etime); err != nil {

		log.Errorc(c, "Create s.dao.Create(%v, %v, %v, %v) failed. error(%v)", sid, sidSource, stime, etime, err)
		return
	}
	return
}

// Detail get rank detail
func (s *Service) Detail(c context.Context, sid int64, sidSource int) (rank *rankmdl.Rank, err error) {
	if rank, err = s.dao.GetRankConfigBySid(c, sid, sidSource); err != nil {
		log.Errorc(c, "Add s.dao.GetRankConfigBySid() failed. error(%v)", err)
	}
	return
}

// Update rank detail
func (s *Service) Update(c context.Context, id int64, rank *rankmdl.Rank, userName string) (err error) {
	var rankOld *rankmdl.Rank
	if rankOld, err = s.dao.GetRankConfigByIDAll(c, id); err != nil {
		log.Errorc(c, "Get s.dao.GetRankConfigByID() failed. error(%v)", err)
		return
	}
	if rankOld == nil {
		return ecode.Error(ecode.RequestErr, "can not find online rank")
	}
	if int64(rankOld.Stime) < time.Now().Unix() && int64(rankOld.Stime) != 0 {
		return ecode.Error(ecode.RequestErr, "排行榜已经开始，不允许修改")
	}
	if !s.hasAuthority(c, userName) {
		return ecode.Error(ecode.RequestErr, "请联系贾花朵进行配置")
	}
	if err = s.dao.UpdateRankConfig(c, id, rank); err != nil {
		log.Errorc(c, "Update s.dao.UpdateRankConfig() failed. error(%v)", err)
	}
	return
}

// Offline rank detail
func (s *Service) Offline(c context.Context, id int64) (err error) {
	var rank *rankmdl.Rank
	if rank, err = s.dao.GetRankConfigByID(c, id); err != nil {
		log.Errorc(c, "Get s.dao.GetRankConfigByID() failed. error(%v)", err)
		return
	}
	if rank == nil {
		return errors.New("can not find online rank")
	}
	rank.State = rankmdl.RankStateOffline
	if err = s.dao.UpdateRankConfig(c, id, rank); err != nil {
		log.Errorc(c, "Update s.dao.UpdateRankConfig() failed. error(%v)", err)
	}
	return
}

// UpdateBlackOrWhite 新增黑白名单
func (s *Service) UpdateBlackOrWhite(c context.Context, id int64, intervention string) (err error) {
	if intervention != "" {
		var list []*rankmdl.Intervention
		if err := json.Unmarshal([]byte(intervention), &list); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", intervention, err)
			return err
		}
		if err = s.dao.BacthInsertOrUpdateBlackOrWhite(c, id, list); err != nil {
			log.Errorc(c, "s.BacthInsertOrUpdateBlackOrWhite(sid: %v, intervention:%v) failed. error(%v)", id, intervention, err)
			return
		}
	}
	return

}

// GetBlackOrWhite 获取黑白名单
func (s *Service) GetBlackOrWhite(c context.Context, id int64, objectType, interventionType, pn, ps int) (*rankmdl.InterventionReply, error) {
	limit := ps
	offset := (pn - 1) * ps
	res := &rankmdl.InterventionReply{}
	list, err := s.dao.AllIntervention(c, id, objectType, interventionType, offset, limit)
	if err != nil {
		return nil, err
	}
	res.List = list
	page := rankmdl.Page{}
	if page.Total, err = s.dao.AllInterventionTotal(c, id, objectType, interventionType); err != nil {
		log.Errorc(c, "s.lot.GiftDraftTotal() failed. error(%v)", err)
		return nil, err
	}
	page.Num = pn
	page.Size = ps
	res.Page = &page
	return res, nil
}

// RankResultExport 排行榜结果
func (s *Service) RankResultExport(ctx context.Context, id int64, attributeType int, username string) (err error) {
	_ = s.cache.SyncDo(ctx, func(ctx context.Context) {
		res, err := s.GetRankDetail(ctx, id, attributeType, 1000, 1)
		if err != nil {
			log.Errorc(ctx, "s.RankResult err(%v)", err)
			return
		}
		if res != nil && len(res.Rank) > 0 {
			url, err := s.export(ctx, res.Rank)
			if err != nil {
				log.Errorc(ctx, "s.export err(%v)", err)
				return
			}
			_ = exporttask.SendWeChatTextMessage(ctx, []string{username}, fmt.Sprintf("榜单已导出，下载链接：%s", url))

		}
	})
	return

}

// export ...
func (s *Service) export(c context.Context, list []*rankmdl.Result) (url string, err error) {
	result := make([][]string, 0)
	var i int64
	for _, item := range list {
		if item.Account == nil {
			continue
		}
		var archive string
		if item.Videos != nil {
			for _, v := range item.Videos {
				archive = archive + fmt.Sprintf("BvID:%s,标题:%s,分数：%d \n", v.Bvid, v.Title, v.Score)
			}
		}
		i++

		result = append(result, []string{strconv.FormatInt(i, 10), strconv.FormatInt(item.Account.Mid, 10), item.Account.Name, strconv.FormatInt(item.Score, 10), archive})
	}
	categoryHeader := []string{"id", "用户mid", "用户昵称", "积分", "稿件信息"}
	b := &bytes.Buffer{}
	b.WriteString("\xEF\xBB\xBF")
	wr := csv.NewWriter(b)
	_ = wr.Write(categoryHeader)
	for i := 0; i < len(result); i++ {
		_ = wr.Write(result[i])
	}
	wr.Flush()
	url, err = boss.Client.UploadObject(c, boss.Bucket, fmt.Sprintf("rank/v1/排行榜导出_%d.csv", time.Now().Unix()), b)
	if err != nil {
		log.Errorc(c, "boss.Client.UploadObject failed. error(%v)", err)
		return "", ecode.Error(ecode.RequestErr, "导出失败请重试")
	}
	return
}

// RankResult 排行榜结果
func (s *Service) RankResult(c context.Context, id int64, attributeType, pn, ps int) (res *rankmdl.ResultReply, err error) {
	var rank *rankmdl.Rank
	res = &rankmdl.ResultReply{}
	page := rankmdl.Page{}
	page.Num = pn
	if ps > 200 {
		ps = 200
	}
	page.Size = ps
	if rank, err = s.dao.GetRankConfigByID(c, id); err != nil {
		log.Errorc(c, "Get s.dao.GetRankConfigByID() failed. error(%v)", err)
		return
	}
	if rank == nil {
		return nil, errors.New("can not find online rank")
	}
	newBatch, err := s.dao.GetLastBatch(c, id, attributeType)
	if err != nil {
		return nil, err
	}
	if newBatch == nil {
		return nil, nil
	}
	res.Batch = newBatch.Batch
	limit := ps
	offset := (pn - 1) * ps
	oidResult, err := s.dao.OidRankInRank(c, id, newBatch.Batch, attributeType, offset, limit)
	if err != nil {
		return nil, err
	}
	var snapshot []*rankmdl.Snapshot
	if oidResult == nil {
		return nil, nil
	}
	oids := make([]int64, 0)
	for _, v := range oidResult {
		oids = append(oids, v.OID)
	}
	aids := make([]int64, 0)
	mids := make([]int64, 0)
	rankTypeUpAids := make(map[int64][]*rankmdl.Video)
	rankTypeAidMid := make(map[int64]*rankmdl.Video)
	if rank.RankType == rankmdl.RankTypeUp {
		mids = oids
		snapshot, err = s.dao.AllSnapshotByMids(c, id, oids, newBatch.Batch, attributeType)
		if err != nil {
			return nil, err
		}
		if len(snapshot) > 0 {
			for _, v := range snapshot {
				aids = append(aids, v.AID)
				if _, ok := rankTypeUpAids[v.MID]; !ok {
					rankTypeUpAids[v.MID] = make([]*rankmdl.Video, 0)
				}
				rankTypeUpAids[v.MID] = append(rankTypeUpAids[v.MID], &rankmdl.Video{ID: v.ID, Aid: v.AID, Score: v.Score, State: v.State, Rank: v.Rank})

			}
		}
	} else {
		for _, v := range oidResult {
			aids = append(aids, v.OID)
		}
		snapshot, err = s.dao.AllSnapshotByAids(c, id, oids, newBatch.Batch, attributeType)
		if err != nil {
			return nil, err
		}
		if len(snapshot) > 0 {
			for _, v := range snapshot {
				mids = append(mids, v.MID)
				rankTypeAidMid[v.AID] = &rankmdl.Video{ID: v.ID, Aid: v.AID, Score: v.Score, State: v.State, Rank: v.Rank, Mid: v.MID}
			}
		}
	}
	if page.Total, err = s.dao.OidRankInRankTotal(c, id, newBatch.Batch, attributeType); err != nil {
		log.Errorc(c, "s.lot.GiftDraftTotal() failed. error(%v)", err)
		return nil, err
	}
	var (
		memberInfo map[int64]*accountapi.Info
		archive    map[int64]*api.Arc
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {

		memberInfo, err = s.account.MemberInfo(c, mids)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		archive, err = s.archive.ArchiveInfo(c, aids)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, err
	}
	var rankList = make([]*rankmdl.Result, 0)

	if rank.RankType == rankmdl.RankTypeUp {
		for _, v := range oidResult {
			replyRank := &rankmdl.Result{}
			replyRank.ID = v.ID
			replyRank.Rank = v.Rank
			replyRank.State = v.State
			if account, ok := memberInfo[v.OID]; ok {
				replyRank.Account = s.accountToRankAccount(c, account)
				video := []*rankmdl.Video{}
				if aids, ok1 := rankTypeUpAids[v.OID]; ok1 {
					if aids != nil && len(aids) > 0 {
						for _, a := range aids {
							if arc, ok := archive[a.Aid]; ok {
								video = append(video, s.archiveInfoToRankVideo(c, arc, a.ID, a.Score, a.State, a.Rank))
							}
						}
						replyRank.Videos = video
					}
				}
			}
			replyRank.Score = v.Score
			rankList = append(rankList, replyRank)
		}
	} else {
		for _, v := range oidResult {
			replyRank := &rankmdl.Result{}
			replyRank.ID = v.ID
			replyRank.Rank = v.Rank
			replyRank.State = v.State
			if arc, ok := archive[v.OID]; ok {
				videos := []*rankmdl.Video{}
				var video = &rankmdl.Video{}
				if videoArc, ok1 := rankTypeAidMid[v.OID]; ok1 {
					if account, ok2 := memberInfo[videoArc.Mid]; ok2 {
						replyRank.Account = s.accountToRankAccount(c, account)
					}
					video = s.archiveInfoToRankVideo(c, arc, videoArc.ID, videoArc.Score, videoArc.State, videoArc.Rank)

				}
				videos = append(videos, video)
				replyRank.Videos = videos

			}
			replyRank.Score = v.Score
			rankList = append(rankList, replyRank)
		}
	}
	res.Rank = rankList
	res.Page = &page
	return res, nil
}

func (s *Service) accountToRankAccount(c context.Context, midInfo *accountapi.Info) *rankmdl.Account {
	return &rankmdl.Account{
		Mid:  midInfo.Mid,
		Name: midInfo.Name,
		Face: midInfo.Face,
		Sign: midInfo.Sign,
		Sex:  midInfo.Sex,
	}
}

func (s *Service) archiveInfoToRankVideo(c context.Context, arc *api.Arc, id, score int64, state int, rank int64) *rankmdl.Video {
	var bvidStr string
	var err error
	if bvidStr, err = bvid.AvToBv(arc.Aid); err != nil || bvidStr == "" {
		return nil
	}
	return &rankmdl.Video{
		ID:       id,
		Bvid:     bvidStr,
		Aid:      arc.Aid,
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
		Score:    score,
		State:    state,
		Rank:     rank,
	}
}

// UpdateRankResult 更新排行结果
func (s *Service) UpdateRankResult(c context.Context, id int64, result []*rankmdl.ResultEdit) (err error) {
	var rank *rankmdl.Rank
	if rank, err = s.dao.GetRankConfigByID(c, id); err != nil {
		log.Errorc(c, "Get s.dao.GetRankConfigByID() failed. error(%v)", err)
		return
	}

	if rank == nil {
		return errors.New("can not find online rank")
	}
	var snapshot = make([]*rankmdl.Snapshot, 0)
	var oidResult = make([]*rankmdl.OidResult, 0)
	black := make([]*rankmdl.Intervention, 0)
	if result != nil {
		for _, v := range result {

			// oidResult处理
			oidResult = append(oidResult, &rankmdl.OidResult{
				ID:    v.ID,
				State: v.State,
				Rank:  v.Rank,
				Score: v.Score,
			})

			// 黑名单处理
			if v.Account != nil && v.State == rankmdl.RankStateOffline {
				black = append(black, &rankmdl.Intervention{
					OID:              v.Account.Mid,
					State:            1,
					InterventionType: rankmdl.InterventionTypeBlack,
					ObjectType:       rankmdl.InterventionObjectUp,
				})
			}

			if v.Video != nil {

				for _, vi := range v.Video {
					if vi.State == rankmdl.RankStateOffline {
						black = append(black, &rankmdl.Intervention{
							OID:              vi.AID,
							State:            1,
							InterventionType: rankmdl.InterventionTypeBlack,
							ObjectType:       rankmdl.InterventionObjectArchive,
						})
					}
					snapshot = append(snapshot, &rankmdl.Snapshot{
						ID:    vi.ID,
						State: vi.State,
						Rank:  vi.Rank,
						Score: vi.Score,
					})
				}
			}
		}
	}
	return s.txSaveEditResult(c, id, snapshot, oidResult, black)
}

func (s *Service) txSaveEditResult(c context.Context, id int64, snapshot []*rankmdl.Snapshot, oidResult []*rankmdl.OidResult, black []*rankmdl.Intervention) (err error) {
	var (
		tx *xsql.Tx
	)
	if tx, err = s.dao.BeginTran(c); err != nil {
		log.Errorc(c, "s.dao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() (id int64, err error) {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			err = errors.New("err")
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
				err = err1
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
		return
	}()
	if len(snapshot) > 0 {
		err = s.dao.BacthInsertOrUpdateSnapshotTx(c, tx, id, snapshot)
		if err != nil {
			return err
		}
	}
	if len(oidResult) > 0 {
		err = s.dao.BacthInsertOrUpdateOidResultTx(c, tx, id, oidResult)
		if err != nil {
			return err
		}
	}
	if len(black) > 0 {
		err = s.dao.BacthInsertOrUpdateBlackOrWhiteTx(c, tx, id, black)
		if err != nil {
			return err
		}
	}
	return nil

}
