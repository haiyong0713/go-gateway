package service

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"go-gateway/app/app-svr/app-player/job/model"
	smodel "go-gateway/app/app-svr/playurl/service/model"

	"go-common/library/ecode"
	"go-common/library/log"

	bcgrpc "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"
)

const (
	_ugc         = "ugc"
	_miniPathLen = 2
)

func (s *Service) SaveOnlineInfo(ctx context.Context, manual bool, aid int64) {
	onlineInfo, err := s.FetchOnlineInfo(ctx, aid)
	if err != nil {
		return
	}
	//不是运营手动操作有触发阈值
	if !manual && onlineInfo.AidTotal < s.c.Custom.SmoothThreshold {
		return
	}
	if err := s.dao.AddOnlineInfo(ctx, aid, onlineInfo); err != nil {
		log.Error("s.dao.AddOnlineInfo error(%+v), aid(%d)", err, aid)
	}
}

// FetchOnlineInfo 返回值为mobi_app + cid + count
func (s *Service) FetchOnlineInfo(ctx context.Context, aid int64) (*smodel.OnlineInfo, error) {
	req := &bcgrpc.OnlineReq{
		Business: _ugc,
		Aid:      aid,
	}
	res, err := s.bcClient.Online(ctx, req)
	if err != nil {
		log.Error("FetchOnlineInfo error(%+v), aid(%d)", err, aid)
		return nil, err
	}
	if res == nil {
		return nil, ecode.NothingFound
	}
	//根据aid取出web和全站端所对应的cid的值
	webCidCount, appCidCount := fetchCidOnlineByPlat(res.Rooms)
	reply := &smodel.OnlineInfo{
		WebCidCount: webCidCount,
		AppCidCount: appCidCount,
		Time:        time.Now().Unix(),
		AidTotal:    res.Total,
	}
	return reply, nil
}

func fetchCidOnlineByPlat(rawOnlineMsg map[string]int64) (map[int64]int64, map[int64]int64) {
	webCidCount := make(map[int64]int64, len(rawOnlineMsg))
	appCidCount := make(map[int64]int64, len(rawOnlineMsg))
	for onlineUrl, count := range rawOnlineMsg {
		ru, err := url.Parse(onlineUrl)
		if err != nil {
			log.Error("FetchOnlineInfo url.Parse error(%+v) roomUrl(%s)", err, onlineUrl)
			continue
		}
		onlineMsg, err := splitCidAndMobiApp(ru)
		if err != nil {
			log.Error("splitCidAndMobiApp error(%+v)", err)
			continue
		}
		if onlineMsg.MobileApp == smodel.Web {
			webCidCount[onlineMsg.Cid] = count
			continue
		}
		appCidCount[onlineMsg.Cid] = appCidCount[onlineMsg.Cid] + count
	}
	return webCidCount, appCidCount
}

func splitCidAndMobiApp(ru *url.URL) (*model.SplitOnlineMsg, error) {
	msg := &model.SplitOnlineMsg{}
	paths := strings.Split(ru.Path, "/")
	if len(paths) < _miniPathLen {
		return nil, errors.Wrapf(ecode.RequestErr, "illegal path len url(%+v)", ru)
	}
	cid, err := strconv.ParseInt(paths[1], 10, 64)
	if err != nil {
		return nil, errors.Wrapf(ecode.RequestErr, "strconv.ParseInt error")
	}
	msg.Cid = cid
	switch ru.Scheme {
	case _ugc:
		msg.MobileApp = smodel.App
	case smodel.Video:
		msg.MobileApp = smodel.Web
	default:
		return nil, errors.Wrapf(ecode.RequestErr, "unknown online type")
	}
	return msg, nil
}
