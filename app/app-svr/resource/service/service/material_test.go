package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	pb2 "go-gateway/app/app-svr/resource/service/api/v2"

	"github.com/glycerine/goconvey/convey"
)

func TestService_GetMaterial(t *testing.T) {
	convey.Convey("GetMaterial", t, WithService(func(s *Service) {
		req := pb2.MaterialReq{
			Id: []int64{1, 2},
		}
		tmp, err := s.GetMaterial(context.Background(), &req)
		if err != nil {
			panic(err)
		}
		byte, _ := json.Marshal(tmp)
		fmt.Println(string(byte))
	}))
}
