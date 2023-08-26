package online

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	v1 "go-gateway/app/app-svr/player-online/api"
	"go-gateway/app/app-svr/player-online/internal/conf"

	"go-common/component/metadata/device"
)

var (
	s *Service
)

func init() {
	dir, _ := filepath.Abs("../../../configs/player-online-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	s = New(conf.Conf)
	time.Sleep(time.Second)
}

func Test_PlayerOnlineGRPC(t *testing.T) {
	req := &v1.PlayerOnlineReq{
		Aid:      47527433,
		Cid:      83244511,
		PlayOpen: true,
	}

	res, err := s.PlayerOnlineGRPC(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func Test_onlineText(t *testing.T) {
	t1 := s.onlineText(999)
	assert.Equal(t, "999", t1)
	t2 := s.onlineText(1000)
	assert.Equal(t, "1000+", t2)
	t3 := s.onlineText(3189)
	assert.Equal(t, "3000+", t3)
	t4 := s.onlineText(9189)
	assert.Equal(t, "9000+", t4)
	t5 := s.onlineText(10000)
	assert.Equal(t, "1万+", t5)
	t6 := s.onlineText(93011)
	assert.Equal(t, "9.3万+", t6)
	t7 := s.onlineText(210983)
	assert.Equal(t, "10万+", t7)
	t8 := s.onlineText(890983)
	assert.Equal(t, "80万+", t8)
	t9 := s.onlineText(1210983)
	assert.Equal(t, "100万+", t9)
}

func Test_bottomInGrey(t *testing.T) {
	var (
		mid   int64 = 0
		buvid       = ""
		ctx         = device.NewContext(context.Background(), device.Device{RawMobiApp: "android_hd", Build: 61600000})
	)
	res := s.bottomInGrey(ctx, mid, buvid)
	assert.Equal(t, false, res)
}
