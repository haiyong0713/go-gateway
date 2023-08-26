package space

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/space"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

func (s *Service) SearchTab(ctx context.Context, mid int64, req *api.SearchTabReq, isIpad bool) (*api.SearchTabReply, error) {
	var archiveTotal, dynamicTotal int64
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		kwFields := []upgrpc.KwField{upgrpc.KwField_title}
		reply, err := s.srchDao.ArcPassedSearch(ctx, req.Mid, req.Keyword, false, kwFields, upgrpc.SearchOrder_pubtime, "desc", int64(1), int64(1), isIpad)
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		archiveTotal = reply.Total
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var err error
		if _, _, dynamicTotal, err = s.srchDao.DynamicSearch(ctx, mid, req.Mid, req.Keyword, int64(1), int64(1)); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if archiveTotal == 0 && dynamicTotal == 0 {
		return &api.SearchTabReply{}, nil
	}
	tabs := []*api.Tab{
		{Title: fmt.Sprintf("视频 %d", archiveTotal), Uri: "bilibili://space/search/av"},
		{Title: fmt.Sprintf("动态 %d", dynamicTotal), Uri: "bilibili://space/search/dynamic"},
	}
	focus := func() int64 {
		switch req.From {
		case api.From_ArchiveTab:
			if archiveTotal > 0 {
				return 0
			}
		case api.From_DynamicTab:
			if dynamicTotal > 0 {
				return 1
			}
		}
		if archiveTotal > 0 {
			return 0
		}
		if dynamicTotal > 0 {
			return 1
		}
		return 0
	}()
	return &api.SearchTabReply{Focus: focus, Tabs: tabs}, nil
}

func (s *Service) SearchArchive(ctx context.Context, req *api.SearchArchiveReq, isIpad bool) (*api.SearchArchiveReply, error) {
	kwFields := []upgrpc.KwField{upgrpc.KwField_title}
	reply, err := s.searchDao.ArcPassedSearch(ctx, req.Mid, req.Keyword, true, kwFields, upgrpc.SearchOrder_pubtime, "desc", req.Pn, req.Ps, isIpad)
	if err != nil {
		return nil, err
	}
	var playAvs []*arcgrpc.PlayAv
	for _, v := range reply.Archives {
		if v == nil {
			continue
		}
		playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: v.Aid})
	}
	arcm, err := s.arcDao.ArcsPlayer(ctx, playAvs, false)
	if err != nil {
		return nil, err
	}
	var archives []*api.Arc
	for _, v := range reply.Archives {
		if v == nil {
			continue
		}
		ap, ok := arcm[v.Aid]
		if !ok {
			continue
		}
		ap.GetArc().Title = v.Title
		ap.GetArc().Desc = v.Desc
		arc := &api.Arc{Archive: ap.GetArc()}
		playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
		arc.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(v.Aid, 10), model.AvPlayHandlerGRPC(ap.Arc, playInfo))
		func() {
			u, err := url.Parse(arc.Uri)
			if err != nil {
				log.Error("s.SearchArchive err:%v", err)
				return
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("s.SearchArchive err:%v", err)
				return
			}
			params.Set("from_spmid", "main.space-search.0.0")
			paramStr := params.Encode()
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			arc.Uri = u.String()
		}()
		// pgc稿件url直接跳转
		if model.AttrVal(v.Attribute, model.AttrBitIsPGC) == model.AttrYes && v.RedirectURL != "" {
			arc.Uri = v.RedirectURL
		}
		// 敏感字段不外露
		arc.GetArchive().Access = 0
		arc.GetArchive().Attribute = 0
		arc.GetArchive().AttributeV2 = 0
		archives = append(archives, arc)
	}
	return &api.SearchArchiveReply{
		Archives: archives,
		Total:    reply.Total,
	}, nil
}

func (s *Service) SearchDynamic(ctx context.Context, mid int64, req *api.SearchDynamicReq, dev device.Device, ip string, net network.Network) (*api.SearchDynamicReply, error) {
	dynamicIDs, searchWords, total, err := s.srchDao.DynamicSearch(ctx, mid, req.Mid, req.Keyword, req.Pn, req.Ps)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if len(dynamicIDs) == 0 {
		return &api.SearchDynamicReply{
			Total: total,
		}, nil
	}
	reply, err := s.srchDao.DynamicDetail(ctx, mid, dynamicIDs, searchWords, req.PlayerArgs, dev, ip, net)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var dynamics []*api.Dynamic
	for _, id := range dynamicIDs {
		val, ok := reply[id]
		if !ok {
			continue
		}
		dynamics = append(dynamics, &api.Dynamic{Dynamic: val})
	}
	return &api.SearchDynamicReply{
		Dynamics: dynamics,
		Total:    total,
	}, nil
}
