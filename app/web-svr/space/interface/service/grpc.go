package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
)

// Official .
func (s *Service) Official(c context.Context, req *pb.OfficialRequest) (res *pb.OfficialReply, err error) {
	if res, err = s.dao.Official(c, req); err != nil {
		log.Error("Official req(%v) err(%v)", req, err)
		return
	}
	if res == nil {
		err = ecode.NothingFound
	}
	return
}
