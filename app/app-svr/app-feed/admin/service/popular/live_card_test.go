package popular

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPopLiveCardList(t *testing.T) {
	Convey("PopLiveCardList", t, WithService(func(s *Service) {
		err, res := s.PopLiveCardList(context.Background(), 0, -1, "", 1, 20)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestPopLiveCardSave(t *testing.T) {
	Convey("PopLargeCardSave", t, WithService(func(s *Service) {
		err := s.PopLiveCardSave(context.Background(), &show.PopLiveCard{
			ID:       12,
			CardType: "",
			RID:      0,
			CreateBy: "",
			Cover:    "",
		}, "", 0)
		So(err, ShouldBeNil)
	}))
}

func TestPopLiveCardOperate(t *testing.T) {
	Convey("PopLiveCardOperate", t, WithService(func(s *Service) {
		err := s.PopLiveCardOperate(context.Background(), 0, 1, "", 0)
		So(err, ShouldBeNil)
	}))
}
