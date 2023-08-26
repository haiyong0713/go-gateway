package service

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	sss *Service
)

func TestRedisMSetWithExp(t *testing.T) {
	kvMap := make(map[string][]byte)
	for i := 0; i < 3; i++ {
		value, _ := json.Marshal("test" + strconv.Itoa(i))
		kvMap[strconv.Itoa(i)] = value
	}
	convey.Convey("TestRedisMSetWithExp", t, func() {
		sss.redisMSetWithExp(context.Background(), kvMap, 100)
	})

}
