package web

import (
	"context"
	"sort"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/web-goblin/interface/model/web"
)

const (
	_typeCount       = 5
	_hint            = "hint"
	_self            = "self"
	_contact         = "contact"
	_guess           = "guess"
	_strategy        = "strategy"
	_hintTitle       = "顶部黄条"
	_selfTitle       = "自助服务"
	_contactTitle    = "联系客服"
	_guessTitle      = "猜你想问"
	_hintTitleTitle  = "产品攻略"
	_tpHint          = 1
	_tpSelf          = 2
	_tpContact       = 3
	_tpGuess         = 4
	_tpStrategy      = 5
	_dateTimeFormart = "2006-01-02 15:04:00"
)

var (
	_emptyCustomer = make([]*web.Customer, 0)
	_emptyBusiness = make([]*web.BusinessList, 0)
)

// CusCenter get customer center.
// nolint: gocognit,gomnd
func (s *Service) CusCenter(c context.Context) (rs map[string]*web.CustomerCenter, err error) {
	if rs, err = s.dao.CustomerCache(c); err != nil {
		err = nil
	} else if len(rs) > 0 {
		return
	}
	tmpRs, err := s.dao.CusCenter(c)
	if err != nil {
		log.Error("Customer error(%+v)", err)
		return nil, err
	}
	rs = make(map[string]*web.CustomerCenter, _typeCount)
	nowTime := time.Now().Format(_dateTimeFormart)
	now, err := time.ParseInLocation(_dateTimeFormart, nowTime, time.Now().Location())
	if err != nil {
		log.Error("Customer error(%+v)", err)
		return nil, err
	}
	for t, customers := range tmpRs {
		//因为会修改business_type，所以要重新排序
		sort.Slice(customers, func(i, j int) bool {
			if customers[i].CustomerRank != customers[j].CustomerRank {
				return customers[i].CustomerRank > customers[j].CustomerRank
			}
			return customers[i].ID < customers[j].ID
		})
		switch t {
		case _tpHint:
			for _, cus := range customers {
				if now.Unix() >= cus.Stime && now.Unix() <= cus.Etime {
					rs[_hint] = &web.CustomerCenter{Title: _hintTitle, List: []*web.Customer{cus}, BusinessList: _emptyBusiness}
				}
			}
		case _tpSelf:
			rs[_self] = &web.CustomerCenter{Title: _selfTitle, List: customers, BusinessList: _emptyBusiness}
		case _tpContact:
			var (
				busIDs  []int64
				busList []*web.BusinessList
			)
			mapBus := make(map[int64]*web.BusinessList, len(customers))
			for _, cus := range customers {
				if cus.BusinessCustomerType != _tpContact {
					continue
				}
				if _, ok := mapBus[cus.BusinessType]; !ok {
					busIDs = append(busIDs, cus.BusinessType)
					mapBus[cus.BusinessType] = &web.BusinessList{BusinessID: cus.BusinessType, BusinessName: cus.BusinessName, BusinessLogo: cus.Logo, BusinessRank: cus.BusinessRank, BusinessList: []*web.Customer{cus}}
					continue
				}
				mapBus[cus.BusinessType].BusinessList = append(mapBus[cus.BusinessType].BusinessList, cus)
			}
			for _, id := range busIDs {
				if list, ok := mapBus[id]; ok {
					busList = append(busList, list)
				}
			}
			sort.Slice(busList, func(i, j int) bool {
				if busList[i].BusinessRank != busList[j].BusinessRank {
					return busList[i].BusinessRank > busList[j].BusinessRank
				}
				return busList[i].BusinessID < busList[j].BusinessID
			})
			rs[_contact] = &web.CustomerCenter{Title: _contactTitle, List: _emptyCustomer, BusinessList: busList}
		case _tpGuess:
			var (
				list    []*web.Customer
				busList []*web.BusinessList
			)
			busm := map[int64]*web.BusinessList{}
			for _, cus := range customers {
				if cus.BusinessCustomerType != _tpGuess {
					continue
				}
				if cus.BusinessType == 8 {
					list = append(list, cus)
				}
				if _, ok := busm[cus.BusinessType]; !ok {
					busm[cus.BusinessType] = &web.BusinessList{BusinessID: cus.BusinessType, BusinessName: cus.BusinessName, BusinessLogo: cus.Logo, BusinessRank: cus.BusinessRank}
				}
				busm[cus.BusinessType].BusinessList = append(busm[cus.BusinessType].BusinessList, cus)
			}
			for _, list := range busm {
				busList = append(busList, list)
			}
			sort.Slice(busList, func(i, j int) bool {
				if busList[i].BusinessRank != busList[j].BusinessRank {
					return busList[i].BusinessRank > busList[j].BusinessRank
				}
				return busList[i].BusinessID < busList[j].BusinessID
			})
			rs[_guess] = &web.CustomerCenter{Title: _guessTitle, List: list, BusinessList: busList}
		case _tpStrategy:
			var strategys []*web.Customer
			for _, cus := range customers {
				if now.Unix() >= cus.Stime && now.Unix() <= cus.Etime {
					strategys = append(strategys, cus)
				}
			}
			if len(strategys) == 0 {
				rs[_strategy] = &web.CustomerCenter{Title: _hintTitleTitle, List: _emptyCustomer, BusinessList: _emptyBusiness}
				continue
			}
			rs[_strategy] = &web.CustomerCenter{Title: _hintTitleTitle, List: strategys, BusinessList: _emptyBusiness}
		}
	}
	if len(rs) > 0 {
		s.cache.Do(c, func(c context.Context) {
			if err := s.dao.SetCustomerCache(c, rs); err != nil {
				log.Error("%+v", err)
			}
		})
	}
	return
}
