package information

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// RecommendCardList
func (s *Service) RecommendCardList(c context.Context, req *show.RecommendCardListReq) (rsp *show.RecommendCardList, err error) {
	var (
		list  []*show.RecommendCard
		count int
	)
	rsp = &show.RecommendCardList{
		Page: common.Page{
			Num:  req.Pn,
			Size: req.Ps,
		},
	}
	if list, count, err = s.showDao.RecommendCardList(c, req); err != nil {
		log.Error("[RecommendCardList] s.showDao.RecommendCardList req(%v) error(%v)", req, err)
		return
	}
	for _, item := range list {
		item.Status = item.StatusVal()
	}

	rsp.Page.Total = count
	rsp.List = list
	return
}

func (s *Service) IntervalCheckRecommendCard(c context.Context, params *show.RecommendCardIntervalCheckReq) (overlap bool, err error) {
	// 是否存在时间段overlap
	if overlap, err = s.showDao.IntervalCheckRecommendCard(c, params); err != nil {
		log.Error("s.showDao.DupCheckRecommendCard error(%v), params(%v)", err, params)
		return
	}
	return
}

// AddRecommendCard
func (s *Service) AddRecommendCard(c context.Context, params *show.RecommendCardAddReq) (err error) {
	if err = s.showDao.AddRecommendCard(c, params); err != nil {
		log.Error("s.showDao.AddRecommendCard error(%v), params(%v)", err, params)
		return
	}
	if err = util.AddLogs(common.LogInformationRecommendCard, params.Uname, params.Uid, params.AvID, common.ActionAdd, params); err != nil {
		log.Error("infoSvc.AddRecommendCard AddLog error(%v)", err)
		return
	}
	return
}

// ModifyRecommendCard
func (s *Service) ModifyRecommendCard(c context.Context, params *show.RecommendCardModifyReq) (err error) {
	var (
		card    *show.RecommendCard
		opUname = params.Uname
		opUid   = params.Uid
	)
	params.AuditStatus = show.AuditStatusToAudit
	params.OfflineStatus = show.OnlineStatus
	validStatusArr := []string{show.StatusToAudit, show.StatusAuditReject, show.StatusAuditPass, show.StatusOffline, show.StatusOnline}
	if card, err = s.cardStatusCheck(c, params.ID, validStatusArr); err != nil {
		return
	}
	params.Uname = card.Uname
	params.Uid = card.Uid
	if err = s.showDao.ModifyRecommendCard(c, params); err != nil {
		return
	}
	if err = util.AddLogs(common.LogInformationRecommendCard, opUname, opUid, params.ID, common.ActionUpdate, params); err != nil {
		log.Error("infoSvc.ModifyRecommendCard AddLog error(%v)", err)
		return
	}
	return
}

// DeleteRecommendCard
func (s *Service) DeleteRecommendCard(c context.Context, params *show.RecommendCardOpReq) (err error) {
	var (
		card    *show.RecommendCard
		opUname = params.Uname
		opUid   = params.Uid
	)

	validStatusArr := []string{show.StatusToAudit, show.StatusAuditReject, show.StatusOffline}
	if card, err = s.cardStatusCheck(c, params.ID, validStatusArr); err != nil {
		return
	}
	if err = s.showDao.DeleteRecommendCard(c, params); err != nil {
		return
	}
	preCardInfo(params, card)
	if err = util.AddLogs(common.LogInformationRecommendCard, opUname, opUid, params.ID, common.ActionDelete, params); err != nil {
		log.Error("infoSvc.DeleteRecommendCard AddLog error(%v)", err)
		return
	}
	return
}

// 补充之前的卡片信息
func preCardInfo(params *show.RecommendCardOpReq, card *show.RecommendCard) {
	params.CardType = card.CardType
	params.CardID = card.CardID
	params.CardPos = card.CardPos
	params.PosIndex = card.PosIndex
	params.Etime = card.Stime
	params.Stime = card.Stime
	params.Uname = card.Uname
	params.Uid = card.Uid
}

// OfflineRecommendCard
func (s *Service) OfflineRecommendCard(c context.Context, params *show.RecommendCardOpReq) (err error) {
	var (
		card    *show.RecommendCard
		opUname = params.Uname
		opUid   = params.Uid
	)
	validStatusArr := []string{show.StatusAuditPass, show.StatusOnline}
	if card, err = s.cardStatusCheck(c, params.ID, validStatusArr); err != nil {
		return
	}
	if err = s.showDao.OfflineRecommendCard(c, params); err != nil {
		return
	}
	preCardInfo(params, card)
	if err = util.AddLogs(common.LogInformationRecommendCard, opUname, opUid, params.ID, common.ActionOffline, params); err != nil {
		log.Error("infoSvc.OfflineRecommendCard AddLog error(%v)", err)
		return
	}
	return
}

// PassRecommendCard
func (s *Service) PassRecommendCard(c context.Context, params *show.RecommendCardOpReq) (err error) {
	var (
		card    *show.RecommendCard
		opUname = params.Uname
		opUid   = params.Uid
		overlap bool
	)

	validStatusArr := []string{show.StatusToAudit}
	if card, err = s.cardStatusCheck(c, params.ID, validStatusArr); err != nil {
		return
	}
	// 生效时间段overlap验证
	checkParams := &show.RecommendCardIntervalCheckReq{
		ID:       card.ID,
		CardPos:  card.CardPos,
		PosIndex: card.PosIndex,
		Stime:    card.Stime,
		Etime:    card.Etime,
	}
	if overlap, err = s.IntervalCheckRecommendCard(c, checkParams); err != nil {
		return
	}
	if overlap {
		err = ecode.Error(ecode.RequestErr, "该位置已有运营卡片")
		return
	}
	if err = s.showDao.PassRecommendCard(c, params); err != nil {
		return
	}
	preCardInfo(params, card)
	if err = util.AddLogs(common.LogInformationRecommendCard, opUname, opUid, params.ID, common.ActionOpt, params); err != nil {
		log.Error("infoSvc.PassRecommendCard AddLog error(%v)", err)
		return
	}
	return
}

// RejectRecommendCard
func (s *Service) RejectRecommendCard(c context.Context, params *show.RecommendCardOpReq) (err error) {
	var (
		card    *show.RecommendCard
		opUname = params.Uname
		opUid   = params.Uid
	)
	validStatusArr := []string{show.StatusToAudit}
	if card, err = s.cardStatusCheck(c, params.ID, validStatusArr); err != nil {
		return
	}
	if err = s.showDao.RejectRecommendCard(c, params); err != nil {
		return
	}
	preCardInfo(params, card)
	if err = util.AddLogs(common.LogInformationRecommendCard, opUname, opUid, params.ID, common.ActionOpt, params); err != nil {
		log.Error("infoSvc.RejectRecommendCard AddLog error(%v)", err)
		return
	}
	return
}

func validStatus(status string, validStatusArr []string) (valid bool) {
	valid = false
	for _, item := range validStatusArr {
		if item == status {
			valid = true
			return
		}
	}
	return
}

func (s *Service) cardStatusCheck(c context.Context, id int64, validStatusArr []string) (card *show.RecommendCard, err error) {
	if card, err = s.showDao.RecommendCardByID(c, id); err != nil {
		log.Error("s.showDao.RecommendCardByID error(%v), id(%d)", err, id)
		return
	}
	if card.ID <= 0 {
		err = ecode.Error(ecode.RequestErr, "卡片不存在")
		return
	}
	if !validStatus(card.StatusVal(), validStatusArr) {
		err = ecode.Error(ecode.RequestErr, "当前状态不允许进行该操作")
		return
	}
	return
}
