package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_BangumiList(t *testing.T) {
	Convey("test bangumi list", t, WithService(func(s *Service) {
		mid := int64(0)
		vmid := int64(883968)
		pn := 1
		ps := 10
		data, cnt, err := s.BangumiList(context.Background(), mid, vmid, pn, ps)
		So(err, ShouldBeNil)
		Printf("%+v,%d", data, cnt)
	}))
}

func TestService_FollowList(t *testing.T) {
	Convey("test pgc follow list", t, WithService(func(s *Service) {
		mid := int64(27515233)
		vmid := int64(27515233)
		typ := int32(1)
		pn := int32(1)
		ps := int32(10)
		followStatus := int32(0)
		data, cnt, err := s.FollowList(context.Background(), mid, vmid, typ, pn, ps, followStatus)
		So(err, ShouldBeNil)
		Printf("anime %+v,%d", data, cnt)
		typ = 2
		data, cnt, err = s.FollowList(context.Background(), mid, vmid, typ, pn, ps, followStatus)
		So(err, ShouldBeNil)
		Printf("cinema %+v,%d", data, cnt)
	}))
}
