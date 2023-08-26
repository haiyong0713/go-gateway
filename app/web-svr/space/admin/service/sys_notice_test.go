package service

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/space/admin/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_SysNoticeUidAdd(t *testing.T) {
	var (
		param = &model.SysNotUidAddDel{
			ID:   159,
			UIDs: []int64{3, 4},
		}
	)
	Convey("SysNoticeUidAdd", t, WithService(func(s *Service) {
		err := s.SysNoticeUidAdd(context.Background(), param)
		So(err, ShouldBeNil)
	}))
}

func TestService_SysNoticeUp(t *testing.T) {
	var (
		param = &model.SysNoticeUp{
			ID:         158,
			NoticeType: 1,
			Content:    "testtest",
			Scopes:     "1,2",
		}
	)
	Convey("SysNoticeUp", t, WithService(func(s *Service) {
		err := s.SysNoticeUp(context.Background(), param)
		So(err, ShouldBeNil)
	}))
}
