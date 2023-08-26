package like

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"strconv"
	"time"
)

func (s *Service) CacheData(ctx context.Context, param *like.CacheData) (res interface{}, err error) {
	switch param.Type {
	case 1:
		res, err = s.dao.CacheReserveOnly(ctx, param.Sid, param.Mid)
	default:
		err = fmt.Errorf("no type %v", param.Type)
	}
	return
}

func (s *Service) AddData(ctx context.Context, sid int64) (res int, err error) {
	type UpActReserve41 struct {
		Sid   int64 `json:"sid"`
		Total int64 `json:"total"`
		Time  int64 `json:"time"`
	}
	preData := UpActReserve41{
		Sid:   sid,
		Time:  time.Now().UnixNano() / 1000,
		Total: time.Now().Unix(),
	}
	data, _ := json.Marshal(preData)
	if err = component.UpActReserveProducer.Send(ctx, strconv.FormatInt(sid, 10), data); err != nil {
		log.Errorc(ctx, "test err %+v", preData)
		return
	}
	return
}
