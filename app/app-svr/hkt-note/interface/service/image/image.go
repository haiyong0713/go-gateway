package image

import (
	"context"
	"fmt"
	"time"

	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"
)

func (s *Service) BfsNoteUpload(c context.Context, mid int64, fileType string, bs []byte) (*note.ImageRes, error) {
	if len(bs) == 0 {
		log.Error("ImageError Upload mid(%d) fileType(%s) err(%+v)", mid, fileType, xecode.ImageStreamEmpty)
		return nil, xecode.ImageStreamEmpty
	}
	if len(bs) >= s.c.Bfs.MaxSize {
		log.Error("ImageError Upload mid(%d) fileType(%s) err(%+v)", mid, fileType, xecode.ImageTooLarge)
		return nil, xecode.ImageTooLarge
	}
	res, err := s.dao.NoteImgUpload(c, mid, fileType, bs)
	if err != nil {
		log.Error("ImageError Upload err(%+v)", err)
		return nil, err
	}
	var imageId int64
	if imageId, err = s.dao.ImgAdd(c, res.Location, mid); err != nil {
		log.Error("ImageError Upload err(%+v)", err)
		return nil, err
	}
	s.infocImg(c, mid, imageId)
	return &note.ImageRes{Location: fmt.Sprintf("%s?image_id=%d", s.c.Bfs.Host, imageId)}, nil
}

func (s *Service) BfsImage(c context.Context, mid int64, imageId int64) (bs []byte, fileType string, err error) {
	location, err := s.dao.Img(c, mid, imageId)
	if err != nil {
		log.Error("BfsImage err(%+v)", err)
		return
	}
	return s.dao.NoteImgDownload(c, mid, location)

}

func (s *Service) infocImg(c context.Context, mid, imageId int64) {
	api := fmt.Sprintf("%s?image_id=%d&mid=%d&token=%s", s.c.Bfs.PublicUrl, imageId, mid, s.c.Bfs.PublicToken)
	payload := infocV2.NewLogStream("006280", mid, imageId, api, time.Now().Format("2006-01-02 15:04:05"))
	log.Info("infocImg %s", payload.Data)
	if err := s.infocV2Log.Info(c, payload); err != nil {
		log.Error("infocImg api(%s) err(%v)", api, err)
	}
}
