package like

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/dao/like"
)

func (s *Service) WatchData(ctx context.Context, typ int64) (res interface{}, err error) {
	switch typ {
	case 1:
		res = s.reserveVideoSourceTags
	case 2:
		res = s.HotActRelationInfoStore
	case 3:
		res = s.HotActSubjectInfoStore
	case 4:
		res = s.HotActSubjectReserveIDsInfoStore
	case 5:
		res = like.MemSubjectRule
	case 6:
		res = s.dao.DynamicArc
	case 7:
		res = s.dao.DynamicLive
	default:
		err = fmt.Errorf("no type %v", typ)
	}
	return
}
