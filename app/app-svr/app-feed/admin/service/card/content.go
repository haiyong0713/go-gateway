package card

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/app-feed/admin/bvav"
	model "go-gateway/app/app-svr/app-feed/admin/model/card"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/util"
	"go-gateway/app/app-svr/app-feed/ecode"

	"go-common/library/log"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) AddContentCard(c context.Context, req *model.AddContentCardReq) (resp *model.AddContentCardResp, err error) {
	var (
		card *model.ResourceCard
	)
	resp = &model.AddContentCardResp{}

	if req.Jump != nil && req.Jump.ReType == model.ContentReTypeAv {
		if req.Jump.ReValue, err = bvav.ToAvStr(req.Jump.ReValue); err != nil {
			log.Error("AddContentCard call bvav.ToAvStr id(%+v) err(%v)", req.Jump.ReValue, err)
			return
		}
	}

	for _, cont := range req.Content {
		switch cont.ReType {
		case model.ContentCTypeAv:
			if cont.ReValue, err = bvav.ToAvStr(cont.ReValue); err != nil {
				log.Error("AddContentCard call bvav.ToAvStr id(%+v) err(%v)", cont.ReValue, err)
				return
			}
		case model.ContentCTypeArticle:
			rId, err := strconv.ParseInt(cont.ReValue, 10, 64)
			if err != nil {
				log.Error("AddContentCard parseInt on id(%+v) err(%v)", cont.ReValue, err)
				return nil, ecode.InvalidResourceId
			}
			article, err := s.articleDao.ArticleRpc(c, rId)
			if err != nil || article == nil {
				log.Error("AddContentCard call ArticleRpc id(%+v) err(%v)", rId, err)
				return nil, ecode.InvalidResourceId
			}
			if (article.Attributes>>1)&1 > 0 {
				log.Error("AddContentCard invalid article id(%+v), err(禁止分发)", rId)
				return nil, ecode.InvalidResourceId
			}
		}
	}

	if card, err = model.ConvertContentCard(req.Title, req.Cover, req.Jump, req.Button, req.Content); err != nil {
		log.Error("ConvertContentCard to resource card req(%+v) err(%v)", req, err)
		return
	}

	card.CUname = req.Username
	card.MUname = req.Username
	if resp.CardId, err = s.dao.ResourceCardAdd(c, card); err != nil {
		log.Error("dao.ResourceCardAdd card(%+v) error(%v)", card, err)
		return
	}

	if err1 := util.AddResourceCardLogs(req.Username, req.Uid, resp.CardId, model.CardContent, common.ActionAdd,
		nil, card, nil); err1 != nil {
		log.Error("AddResourceCardLogs create error(%v)", err1)
	}
	return
}

func (s *Service) UpdateContentCard(c context.Context, req *model.UpdateContentCardReq) (resp *empty.Empty, err error) {
	var (
		card    *model.ResourceCard
		oldCard *model.ResourceCard
	)

	if req.Jump != nil && req.Jump.ReType == model.ContentReTypeAv {
		if req.Jump.ReValue, err = bvav.ToAvStr(req.Jump.ReValue); err != nil {
			log.Error("AddContentCard call bvav.ToAvStr id(%+v) err(%v)", req.Jump.ReValue, err)
			return
		}
	}

	for _, cont := range req.Content {
		rId, _ := strconv.ParseInt(cont.ReValue, 10, 64)
		switch cont.ReType {
		case model.ContentCTypeAv:
			if cont.ReValue, err = bvav.ToAvStr(cont.ReValue); err != nil {
				log.Error("AddContentCard call bvav.ToAvStr id(%+v) err(%v)", cont.ReValue, err)
				return
			}
		case model.ContentCTypeArticle:
			article, err := s.articleDao.ArticleRpc(c, rId)
			if err != nil || article == nil {
				log.Error("AddContentCard call ArticleRpc id(%+v) err(%v)", rId, err)
				return nil, ecode.InvalidResourceId
			}
			if (article.Attributes>>1)&1 > 0 {
				log.Error("AddContentCard invalid article id(%+v), err(禁止分发)", rId)
				return nil, ecode.InvalidResourceId
			}
		}
	}

	if card, err = model.ConvertContentCard(req.Title, req.Cover, req.Jump, req.Button, req.Content); err != nil {
		log.Error("ConvertContentCard to resource card req(%+v) err(%v)", req, err)
		return
	}

	card.Id = req.CardId
	card.MUname = req.Username
	if oldCard, err = s.dao.ResourceCardUpdate(c, card); err != nil {
		log.Error("dao.ResourceCardUpdate card(%+v) error(%v)", card, err)
		return
	}

	if err1 := util.AddResourceCardLogs(req.Username, req.Uid, req.CardId, model.CardContent, common.ActionUpdate,
		oldCard, card, nil); err1 != nil {
		log.Error("AddResourceCardLogs create error(%v)", err1)
	}
	return
}

func (s *Service) DeleteContentCard(c context.Context, req *model.DeleteContentCardReq) (resp *empty.Empty, err error) {
	var (
		oldCard *model.ResourceCard
	)
	if oldCard, err = s.dao.ResourceCardDelete(c, req.Username, req.CardId, model.CardTypeContent); err != nil {
		log.Error("dao.ResourceCardDelete cardId(%+v) error(%v)", req.CardId, err)
		return
	}

	if err1 := util.AddResourceCardLogs(req.Username, req.Uid, req.CardId, model.CardContent, common.ActionDelete,
		oldCard, nil, nil); err1 != nil {
		log.Error("AddResourceCardLogs delete error(%v)", err1)
	}
	return
}

func (s *Service) QueryContentCard(c context.Context, req *model.QueryContentCardReq) (resp *model.QueryContentCardResp, err error) {
	var (
		rCard    *model.ResourceCard
		contCard *model.ContentCard
	)
	if rCard, err = s.dao.ResourceCardQuery(c, req.CardId, model.CardTypeContent); err != nil {
		log.Error("dao.ResourceCardQuery req(%+v) error(%v)", req, err)
		return
	}
	if contCard, err = model.ParseContentCard(rCard); err != nil {
		log.Error("model.ParseContentCard rCard(%+v) error(%v)", rCard, err)
		return
	}
	resp = &model.QueryContentCardResp{
		CardId:  contCard.Id,
		Title:   contCard.Title,
		Cover:   contCard.Cover,
		Jump:    contCard.Jump,
		Button:  contCard.Button,
		Content: contCard.Content,
		Ctime:   contCard.Ctime,
		Mtime:   contCard.Mtime,
		CUname:  contCard.CUname,
		MUname:  contCard.MUname,
	}
	return
}

func (s *Service) ListContentCard(c context.Context, req *model.ListContentCardReq) (resp *model.ListContentCardResp, err error) {
	var (
		rawList []*model.ResourceCard
	)
	resp = &model.ListContentCardResp{Page: &model.Page{Pn: req.Pn, Ps: req.Ps}}
	if resp.Page.Total, rawList, err = s.dao.ResourceCardList(c, req.CardId, model.CardTypeContent, req.Keyword, req.Pn, req.Ps); err != nil {
		log.Error("dao.ResourceCardList req(%+v) error(%v)", req, err)
		return
	}

	list := make([]*model.ContListItem, 0, len(rawList))
	for _, raw := range rawList {
		contCard, err1 := model.ParseContentCard(raw)
		if err != nil {
			log.Error("failed to parse content card(%+v) error(%v)", raw, err1)
			continue
		}

		list = append(list, &model.ContListItem{
			CardId:  contCard.Id,
			Title:   contCard.Title,
			Cover:   contCard.Cover,
			Jump:    contCard.Jump,
			Button:  contCard.Button,
			Content: contCard.Content,
			Ctime:   contCard.Ctime,
			Mtime:   contCard.Mtime,
			CUname:  contCard.CUname,
			MUname:  contCard.MUname,
		})
	}

	resp.List = list
	return
}
