package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	pb2 "go-gateway/app/app-svr/resource/service/api/v2"

	"github.com/glycerine/goconvey/convey"
)

func TestService_GetWebSpecialCard(t *testing.T) {
	convey.Convey("get web special card", t, WithService(func(s *Service) {
		req := pb2.NoArgRequest{}
		tmp, err := s.GetWebSpecialCard(context.Background(), &req)
		if err != nil {
			panic(err)
		}
		byte, _ := json.Marshal(tmp)
		fmt.Println(string(byte))
	}))
}

func TestService_GetAppSpecialCard(t *testing.T) {
	convey.Convey("get app special card", t, WithService(func(s *Service) {
		req := pb2.NoArgRequest{}
		tmp, err := s.GetAppSpecialCard(context.Background(), &req)
		if err != nil {
			panic(err)
		}
		byte, _ := json.Marshal(tmp)
		fmt.Println(string(byte))
	}))
}

func TestService_GetAppRcmdRelatePgc(t *testing.T) {
	convey.Convey("get app rcmd relate pgc", t, WithService(func(s *Service) {
		req := &pb2.AppRcmdRelatePgcRequest{
			Id:      0,
			MobiApp: "",
			Device:  "",
			Build:   0,
		}
		res, err := s.GetAppRcmdRelatePgc(context.Background(), req)
		if err != nil {
			panic(err)
		}
		byte, _ := json.Marshal(res)
		fmt.Println(string(byte))
	}))
}
