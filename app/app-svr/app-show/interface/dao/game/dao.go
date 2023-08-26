package game

import (
	"context"
	"net/url"
	"strconv"
	"sync"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-show/interface/conf"
	gamdl "go-gateway/app/app-svr/app-show/interface/model/game"

	"github.com/pkg/errors"
)

const (
	_gameURI = "/dynamic_card/multi_game_info"
)

type Dao struct {
	c           *conf.Config
	client      *httpx.Client
	gameInfoURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:           c,
		client:      httpx.NewClient(c.HTTPGameCo),
		gameInfoURL: c.Host.GameCo + _gameURI,
	}
	return
}

// nolint:gomnd
func (d *Dao) GetPlatformType(mobiApp string) int {
	if mobiApp == "android" {
		return 1
	} else if mobiApp == "iphone" {
		return 2
	} else {
		return 0
	}
}

func (d *Dao) BatchMultiGameInfo(c context.Context, gameIDs []int64, mid int64, mobiApp string) map[int64]*gamdl.Item {
	// 去重
	idsSet := make(map[int64]struct{})
	for _, v := range gameIDs {
		idsSet[v] = struct{}{}
	}
	var ids []int64
	for id := range idsSet {
		if id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return make(map[int64]*gamdl.Item)
	}
	var (
		idsLen = len(ids)
		mutex  = sync.Mutex{}
		maxIDs = 20
	)
	aidMap := make([]*gamdl.Item, 0)
	gp := errgroup.WithContext(c)
	for i := 0; i < idsLen; i += maxIDs {
		var partAids []int64
		if i+maxIDs > idsLen {
			partAids = ids[i:]
		} else {
			partAids = ids[i : i+maxIDs]
		}
		gp.Go(func(ctx context.Context) error {
			tmpRes, err := d.MultiGameInfo(ctx, partAids, mid, mobiApp)
			if err != nil { //错误忽略，降级处理
				log.Error("d.MultiGameInfo(%v,%d,%s) error(%v)", partAids, mid, mobiApp, err)
				return nil
			}
			if len(tmpRes) > 0 {
				mutex.Lock()
				aidMap = append(aidMap, tmpRes...)
				mutex.Unlock()
			}
			return nil
		})
	}
	_ = gp.Wait()
	rly := make(map[int64]*gamdl.Item)
	for _, v := range aidMap {
		if v == nil || v.GameBaseId == 0 {
			continue
		}
		rly[v.GameBaseId] = v
	}
	return rly
}

func (d *Dao) MultiGameInfo(c context.Context, gameIDs []int64, mid int64, mobiApp string) ([]*gamdl.Item, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	platformType := strconv.Itoa(d.GetPlatformType(mobiApp))
	params.Set("game_base_ids", xstr.JoinInts(gameIDs))
	params.Set("uid", strconv.FormatInt(mid, 10))
	ts := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	params.Set("ts", ts)
	//平台类型：0=PC，1=安卓，2=IOS
	params.Set("platform_type", platformType)
	params.Set("source", "1009")
	var res struct {
		Code int           `json:"code"`
		Data []*gamdl.Item `json:"data"`
	}
	if err := d.client.Get(c, d.gameInfoURL, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.gameInfoURL+"?"+params.Encode())
		return nil, err
	}
	return res.Data, nil
}
