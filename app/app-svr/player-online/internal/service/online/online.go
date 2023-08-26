package online

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/stat/prom"

	v1 "go-gateway/app/app-svr/player-online/api"
	"go-gateway/app/app-svr/player-online/internal/conf"
	"go-gateway/app/app-svr/player-online/internal/dao/archive"
	"go-gateway/app/app-svr/player-online/internal/dao/pgc"
	"go-gateway/app/app-svr/player-online/internal/dao/playurl"
	redisDao "go-gateway/app/app-svr/player-online/internal/dao/redis"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

const (
	PercentageGreyMaxNum = 100
)

type Service struct {
	c        *conf.Config
	psDao    *playurl.Dao
	pgcDao   *pgc.Dao
	arcDao   *archive.Dao
	redisDao *redisDao.Dao
	errProm  *prom.Prom
	infoProm *prom.Prom
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:        c,
		psDao:    playurl.New(c),
		pgcDao:   pgc.New(c),
		arcDao:   archive.New(c),
		redisDao: redisDao.New(c),
		errProm:  prom.BusinessErrCount,
		infoProm: prom.BusinessInfoCount,
	}
	return
}

// Ping Service
func (s *Service) Ping(c context.Context) (err error) {
	return
}

// PlayerOnlineGRPC warden server list
func (s *Service) PlayerOnlineGRPC(c context.Context, req *v1.PlayerOnlineReq) (*v1.PlayerOnlineReply, error) {
	//功能开关
	if s.c.Online == nil || !s.c.Online.SwitchOn {
		return defaultPlayerOnlineReply(), nil
	}

	var (
		buvid     string
		mid       int64
		totalText string
		sdmText   string
	)

	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}

	if d, ok := device.FromContext(c); ok {
		buvid = d.Buvid
	}

	//获取在看人数
	onlineCnt, err := s.getOnlineCount(c, req.Aid, req.Cid)
	if err != nil {
		log.Error("s.PlayerOnlineGRPC aid(+%v) cid(+%v) error(+%v)", req.Aid, req.Cid, err)
	}

	//由于统计有延迟，如果正常返回0，处理为1人在看
	if err == nil {
		if onlineCnt == 0 {
			onlineCnt = 1
		}
		totalText = fmt.Sprintf(s.c.Online.Text, s.onlineText(onlineCnt))
		sdmText = fmt.Sprintf(s.c.OnlineSpecialDM.Text, s.onlineText(onlineCnt))
	}

	res := &v1.PlayerOnlineReply{
		SecNext:         s.c.Online.SecNext,
		TotalText:       totalText,
		BottomShow:      s.bottomInGrey(c, mid, buvid),
		SdmShow:         s.sdmInGrey(mid, buvid) && s.canSdmShowByBuvid(c, req.PlayOpen, onlineCnt, buvid, req.Aid, req.Cid),
		SdmText:         sdmText,
		TotalNumberText: s.onlineText(onlineCnt),
	}

	return res, nil
}

// hasShow true未展示过，可以展示
func (s *Service) canSdmShowByBuvid(c context.Context, playOpen bool, onlineCnt int64, buvid string, aid int64, cid int64) bool {
	//新的播放行为
	if playOpen {
		_ = s.redisDao.DelSdmCache(c, aid, cid, buvid)
		s.infoProm.Incr("sdm_play_open")
	}

	if onlineCnt < s.c.OnlineSpecialDM.ShowCount {
		s.infoProm.Incr("sdm_under_cnt")
		return false
	}

	if playOpen {
		_ = s.redisDao.SetSdmCache(c, aid, cid, buvid, 86400, 1)
		s.infoProm.Incr("sdm_show")
		return true
	}

	_, err := s.redisDao.GetSdmCache(c, aid, cid, buvid)
	if err == redis.ErrNil {
		_ = s.redisDao.SetSdmCache(c, aid, cid, buvid, 86400, 1)
		s.infoProm.Incr("sdm_show")
		return true
	}

	s.infoProm.Incr("sdm_not_show")
	return false
}

func (s *Service) getOnlineCount(c context.Context, aid int64, cid int64) (int64, error) {
	if cid <= 0 || aid <= 0 {
		return 0, fmt.Errorf("s.psDao.PlayOnline lack of aid(%d) cid(%d)", aid, cid)
	}

	cnt, err := s.redisDao.GetOnlineCountCache(c, aid, cid)
	if err == nil {
		return cnt, nil
	}

	onlineCnt, err := s.psDao.PlayOnline(c, aid, cid)
	if err != nil {
		log.Error("s.psDao.PlayOnline aid(+%v) cid(+%v) error(+%v)", aid, cid, err)
		return 0, err
	}

	_ = s.redisDao.SetOnlineCountCache(c, aid, cid, 30, onlineCnt)

	return onlineCnt, nil
}

// nolint:gomnd
func (s *Service) onlineText(number int64) string {
	if number < 1000 {
		return strconv.FormatInt(number, 10)
	}
	if number < 10000 {
		return strconv.FormatInt(number/1000*1000, 10) + "+"
	}
	if number < 100000 {
		return strings.TrimSuffix(strconv.FormatFloat(float64(number)/10000, 'f', 1, 64), ".0") + "万+"
	}
	if number < 1000000 {
		return strings.TrimSuffix(strconv.FormatFloat(float64(number)/10000, 'f', 0, 64), ".0") + "万+"
	}
	return "100万+"
}

func (s *Service) bottomInGrey(c context.Context, mid int64, buvid string) bool {
	if s.c.OnlineBottom == nil || !s.c.OnlineBottom.SwitchOn {
		return false
	}

	// 隐藏常驻在线人数及开关，适用范围：iPad粉、iPad HD、安卓HD，后续优化后再放开
	if pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD()
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidHD()
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPad()
	}).FinishOr(true) {
		return false
	}

	_, ok := s.c.OnlineBottom.Mid[strconv.FormatInt(mid, 10)]
	group := crc32.ChecksumIEEE([]byte(buvid+"_online_ctrl")) % PercentageGreyMaxNum

	return ok || group < uint32(s.c.OnlineBottom.Gray)
}

func (s *Service) sdmInGrey(mid int64, buvid string) bool {
	if s.c.OnlineSpecialDM == nil || !s.c.OnlineSpecialDM.SwitchOn {
		return false
	}

	_, ok := s.c.OnlineSpecialDM.Mid[strconv.FormatInt(mid, 10)]
	group := crc32.ChecksumIEEE([]byte(buvid+"_online_ctrl")) % PercentageGreyMaxNum

	return ok || group < uint32(s.c.OnlineSpecialDM.Gray)
}

func defaultPlayerOnlineReply() *v1.PlayerOnlineReply {
	return &v1.PlayerOnlineReply{
		TotalText:  "",
		SecNext:    60,
		BottomShow: false,
		SdmShow:    false,
	}
}
