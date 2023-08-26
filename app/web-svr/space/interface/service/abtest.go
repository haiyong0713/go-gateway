package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"

	accwar "git.bilibili.co/bapis/bapis-go/account/service"
)

const (
	//已登录
	_login = 1
	//客人态
	_guestState = 1
	//未关注
	_spaceNoFollow = 1
	//_testNewPublishA default Order
	_testNewPublishA = "1"
	_testMostView    = "3"
)

// AbtestVideoSearch 已登录 客人态 未关注
// nolint:gomnd
func (s *Service) AbtestVideoSearch(c context.Context, mid, vmid int64) (res *model.AbtestVideoSearch) {
	var (
		relate                              *accwar.RelationReply
		order                               string
		spaceLogin, spaceState, spaceFollow int64
		err                                 error
	)
	order = _testNewPublishA
	res = &model.AbtestVideoSearch{
		VideoOrder: order,
	}
	if mid != 0 {
		//已登录
		spaceLogin = _login
		if mid != vmid {
			spaceState = _guestState
		}
		if relate, err = s.accClient.Relation3(c, &accwar.RelationReq{Mid: mid, Owner: vmid}); err != nil {
			log.Error("AbtestVideoSearch mid(%d) vmid(%d) error (%v)", mid, vmid, err)
		}
		if err != nil || relate == nil || !relate.Following {
			spaceFollow = _spaceNoFollow
		}
		//已登录 客人态 未关注
		if spaceLogin == _login && spaceState == _guestState && spaceFollow == _spaceNoFollow {
			remainder := mid % 100
			if remainder < 25 {
				order = _testNewPublishA
			} else if remainder < 50 {
				order = _testMostView
			} else if remainder < 75 {
				order = _testMostView
			} else {
				order = _testNewPublishA
			}
		} else {
			order = _testNewPublishA
		}
	} else {
		order = _testNewPublishA
	}
	log.Info("AbtestVideoSearch mid(%d) vmid(%d) _login(%d) _guestState(%d) _spaceNoFollow(%d) order(%s)", mid, vmid, _login, _guestState, _spaceNoFollow, order)
	res.VideoOrder = order
	return
}
