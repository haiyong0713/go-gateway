package rank

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/admin/component/boss"
	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank_v3"
	"go-gateway/app/web-svr/activity/admin/service/exporttask"
	"go-gateway/pkg/idsafe/bvid"
	"strconv"
	"strings"
	"time"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"

	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"
	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

const (
	// concurrency 并发为1
	concurrency = 1
	// tagBatch  批量tag
	tagBatch = 50
	// allLimit
	allLimit    = 1000
	writeString = "\xEF\xBB\xBF"
)

// Export ...
func (s *Service) Export(c context.Context, req *rankmdl.ExportReq, userName string) (err error) {
	// 查看baseid和rankid是否正确
	rule, err := s.dao.GetRuleByID(c, req.RankID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRuleByID err(%v)", err)
		return
	}
	if rule == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("无相关榜单"))
		return
	}
	if rule.State == rankmdl.RankStateNotStart {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("未开始"))
		return
	}
	eg := errgroup.WithContext(c)
	var (
		oidResult []*rankmdl.ResultOid
		aidResult []*rankmdl.ResultOidArchive
	)
	eg.Go(func(ctx context.Context) (err error) {
		oidResult, err = s.dao.GetRankOid(ctx, rule.BaseID, rule.ID, rule.ShowBatch)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetRankOid err(%v)", err)
			return
		}
		return nil
	})
	if rule.StatisticsType == rankmdl.StatisticsTypeTag || rule.StatisticsType == rankmdl.StatisticsTypeUp {
		eg.Go(func(ctx context.Context) (err error) {
			aidResult, err = s.dao.GetRankArchive(ctx, rule.BaseID, rule.ID, rule.ShowBatch)
			if err != nil {
				log.Errorc(ctx, "s.dao.GetRankOid err(%v)", err)
				return
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	go func() {
		c := context.Background()
		var oidResultStr, aidResultStr string
		if len(oidResult) > 0 {
			oidResultStr, err = s.oidResultToList(c, rule, oidResult)
			if err != nil {
				log.Errorc(c, "s.oidResultToList err(%v)", err)
				return
			}
		}
		if len(aidResult) > 0 {
			aidResultStr, err = s.oidResultArchiveToList(c, rule, aidResult)
			if err != nil {
				log.Errorc(c, "s.oidResultArchiveToList err(%v)", err)
				return
			}
		}
		_ = exporttask.SendWeChatTextMessage(c, []string{userName}, fmt.Sprintf("子榜%d导出数据成功，对象榜下载链接:%s，若有附加稿件榜下载链接：%s", rule.ID, oidResultStr, aidResultStr))
	}()
	return
}

func (s *Service) oidResultToList(c context.Context, rank *rankmdl.Rule, list []*rankmdl.ResultOid) (url string, err error) {
	result := make([][]string, 0)
	for _, item := range list {
		result = append(result, []string{strconv.FormatInt(item.ID, 10), strconv.FormatInt(item.BaseID, 10), strconv.FormatInt(item.RankID, 10), strconv.FormatInt(item.OID, 10), strconv.FormatInt(item.Rank, 10), strconv.FormatInt(item.Score, 10), strconv.Itoa(item.Batch), strconv.Itoa(item.State)})
	}
	categoryHeader := []string{"id", "排行榜id", "子榜id", "对象id", "排名", "对外分数", "批次", "状态"}
	b := &bytes.Buffer{}
	b.WriteString(writeString)
	wr := csv.NewWriter(b)
	_ = wr.Write(categoryHeader)
	for i := 0; i < len(result); i++ {
		_ = wr.Write(result[i])
	}
	wr.Flush()
	url, err = boss.Client.UploadObject(c, boss.Bucket, fmt.Sprintf("rank/result/%d_%d_%d_oid.csv", rank.BaseID, rank.ID, rank.ShowBatch), b)
	if err != nil {
		log.Errorc(c, "boss.Client.UploadObject (baseid:%d) failed. error(%v)", rank.BaseID, err)
		return "", ecode.Error(ecode.RequestErr, "导出失败请重试")
	}
	return
}

func (s *Service) oidResultArchiveToList(c context.Context, rank *rankmdl.Rule, list []*rankmdl.ResultOidArchive) (url string, err error) {
	result := make([][]string, 0)
	for _, item := range list {
		result = append(result, []string{strconv.FormatInt(item.ID, 10), strconv.FormatInt(item.BaseID, 10), strconv.FormatInt(item.RankID, 10), strconv.FormatInt(item.AID, 10), strconv.FormatInt(item.OID, 10), strconv.FormatInt(item.Rank, 10), strconv.FormatInt(item.Score, 10), strconv.Itoa(item.Batch), strconv.Itoa(item.State)})
	}
	categoryHeader := []string{"id", "排行榜id", "子榜id", "稿件aid", "对象id", "排名", "对外分数", "批次", "状态"}
	b := &bytes.Buffer{}
	b.WriteString(writeString)
	wr := csv.NewWriter(b)
	_ = wr.Write(categoryHeader)
	for i := 0; i < len(result); i++ {
		_ = wr.Write(result[i])
	}
	wr.Flush()
	url, err = boss.Client.UploadObject(c, boss.Bucket, fmt.Sprintf("rank/result/%d_%d_%d_aid.csv", rank.BaseID, rank.ID, rank.ShowBatch), b)
	if err != nil {
		log.Errorc(c, "UploadObject(baseid:%d) failed. error(%v)", rank.BaseID, err)
		return "", ecode.Error(ecode.RequestErr, "导出失败请重试")
	}
	return
}

// ExportRankResult ...
func (s *Service) ExportRankResult(ctx context.Context, req *rankmdl.ExportRankResultReq, username string) (err error) {
	// 查看baseid和rankid是否正确
	rule, err := s.dao.GetRuleByID(ctx, req.RankID)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetRuleByID err(%v)", err)
		return
	}
	if rule == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("无相关榜单"))
		return
	}
	resultReq := &rankmdl.ResultReq{
		BaseID:     req.BaseID,
		RankID:     req.RankID,
		IsShow:     1,
		ObjectType: req.ObjectType,
		Aid:        req.Aid,
		MID:        req.MID,
		TagID:      req.TagID,
		Ps:         1000,
		Pn:         1,
	}
	res, err := s.GetRankResult(ctx, resultReq)
	if err != nil {
		log.Errorc(ctx, "s.GetRankResult err(%v)", err)
		return
	}
	// 异步处理
	_ = s.cache.SyncDo(ctx, func(ctx context.Context) {
		if len(res.List) > 0 {
			var url string
			if req.ObjectType == rankmdl.ObjectTypeArchive {
				url, err = s.exportArchiveRankResult(ctx, rule, res.List)
			}

			if req.ObjectType == rankmdl.ObjectTypeUp {
				url, err = s.exportAccountRankResult(ctx, rule, res.List)

			}
			if req.ObjectType == rankmdl.ObjectTypeTag {
				url, err = s.exportTagRankResult(ctx, rule, res.List)
			}
			if err != nil {
				log.Errorc(ctx, "s.export() err(%v)", err)
				return
			}
			_ = exporttask.SendWeChatTextMessage(ctx, []string{username}, fmt.Sprintf("子榜%d导出数据成功，榜单下载链接:%s", rule.ID, url))
		}
	})
	return
}

func (s *Service) exportArchiveRankResult(c context.Context, rank *rankmdl.Rule, list []*rankmdl.ResultRes) (url string, err error) {
	result := make([][]string, 0)
	for _, item := range list {
		if item.Archive == nil || item.Account == nil {
			continue
		}
		result = append(result, []string{strconv.FormatInt(item.ShowRank, 10), strconv.FormatInt(item.Account.Mid, 10), item.Account.Name, strconv.FormatInt(item.Archive.Aid, 10),
			item.Archive.Bvid, item.Archive.Title, strconv.FormatInt(item.Score.Total, 10), item.Score.ShowScore, strconv.FormatInt(item.Score.Extra, 10), strconv.FormatInt(item.Score.Play, 10),
			strconv.FormatInt(item.Score.Like, 10), strconv.FormatInt(item.Score.Coin, 10), strconv.FormatInt(item.Score.Share, 10)})
	}
	categoryHeader := []string{"榜单排名", "用户id", "用户信息", "稿件aid", "稿件bvid", "稿件标题", "总分", "展示分", "评委分", "播放分", "点赞分", "投币分", "分享分"}
	b := &bytes.Buffer{}
	b.WriteString(writeString)
	wr := csv.NewWriter(b)
	_ = wr.Write(categoryHeader)
	for i := 0; i < len(result); i++ {
		_ = wr.Write(result[i])
	}
	wr.Flush()
	url, err = boss.Client.UploadObject(c, boss.Bucket, fmt.Sprintf("rank/rank_detail_result/%d_%d_%d_aid.csv", rank.BaseID, rank.ID, rank.ShowBatch), b)
	if err != nil {
		log.Errorc(c, "UploadObject (baseid:%d) failed. error(%v)", rank.BaseID, err)
		return "", ecode.Error(ecode.RequestErr, "导出失败请重试")
	}
	return
}

func (s *Service) exportAccountRankResult(c context.Context, rank *rankmdl.Rule, list []*rankmdl.ResultRes) (url string, err error) {
	result := make([][]string, 0)
	for _, item := range list {
		if item.Account == nil {
			continue
		}
		result = append(result, []string{strconv.FormatInt(item.ShowRank, 10), strconv.FormatInt(item.Account.Mid, 10), item.Account.Name,
			strconv.FormatInt(item.Score.Total, 10), item.Score.ShowScore, strconv.FormatInt(item.Score.Extra, 10), strconv.FormatInt(item.Score.Play, 10),
			strconv.FormatInt(item.Score.Like, 10), strconv.FormatInt(item.Score.Coin, 10), strconv.FormatInt(item.Score.Share, 10)})
	}
	categoryHeader := []string{"榜单排名", "用户id", "用户信息", "总分", "展示分", "评委分", "播放分", "点赞分", "投币分", "分享分"}
	b := &bytes.Buffer{}
	b.WriteString(writeString)
	wr := csv.NewWriter(b)
	_ = wr.Write(categoryHeader)
	for i := 0; i < len(result); i++ {
		wr.Write(result[i])
	}
	wr.Flush()
	url, err = boss.Client.UploadObject(c, boss.Bucket, fmt.Sprintf("rank/rank_detail_result/%d_%d_%d_mid.csv", rank.BaseID, rank.ID, rank.ShowBatch), b)
	if err != nil {
		log.Errorc(c, "boss.Client.UploadObject (baseid:%d) failed. error(%v)", rank.BaseID, err)
		return "", ecode.Error(ecode.RequestErr, "导出失败请重试")
	}
	return
}

func (s *Service) exportTagRankResult(c context.Context, rank *rankmdl.Rule, list []*rankmdl.ResultRes) (url string, err error) {
	result := make([][]string, 0)
	for _, item := range list {
		if item.Tag == nil {
			continue
		}
		result = append(result, []string{strconv.FormatInt(item.ShowRank, 10), strconv.FormatInt(item.Tag.TID, 10), item.Tag.Name,
			strconv.FormatInt(item.Score.Total, 10), item.Score.ShowScore, strconv.FormatInt(item.Score.Extra, 10), strconv.FormatInt(item.Score.Play, 10),
			strconv.FormatInt(item.Score.Like, 10), strconv.FormatInt(item.Score.Coin, 10), strconv.FormatInt(item.Score.Share, 10)})
	}
	categoryHeader := []string{"榜单排名", "tagID", "Tag信息", "总分", "展示分", "评委分", "播放分", "点赞分", "投币分", "分享分"}
	b := &bytes.Buffer{}
	b.WriteString(writeString)
	wr := csv.NewWriter(b)
	_ = wr.Write(categoryHeader)
	for i := 0; i < len(result); i++ {
		wr.Write(result[i])
	}
	wr.Flush()
	url, err = boss.Client.UploadObject(c, boss.Bucket, fmt.Sprintf("rank/rank_detail_result/%d_%d_%d_tag.csv", rank.BaseID, rank.ID, rank.ShowBatch), b)
	if err != nil {
		log.Errorc(c, "boss.Client. UploadObject(baseid:%d) failed. error(%v)", rank.BaseID, err)
		return "", ecode.Error(ecode.RequestErr, "导出失败请重试")
	}
	return
}

// Publish ...
func (s *Service) Publish(c context.Context, req *rankmdl.PublishReq, username string) (err error) {

	base, err := s.dao.GetRankByID(c, req.BaseID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByID err(%v)", err)
		return
	}
	// 查看baseid和rankid是否正确
	rule, err := s.dao.GetRuleByID(c, req.RankID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRuleByID err(%v)", err)
		return
	}
	if rule == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("无相关榜单"))
		return
	}

	if rule.State == rankmdl.RankStateNotStart {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("未开始"))
		return
	}
	if rule.BaseID != req.BaseID {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("数据有误"))
		return
	}
	if !s.checkAuthority(c, username, base.Author, base.Authority) {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("没有配置权限，请联系%s增加权限", base.Author))
		return
	}

	var objectType int
	var needParent bool
	if rule.StatisticsType == rankmdl.StatisticsTypeArchive || rule.StatisticsType == rankmdl.StatisticsTypeDistinctArchive {
		objectType = rankmdl.ObjectTypeArchive
	}
	if rule.StatisticsType == rankmdl.StatisticsTypeUp {
		objectType = rankmdl.ObjectTypeUp

	}
	if rule.StatisticsType == rankmdl.StatisticsTypeTag {
		objectType = rankmdl.ObjectTypeTag
	}
	// 异步处理
	_ = s.cache.SyncDo(c, func(c context.Context) {
		adjustOid, err := s.dao.GetAdjust(c, base.ID, rule.ID, objectType)
		if err != nil {
			log.Errorc(c, "s.dao.GetRuleByID err(%v)", err)
			return
		}
		res, err := s.getRankResultDB(c, adjustOid, base, needParent, rule, rule.LastBatch, objectType, []int64{}, []int64{})
		if err != nil {
			log.Errorc(c, "s.getRankResultDB err(%v)", err)
			return
		}
		if res == nil {
			err = ecode.Error(ecode.RequestErr, "无可发布内容")
			return
		}
		// 处理原始榜单
		showBatch := rule.ShowBatch + 1
		oidRank, oidArchiveRank, err := s.getPublishData(c, base, rule, res, rankmdl.ArchiveNums, showBatch, objectType)
		if err != nil {
			log.Errorc(c, "s.getPublishData err(%v)", err)
			return
		}
		s.scorePrecision(c, rule.Precision, rule.Unit, rule.Description, oidRank, oidArchiveRank)
		// 更新展示batch
		var (
			tx *xsql.Tx
		)
		if tx, err = s.dao.BeginTran(c); err != nil {
			log.Errorc(c, "s.dao.BeginTran() failed. error(%v)", err)
			return
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				log.Errorc(c, "%v", r)
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

		err = s.dao.BatchOidResult(c, tx, req.BaseID, oidRank)
		if err != nil {
			log.Errorc(c, "s.dao.BatchOidResult err(%v)", err)
			return
		}
		if len(oidArchiveRank) > 0 {
			err = s.dao.BatchOidArchiveResult(c, tx, req.BaseID, oidArchiveRank)
			if err != nil {
				log.Errorc(c, "s.dao.BatchOidArchiveResult err(%v)", err)
				return
			}
		}
		// 更新结果
		err = s.dao.UpdateRuleBatch(c, tx, req.RankID, showBatch, time.Now().Unix())
		if err != nil {
			log.Errorc(c, "s.dao.UpdateRuleBatch err(%v)", err)
			return
		}
		return
	})
	return
}

func (s *Service) showScoreText(c context.Context, score int64, precision int, unit int, description string) string {
	var newScore float64
	var newScoreStr string
	var units string
	if unit == rankmdl.UnitOne {
		newScore = float64(score)
		units = ""
	}
	if unit == rankmdl.UnitTenThousands {
		newScore = float64(score) / 10000
		units = "万"
	}
	switch precision {

	case rankmdl.PrecisionZero:
		newScoreStr = fmt.Sprintf("%.0f", newScore)
	case rankmdl.PrecisionOne:
		newScoreStr = fmt.Sprintf("%.1f", newScore)

	case rankmdl.PrecisionTwo:
		newScoreStr = fmt.Sprintf("%.2f", newScore)

	case rankmdl.PrecisionThree:
		newScoreStr = fmt.Sprintf("%.3f", newScore)
	}
	if description == "" {
		return fmt.Sprintf("%s%s", newScoreStr, units)

	} else {
		return fmt.Sprintf("%s%s%s", description, newScoreStr, units)
	}

}

func (s *Service) scorePrecision(c context.Context, precision int, unit int, description string, oidRank []*rankmdl.ResultOid, oidArchiveRank []*rankmdl.ResultOidArchive) {
	if len(oidRank) > 0 {
		for i, v := range oidRank {
			score := v.Score
			oidRank[i].ShowScore = s.showScoreText(c, score, precision, unit, description)
		}
	}
	if len(oidArchiveRank) > 0 {
		for i, v := range oidArchiveRank {
			score := v.Score
			oidArchiveRank[i].ShowScore = s.showScoreText(c, score, precision, unit, description)

		}
	}
	return
}

func (s *Service) getPublishData(c context.Context, base *rankmdl.Base, rule *rankmdl.Rule, result []*rankmdl.ResultDetail, archiveNums int, showBatch int, objectType int) (oidRank []*rankmdl.ResultOid, oidArchiveRank []*rankmdl.ResultOidArchive, err error) {
	oidRank = make([]*rankmdl.ResultOid, 0)
	oidArchiveRank = make([]*rankmdl.ResultOidArchive, 0)
	oidList := make([]int64, 0)
	var nums int
	for _, v := range result {
		if v.IsHidden == 1 {
			continue
		}
		score := v.Score.Total
		if base.IsShowScore == rankmdl.IsNotShow {
			score = 0
		}
		oidRank = append(oidRank, &rankmdl.ResultOid{
			BaseID: rule.BaseID,
			RankID: rule.ID,
			OID:    v.OID,
			Rank:   v.ShowRank,
			Score:  score,
			Batch:  showBatch,
		})
		oidList = append(oidList, v.OID)
		nums++
		if nums == rule.Nums {
			break
		}
	}
	if len(oidList) == 0 {
		err = ecode.Error(ecode.RequestErr, "无可发布内容")
		return
	}
	// 处理榜单相关稿件
	if objectType == rankmdl.ObjectTypeUp || objectType == rankmdl.ObjectTypeTag {
		mids := make([]int64, 0)
		tagIds := make([]int64, 0)
		if objectType == rankmdl.ObjectTypeUp {
			mids = oidList
		}
		if objectType == rankmdl.ObjectTypeTag {
			tagIds = oidList
		}
		adjustOid, err := s.dao.GetAdjust(c, base.ID, rule.ID, rankmdl.ObjectTypeArchive)
		if err != nil {
			log.Errorc(c, "s.dao.GetRuleByID err(%v)", err)
			return nil, nil, err
		}
		oldRank, err := s.getRankResultDB(c, adjustOid, base, true, rule, rule.LastBatch, rankmdl.ObjectTypeArchive, tagIds, mids)
		if err != nil {
			log.Errorc(c, "s.getActRank err(%v)", err)
			return nil, nil, err
		}
		mapOid := make(map[int64][]*rankmdl.ResultDetail)
		if oldRank != nil {
			for _, v := range oldRank {
				if v.IsHidden == 1 {
					continue
				}
				score := v.Score.Total
				var oid int64
				if objectType == rankmdl.ObjectTypeUp {
					oid = v.Mid
				}
				if objectType == rankmdl.ObjectTypeTag {
					oid = v.TagID
				}
				if base.IsShowScore == rankmdl.IsNotShow {
					score = 0
				}
				object := &rankmdl.ResultOidArchive{
					BaseID: rule.BaseID,
					RankID: rule.ID,
					AID:    v.OID,
					OID:    oid,
					Rank:   v.ShowRank,
					Score:  score,
					Batch:  showBatch,
				}
				if mapObject, ok := mapOid[oid]; ok {
					if len(mapObject) == archiveNums {
						continue
					}
					mapOid[oid] = append(mapOid[oid], v)
					oidArchiveRank = append(oidArchiveRank, object)

					continue
				}
				mapOid[oid] = make([]*rankmdl.ResultDetail, 0)
				mapOid[oid] = append(mapOid[oid], v)
				oidArchiveRank = append(oidArchiveRank, object)
			}
		}
	}
	return oidRank, oidArchiveRank, nil
}

// GetRankResult ...
func (s *Service) GetRankResult(c context.Context, req *rankmdl.ResultReq) (rankRes *rankmdl.ResultList, err error) {

	// 查看baseid和rankid是否正确
	rule, err := s.dao.GetRuleByID(c, req.RankID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRuleByID err(%v)", err)
		return
	}
	if rule == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("无相关榜单"))
		return
	}
	if rule.State == rankmdl.RankStateNotStart {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("未开始"))
		return
	}
	if rule.BaseID != req.BaseID {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("数据有误"))
		return
	}
	if req.ObjectType == rankmdl.ObjectTypeUp {
		if rule.StatisticsType != rankmdl.StatisticsTypeUp {
			err = ecode.Error(ecode.RequestErr, fmt.Sprintf("无相关榜单"))
			return
		}
	}
	if req.ObjectType == rankmdl.ObjectTypeTag {
		if rule.StatisticsType != rankmdl.StatisticsTypeTag {
			err = ecode.Error(ecode.RequestErr, fmt.Sprintf("无相关榜单"))
			return
		}
	}
	base, err := s.dao.GetRankByID(c, req.BaseID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByID err(%v)", err)
		return
	}
	rankRes, err = s.getRankDetail(c, base, rule, req)
	if err != nil {
		log.Errorc(c, "s.getRankDetail err(%v)", err)
		return
	}
	if rankRes != nil {
		rankRes.Rule = rule
	}
	return

}

func (s *Service) getRankDetail(c context.Context, base *rankmdl.Base, rule *rankmdl.Rule, req *rankmdl.ResultReq) (rankRes *rankmdl.ResultList, err error) {
	rankRes = &rankmdl.ResultList{}
	rankRes.Page = &rankmdl.Page{Num: int(req.Pn), Size: int(req.Ps)}
	rankRes.List = make([]*rankmdl.ResultRes, 0)
	adjustOid, err := s.dao.GetAdjust(c, base.ID, rule.ID, req.ObjectType)
	if err != nil {
		log.Errorc(c, "s.dao.GetRuleByID err(%v)", err)
		return
	}
	newAdjust := make([]*rankmdl.Adjust, 0)
	newAdjust = adjustOid
	if rule.StatisticsType == rankmdl.StatisticsTypeUp || rule.StatisticsType == rankmdl.StatisticsTypeTag {
		if req.ObjectType == rankmdl.ObjectTypeArchive {
			newAdjust = make([]*rankmdl.Adjust, 0)
			if len(req.TagID) > 0 || len(req.MID) > 0 {
				mapTagID := make(map[int64]struct{})
				mapMID := make(map[int64]struct{})
				if len(req.TagID) > 0 {
					for _, v := range req.TagID {
						mapTagID[v] = struct{}{}
					}
				}
				if len(req.MID) > 0 {
					for _, v := range req.MID {
						mapMID[v] = struct{}{}
					}
				}
				for _, v := range adjustOid {
					if _, ok := mapMID[v.ParentID]; ok {
						newAdjust = append(newAdjust, v)
					}
					if _, ok := mapTagID[v.ParentID]; ok {
						newAdjust = append(newAdjust, v)
					}
				}

			}
		}
	}
	res, err := s.getRankResultDB(c, newAdjust, base, false, rule, rule.LastBatch, req.ObjectType, req.TagID, req.MID)
	if err != nil {
		log.Errorc(c, "s.getRankResultDB err(%v)", err)
		return
	}
	new := make([]*rankmdl.ResultDetail, 0)
	mapAidFilter := make(map[int64]struct{})
	if len(req.Aid) > 0 {
		for _, v := range req.Aid {
			mapAidFilter[v] = struct{}{}
		}
	}
	for _, v := range res {
		if len(req.Aid) > 0 {
			if _, ok := mapAidFilter[v.AID]; !ok {
				continue
			}
		}
		if req.IsShow != 0 {
			if req.IsShow == rankmdl.IsShow && v.IsHidden == 0 {
				new = append(new, v)
				continue
			}
			if req.IsShow == rankmdl.IsNotShow && v.IsHidden == 1 {
				new = append(new, v)
				continue
			}
			continue
		}
		new = append(new, v)
	}

	rankRes.Page.Total = len(new)
	start := (req.Pn - 1) * req.Ps
	if start >= int64(len(new)) {
		return
	}
	newRes := new[start:]
	end := start + req.Ps

	if end < int64(len(new)) {
		newRes = new[start:end]
	}
	aids := make([]int64, 0)
	mapAids := make(map[int64]struct{})
	mids := make([]int64, 0)
	mapMids := make(map[int64]struct{})

	tagIDs := make([]int64, 0)
	mapTagIDs := make(map[int64]struct{})

	for _, v := range newRes {
		if v.AID > 0 {
			if _, ok := mapAids[v.AID]; !ok {
				aids = append(aids, v.AID)
				mapAids[v.AID] = struct{}{}
			}
		}
		if v.Mid > 0 {
			if _, ok := mapMids[v.Mid]; !ok {
				mids = append(mids, v.Mid)
				mapMids[v.Mid] = struct{}{}
			}

		}
		if v.TagID > 0 {
			if _, ok := mapTagIDs[v.TagID]; !ok {
				tagIDs = append(tagIDs, v.TagID)
				mapTagIDs[v.TagID] = struct{}{}
			}
		}
	}
	var (
		memberInfo map[int64]*accountapi.Info
		archive    map[int64]*api.Arc
		tag        map[int64]*tagrpc.Tag
	)
	eg := errgroup.WithContext(c)
	if len(mids) > 0 {
		eg.Go(func(ctx context.Context) error {

			memberInfo, err = s.account.MemberInfo(c, mids)
			return err
		})
	}
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			archive, err = s.archive.ArchiveInfo(c, aids)
			return err
		})
	}
	if len(tagIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			tag, err = s.tag.TagInfo(c, tagIDs)
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, err
	}

	result := make([]*rankmdl.ResultRes, 0)
	for _, v := range newRes {
		newResult := &rankmdl.ResultRes{
			AID:          v.AID,
			TagID:        v.TagID,
			IsHidden:     v.IsHidden,
			HiddenReason: v.HiddenReason,
			ManualRank:   v.ManualRank,
			Score:        v.Score,
			ShowRank:     v.ShowRank,
			Mid:          v.Mid,
			ObjectType:   v.ObjectType,
			Archive:      &rankmdl.Archive{},
			Tag:          &rankmdl.Tag{},
			Account:      &rankmdl.Account{},
		}
		if v.AID > 0 {

			if arc, ok := archive[v.AID]; ok {
				var bvidStr string
				if bvidStr, err = bvid.AvToBv(arc.Aid); err != nil || bvidStr == "" {
					continue
				}
				newResult.Archive = &rankmdl.Archive{
					Aid:   v.AID,
					Title: arc.Title,
					Bvid:  bvidStr,
				}
			}
		}
		if v.Mid > 0 {
			if arc, ok := memberInfo[v.Mid]; ok {
				newResult.Account = &rankmdl.Account{
					Mid:  v.Mid,
					Name: arc.Name,
				}
			}
		}
		if v.TagID > 0 {
			if arc, ok := tag[v.TagID]; ok {
				newResult.Tag = &rankmdl.Tag{
					TID:  v.TagID,
					Name: arc.Name,
				}
			}
		}
		result = append(result, newResult)
	}
	rankRes.List = result
	return

}

func (s *Service) archiveRank(c context.Context, oldRank []*rankmdl.Result, rule *rankmdl.Rule, adjustOid []*rankmdl.Adjust) (newRank []*rankmdl.ResultDetail, err error) {
	newRank = make([]*rankmdl.ResultDetail, 0)
	mapRankAid := make(map[int64]*rankmdl.Adjust)
	mapRankNotShow := make(map[int64]struct{})
	mapAdjust := make(map[int64]*rankmdl.Adjust)
	for _, v := range adjustOid {
		mapRankAid[v.Rank] = v
		mapAdjust[v.OID] = v
		if v.IsShow == rankmdl.IsNotShow {
			mapRankNotShow[v.OID] = struct{}{}
		}
	}
	oldRankAidMap := make(map[int64]*rankmdl.Result)
	for _, v := range oldRank {
		oldRankAidMap[v.OID] = v
	}
	alreadyInRank := make(map[int64]struct{})
	mapNewRank := make(map[int64]*rankmdl.ResultDetail)

	for i := 1; i <= len(oldRank); i++ {
		rankIndex := int64(i)
		isHidden := 0
		hiddenReason := make([]string, 0)

		if adjust, ok := mapRankAid[rankIndex]; ok {
			if _, ok := mapRankNotShow[adjust.OID]; ok {
				isHidden = 1
				hiddenReason = append(hiddenReason, "前端隐藏")
			}
			if oldRankAid, ok := oldRankAidMap[adjust.OID]; ok {
				score := &rankmdl.Score{
					Rank:      oldRankAid.Rank,
					Total:     oldRankAid.Score,
					Play:      oldRankAid.PlayScore,
					Like:      oldRankAid.LikesScore,
					Coin:      oldRankAid.CoinScore,
					Share:     oldRankAid.ShareScore,
					Extra:     oldRankAid.WhiteScore,
					Fans:      oldRankAid.FansScore,
					ShowScore: s.showScoreText(c, oldRankAid.Score, rule.Precision, rule.Unit, rule.Description),
				}

				mapNewRank[rankIndex] = &rankmdl.ResultDetail{
					AID:          oldRankAid.AID,
					OID:          oldRankAid.OID,
					IsHidden:     isHidden,
					HiddenReason: hiddenReason,
					ManualRank:   rankIndex,
					Score:        score,
					Mid:          oldRankAid.MID,
					TagID:        oldRankAid.TagID,
					ObjectType:   oldRankAid.ObjectType,
				}
				alreadyInRank[adjust.OID] = struct{}{}
			}
		}
	}
	normalRank := make([]*rankmdl.ResultDetail, 0)
	for _, v := range oldRank {
		if _, ok := alreadyInRank[v.OID]; ok {
			continue
		}
		var rankIndex int64
		rankIndex = 1
		for i := 1; i <= len(oldRank); i++ {
			index := i
			rankIndex = int64(index)
			if _, ok := mapNewRank[rankIndex]; !ok {
				break
			}
		}

		isHidden := 0
		hiddenReason := make([]string, 0)
		if _, ok := mapRankNotShow[v.OID]; ok {
			isHidden = 1
			hiddenReason = append(hiddenReason, "前端隐藏")
		}
		score := &rankmdl.Score{
			Rank:      v.Rank,
			Total:     v.Score,
			Play:      v.PlayScore,
			Like:      v.LikesScore,
			Coin:      v.CoinScore,
			Share:     v.ShareScore,
			Extra:     v.WhiteScore,
			Fans:      v.FansScore,
			ShowScore: s.showScoreText(c, v.Score, rule.Precision, rule.Unit, rule.Description),
		}
		normalRank = append(normalRank, &rankmdl.ResultDetail{
			AID:          v.AID,
			OID:          v.OID,
			IsHidden:     isHidden,
			HiddenReason: hiddenReason,
			ManualRank:   0,
			Score:        score,
			Mid:          v.MID,
			TagID:        v.TagID,
			ObjectType:   v.ObjectType,
		})
		alreadyInRank[v.OID] = struct{}{}
	}
	var showRank, adjustRankIndex, normalRankIndex int64
	showRank = 1
	adjustRankIndex = 1
	normalRankIndex = 0
	for i := 1; i <= len(oldRank); i++ {
		newRankDetail, ok := mapNewRank[adjustRankIndex]
		if ok {
			if newRankDetail.IsHidden == 1 {
				newRankDetail.ShowRank = 0
			} else {
				newRankDetail.ShowRank = showRank
				showRank++
			}
			newRank = append(newRank, newRankDetail)
			adjustRankIndex++
		} else {
			if normalRankIndex < int64(len(normalRank)) {
				normalRankDetail := normalRank[normalRankIndex]
				if normalRankDetail.IsHidden == 1 {
					normalRankDetail.ShowRank = 0
				} else {
					normalRankDetail.ShowRank = showRank
					showRank++
					adjustRankIndex++
				}
				newRank = append(newRank, normalRankDetail)
				normalRankIndex++
			}

		}

	}
	return newRank, nil
}

func (s *Service) getRankResultDB(c context.Context, adjustOid []*rankmdl.Adjust, base *rankmdl.Base, needParent bool, rule *rankmdl.Rule, batch, objectType int, tagID, mid []int64) (result []*rankmdl.ResultDetail, err error) {

	oldRank := make([]*rankmdl.Result, 0)
	if objectType == rankmdl.ObjectTypeArchive && needParent {
		oldRankGroup := make(map[int64][]*rankmdl.Result)
		sortList := make([]int64, 0)
		filterAdjustOid := make(map[int64][]*rankmdl.Adjust, 0)
		if len(adjustOid) > 0 {
			for _, v := range adjustOid {
				if _, ok := filterAdjustOid[v.ParentID]; !ok {
					filterAdjustOid[v.ParentID] = make([]*rankmdl.Adjust, 0)
				}
				filterAdjustOid[v.ParentID] = append(filterAdjustOid[v.ParentID], v)
			}
		}
		oldRank, err = s.getActRank(c, rule, batch, tagID, mid)
		if err != nil {
			log.Errorc(c, "s.getActRank err(%v)", err)
			return
		}
		sortMap := make(map[int64]struct{})
		for _, v := range oldRank {

			if rule.StatisticsType == rankmdl.StatisticsTypeUp {
				v.ParentID = v.MID
			}
			if rule.StatisticsType == rankmdl.StatisticsTypeTag {
				v.ParentID = v.TagID
			}
			if _, ok := oldRankGroup[v.ParentID]; !ok {
				oldRankGroup[v.ParentID] = make([]*rankmdl.Result, 0)
			}
			oldRankGroup[v.ParentID] = append(oldRankGroup[v.ParentID], v)
			if _, ok := sortMap[v.ParentID]; !ok {
				sortList = append(sortList, v.ParentID)
				sortMap[v.ParentID] = struct{}{}
			}
		}
		result = make([]*rankmdl.ResultDetail, 0)
		for _, sort := range sortList {
			if v, ok := oldRankGroup[sort]; ok {
				if adjust, ok := filterAdjustOid[sort]; ok {
					res, err := s.archiveRank(c, v, rule, adjust)
					if err != nil {
						return nil, err
					}
					result = append(result, res...)
				} else {
					res, err := s.archiveRank(c, v, rule, []*rankmdl.Adjust{})
					if err != nil {
						return nil, err
					}
					result = append(result, res...)
				}
			}
		}
		return
	}

	if objectType == rankmdl.ObjectTypeArchive {
		oldRank, err = s.getActRank(c, rule, batch, tagID, mid)
		if err != nil {
			log.Errorc(c, "s.getActRank err(%v)", err)
			return
		}

	}
	if objectType == rankmdl.ObjectTypeUp {
		oldRank, err = s.getUpRank(c, rule, batch, mid)
		if err != nil {
			log.Errorc(c, "s.getActRank err(%v)", err)
			return
		}

	}
	if objectType == rankmdl.ObjectTypeTag {
		oldRank, err = s.getTagRank(c, rule, batch, tagID)
		if err != nil {
			log.Errorc(c, "s.getTagRank err(%v)", err)
			return
		}
	}
	result, err = s.archiveRank(c, oldRank, rule, adjustOid)
	return
}

func (s *Service) getUpRank(c context.Context, rule *rankmdl.Rule, batch int, mid []int64) (result []*rankmdl.Result, err error) {
	result, err = s.dao.RankUp(c, rule.BaseID, rule.ID, batch, mid)
	if err != nil {
		log.Errorc(c, "s.dao.RankUp err(%v)", err)
		return
	}
	for i := range result {
		result[i].OID = result[i].MID
		result[i].ObjectType = rankmdl.ObjectTypeUp

	}
	return
}

func (s *Service) getTagRank(c context.Context, rule *rankmdl.Rule, batch int, tagID []int64) (result []*rankmdl.Result, err error) {
	result, err = s.dao.RankTag(c, rule.BaseID, rule.ID, batch, tagID)
	if err != nil {
		log.Errorc(c, "s.dao.RankTag err(%v)", err)
		return
	}
	for i := range result {
		result[i].OID = result[i].TagID
		result[i].ObjectType = rankmdl.ObjectTypeTag

	}
	return
}
func (s *Service) getActRank(c context.Context, rule *rankmdl.Rule, batch int, tagID []int64, mid []int64) (result []*rankmdl.Result, err error) {
	var offset int
	resList := make([]*rankmdl.Result, 0)
	for {
		result, err = s.dao.RankArchive(c, rule.BaseID, rule.ID, batch, tagID, mid, offset, allLimit)
		if err != nil {
			log.Errorc(c, "s.dao.RankArchive err(%v)", err)
			return
		}
		for i := range result {
			result[i].OID = result[i].AID
			result[i].ObjectType = rankmdl.ObjectTypeArchive
		}
		if len(result) > 0 {
			resList = append(resList, result...)
		}
		if len(result) < allLimit {
			break
		}
		offset += allLimit

	}
	return resList, nil

}

// UpdateAdjust ...
func (s *Service) UpdateAdjust(c context.Context, req *rankmdl.UpdateAdjustObject, username string) (err error) {
	// 先查询
	oldRank, err := s.dao.GetRankByID(c, req.BaseID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByID (%d) err(%v)", req.BaseID, err)
		return err
	}
	if !s.checkAuthority(c, username, oldRank.Author, oldRank.Authority) {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("没有配置权限，请联系%s增加权限", oldRank.Author))
		return
	}
	adjustOid, err := s.dao.GetAdjust(c, req.BaseID, req.RankID, req.ObjectType)
	if err != nil {
		log.Errorc(c, "s.dao.GetRuleByID err(%v)", err)
		return
	}
	if req.Rank != 0 {
		for _, v := range adjustOid {
			if v.Rank == req.Rank && v.OID != req.OID && v.ParentID == req.ParentID {
				err = ecode.Error(ecode.RequestErr, fmt.Sprintf("已设置相同排名，请取消原有排名再进行设置"))
				return
			}
		}
	}
	return s.dao.UpdateRankAdjust(c, req.BaseID, &rankmdl.Adjust{
		RankID:     req.RankID,
		OID:        req.OID,
		Rank:       req.Rank,
		IsShow:     req.IsShow,
		ObjectType: req.ObjectType,
		State:      req.State,
		ParentID:   req.ParentID,
	})
}

// UpdateBlackWhite ...
func (s *Service) UpdateBlackWhite(c context.Context, blackWhiteReq *rankmdl.BlackWhiteReq) (err error) {
	update := make([]*rankmdl.BlackWhite, 0)
	if blackWhiteReq.BlackWhite != nil && len(blackWhiteReq.BlackWhite) > 0 {
		for _, v := range blackWhiteReq.BlackWhite {
			update = append(update, &rankmdl.BlackWhite{
				Oid:              v.Oid,
				Score:            v.Score,
				State:            v.State,
				InterventionType: v.InterventionType,
				ObjectType:       v.ObjectType,
			})
		}
		if len(update) == 0 {
			return
		}
		aids := make([]int64, 0)
		var archive map[int64]*api.Arc

		// 新增黑白名单
		err = s.dao.AddBlackWhite(c, blackWhiteReq.BaseID, update)
		if err != nil {
			log.Errorc(c, " s.dao.AddBlackWhite err(%v)", err)
			return
		}
		// 若rank_id>0 并且为白名单，则干预某个子榜的稿件积分，需要修改相应对象的粗排表
		if blackWhiteReq.RankID > 0 {
			rule, err := s.dao.GetRuleByID(c, blackWhiteReq.RankID)
			if err != nil {
				log.Errorc(c, "s.getRule err(%v)", err)
				return err
			}
			if rule == nil {
				return err
			}

			for _, v := range blackWhiteReq.BlackWhite {

				if v.ObjectType == rankmdl.ObjectTypeArchive {
					aids = append(aids, v.Oid)
				}
			}

			if len(aids) > 0 {
				archive, err = s.archive.ArchiveInfo(c, aids)
				if err != nil {
					return err
				}
			}

			archiveUpdate := make([]*rankmdl.Result, 0)
			upUpdate := make([]*rankmdl.Result, 0)
			tagUpdate := make([]*rankmdl.Result, 0)

			for _, v := range blackWhiteReq.BlackWhite {
				if v.InterventionType == rankmdl.BlackWhiteInterventionTypeBlack {
					continue
				}
				object := &rankmdl.Result{
					ObjectType: v.ObjectType,
					BaseID:     blackWhiteReq.BaseID,
					RankID:     blackWhiteReq.RankID,
					Batch:      rule.LastBatch,
					WhiteScore: v.Score,
				}
				if v.ObjectType == rankmdl.ObjectTypeArchive {
					object.AID = v.Oid
					if arc, ok := archive[v.Oid]; ok {
						object.MID = arc.GetAuthor().Mid

					}
					archiveUpdate = append(archiveUpdate, object)
				}
				if v.ObjectType == rankmdl.ObjectTypeUp {
					object.MID = v.Oid
					upUpdate = append(upUpdate, object)
				}
				if v.ObjectType == rankmdl.ObjectTypeTag {
					object.TagID = v.Oid
					tagUpdate = append(tagUpdate, object)
				}
			}
			var (
				tx *xsql.Tx
			)
			if tx, err = s.dao.BeginTran(c); err != nil {
				log.Errorc(c, "s.dao.BeginTran() failed. error(%v)", err)
				return err
			}
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					log.Errorc(c, "%v", r)
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
			eg := errgroup.WithContext(c)

			if len(archiveUpdate) > 0 {
				eg.Go(func(ctx context.Context) error {
					for _, v := range archiveUpdate {
						if v != nil {
							err = s.dao.UpdateRankArchive(c, tx, v)
							if err != nil {
								return err
							}
						}
					}
					return nil
				})
			}
			if len(upUpdate) > 0 {
				eg.Go(func(ctx context.Context) error {
					for _, v := range upUpdate {
						if v != nil {
							err = s.dao.UpdateRankUp(c, tx, v)
							if err != nil {
								return err
							}
						}
					}
					return nil
				})

			}
			if len(tagUpdate) > 0 {
				eg.Go(func(ctx context.Context) error {
					for _, v := range tagUpdate {
						if v != nil {
							err = s.dao.UpdateRankTag(c, tx, v)
							if err != nil {
								return err
							}
						}
					}
					return nil
				})
			}
			if err := eg.Wait(); err != nil {
				log.Errorc(c, "eg.Wait error(%v)", err)
				return err
			}

		}

		return
	}
	return

}

// Create rank
func (s *Service) Create(c context.Context, rankReq *rankmdl.CreateReq, userName string) (res *rankmdl.CreateRes, err error) {

	var (
		tx *xsql.Tx
		id int64
	)
	if tx, err = s.dao.BeginTran(c); err != nil {
		log.Errorc(c, "s.dao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
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
	if rankReq.IsType == rankmdl.IsNotType {
		rankReq.Tids = make([]int64, 0)
	}
	rank := &rankmdl.Base{
		Name:         rankReq.Name,
		RankType:     rankReq.RankType,
		IsType:       rankReq.IsType,
		IsShowScore:  rankReq.IsShowScore,
		TidsStruct:   rankReq.Tids,
		Tids:         xstr.JoinInts(rankReq.Tids),
		ArchiveStime: rankReq.ArchiveStime,
		ArchiveEtime: rankReq.ArchiveEtime,
		Author:       userName,
		Authority:    rankReq.Authority,
	}
	source := make([]*rankmdl.Source, 0)
	if len(rankReq.Source) > 0 {
		for _, v := range rankReq.Source {
			if v.SourceID == 0 {
				continue
			}
			source = append(source, v)
		}
	}
	if id, err = s.dao.Create(c, tx, rank, source, rankReq.BlackWhite); err != nil {

		log.Errorc(c, "Create s.dao.Create failed. error(%v)", err)
		return
	}
	res = &rankmdl.CreateRes{
		ID: id,
	}
	return
}

// GetSourceList
func (s *Service) GetSourceList(ctx context.Context, request *rankmdl.SourceListReq) (res *rankmdl.SourceListRsp, err error) {
	var (
		total int
		Page  = &rankmdl.Page{}
		list  []*rankmdl.Source
	)
	res = new(rankmdl.SourceListRsp)

	// 先查询
	oldRank, err := s.dao.GetRankByID(ctx, request.BaseID)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetRankByID (%d) err(%v)", request.BaseID, err)
		return
	}
	if oldRank == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("未找到相应榜单信息"))
		return
	}
	if total, err = s.dao.SourceListTotal(ctx, request.BaseID); err != nil {
		log.Errorc(ctx, "s.dao.SourceListTotal() failed. error(%v)", err)
		return
	}
	if list, err = s.dao.GetSourceList(ctx, request.Pn, request.Ps, request.BaseID); err != nil {
		log.Errorc(ctx, "s.dao.GetSourceList(%d) failed. error(%v)", request.BaseID, err)
		return
	}
	sourceId := make([]int64, 0)
	if len(list) > 0 {
		var sourceInfo []*rankmdl.IdAndName
		for _, v := range list {
			if v.SourceID == 0 {
				continue
			}
			sourceId = append(sourceId, v.SourceID)
		}
		sourceInfo, err = s.getSourceInfo(ctx, oldRank.RankType, sourceId)
		if err != nil {
			log.Errorc(ctx, "s.getSourceInfo err(%v)", err)
			return
		}
		sourceMap := make(map[int64]*rankmdl.IdAndName)
		if len(sourceInfo) > 0 {
			for _, v := range sourceInfo {
				sourceMap[v.ID] = v
			}
		}
		for i, v := range list {
			index := i
			if s, ok := sourceMap[v.SourceID]; ok {
				list[index].Name = s.Name
			}
		}
	}
	Page.Num = request.Pn
	Page.Size = request.Ps
	Page.Total = total
	res.Page = Page
	res.List = list
	return
}

// UploadSource 上传来源列表
func (s *Service) UploadSource(ctx context.Context, baseID int64, sourceID []int64) (err error) {

	// 先查询
	oldRank, err := s.dao.GetRankByID(ctx, baseID)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetRankByID (%d) err(%v)", baseID, err)
		return err
	}
	if oldRank == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("未找到相应榜单信息"))
		return
	}
	var (
		tx *xsql.Tx
	)
	if tx, err = s.dao.BeginTran(ctx); err != nil {
		log.Errorc(ctx, "s.dao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(ctx, "%v", r)
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(ctx, "tx.Rollback() error(%v)", err1)
				err = err1
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(ctx, "tx.Commit() error(%v)", err)
		}
		return
	}()

	source := make([]*rankmdl.Source, 0)
	for _, v := range sourceID {
		source = append(source, &rankmdl.Source{
			SourceType: oldRank.RankType,
			SourceID:   v,
			State:      rankmdl.SourceStateOnline,
		})
	}
	err = s.dao.AddRankSourceEliminateOld(ctx, tx, baseID, source)
	if err != nil {
		log.Errorc(ctx, "s.dao.AddRankSourceEliminateOld err(%v)", err)
		return
	}
	return
}

// checkAuthority
func (s *Service) checkAuthority(c context.Context, userName string, author string, authority string) bool {
	if userName == author {
		return true
	}

	for _, v := range s.c.Rank.Admin {
		if v == userName {
			return true
		}
	}
	if authority == "" {
		return false
	}
	for _, v := range strings.Split(authority, ",") {
		if v == userName {
			return true
		}
	}
	return false

}

// changeRankState ...
func (s *Service) changeRankState(c context.Context) (err error) {
	err = s.changeNotStartRank(c)
	if err != nil {
		log.Errorc(c, "s.changeNotStartRank err(%v)", err)
	}
	err = s.changeNotEndRank(c)
	if err != nil {
		log.Errorc(c, "s.changeNotEndRank err(%v)", err)
	}
	return
}

func (s *Service) changeNotStartRank(c context.Context) (err error) {
	rank, err := s.getAlreadyStartRank(c)
	if err != nil {
		return
	}
	if rank != nil && len(rank) > 0 {
		ids := make([]int64, 0)
		for _, v := range rank {
			ids = append(ids, v.ID)
		}
		err = s.updateRankState(c, ids, rankmdl.RankStateStart)
		if err != nil {
			return
		}
	}
	return
}

// getAlreadyStartRank 获取状态未开始实际已经开始的榜单
func (s *Service) getAlreadyStartRank(c context.Context) (rank []*rankmdl.Rule, err error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	rank, err = s.dao.GetRankByStateAndTime(c, rankmdl.RankStateNotStart, fmt.Sprintf("and stime<'%s' and etime>'%s'", now, now))
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByStateAndTime err(%v)", err)
		return
	}
	return
}

func (s *Service) changeNotEndRank(c context.Context) (err error) {
	rank, err := s.getAlreadyEndRank(c)
	if err != nil {
		return
	}
	if rank != nil && len(rank) > 0 {
		ids := make([]int64, 0)
		for _, v := range rank {
			ids = append(ids, v.ID)
		}
		err = s.updateRankState(c, ids, rankmdl.RankStateEnd)
		if err != nil {
			return
		}
	}
	return
}

// getAlreadyEndRank 获取状态已开始实际已经结束的榜单
func (s *Service) getAlreadyEndRank(c context.Context) (rank []*rankmdl.Rule, err error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	rank, err = s.dao.GetRankByStateAndTime(c, rankmdl.RankStateStart, fmt.Sprintf("and etime<'%s'", now))
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByStateAndTime err(%v)", err)
		return
	}
	return
}

// updateRankState 更新榜单状态
func (s *Service) updateRankState(c context.Context, ruleIDs []int64, state int) (err error) {
	return s.dao.UpdateRuleState(c, ruleIDs, state)
}

// Update ...
func (s *Service) Update(c context.Context, rankReq *rankmdl.UpdateReq, userName string) (err error) {
	var (
		tx *xsql.Tx
	)
	// 先查询
	oldRank, err := s.dao.GetRankByID(c, rankReq.ID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByID (%d) err(%v)", rankReq.ID, err)
		return err
	}
	if !s.checkAuthority(c, userName, oldRank.Author, oldRank.Authority) {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("没有配置权限，请联系%s增加权限", oldRank.Author))
		return
	}
	if rankReq.RankType != oldRank.RankType {
		err = ecode.Error(ecode.RequestErr, "无法修改排行榜类型")
		return
	}
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
	if rankReq.IsType == rankmdl.IsNotType {
		rankReq.Tids = make([]int64, 0)
	}
	rank := &rankmdl.Base{
		ID:           rankReq.ID,
		Name:         rankReq.Name,
		IsShowScore:  rankReq.IsShowScore,
		RankType:     rankReq.RankType,
		IsType:       rankReq.IsType,
		TidsStruct:   rankReq.Tids,
		Tids:         xstr.JoinInts(rankReq.Tids),
		ArchiveStime: rankReq.ArchiveStime,
		ArchiveEtime: rankReq.ArchiveEtime,
		Authority:    rankReq.Authority,
		State:        rankReq.State,
	}
	source := make([]*rankmdl.Source, 0)
	if len(rankReq.Source) > 0 {
		for _, v := range rankReq.Source {
			if v.SourceID != 0 {
				source = append(source, v)
			}
		}
	}

	if err = s.dao.Update(c, tx, rank, source, rankReq.BlackWhite); err != nil {

		log.Errorc(c, "Create s.dao.Update failed. error(%v)", err)
		return
	}
	return
}

// List get lottery information list
func (s *Service) List(c context.Context, request *rankmdl.ListReq) (rsp rankmdl.ListRsp, err error) {
	var (
		total int
		Page  = &rankmdl.Page{}
		list  []*rankmdl.Base
	)
	res := make([]*rankmdl.Res, 0)

	if total, err = s.dao.ListTotal(c, request.State, request.Keyword, request.RankType, request.ValidTime); err != nil {
		log.Errorc(c, "s.dao.ListTotal() failed. error(%v)", err)
		return
	}
	if list, err = s.dao.GetRankList(c, request.Pn, request.Ps, request.State, request.Keyword, request.RankType, request.ValidTime); err != nil {
		log.Errorc(c, "s.dao.GetRankList(%v,%v,%v,%v) failed. error(%v)", request.Pn, request.Ps, request.State, request.Keyword, err)
		return
	}
	if list != nil && len(list) > 0 {
		sourcesBatch, err := s.getSourceBatch(c, list)
		if err != nil {
			log.Errorc(c, "s.getSourceBatch err(%v)", err)
		}
		for _, v := range list {
			var tid = make([]*rankmdl.IdAndName, 0)
			var source = make([]*rankmdl.IdAndName, 0)
			sources, ok := sourcesBatch[v.ID]
			if ok {
				source = sources
			}
			if v.IsType == rankmdl.IsType {
				var tids = make([]*rankmdl.IdAndName, 0)
				if v.TidsStruct != nil && len(v.TidsStruct) > 0 {
					tids, err = s.getTidInfo(c, v.TidsStruct)
					if err != nil {
						log.Errorc(c, "s.getTidInfo err(%v)", err)
					}
				}
				tid = tids
			}
			base := &rankmdl.BaseRes{
				ID:           v.ID,
				Name:         v.Name,
				RankType:     v.RankType,
				IsType:       v.IsType,
				Tids:         tid,
				ArchiveStime: v.ArchiveStime,
				ArchiveEtime: v.ArchiveEtime,
				State:        v.State,
				Author:       v.Author,
				Authority:    v.Authority,
				Ctime:        v.Ctime,
				Mtime:        v.Mtime,
			}

			res = append(res, &rankmdl.Res{
				BaseRes: base,
				Source:  source,
			})
		}

	}

	Page.Num = request.Pn
	Page.Size = request.Ps
	Page.Total = total
	rsp.Page = Page
	rsp.List = res
	return
}

// RankRuleOffline ...
func (s *Service) RankRuleOffline(c context.Context, ruleReq *rankmdl.RuleOfflineReq, userName string) (err error) {
	// 先查询
	rule, err := s.dao.GetRuleByID(c, ruleReq.RankID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRuleByID (%d) err(%v)", ruleReq.RankID, err)
		return err
	}
	// 先查询
	oldRank, err := s.dao.GetRankByID(c, rule.BaseID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByID (%d) err(%v)", rule.BaseID, err)
		return err
	}
	if !s.checkAuthority(c, userName, oldRank.Author, oldRank.Authority) {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("没有配置权限，请联系%s增加权限", oldRank.Author))
		return
	}
	rule.State = rankmdl.RankStateDelete
	err = s.dao.UpdateRankRuleState(c, rule)
	if err != nil {
		log.Errorc(c, "s.dao.UpdateRankRuleState err(%v)", err)
	}
	return
}

// UpdateRules 更新rules
func (s *Service) UpdateRules(c context.Context, ruleReq *rankmdl.RuleUpdateReq, userName string) (err error) {
	// 先查询
	oldRank, err := s.dao.GetRankByID(c, ruleReq.BaseID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByID (%d) err(%v)", ruleReq.BaseID, err)
		return err
	}
	if !s.checkAuthority(c, userName, oldRank.Author, oldRank.Authority) {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("没有配置权限，请联系%s增加权限", oldRank.Author))
		return
	}
	var state = rankmdl.RankStateNotStart
	if int64(ruleReq.Stime) < time.Now().Unix() {
		state = rankmdl.RankStateStart
	}
	if int64(ruleReq.Etime) < time.Now().Unix() {
		state = rankmdl.RankStateEnd
	}

	var (
		tx *xsql.Tx
	)
	if tx, err = s.dao.BeginTran(c); err != nil {
		log.Errorc(c, "s.dao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
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

	rule := &rankmdl.Rule{
		ID:              ruleReq.ID,
		BaseID:          ruleReq.BaseID,
		Name:            ruleReq.Name,
		StatisticsType:  ruleReq.StatisticsType,
		Nums:            ruleReq.Nums,
		UpdateFrequency: ruleReq.UpdateFrequency,
		Unit:            ruleReq.Unit,
		Precision:       ruleReq.Precision,
		Description:     ruleReq.Description,
		UpdateScope:     ruleReq.UpdateScope,
		State:           state,
		Stime:           ruleReq.Stime,
		Etime:           ruleReq.Etime,
	}
	if ruleReq.ID == 0 {
		err = s.dao.AddRankRule(c, tx, rule, ruleReq.ScoreConfig)
		if err != nil {
			log.Errorc(c, "s.dao.AddRankRule err(%v)", err)
			return
		}
		return
	}
	if ruleReq.ID != 0 {

		oldRule, err := s.dao.GetRuleByID(c, ruleReq.ID)
		if err != nil {
			log.Errorc(c, "s.dao.GetRuleByID err(%v)", err)
			return err
		}
		if ruleReq.StatisticsType != oldRule.StatisticsType {
			err = ecode.Error(ecode.RequestErr, "不能修改榜单排名维度")
			return err
		}
		err = s.dao.UpdateRankRule(c, tx, rule, ruleReq.ScoreConfig)
		if err != nil {
			log.Errorc(c, "s.dao.AddRankRule err(%v)", err)
			return err
		}
	}

	return nil
}

// UpdateRules 更新rules
func (s *Service) UpdateRulesShowInfo(c context.Context, ruleReq *rankmdl.RuleUpdateShowReq, userName string) (err error) {
	// 先查询
	oldRank, err := s.dao.GetRankByID(c, ruleReq.BaseID)
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByID (%d) err(%v)", ruleReq.BaseID, err)
		return err
	}
	if !s.checkAuthority(c, userName, oldRank.Author, oldRank.Authority) {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("没有配置权限，请联系%s增加权限", oldRank.Author))
		return
	}

	err = s.dao.UpdateRankRuleShow(c, ruleReq.ID, ruleReq.Unit, ruleReq.Precision, ruleReq.Description)
	if err != nil {
		log.Errorc(c, "s.dao.AddRankRule err(%v)", err)
		return err
	}

	return nil
}

// getSourceBatch
func (s *Service) getSourceBatch(c context.Context, bases []*rankmdl.Base) (res map[int64][]*rankmdl.IdAndName, err error) {
	baseIds := make([]int64, 0)
	baseSourceType := make(map[int64]int)
	filterTagSource := make(map[int64][]*rankmdl.Source)
	filterActivitySource := make(map[int64][]*rankmdl.Source)
	filterMidSource := make(map[int64][]*rankmdl.Source)
	res = make(map[int64][]*rankmdl.IdAndName)
	for _, v := range bases {
		baseIds = append(baseIds, v.ID)
		baseSourceType[v.ID] = v.RankType
		if v.RankType == rankmdl.SourceTypeActivity {
			filterActivitySource[v.ID] = make([]*rankmdl.Source, 0)
			continue
		}
		if v.RankType == rankmdl.SourceTypeTag {
			filterTagSource[v.ID] = make([]*rankmdl.Source, 0)
			continue
		}
		if v.RankType == rankmdl.SourceTypeMid {
			filterMidSource[v.ID] = make([]*rankmdl.Source, 0)
			continue
		}
	}
	source, err := s.dao.GetSourceBatch(c, baseIds)
	if err != nil {
		log.Errorc(c, "s.dao.GetSourceBatch err(%v)", err)
		return
	}
	mapTag := make(map[int64]struct{})
	mapActivity := make(map[int64]struct{})
	mapMid := make(map[int64]struct{})
	tagID := make([]int64, 0)
	activityID := make([]int64, 0)
	mids := make([]int64, 0)
	for _, v := range source {
		if v.SourceID == 0 {
			continue
		}
		sourceType, ok := baseSourceType[v.BaseID]
		if ok && sourceType == v.SourceType {
			if sourceType == rankmdl.SourceTypeTag {
				filterTagSource[v.BaseID] = append(filterTagSource[v.BaseID], v)
				if _, ok := mapTag[v.SourceID]; !ok {
					if v.SourceID > 0 {
						tagID = append(tagID, v.SourceID)
					}
					mapTag[v.SourceID] = struct{}{}
				}
				continue
			}
			if sourceType == rankmdl.SourceTypeActivity {
				filterActivitySource[v.BaseID] = append(filterActivitySource[v.BaseID], v)
				if _, ok := mapActivity[v.SourceID]; !ok {
					if v.SourceID > 0 {
						activityID = append(activityID, v.SourceID)
					}
					mapActivity[v.SourceID] = struct{}{}
				}
				continue
			}
			if sourceType == rankmdl.SourceTypeMid {
				filterMidSource[v.BaseID] = append(filterMidSource[v.BaseID], v)
				if _, ok := mapMid[v.SourceID]; !ok {
					if v.SourceID > 0 {
						mids = append(mids, v.SourceID)
					}
					mapActivity[v.SourceID] = struct{}{}
				}
				continue
			}
		}
	}
	var tags = make([]*rankmdl.IdAndName, 0)
	mapTags := make(map[int64]*rankmdl.IdAndName)
	var activity = make([]*rankmdl.IdAndName, 0)
	mapActivitys := make(map[int64]*rankmdl.IdAndName)
	var mid = make([]*rankmdl.IdAndName, 0)
	mapMids := make(map[int64]*rankmdl.IdAndName)

	if len(tagID) > 0 {
		tags, err = s.getTagInfo(c, tagID)
		if err != nil {
			log.Errorc(c, "s.getTagInfo (%v)", err)
			return
		}
		if len(tags) > 0 {
			for _, v := range tags {
				mapTags[v.ID] = v
			}
		}

	}
	if len(activityID) > 0 {
		activity, err = s.getActivityName(c, activityID)
		if err != nil {
			log.Errorc(c, "s.getActivityName (%v)", err)
			return
		}
		if len(activity) > 0 {
			for _, v := range activity {
				mapActivitys[v.ID] = v
			}
		}
	}
	if len(mids) > 0 {
		mid, err = s.getMidName(c, mids)
		if err != nil {
			log.Errorc(c, "s.getMidName (%v)", err)
			return
		}
		if len(mid) > 0 {
			for _, v := range mid {
				mapMids[v.ID] = v
			}
		}
	}

	for baseID, v := range filterTagSource {
		resData := make([]*rankmdl.IdAndName, 0)
		for _, source := range v {
			if data, ok := mapTags[source.SourceID]; ok {
				resData = append(resData, data)
			}
		}
		res[baseID] = resData
	}
	for baseID, v := range filterActivitySource {
		resData := make([]*rankmdl.IdAndName, 0)
		for _, source := range v {
			if data, ok := mapActivitys[source.SourceID]; ok {
				resData = append(resData, data)
			}
		}
		res[baseID] = resData
	}
	for baseID, v := range filterMidSource {
		resData := make([]*rankmdl.IdAndName, 0)
		for _, source := range v {
			if data, ok := mapMids[source.SourceID]; ok {
				resData = append(resData, data)
			}
		}
		res[baseID] = resData
	}
	return

}

func (s *Service) getSource(c context.Context, id int64, sourceType int) (res []*rankmdl.IdAndName, err error) {
	source, err := s.dao.GetSource(c, id, sourceType)
	if err != nil {
		log.Errorc(c, "s.dao.GetSource err(%v)", err)
		return
	}
	sourceIds := make([]int64, 0)
	if source != nil {
		for _, v := range source {
			sourceIds = append(sourceIds, v.SourceID)
		}
	}
	if len(sourceIds) > 0 {
		return s.getSourceInfo(c, sourceType, sourceIds)
	}
	return
}

func (s *Service) getSourceInfo(ctx context.Context, sourceType int, sourceIDs []int64) (res []*rankmdl.IdAndName, err error) {
	if sourceType == rankmdl.SourceTypeActivity {
		return s.getActivityName(ctx, sourceIDs)
	}
	if sourceType == rankmdl.SourceTypeTag {
		return s.getTagInfo(ctx, sourceIDs)
	}
	if sourceType == rankmdl.SourceTypeMid {
		return s.getMidName(ctx, sourceIDs)
	}
	return
}

// getTagInfo ...
func (s *Service) getTagInfo(c context.Context, tagsID []int64) (res []*rankmdl.IdAndName, err error) {
	var times int
	patch := tagBatch
	times = len(tagsID) / patch / concurrency
	tagInfo := make(map[int64]string)
	tagAllRes := make([]*tagrpc.TagsReply, 0)
	res = make([]*rankmdl.IdAndName, 0)
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(tagsID) {
					return nil
				}
				reqTag := tagsID[start:]
				end := start + patch
				if end < len(tagsID) {
					reqTag = tagsID[start:end]
				}
				if len(reqTag) > 0 {
					tagRes, err := s.tagRPC.Tags(c, &tagrpc.TagsReq{Tids: reqTag})
					if err != nil || tagRes == nil || tagRes.Tags == nil {
						err = errors.Wrapf(err, "s.tagRPC.TagByNames")
						return err
					}
					tagAllRes = append(tagAllRes, tagRes)
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Errorc(c, "eg.Wait error(%v)", err)
			return nil, err
		}
	}
	for _, tagRes := range tagAllRes {
		if tagRes != nil {
			for _, value := range tagRes.Tags {
				if value != nil {
					tagInfo[value.Id] = value.Name
				}
			}
		}
	}
	for _, v := range tagsID {
		types, ok := tagInfo[v]
		if ok {
			res = append(res, &rankmdl.IdAndName{
				ID:   v,
				Name: types,
			})
		} else {
			res = append(res, &rankmdl.IdAndName{
				ID:   v,
				Name: "错误!未能获取TID信息!",
			})
		}
	}
	return
}

func (s *Service) getActivityName(c context.Context, sid []int64) (res []*rankmdl.IdAndName, err error) {
	res = make([]*rankmdl.IdAndName, 0)
	act, err := s.likedao.ActSubjects(c, sid)
	if err != nil {
		log.Errorc(c, "s.likedao.ActSubjects err(%v)", err)
		return
	}
	if act != nil {
		actMap := make(map[int64]string)
		for _, v := range act {
			actMap[v.ID] = v.Name
		}
		for _, v := range sid {
			types, ok := actMap[v]
			if ok {
				res = append(res, &rankmdl.IdAndName{
					ID:   v,
					Name: types,
				})
			} else {
				res = append(res, &rankmdl.IdAndName{
					ID:   v,
					Name: "错误!未能获取活动信息!",
				})
			}
		}
	}

	return
}

func (s *Service) getMidName(ctx context.Context, mids []int64) (res []*rankmdl.IdAndName, err error) {
	res = make([]*rankmdl.IdAndName, 0)
	acc, err := s.account.MemberInfo(ctx, mids)
	if err != nil {
		log.Errorc(ctx, "s.account.MemberInfo err(%v)", err)
		return
	}
	if acc != nil {
		actMap := make(map[int64]string)
		for _, v := range acc {
			actMap[v.Mid] = v.Name
		}
		for _, v := range mids {
			types, ok := actMap[v]
			if ok {
				res = append(res, &rankmdl.IdAndName{
					ID:   v,
					Name: types,
				})
			} else {
				res = append(res, &rankmdl.IdAndName{
					ID:   v,
					Name: "错误!未能获取用户信息!",
				})
			}
		}
	}
	return
}

// GetRank ...
func (s *Service) GetRank(c context.Context, id int64) (res *rankmdl.Res, err error) {
	res = &rankmdl.Res{}
	base, err := s.getBase(c, id)
	if err != nil {
		log.Errorc(c, "s.getBase err(%v)", err)
		return
	}
	source, err := s.getSource(c, id, base.RankType)
	if err != nil {
		log.Errorc(c, "s.getSource err(%v)", err)
		return
	}

	rule, err := s.getRule(c, id)
	if err != nil {
		log.Errorc(c, "s.getRule err(%v)", err)
		return
	}
	black, err := s.getUpBlack(c, id)
	if err != nil {
		log.Errorc(c, "s.getRule err(%v)", err)
		return
	}
	res.BaseRes = base
	res.Source = source
	res.Rule = rule
	res.Black = black
	return

}

func (s *Service) getUpBlack(c context.Context, id int64) (res []*rankmdl.BlackWhiteRes, err error) {
	black, err := s.dao.GetBlackOrWhite(c, id, rankmdl.BlackWhiteInterventionTypeBlack, rankmdl.BlackWhiteObjectTypeUp)
	if err != nil {
		log.Errorc(c, "s.dao.GetBlackOrWhite err(%v)", err)
		return
	}
	res = make([]*rankmdl.BlackWhiteRes, 0)
	mids := make([]int64, 0)
	if black != nil && len(black) > 0 {
		for _, v := range black {
			mids = append(mids, v.Oid)
		}
		mapMidAccount, err := s.account.MemberInfo(c, mids)
		if err != nil {
			log.Errorc(c, "s.account.MemberInfo (%v)", err)
			return nil, err
		}
		for _, v := range black {
			if acc, ok := mapMidAccount[v.Oid]; ok {
				res = append(res, &rankmdl.BlackWhiteRes{
					Name:             acc.Name,
					ID:               v.ID,
					BaseID:           v.BaseID,
					Oid:              v.Oid,
					Score:            v.Score,
					State:            v.State,
					InterventionType: v.InterventionType,
					ObjectType:       v.ObjectType,
					Ctime:            v.Ctime,
					Mtime:            v.Mtime,
				})
			}
		}
	}
	return

}

func (s *Service) getRule(c context.Context, id int64) (res []*rankmdl.Rule, err error) {
	res = make([]*rankmdl.Rule, 0)
	data, err := s.dao.GetRule(c, id)
	if err != nil {
		log.Errorc(c, "s.dao.GetRule err(%v)", err)
		return
	}
	if data == nil {
		return res, nil
	}
	return data, nil
}

func (s *Service) getBase(c context.Context, id int64) (base *rankmdl.BaseRes, err error) {
	// 先查询
	rank, err := s.dao.GetRankByID(c, id)
	if err != nil {
		log.Errorc(c, "s.dao.GetRankByID (%d) err(%v)", id, err)
		return nil, err
	}
	base = &rankmdl.BaseRes{
		ID:           rank.ID,
		Name:         rank.Name,
		RankType:     rank.RankType,
		IsType:       rank.IsType,
		ArchiveStime: rank.ArchiveStime,
		ArchiveEtime: rank.ArchiveEtime,
		State:        rank.State,
		Author:       rank.Author,
		Authority:    rank.Authority,
		Ctime:        rank.Ctime,
		Mtime:        rank.Mtime,
	}
	var tids = make([]*rankmdl.IdAndName, 0)
	if rank.TidsStruct != nil && len(rank.TidsStruct) > 0 {
		tids, err = s.getTidInfo(c, rank.TidsStruct)
		if err != nil {
			log.Errorc(c, "s.getTidInfo err(%v)", err)
		}
	}
	base.Tids = tids
	return
}

func (s *Service) getTidInfo(c context.Context, tid []int64) (tids []*rankmdl.IdAndName, err error) {
	var TypesReply *arcgrpc.TypesReply
	tids = make([]*rankmdl.IdAndName, 0)

	if TypesReply, err = s.arcClient.Types(c, &arcgrpc.NoArgRequest{}); err != nil {
		return tids, err
	}
	for _, v := range tid {
		types, ok := TypesReply.Types[int32(v)]
		if ok {
			tids = append(tids, &rankmdl.IdAndName{
				ID:   v,
				Name: types.Name,
			})
		} else {
			tids = append(tids, &rankmdl.IdAndName{
				ID:   v,
				Name: "错误!未能获取TID信息!",
			})
		}
	}
	return
}
