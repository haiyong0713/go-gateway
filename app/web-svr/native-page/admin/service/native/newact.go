package native

import (
	"context"

	"go-common/library/ecode"

	"go-gateway/app/web-svr/native-page/admin/model"
	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

func (s *Service) AddNewact(c context.Context, req *natmdl.AddNewactReq) (*natmdl.AddNewactRly, error) {
	subject, err := s.dao.ActSubject(c, req.Sid)
	if err != nil {
		return nil, ecode.Error(ecode.ServerErr, "数据源获取失败")
	}
	if subject == nil {
		return nil, ecode.Error(ecode.RequestErr, "无效的数据源")
	}
	fromType, ok := model.ActType2FromType[subject.Type]
	if !ok {
		return nil, ecode.Error(ecode.RequestErr, "不支持该数据源类型")
	}
	pages, err := s.dao.PageByFID(c, req.Sid, natpagegrpc.NewactType)
	if err != nil {
		return nil, ecode.Error(ecode.ServerErr, "数据查询失败")
	}
	if len(pages) > 0 {
		return nil, ecode.Error(ecode.RequestErr, "NA页已被创建")
	}
	id, err := s.dao.AddPageFromNewact(c, subject, fromType)
	if err != nil {
		return nil, ecode.Error(ecode.ServerErr, "NA页创建失败")
	}
	return &natmdl.AddNewactRly{ID: id}, nil
}
