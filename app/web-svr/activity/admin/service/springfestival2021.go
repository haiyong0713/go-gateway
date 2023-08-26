package service

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	actapi "go-gateway/app/web-svr/activity/interface/api"
)

// SpringFestival2021MidCardReply ...
type SpringFestival2021MidCardReply struct {
	CardID1 int64 `json:"CardID1"`
	CardID2 int64 `json:"CardID2"`
	CardID3 int64 `json:"CardID3"`
	CardID4 int64 `json:"CardID4"`
	CardID5 int64 `json:"CardID5"`
	Compose int64 `json:"Compose"`
}

// CardsMidCardReply ...
type CardsMidCardReply struct {
	CardID1 int64 `json:"CardID1"`
	CardID2 int64 `json:"CardID2"`
	CardID3 int64 `json:"CardID3"`
	CardID4 int64 `json:"CardID4"`
	CardID5 int64 `json:"CardID5"`
	CardID6 int64 `json:"CardID6"`
	CardID7 int64 `json:"CardID7"`
	CardID8 int64 `json:"CardID8"`
	CardID9 int64 `json:"CardID9"`
	Compose int64 `json:"Compose"`
}

// SpringFestival2021Mid ...
func (s *Service) SpringFestival2021Mid(c context.Context, mid int64) (reply *SpringFestival2021MidCardReply, err error) {
	res := &actapi.SpringFestival2021MidCardReply{}
	reply = &SpringFestival2021MidCardReply{}
	req := &actapi.SpringFestival2021MidCardReq{
		Mid: mid,
	}
	if res, err = s.actClient.SpringFestival2021MidCard(c, req); err != nil {
		log.Errorc(c, "SpringFestival2021MidCard mid(%d )Err(%v)", mid, err)
		err = ecode.Error(ecode.RequestErr, "请求错误")
	}
	if res != nil {
		reply = &SpringFestival2021MidCardReply{
			CardID1: res.CardID1,
			CardID2: res.CardID2,
			CardID3: res.CardID3,
			CardID4: res.CardID4,
			CardID5: res.CardID5,
			Compose: res.Compose,
		}
	}
	return
}

// CardsMid ...
func (s *Service) CardsMid(c context.Context, mid int64) (reply *CardsMidCardReply, err error) {
	res := &actapi.CardsMidCardReply{}
	reply = &CardsMidCardReply{}
	req := &actapi.CardsMidCardReq{
		Mid:      mid,
		Activity: s.c.Cards.Activity,
	}
	if res, err = s.actClient.Cards2021MidCard(c, req); err != nil {
		log.Errorc(c, "SpringFestival2021MidCard mid(%d )Err(%v)", mid, err)
		err = ecode.Error(ecode.RequestErr, "请求错误")
	}
	if res != nil {
		reply = &CardsMidCardReply{
			CardID1: res.CardID1,
			CardID2: res.CardID2,
			CardID3: res.CardID3,
			CardID4: res.CardID4,
			CardID5: res.CardID5,
			CardID6: res.CardID6,
			CardID7: res.CardID7,
			CardID8: res.CardID8,
			CardID9: res.CardID9,
			Compose: res.Compose,
		}
	}
	return
}
