package grpc

import (
	"context"
	"go-common/library/conf/paladin.v2"

	"go-common/library/net/rpc/warden"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/conf"
	"go-gateway/app/app-svr/hkt-note/service/server/http"
	"go-gateway/app/app-svr/hkt-note/service/service/article"
	"go-gateway/app/app-svr/hkt-note/service/service/image"
	"go-gateway/app/app-svr/hkt-note/service/service/note"
)

type server struct {
	noteSrv *note.Service
	imgSrv  *image.Service
	artSrv  *article.Service
	c       *conf.Config
}

// New grpc server
func New(cfg *warden.ServerConfig, srv *http.Server) (wsvr *warden.Server, err error) {
	conf := &conf.Config{}
	if err := paladin.Get("hkt-note-service.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	wsvr = warden.NewServer(cfg)
	api.RegisterHktNoteServer(wsvr.Server(), &server{noteSrv: srv.NoteSvr, imgSrv: srv.ImgSvr, artSrv: srv.ArtSvr, c: conf})
	wsvr, err = wsvr.Start()
	return
}

func (s *server) NoteInfo(ctx context.Context, req *api.NoteInfoReq) (resp *api.NoteInfoReply, err error) {
	return s.noteSrv.NoteInfo(ctx, req)
}

func (s *server) NoteList(ctx context.Context, req *api.NoteListReq) (resp *api.NoteListReply, err error) {
	switch req.Type {
	case api.NoteListType_USER_ALL:
		if req.Mid == 0 {
			return nil, xecode.NoteListTypeInvalid
		}
		return s.noteSrv.NoteList(ctx, req)
	case api.NoteListType_USER_PUBLISHED:
		if req.Mid == 0 {
			return nil, xecode.NoteListTypeInvalid
		}
		return s.artSrv.PublishListInUser(ctx, req)
	case api.NoteListType_ARCHIVE_PUBLISHED:
		if req.Oid == 0 {
			return nil, xecode.NoteListTypeInvalid
		}
		return s.artSrv.PublishListInArc(ctx, req)
	default:
		return nil, xecode.NoteListTypeInvalid
	}
}

func (s *server) ImgAdd(ctx context.Context, req *api.ImgAddReq) (resp *api.ImgAddReply, err error) {
	return s.imgSrv.ImgAdd(ctx, req)
}

func (s *server) Img(ctx context.Context, req *api.ImgReq) (resp *api.ImgReply, err error) {
	return s.imgSrv.Img(ctx, req)
}

func (s *server) NoteSize(ctx context.Context, req *api.NoteSizeReq) (resp *api.NoteSizeReply, err error) {
	return s.noteSrv.NoteSize(ctx, req)
}

func (s *server) NoteCount(ctx context.Context, req *api.NoteCountReq) (resp *api.NoteCountReply, err error) {
	return s.noteSrv.NoteCount(ctx, req)
}

func (s *server) NoteListInArc(ctx context.Context, req *api.NoteListInArcReq) (resp *api.NoteListInArcReply, err error) {
	return s.noteSrv.NoteListInArc(ctx, req)
}

func (s *server) SimpleNotes(ctx context.Context, req *api.SimpleNotesReq) (resp *api.SimpleNotesReply, err error) {
	return s.noteSrv.SimpleNotes(ctx, req)
}

func (s *server) PublishImgs(ctx context.Context, req *api.PublishImgsReq) (resp *api.PublishImgsReply, err error) {
	return s.imgSrv.PublishImgs(ctx, req)
}

func (s *server) PublishNoteInfo(ctx context.Context, req *api.PublishNoteInfoReq) (resp *api.PublishNoteInfoReply, err error) {
	return s.artSrv.PublishNoteInfo(ctx, req)
}

func (s *server) SimpleArticles(ctx context.Context, req *api.SimpleArticlesReq) (resp *api.SimpleArticlesReply, err error) {
	return s.artSrv.SimpleArticles(ctx, req)
}

func (s *server) ArcsForbid(ctx context.Context, req *api.ArcsForbidReq) (resp *api.ArcsForbidReply, err error) {
	return s.noteSrv.ArcsForbid(ctx, req)
}

func (s *server) UpArc(ctx context.Context, req *api.UpArcReq) (resp *api.UpArcReply, err error) {
	return s.artSrv.UpArc(ctx, req)
}

func (s *server) ArcTag(ctx context.Context, req *api.ArcTagReq) (resp *api.ArcTagReply, err error) {
	return s.artSrv.ArcTag(ctx, req, s.noteSrv)
}

func (s *server) AutoPullCvid(ctx context.Context, req *api.AutoPullAidCivdReq) (resp *api.AutoPullAidCivdReply, err error) {
	return s.artSrv.AutoPullCvid(ctx, req)
}

func (s *server) ArcNotesCount(ctx context.Context, req *api.ArcNotesCountReq) (resp *api.ArcNotesCountReply, err error) {
	return s.artSrv.ArcNotesCount(ctx, req)
}

func (s *server) BatchGetReplyRenderInfo(ctx context.Context, req *api.BatchGetReplyRenderInfoReq) (resp *api.BatchGetReplyRenderInfoRes, err error) {
	return s.artSrv.BatchGetReplyRenderInfo(ctx, req)
}

func (s *server) GetAttachedRpid(ctx context.Context, req *api.GetAttachedRpidReq) (*api.GetAttachedRpidReply, error) {
	return s.artSrv.GetAttachedRpid(ctx, req)
}
