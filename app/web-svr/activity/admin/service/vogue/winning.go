package vogue

import (
	"context"
	"strconv"

	"go-common/library/log"

	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

const (
	_sid = "ea00bdc0-a64e-11ea-8597-246e966235d8"
)

// 获取全部中奖纪录
func (s *Service) WinningList(c context.Context, search *voguemdl.CreditSearch) (rsp *voguemdl.WinningList, err error) {
	var (
		lotInfo *lotmdl.LotInfo
		giftWin []*lotmdl.GiftWinInfo
		uidList []int64
		rspList []*voguemdl.WinningItem
	)

	if lotInfo, err = s.lotDao.LotDetailBySID(c, _sid); err != nil {
		log.Errorc(c, "s.lotDao.LotDetailBySID() failed. error(%v)", err)
		return
	}

	var (
		count int64
	)
	if giftWin, count, err = s.dao.GiftWinList(c, lotInfo.ID, search.Uid, search.Pn, search.Ps); err != nil {
		log.Errorc(c, "s.lotDao.GiftWinList() failed. error(%v)", err)
		return
	}

	for _, item := range giftWin {
		uidList = append(uidList, int64(item.Mid))
	}

	// 获取 uname
	users, err := s.dao.UserInfos(context.TODO(), uidList)
	if err != nil {
		log.Error("[ListPrizes] Fetch users error, uidList:%v, error(%v)", uidList, err)
		return
	}

	rspList = make([]*voguemdl.WinningItem, len(giftWin))

	for i, item := range giftWin {

		var (
			addrTmp      = &lotmdl.Address{}
			uname        string
			goodsAddress string
			errorMsg     string
			goodsName    string
			gift         *lotmdl.GiftInfo
			goodsType    int
		)

		// 获取礼物信息
		if gift, err = s.lotDao.GiftDetailByID(c, item.GiftId); err != nil {
			log.Errorc(c, "s.lotDao.GiftDetailByID(%v) failed. error(%v)", item.GiftAddrID, err)
			goodsName = "商品名称获取失败"
			goodsType = gift.Type
		} else {
			goodsName = gift.Name
		}

		// type == 1 表示实物商品
		if item.GiftAddrID != 0 && gift.Type == 1 {
			// 获取地址
			// todo: 需要批量获取
			if addrTmp, err = s.lotDao.GetAddressByID(c, item.GiftAddrID, item.Mid); err != nil {
				log.Errorc(c, "s.lotDao.GetAddressByID() failed. error(%v)", err)
				goodsAddress = "地址获取失败"
			} else {
				goodsAddress = addrTmp.Prov + addrTmp.City + addrTmp.Area + addrTmp.Addr + addrTmp.Name + addrTmp.Phone
			}
		} else {
			goodsAddress = "非实物商品"
		}

		// 获取风控数据
		if _, errorMsg, err = s.dao.RiskInfo(c, int64(item.Mid)); err != nil {
			log.Errorc(c, "s.lotDao.GetAddressByID() failed. error(%v)", err)
			return
		}

		user, ok := users[int64(item.Mid)]
		if ok {
			uname = user.GetName()
		}

		rspList[i] = &voguemdl.WinningItem{
			Uid:          item.Mid,
			UName:        uname,
			GoodsName:    goodsName,
			GoodsAddress: goodsAddress,
			GoodsType:    goodsType,
			WinningTime:  item.CTime,
		}
		rspList[i].HasError = errorMsg
	}

	rsp = &voguemdl.WinningList{}

	rsp.List = rspList
	rsp.Page = voguemdl.Page{
		Size:  search.Ps,
		Num:   search.Pn,
		Total: count,
	}

	return
}

// 导出中奖纪录
func (s *Service) ExportWinningList(c context.Context, search *voguemdl.CreditSearch) (result [][]string, err error) {
	var (
		rsp *voguemdl.WinningList
	)

	search.Pn = 0
	search.Ps = 0
	if rsp, err = s.WinningList(c, search); err != nil {
		log.Error("[WinningList] s.WinningList error(%v)", err)
		return
	}

	for _, item := range rsp.List {
		result = append(result, []string{
			item.UName,
			strconv.FormatInt(int64(item.Uid), 10),
			item.GoodsName,
			item.WinningTime.Time().Format(voguemdl.TimeFormat),
			item.GoodsAddress,
			item.HasError,
		})
	}
	return
}
