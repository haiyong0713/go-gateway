package common

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_GameInfo(t *testing.T) {
	Convey("test service GameInfo", t, WithService(func(s *Service) {
		res, err := s.GameInfo(c, 49)
		if err != nil {
			fmt.Println(err)
			return
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
