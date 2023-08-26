package service

import (
	"encoding/json"
	"testing"

	"go-gateway/app/web-svr/esports/admin/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_GameList(t *testing.T) {
	Convey("test game list", t, WithService(func(s *Service) {
		oids := []int64{21, 16}
		data, err := s.gameList(model.TypeTeam, oids)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(data)
		Printf(string(bs))
	}))
}
