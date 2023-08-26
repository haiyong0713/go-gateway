package service

import (
	"context"

	"go-gateway/app/app-svr/kvo/job/internal/dao"

	"go-common/library/conf/paladin"
	"go-common/library/database/taishan"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"
	"go-common/library/queue/databus/databusutil"

	"github.com/golang/protobuf/ptypes/empty"
)

// Service service.
type Service struct {
	cfg     *Config
	dao     dao.Dao
	player  *databusutil.Group
	buvid   *databusutil.Group
	taishan taishan.TaishanProxyClient
}

type Config struct {
	DoInterval   int
	TaishanToken string
	TaishanTable string
}

// New new a service and return.
func New(d dao.Dao) (s *Service, err error) {
	var (
		cfg *Config
		// databus config.
		playerCsmr struct {
			Playerutil *databusutil.Config
			Player     *databus.Config
			Buvidutil  *databusutil.Config
			Buvid      *databus.Config
		}
		taishanCfg struct {
			TaishanRPC *warden.ClientConfig
		}
	)
	checkErr(paladin.Get("application.toml").UnmarshalTOML(&cfg))
	checkErr(paladin.Get("databus.toml").UnmarshalTOML(&playerCsmr))
	checkErr(paladin.Get("taishan.toml").UnmarshalTOML(&taishanCfg))
	s = &Service{
		dao: d,
		cfg: cfg,
	}
	if s.taishan, err = taishan.NewClient(taishanCfg.TaishanRPC); err != nil {
		panic(err)
	}
	s.initPlayer(playerCsmr.Player, playerCsmr.Playerutil)
	s.initBuvid(playerCsmr.Buvid, playerCsmr.Buvidutil)
	return
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
	s.player.Close()
}

func (s *Service) userDoc(ctx context.Context, mid int64, buvid string, moduleKey int) (bs []byte, err error) {
	var taishanbs []byte
	if bs, err = s.dao.UserDocRds(ctx, mid, buvid, moduleKey); err != nil {
		err = nil
	}
	if bs != nil {
		if len(bs) == 0 {
			bs = nil
			return
		}
	}
	if taishanbs, err = s.userDocTaiShan(ctx, mid, buvid, moduleKey); err != nil {
		log.Error("d.userConfDB(mid:%d, buvid:%s, modulekey:%d) err(%v)", mid, buvid, moduleKey, err)
		return
	}
	if bs != nil && taishanbs != nil {
		log.Warn("taishan redis diff start (mid:%d,buvid:%s,modulekey:%d)", mid, buvid, moduleKey)
		if string(bs) != string(taishanbs) {
			log.Warn("taishan redis diff start failer (mid:%d,buvid:%s,modulekey:%d) (bs:%s,taishan:%s)", mid, buvid, moduleKey, string(bs), string(taishanbs))
		}
	}
	if bs == nil {
		bs = taishanbs
	}
	return
}
