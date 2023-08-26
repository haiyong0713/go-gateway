package unicom

import (
	"context"
	"strconv"

	"go-common/library/ecode"
	log "go-common/library/log"
	xecode "go-gateway/app/app-svr/app-wall/ecode"
	v1 "go-gateway/app/app-svr/app-wall/interface/api"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"
)

// UnicomBindInfosGRPC
func (s *Service) UnicomBindInfosGRPC(c context.Context, mids []int64) (res *v1.UsersReply, err error) {
	var (
		max       = 100
		userInfos = map[int64]*v1.UserInfo{}
		ubs       map[int64]*unicom.UserBind
		missMids  []int64
	)
	if len(mids) > max {
		err = xecode.AppQueryExceededLimit
		return
	}
	if ubs, err = s.dao.UsersBindCache(c, mids); err != nil {
		log.Error("UnicomBindInfosGRPC s.dao.UsersBindCache error(%v)", err)
		return
	}
	for _, mid := range mids {
		if mid <= 0 {
			err = ecode.RequestErr
			return
		}
		if ub, ok := ubs[mid]; ok {
			userInfos[mid] = &v1.UserInfo{Phone: strconv.Itoa(ub.Phone)}
		} else {
			missMids = append(missMids, mid)
		}
	}
	if len(missMids) > 0 {
		if ubs, err = s.dao.UserBindByMids(c, mids); err != nil {
			log.Error("UnicomBindInfosGRPC s.dao.UserBindByMids error(%v)", err)
			return
		}
		for _, mid := range mids {
			if ub, ok := ubs[mid]; ok {
				userInfos[mid] = &v1.UserInfo{Phone: strconv.Itoa(ub.Phone)}
			}
		}
	}
	res = &v1.UsersReply{
		UsersInfo: userInfos,
	}
	return
}
