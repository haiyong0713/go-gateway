package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/component/boss"
	"go-gateway/app/web-svr/activity/admin/service/exporttask"
	"go-gateway/app/web-svr/activity/interface/api"
	"strconv"
	"time"
)

func (s *Service) ExportNewVoteDetail(ctx context.Context, userName string, activityId int64) {
	res, err := s.dao.GetNewVoteUserHistory(ctx, activityId)
	if err != nil {
		log.Errorc(ctx, "actSvr.GetNewVoteUserHistory(aid:%v) failed. error(%v)", activityId, err)
		_ = exporttask.SendWeChatTextMessage(ctx, []string{userName}, fmt.Sprintf("投票明细 %d 导出数据失败，请重试或者在救火大队反馈跟进。", activityId))
		return
	}
	b := &bytes.Buffer{}
	b.WriteString("\xEF\xBB\xBF")
	wr := csv.NewWriter(b)
	header := []string{"投票数据源类型", "数据源ID", "内容ID", "Mid", "投票时间", "状态"}
	_ = wr.Write(header)
	for _, r := range res {
		_ = wr.Write([]string{r.SourceType, strconv.FormatInt(r.SourceGroupId, 10),
			strconv.FormatInt(r.SourceItemId, 10), strconv.FormatInt(r.Mid, 10),
			r.Ctime.Time().Format(time.RFC3339), r.State})
	}
	wr.Flush()
	url, err := boss.Client.UploadObject(ctx, boss.Bucket, fmt.Sprintf("newvotedata/%s/投票明细-%d.csv", time.Now().Format("20060102150405"), activityId), b)
	if err != nil {
		log.Errorc(ctx, "actSvr.ExportNewVoteDetail(aid:%v) failed. error(%v)", activityId, err)
		_ = exporttask.SendWeChatTextMessage(ctx, []string{userName}, fmt.Sprintf("投票明细 %d 导出数据失败，请重试或者在救火大队反馈跟进。", activityId))
		return
	}
	_ = exporttask.SendWeChatTextMessage(ctx, []string{userName}, fmt.Sprintf("投票明细 %d 导出成功，下载链接:%s", activityId, url))
	return
}

func (s *Service) ExportNewVoteRank(ctx context.Context, userName string, sourceGroupId int64) {
	res, err := s.actClient.GetVoteActivityRankInternal(ctx, &api.GetVoteActivityRankInternalReq{
		SourceGroupId: sourceGroupId,
		Pn:            1,
		Ps:            20000,
	})
	if err != nil {
		log.Errorc(ctx, "actSvr.GetNewVoteUserHistory(aid:%v) failed. error(%v)", sourceGroupId, err)
		_ = exporttask.SendWeChatTextMessage(ctx, []string{userName}, fmt.Sprintf("投票数据组 %d 导出排行榜失败，请重试或者在救火大队反馈跟进。", sourceGroupId))
		return
	}
	b := &bytes.Buffer{}
	b.WriteString("\xEF\xBB\xBF")
	wr := csv.NewWriter(b)
	header := []string{"排名", "稿件ID", "稿件名", "总票数", "普通票数", "干预票数", "风控票数"}
	_ = wr.Write(header)
	for idx, r := range res.Rank {
		_ = wr.Write([]string{
			strconv.FormatInt(int64(idx+1), 10), strconv.FormatInt(r.SourceItemId, 10),
			r.SourceItemName,
			strconv.FormatInt(r.TotalVoteCount, 10), strconv.FormatInt(r.UserVoteCount, 10),
			strconv.FormatInt(r.InterveneVoteCount, 10), strconv.FormatInt(r.RiskVoteCount, 10)})
	}
	wr.Flush()
	url, err := boss.Client.UploadObject(ctx, boss.Bucket, fmt.Sprintf("newvotedata/%s/投票排行-%d.csv", time.Now().Format("20060102150405"), sourceGroupId), b)
	if err != nil {
		log.Errorc(ctx, "actSvr.ExportNewVoteDetail (gid:%v) failed. error(%v)", sourceGroupId, err)
		_ = exporttask.SendWeChatTextMessage(ctx, []string{userName}, fmt.Sprintf("投票数据组 %d 导出排行榜失败，请重试或者在救火大队反馈跟进。", sourceGroupId))
		return
	}
	_ = exporttask.SendWeChatTextMessage(ctx, []string{userName}, fmt.Sprintf("投票数据组 %d 导出排行榜成功，下载链接:%s", sourceGroupId, url))
	return
}
