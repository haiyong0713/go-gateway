package popular

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPopLargeCardList(t *testing.T) {
	Convey("PopLargeCardList", t, WithService(func(s *Service) {
		err, res := s.PopLargeCardList(context.Background(), 0, "", 0, 1, 20)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestPopLargeCardSave(t *testing.T) {
	Convey("PopLargeCardSave", t, WithService(func(s *Service) {
		err := s.PopLargeCardSave(context.Background(), &show.PopLargeCard{
			ID:         0,
			Title:      "123",
			CardType:   "",
			RID:        0,
			Bvid:       "",
			WhiteList:  "",
			CreateBy:   "",
			Auto:       0,
			Deleted:    0,
			VideoTitle: "",
			Author:     "",
		})
		So(err, ShouldBeNil)
	}))
}

func TestPopLargeCardOperate(t *testing.T) {
	Convey("PopLargeCardOperate", t, WithService(func(s *Service) {
		err := s.PopLargeCardOperate(context.Background(), 0, 1)
		So(err, ShouldBeNil)
	}))
}
