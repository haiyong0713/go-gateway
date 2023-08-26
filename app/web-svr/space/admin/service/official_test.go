package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/space/admin/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_UpdateOfficial(t *testing.T) {
	Convey("UpdateOfficial", t, WithService(func(s *Service) {
		var (
			param = &model.SpaceOfficial{
				ID:     4,
				Uid:    3,
				Name:   "test",
				Icon:   "url",
				Scheme: "test",
				Rcmd:   "test",
				IosUrl: "ioss url",
			}
		)
		err := s.UpdateOfficial(param)
		So(err, ShouldBeNil)
	}))
}

func TestService_InsertOfficial(t *testing.T) {
	Convey("AddOfficial", t, WithService(func(s *Service) {
		var (
			param = &model.SpaceOfficial{
				Uid:    1,
				Name:   "test",
				Icon:   "url",
				Scheme: "test",
				Rcmd:   "test",
				IosUrl: "ioss url",
			}
		)
		err := s.AddOfficial(param)
		So(err, ShouldBeNil)
	}))
}

func TestService_Delete(t *testing.T) {
	Convey("Delete", t, WithService(func(s *Service) {
		err := s.DeleteOfficial(1)
		So(err, ShouldBeNil)
	}))
}

func TestService_Official(t *testing.T) {
	Convey("Official", t, WithService(func(s *Service) {
		req := &model.SpaceOfficialParam{Pn: 1, Ps: 10}
		res, err := s.Official(req)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
