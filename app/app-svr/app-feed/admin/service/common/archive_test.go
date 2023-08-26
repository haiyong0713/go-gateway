package common

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestService_SearchArchiveCheck(t *testing.T) {
	convey.Convey("test service Archives", t, WithService(func(s *Service) {
		//640001692 禁止搜索
		res, err := s.SearchArchiveCheck(c, 10100680)
		if err != nil {
			fmt.Println(err)
			return
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
