package lol

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/lol"
	"go-gateway/app/web-svr/activity/interface/service/s10"

	formula "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

func (s *Service) PointList(c context.Context, mid int64, timestamp, ps int64) (res []*lol.PointMsg, err error) {
	if ps > 200 {
		ps = 200
	}
	s10Conf := conf.LoadS10CoinCfg()
	var mcList, hisList, record []*lol.PointMsg
	var mcErr, grpcErr error
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}
	hisReq := &formula.GetFormulaHistoryReq{
		Activity: s10Conf.ActivityName,
		Formula:  lol.FormulaName,
		Mid:      mid,
	}
	var srvRes interface{}
	if srvRes, grpcErr = client.FetchResourceFromActPlatform(c, client.Path4ActPlatformOfHistory,
		client.FetchActPlatformHistory, hisReq); grpcErr == nil {
		for _, his := range srvRes.(*formula.GetFormulaHistoryResp).History {
			if _, ok := lol.OpTypeTipMap[his.Counter]; !ok && his.Activity != s10Conf.ActivityName {
				continue
			}
			name := lol.OpTypeTipMap[his.Counter]
			if name == "" {
				name = his.Counter
			}
			hisList = append(hisList, &lol.PointMsg{
				Timestamp: his.Timestamp,
				OpTypeTip: name,
				OpType:    his.Counter,
				Point:     his.Count,
			})
		}
	}

	if cs, csErr := s10.UserCostRecord(c, mid); csErr == nil {
		for _, r := range cs {
			name := "goods"
			record = append(record, &lol.PointMsg{
				Timestamp: int64(r.Ctime),
				OpTypeTip: fmt.Sprintf(lol.OpTypeTipMap[name], r.Name),
				OpType:    name,
				Point:     -int64(r.Cost),
			})
		}
	}

	if s10Conf.McSwitch {
		mcList, mcErr = s.dao.PointList(c, mid)
	}

	if grpcErr != nil && ((s10Conf.McSwitch && mcErr != nil) || !s10Conf.McSwitch) {
		err = ecode.ActivityPointDetailFetchFailed
		return
	}

	mcList = trimRecordList(hisList, mcList, record)
	for i, l := range mcList {
		if l.Timestamp < timestamp {
			res = make([]*lol.PointMsg, 0)
			size := len(mcList) - i
			switch {
			case size > int(ps):
				res = mcList[i : i+int(ps)]
			case size <= int(ps):
				res = mcList[i:]
			}
			break
		}
	}
	if s10Conf.McSwitch && (len(hisList) != 0 || len(record) != 0) {
		_ = s.cache.Do(c, func(c context.Context) {
			if bs, err := json.Marshal(mcList); err == nil {
				_ = s.dao.SetPointList(c, mid, bs)
			}
		})
	}
	return
}

func trimRecordList(hisList, mcList, record []*lol.PointMsg) (res []*lol.PointMsg) {
	listMap := make(map[int64]*lol.PointMsg)
	var list []*lol.PointMsg
	for _, l := range hisList {
		listMap[l.Timestamp] = l
	}
	for _, l := range record {
		listMap[l.Timestamp] = l
	}
	for _, l := range mcList {
		if _, ok := listMap[l.Timestamp]; ok {
			continue
		}
		listMap[l.Timestamp] = l
	}
	for _, l := range listMap {
		list = append(list, l)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp > list[j].Timestamp
	})
	size := len(list)
	res = make([]*lol.PointMsg, 0)
	switch {
	case size > 1000:
		res = list[:1000]
	case size <= 1000:
		res = list[:]
	}
	return
}
