package like

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_ClearCache(t *testing.T) {
	Convey("test service ClearCache", t, WithService(func(s *Service) {
		msg := `{"action":"update","table":"act_matchs_object","old":{"name":0},"new":{"id":2,"sid":12,"match_id":2}}`
		err := s.ClearCache(context.Background(), msg)
		So(err, ShouldBeNil)
	}))
}
