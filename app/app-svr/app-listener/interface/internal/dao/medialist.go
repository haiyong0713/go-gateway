package dao

import (
	"context"

	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
)

type MediaListDetailOpt struct {
	Typ    int64
	BizId  int64
	Anchor *v1.PlayItem
}

func (d *dao) MediaListDetail(ctx context.Context, opt MediaListDetailOpt) ([]model.MediaListItem, error) {
	const maxWant = 300

	c := &MediaListReqContext{
		Ctx: ctx, MaxWant: maxWant, FetchAll: true, Anchor: opt.Anchor,
		FnDo: d.mediaListHTTP.Do, FnUri: d.mediaListHTTP.composeURI,
	}
	return c.DoList(MediaListDoListOpt{
		Typ:   int(opt.Typ),
		BizId: opt.BizId,
	})
}

type MediaListPagedOpt struct {
	Typ, BizId int64
	//Anchor     *v1.PlayItem
	Offset string
}

type MediaListPagedResp struct {
	Items   []model.MediaListItem
	Offset  string
	HasMore bool
	Total   int64
}

func (d *dao) MediaListPaged(ctx context.Context, opt MediaListPagedOpt) (resp *MediaListPagedResp, err error) {
	const pageSize = 20
	c := &MediaListReqContext{
		Ctx: ctx, PageSize: pageSize, FetchAll: false, Anchor: nil,
		FnDo: d.mediaListHTTP.Do, FnUri: d.mediaListHTTP.composeURI,
	}

	data, err := c.DoPage(MediaListDoPageOpt{
		Typ: int(opt.Typ), BizId: opt.BizId, Offset: opt.Offset,
	})
	if err != nil {
		return nil, err
	}
	resp = &MediaListPagedResp{
		Items:   data.Items,
		Offset:  data.Offset,
		HasMore: data.HasMore,
		Total:   int64(c.total),
	}
	return
}
