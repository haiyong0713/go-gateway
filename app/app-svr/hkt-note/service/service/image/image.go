package image

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/note"
)

// Upload upload img/png/jpeg/bmp file into bfs.
func (s *Service) ImgAdd(c context.Context, req *api.ImgAddReq) (*api.ImgAddReply, error) {
	id, err := s.dao.AddImage(c, req.Mid, req.Location)
	if err != nil {
		return nil, err
	}
	if err := s.dao.AddCacheImg(c, req.Mid, id, &note.ImgInfo{ImageId: id, Location: req.Location}); err != nil {
		log.Warn("noteWarn ImgAdd err(%+v)", err)
	}
	return &api.ImgAddReply{ImageId: id}, nil
}

func (s *Service) Img(c context.Context, req *api.ImgReq) (*api.ImgReply, error) {
	res, err := s.dao.Image(c, req.Mid, req.ImageId)
	if err != nil {
		return nil, err
	}
	return &api.ImgReply{Location: res.Location}, nil
}

func (s *Service) PublishImgs(c context.Context, req *api.PublishImgsReq) (*api.PublishImgsReply, error) {
	items, err := s.dao.Images(c, req.ImageIds, req.Mid)
	if err != nil {
		log.Error("artError err(%+v)", err)
		return nil, err
	}
	return &api.PublishImgsReply{Items: items, Host: s.c.NoteCfg.BfsHost}, nil
}
