package card

import (
	"context"

	model "go-gateway/app/app-svr/app-feed/admin/model/card"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/util"

	"go-common/library/log"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) AddNavigationCard(c context.Context, req *model.AddNavigationCardReq) (resp *model.AddNavigationCardResp, err error) {
	var (
		card *model.ResourceCard
	)
	resp = &model.AddNavigationCardResp{}

	if card, err = model.ConvertNavigationCard(req.Title, req.Desc, req.Cover, req.Corner, req.Button, req.Navigation); err != nil {
		log.Error("ConvertNavigationCard to resource card req(%+v) err(%v)", req, err)
		return
	}

	card.CUname = req.Username
	card.MUname = req.Username
	if resp.CardId, err = s.dao.ResourceCardAdd(c, card); err != nil {
		log.Error("dao.ResourceCardAdd card(%+v) error(%v)", card, err)
		return
	}

	if err1 := util.AddResourceCardLogs(req.Username, req.Uid, resp.CardId, model.CardNavigation, common.ActionAdd,
		nil, card, nil); err1 != nil {
		log.Error("AddResourceCardLogs create error(%v)", err1)
	}
	return
}

func (s *Service) UpdateNavigationCard(c context.Context, req *model.UpdateNavigationCardReq) (resp *empty.Empty, err error) {
	var (
		card    *model.ResourceCard
		oldCard *model.ResourceCard
	)

	if card, err = model.ConvertNavigationCard(req.Title, req.Desc, req.Cover, req.Corner, req.Button, req.Navigation); err != nil {
		log.Error("ConvertNavigationCard to resource card req(%+v) err(%v)", req, err)
		return
	}

	card.Id = req.CardId
	card.MUname = req.Username
	if oldCard, err = s.dao.ResourceCardUpdate(c, card); err != nil {
		log.Error("dao.ResourceCardUpdate card(%+v) error(%v)", card, err)
		return
	}

	if err1 := util.AddResourceCardLogs(req.Username, req.Uid, req.CardId, model.CardNavigation, common.ActionUpdate,
		oldCard, card, nil); err1 != nil {
		log.Error("AddResourceCardLogs update error(%v)", err1)
	}
	return
}

func (s *Service) DeleteNavigationCard(c context.Context, req *model.DeleteNavigationCardReq) (resp *empty.Empty, err error) {
	var (
		oldCard *model.ResourceCard
	)
	if oldCard, err = s.dao.ResourceCardDelete(c, req.Username, req.CardId, model.CardTypeNavigation); err != nil {
		log.Error("dao.ResourceCardDelete cardId(%+v) error(%v)", req.CardId, err)
		return
	}

	if err1 := util.AddResourceCardLogs(req.Username, req.Uid, req.CardId, model.CardNavigation, common.ActionDelete,
		oldCard, nil, nil); err1 != nil {
		log.Error("AddResourceCardLogs delete error(%v)", err1)
	}
	return
}

func (s *Service) QueryNavigationCard(c context.Context, req *model.QueryNavigationCardReq) (resp *model.QueryNavigationCardResp, err error) {
	var (
		rCard   *model.ResourceCard
		navCard *model.NavigationCard
	)
	if rCard, err = s.dao.ResourceCardQuery(c, req.CardId, model.CardTypeNavigation); err != nil {
		log.Error("dao.ResourceCardQuery req(%+v) error(%v)", req, err)
		return
	}
	if navCard, err = model.ParseNavigationCard(rCard); err != nil {
		log.Error("model.ParseNavigationCard rCard(%+v) error(%v)", rCard, err)
		return
	}
	resp = &model.QueryNavigationCardResp{
		CardId:     navCard.Id,
		Title:      navCard.Title,
		Desc:       navCard.Desc,
		Cover:      navCard.Cover,
		Corner:     navCard.Corner,
		Button:     navCard.Button,
		Navigation: navCard.Navigation,
		Ctime:      navCard.Ctime,
		Mtime:      navCard.Mtime,
		CUname:     navCard.CUname,
		MUname:     navCard.MUname,
	}
	return
}

func (s *Service) ListNavigationCard(c context.Context, req *model.ListNavigationCardReq) (resp *model.ListNavigationCardResp, err error) {
	var (
		rawList []*model.ResourceCard
	)
	resp = &model.ListNavigationCardResp{Page: &model.Page{Pn: req.Pn, Ps: req.Ps}}
	if resp.Page.Total, rawList, err = s.dao.ResourceCardList(c, req.CardId, model.CardTypeNavigation, req.Keyword, req.Pn, req.Ps); err != nil {
		log.Error("dao.ResourceCardList req(%+v) error(%v)", req, err)
		return
	}

	resp.List = make([]*model.NavigationListItem, len(rawList))
	for idx, raw := range rawList {
		navCard, err1 := model.ParseNavigationCard(raw)
		if err1 != nil {
			return resp, err1
		}

		resp.List[idx] = &model.NavigationListItem{
			CardId:          navCard.Id,
			Title:           navCard.Title,
			Desc:            navCard.Desc,
			Cover:           navCard.Cover,
			Ctime:           navCard.Ctime,
			Mtime:           navCard.Mtime,
			CUname:          navCard.CUname,
			MUname:          navCard.MUname,
			NavigationCount: navCard.Count,
		}
	}
	return
}
