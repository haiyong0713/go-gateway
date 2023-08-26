package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"

	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
	"go-gateway/app/app-svr/resource/service/model"
)

const specialIdMaxNum = 20

// WebSpecialCard all web special card.
func (s *Service) GetWebSpecialCard(c context.Context, req *pb2.NoArgRequest) (res *pb2.WebSpecialCardResp, err error) {
	res = &pb2.WebSpecialCardResp{
		Card: s.webSpecialCard,
	}
	return
}

func (s *Service) GetAppSpecialCard(c context.Context, req *pb2.NoArgRequest) (res *pb2.AppSpecialCardResp, err error) {
	res = &pb2.AppSpecialCardResp{
		Card: s.appSpecialCard,
	}
	return
}

// 获取特殊卡信息
func (s *Service) GetSpecialCard(c context.Context, req *pb2.SpecialCardReq) (res *pb2.SpecialCardResp, err error) {
	res = &pb2.SpecialCardResp{}

	if len(req.Ids) > specialIdMaxNum {
		return nil, ecode.Errorf(ecode.RequestErr, "批量请求id个数不能超过%d个", specialIdMaxNum)
	}

	specialCardMap := make(map[int64]*pb2.AppSpecialCard, len(req.Ids))
	for _, id := range req.Ids {
		specialCard, _ := s.GetSpecialCardById(c, id)
		if specialCard != nil {
			specialCardMap[id] = specialCard
		}
	}

	if len(specialCardMap) == 0 {
		return nil, ecode.Error(ecode.NothingFound, "特殊卡不存在！")
	}

	res.SpecialCard = specialCardMap
	return
}

func (s *Service) GetAppRcmdRelatePgc(c context.Context, req *pb2.AppRcmdRelatePgcRequest) (res *pb2.AppRcmdRelatePgcResp, err error) {
	if req == nil || req.Id == 0 || req.MobiApp == "" || req.Build == 0 {
		err = ecode.RequestErr
		return
	}
	var (
		appRcmdId int64
		ok        bool
		appRcmd   *model.AppRcmd
		platVer   []*model.PlatVer
	)
	//判断seasonID 是否有配置的相关推荐卡片
	if appRcmdId, ok = s.appRcmdRelatePgcMapCache[req.Id]; !ok {
		err = ecode.NothingFound
		log.Info("service.GetAppRcmdRelatePgc appRcmdRelatePgcMapCache pgcSeasonID(%v) NothingFound", req.Id)
		return
	}

	//取出seasonID 对应的 相关推荐的卡片数据
	if appRcmd, ok = s.appRcmdRelatePgcCache[appRcmdId]; !ok {
		err = ecode.NothingFound
		log.Info("service.GetAppRcmdRelatePgc appRcmdRelatePgcCache pgcSeasonID(%v) appRcmdId(%v) NothingFound", req.Id, appRcmdId)
		return
	}
	p := model.Plat(req.MobiApp, req.Device)
	//判断APP是否存在
	if platVer, ok = appRcmd.PlatVer[p]; !ok {
		err = ecode.NothingFound
		return
	}
	//判断APP版本是否存在
	if len(platVer) == 0 {
		err = ecode.NothingFound
		return
	}

	//判断版本信息是否匹配
	for _, v := range platVer {
		if model.InvalidBuild(int(req.Build), v.Build, v.Cond) {
			err = ecode.NothingFound
			return
		}
	}
	var specialTmp *pb2.AppSpecialCard

	if specialTmp, ok = s.appSpecialCardMap[appRcmd.CardValue]; ok && specialTmp != nil {
		res = &pb2.AppRcmdRelatePgcResp{
			Id:        specialTmp.Id,
			Title:     specialTmp.Title,
			Desc:      specialTmp.Desc,
			Cover:     specialTmp.Cover,
			Scover:    specialTmp.Scover,
			ReType:    specialTmp.ReType,
			ReValue:   specialTmp.ReValue,
			Corner:    specialTmp.Corner,
			Card:      specialTmp.Card,
			Size_:     specialTmp.Size_,
			Position:  appRcmd.Position,
			RecReason: appRcmd.RecReason,
		}
	} else {
		res = &pb2.AppRcmdRelatePgcResp{}
	}
	return
}
