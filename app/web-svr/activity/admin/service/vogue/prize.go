package vogue

import (
	"context"
	"strconv"

	"go-common/library/log"
	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

// List get goods information list
func (s *Service) ListPrizes(c context.Context, search *voguemdl.PrizeSearch) (rsp *voguemdl.PrizesListRsp, err error) {
	var (
		list    []*voguemdl.PrizeData
		count   int64
		uidList []int64
	)
	rsp = &voguemdl.PrizesListRsp{}
	if list, count, err = s.dao.PrizesList(c, search); err != nil {
		log.Error("[ListPrizes] s.dao.PrizesList error(%v)", err)
	}

	for _, item := range list {
		uidList = append(uidList, item.Uid)
	}
	users, err := s.batchUserInfos(c, uidList)
	if err != nil {
		log.Error("[ListPrizes] Fetch users error, uidList:%v, error(%v)", uidList, err)
		return
	}
	var addrTmp = &lotmdl.Address{}
	for _, item := range list {
		var nickname string
		user, ok := users[item.Uid]
		if ok {
			nickname = user.GetName()
		}
		item.NickName = nickname
		item.GoodsAttrReal = item.GoodsAttr & 1
		item.Risk, item.RiskMsg, _ = s.dao.RiskInfo(context.TODO(), item.Uid)
		if item.GoodsAddressId > 0 {
			item.GoodsAddress = "有地址，待获取"
			// 获取地址
			if addrTmp, err = s.lotDao.GetAddressByID(c, item.GoodsAddressId, int(item.Uid)); err != nil {
				log.Errorc(c, "s.lotDao.GetAddressByID() failed. error(%v)", err)
				item.GoodsAddress = "地址获取失败"
			} else {
				item.GoodsAddress = addrTmp.Prov + addrTmp.City + addrTmp.Area + addrTmp.Addr + addrTmp.Name + addrTmp.Phone
			}
		}
		if item.GoodsState == voguemdl.UserTaskStatusExchangeDone {
			item.ExchangeTime = item.Mtime.Time().Format(voguemdl.TimeFormat)
			duration := item.Mtime.Time().Sub(item.Ctime.Time())
			item.TimeCost = duration.String()
		}
	}

	rsp.List = list
	rsp.Page = &voguemdl.Page{
		Size:  search.Ps,
		Num:   search.Pn,
		Total: count,
	}
	return
}

// Export goods
func (s *Service) ExportPrizes(c context.Context, search *voguemdl.PrizeExportSearch) (result [][]string, err error) {
	var (
		rsp *voguemdl.PrizesListRsp
	)
	params := &voguemdl.PrizeSearch{Uid: search.Uid, Pn: 1, Ps: -1}

	log.Info("ExportPrizes, params: %v", params)
	if rsp, err = s.ListPrizes(c, params); err != nil {
		log.Errorc(c, "[ListPrizes] s.dao.PrizesList error(%v)", err)
		return
	}
	for _, item := range rsp.List {
		result = append(result, []string{
			item.NickName,
			strconv.FormatInt(item.Uid, 10),
			item.GoodsName,
			goodsAttrRealToStr[item.GoodsAttrReal],
			item.ExchangeTime,
			item.TimeCost,
			strconv.Itoa(item.GoodsScore),
			item.GoodsAddress,
			item.RiskMsg,
		})
	}
	return
}
