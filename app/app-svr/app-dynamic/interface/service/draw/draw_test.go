package draw

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"

	"go-common/library/log"
)

var (
	s *Service
)

func TestAll(t *testing.T) {
	flag.Set("conf", "../../cmd/app-dynamic-test.toml")
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	} // init log

	s = New(conf.Conf)
	resp, _ := s.SearchAll(context.Background(), &model.SearchAllReq{
		Uid:      uint64(88895133),
		Keyword:  "手办",
		Page:     0,
		PageSize: 10,
		Lat:      float64(30.67807),
		Lng:      float64(104.151805),
	})
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
	time.Sleep(1 * time.Second)
}
