package lottery

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/component/boss"
	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"
	actapi "go-gateway/app/web-svr/activity/interface/api"

	"strconv"
	"strings"
)

// MidAddTimesLog ...
func (s *Service) MidAddTimesLog(c context.Context, mid int64, sid string, cid int64, pn, ps int) (res *lotmdl.LotteryAddTimesReply, err error) {
	var (
		list []*lotmdl.LotteryAddTimes
		page = lotmdl.Page{}
		//giftNum  map[int64]int
	)
	lottery, err := s.lotDao.LotDetailBySID(c, sid)
	if err != nil || lottery == nil {
		log.Errorc(c, "s.lotDao.LotDetailBySID (%s) err(%v)", sid, err)
		return nil, err
	}

	if page.Total, err = s.lotDao.RawLotteryAddTimesTotal(c, lottery.ID, mid, cid); err != nil {
		log.Errorc(c, "s.lot.RawLotteryAddTimesTotal() failed. error(%v)", err)
		return
	}
	page.Num = pn
	page.Size = ps
	if list, err = s.lotDao.RawLotteryAddTimes(c, lottery.ID, mid, cid, pn, ps); err != nil {
		log.Errorc(c, "s.lotDao.RawLotteryAddTimes() failed. error(%v)", err)
		return
	}
	res = &lotmdl.LotteryAddTimesReply{}
	res.List = list
	res.Page = page
	return

}

// MidTimes 用户剩余次数
func (s *Service) MidTimes(c context.Context, mid int64, sid string) (res *actapi.LotteryUnusedTimesReply, err error) {
	return s.actClient.LotteryUnusedTimes(c, &actapi.LotteryUnusedTimesdReq{
		Mid: mid,
		Sid: sid,
	})

}

// AddTimesBatchList get gift list
func (s *Service) AddTimesBatchList(c context.Context, request *lotmdl.BatchAddTimesParams) (rsp *lotmdl.AddTimesBatchLogList, err error) {
	var (
		list []*lotmdl.AddTimesLog
		page = lotmdl.Page{}
		//giftNum  map[int64]int
	)
	if page.Total, err = s.lotDao.BatchAddTimesLogTotal(c, request.SID); err != nil {
		log.Errorc(c, "s.lot.GiftTotal() failed. error(%v)", err)
		return
	}
	page.Num = request.Pn
	page.Size = request.Ps
	if list, err = s.lotDao.BatchAddTimesLogList(c, request.SID, request.Pn, request.Ps); err != nil {
		log.Errorc(c, "s.lotDao.BatchAddTimesLogList() failed. error(%v)", err)
		return
	}
	rsp = &lotmdl.AddTimesBatchLogList{}
	rsp.List = list
	rsp.Page = page
	return
}

// MidAddTimes 用户剩余次数
func (s *Service) MidAddTimes(c context.Context, mid, cid, actionType, batch int64, sid string) (res *actapi.LotteryAddTimesReply, err error) {
	orderNo := fmt.Sprintf("%d_%d_%d_%d", mid, cid, actionType, batch)
	return s.actClient.LotteryAddTimes(c, &actapi.LotteryAddTimesReq{
		Cid:        cid,
		Mid:        mid,
		OrderNo:    orderNo,
		ActionType: actionType,
		Sid:        sid,
	})
}

// DoBatchMidAddTimes ...
func (s *Service) DoBatchMidAddTimes(c context.Context, mids []int64, author string, cid int64, sid string) (err error) {
	batchID, err := s.lotDao.BatchAddTimesLog(c, sid, author, cid)
	if err != nil {
		log.Errorc(c, "s.lotDao.BatchAddTimesLog err(%v)", err)
		return ecode.Error(ecode.RequestErr, "增加日志失败，请重试")
	}
	// 导入至boss
	b := &bytes.Buffer{}
	categoryHeader := []string{"mid"}
	b.WriteString("\xEF\xBB\xBF")
	wr := csv.NewWriter(b)
	_ = wr.Write(categoryHeader)

	for i := 0; i < len(mids); i++ {
		midStrs := make([]string, 0)
		midStrs = append(midStrs, strconv.FormatInt(mids[i], 10))
		wr.Write(midStrs)
	}
	wr.Flush()
	url, err := boss.Client.UploadObject(c, boss.Bucket, fmt.Sprintf("lotterydata/addTimes/%s_%d_%d.csv", sid, cid, batchID), b)
	if err != nil {
		log.Errorc(c, "lotterySrv.UploadObject(sid:%v) failed. error(%v)", sid, err)
		err = s.lotDao.UpdateBatchAddTimesLog(c, batchID, lotmdl.AddTimesBatchLogStateFileError, url)
		if err != nil {
			log.Errorc(c, "s.lotDao.UpdateBatchAddTimesLog err(%v)", err)
			return
		}
		return ecode.Error(ecode.RequestErr, "用户mid导入失败请重试")
	}
	err = s.BatchMidAddTimes(c, mids, batchID, author, cid, sid, url)
	s.doWechat(c, err, sid, cid, author)
	return
}

func (s *Service) doWechat(c context.Context, err error, sid string, cid int64, author string) {
	if err != nil {
		title := fmt.Sprintf("抽奖次数增加失败，请查看状态，抽奖id：%s", sid)
		message := fmt.Sprintf("author (%s) cid (%d)", author, cid)
		err = s.sendWechat(c, title, message, author)
		if err != nil {
			log.Errorc(c, "s.sendWechat err(%v)", err)
			return
		}
	} else {
		title := fmt.Sprintf("增加抽奖次数已完成，抽奖id：%s", sid)
		message := fmt.Sprintf("author (%s) cid (%d) sid(%s) error:(%v)", author, cid, sid, err)
		err = s.sendWechat(c, title, message, author)
		if err != nil {
			log.Errorc(c, "s.sendWechat err(%v)", err)
			return
		}
	}
}

// AddtimesRetry ...
func (s *Service) AddtimesRetry(c context.Context, id int64, author string) (err error) {
	addTimesLog, err := s.lotDao.AddTimesLogByID(c, id)
	if err != nil {
		log.Errorc(c, "s.lotDao.AddTimesLogByID id(%d)", id)
		return err
	}
	object, err := boss.Client.GetObject(c, boss.Bucket, fmt.Sprintf("lotterydata/addTimes/%s_%d_%d.csv", addTimesLog.Sid, addTimesLog.Cid, id))
	if err != nil {
		log.Errorc(c, "boss.Client.GetObject(sid:%v) failed. error(%v)", addTimesLog.Sid, err)
		return ecode.Error(ecode.RequestErr, "获取文件失败")
	}
	if object.Body == nil {
		return ecode.Error(ecode.RequestErr, "获取文件失败")
	}
	r := csv.NewReader(object.Body)
	records, err := r.ReadAll()
	if err != nil {
		log.Errorc(c, "importDetailCSV r.ReadAll() err(%v)", err)
		return ecode.Error(ecode.RequestErr, "获取文件失败")
	}
	var args []int64
Loop:
	for i, row := range records {
		if i == 0 {
			continue
		}
		// import csv state online
		var arg int64
		for field, value := range row {
			value = strings.TrimSpace(value)
			switch field {
			case 0:
				if value == "" {
					log.Warn("importDetailCSV name provinceID(%s)", value)
					continue Loop
				}
				mid, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					// continue Loop
				}
				arg = mid

			}

		}
		args = append(args, arg)
	}
	err = s.BatchMidAddTimes(c, args, id, author, addTimesLog.Cid, addTimesLog.Sid, "")
	s.doWechat(c, err, addTimesLog.Sid, addTimesLog.Cid, author)
	return nil

}

// BatchMidAddTimes 批量增加抽奖次数
func (s *Service) BatchMidAddTimes(c context.Context, mids []int64, batchID int64, author string, cid int64, sid string, url string) (err error) {
	defer func() {
		if err != nil {
			if url == "" {
				err = s.lotDao.UpdateBatchAddTimesLogState(c, batchID, lotmdl.AddTimesBatchLogStateError)
				if err != nil {
					log.Errorc(c, "s.lotDao.UpdateBatchAddTimesLog err(%v)", err)
					return
				}
				return
			}
			err = s.lotDao.UpdateBatchAddTimesLog(c, batchID, lotmdl.AddTimesBatchLogStateError, url)
			if err != nil {
				log.Errorc(c, "s.lotDao.UpdateBatchAddTimesLog err(%v)", err)
				return
			}
			return
		}
		if url == "" {
			err = s.lotDao.UpdateBatchAddTimesLogState(c, batchID, lotmdl.AddTimesBatchLogStateFinish)
			if err != nil {
				log.Errorc(c, "s.lotDao.UpdateBatchAddTimesLog err(%v)", err)
				return
			}
			return
		}
		err = s.lotDao.UpdateBatchAddTimesLog(c, batchID, lotmdl.AddTimesBatchLogStateFinish, url)
		if err != nil {
			log.Errorc(c, "s.lotDao.UpdateBatchAddTimesLog err(%v)", err)
			return
		}
	}()

	times, err := s.lotDao.TimesConfigByID(c, cid, sid)
	if err != nil {
		log.Errorc(c, "s.lotDao.TimesConfigByID err(%v)", err)
		return ecode.Error(ecode.RequestErr, "获取次数失败")

	}
	if times == nil {
		return ecode.Error(ecode.RequestErr, "获取次数失败")
	}
	for _, v := range mids {
		_, err = s.MidAddTimes(c, v, cid, int64(times.Type), batchID, sid)
		if err != nil {
			log.Errorc(c, " s.MidAddTimes err(%v)", err)
			return ecode.Error(ecode.RequestErr, fmt.Sprintf("添加次数失败，请重试 batchID(%d) err(%v)", batchID, err))
		}
	}
	return err
}
