package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	pb "go-gateway/app/app-svr/resource/service/api/v1"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_SearchOgv(t *testing.T) {
	Convey("get app banner", t, WithService(func(s *Service) {
		req := pb.SearchOgvReq{Id: 1}
		tmp, err := s.SearchOgv(context.Background(), &req)
		if err != nil {
			panic(err)
		}
		byte, _ := json.Marshal(tmp)
		fmt.Println(string(byte))
	}))
}
