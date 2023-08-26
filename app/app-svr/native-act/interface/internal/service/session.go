package service

import (
	"context"

	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	ptypes "github.com/gogo/protobuf/types"
	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

func NewSessionOfIndex(c context.Context, req *api.IndexReq) *kernel.Session {
	ss := kernel.NewSession(c)
	ss.ReqFrom = model.ReqFromIndex
	ss.ShareReq = &kernel.ShareReq{ShareOrigin: req.ShareOrigin, TabID: req.TabId, TabModuleID: req.TabModuleId}
	ss.FromSpmid = req.FromSpmid
	ss.IsColdStart = req.IsColdStart
	ss.LocalTime = req.LocalTime
	ss.HttpsUrlReq = req.HttpsUrlReq
	ss.TabFrom = req.TabFrom
	ss.CurrentTab = req.CurrentTab
	return ss
}

func NewSessionOfDynamic(c context.Context, req *api.DynamicReq) *kernel.Session {
	ss := kernel.NewSession(c)
	ss.ReqFrom = model.ReqFromSubPage
	ss.FromSpmid = req.FromSpmid
	ss.IsColdStart = req.IsColdStart
	ss.LocalTime = req.LocalTime
	func() {
		if req.Params.FeedOffset == nil {
			return
		}
		offset := &dyntopicgrpc.FeedOffset{}
		if err := ptypes.UnmarshalAny(req.Params.FeedOffset, offset); err != nil {
			log.Error("Fail to UnmarshalAny FeedOffset, offset=%+v error=%+v", req.Params.Offset, err)
			return
		}
		ss.FeedOffset = offset
	}()
	ss.Offset = req.Params.Offset
	ss.LastGroup = req.Params.LastGroup
	ss.SortType = req.Params.SortType
	return ss
}

func NewSessionOfEditor(c context.Context, req *api.EditorReq) *kernel.Session {
	ss := kernel.NewSession(c)
	ss.ReqFrom = model.ReqFromSubPage
	ss.Offset = req.Params.Offset
	return ss
}

func NewSessionOfResource(c context.Context, req *api.ResourceReq) *kernel.Session {
	ss := kernel.NewSession(c)
	ss.ReqFrom = model.ReqFromSubPage
	ss.Offset = req.Params.Offset
	ss.OffsetStr = req.Params.TopicOffset
	ss.SortType = req.Params.SortType
	return ss
}

func NewSessionOfVideo(c context.Context, req *api.VideoReq) *kernel.Session {
	ss := kernel.NewSession(c)
	ss.ReqFrom = model.ReqFromSubPage
	ss.Offset = req.Params.Offset
	ss.OffsetStr = req.Params.TopicOffset
	ss.SortType = req.Params.SortType
	return ss
}

func NewSessionOfTimelineSupernatant(c context.Context, req *api.TimelineSupernatantReq) *kernel.Session {
	ss := kernel.NewSession(c)
	ss.ReqFrom = model.ReqFromSubPage
	ss.Index = req.Params.LastIndex
	ss.Offset = req.Params.Offset
	return ss
}

func NewSessionOfOgvSupernatant(c context.Context, req *api.OgvSupernatantReq) *kernel.Session {
	ss := kernel.NewSession(c)
	ss.ReqFrom = model.ReqFromSubPage
	ss.Index = req.Params.LastIndex
	ss.Offset = req.Params.Offset
	return ss
}
