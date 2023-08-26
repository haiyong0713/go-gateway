package appstore

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/appstore"
	appstoremdl "go-gateway/app/web-svr/activity/interface/model/appstore"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	passapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	silverapi "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
	resourceapi "git.bilibili.co/bapis/bapis-go/vip/resource/service"

	"github.com/robfig/cron"
)

// Service struct
type Service struct {
	c              *conf.Config
	dao            *appstore.Dao
	cache          *fanout.Fanout
	cron           *cron.Cron
	rnd            *rand.Rand
	accountClient  accountapi.AccountClient
	passportClient passapi.PassportUserClient
	silverClient   silverapi.SilverbulletProxyClient
	resourceClient resourceapi.ResourceClient
	//couponClient
	activityAppstore map[string]*appstoremdl.ActivityAppstore // map[batchToken]*appstoremdl.ActivityAppstore
}

// Close service
func (s *Service) Close() {
	s.cron.Stop()
	s.dao.Close()
	s.cache.Close()
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   appstore.New(c),
		cache: fanout.New("cache", fanout.Worker(5), fanout.Buffer(10240)),
		cron:  cron.New(),
		rnd:   rand.New(rand.NewSource(time.Now().Unix())),
	}
	var err error
	if s.accountClient, err = accountapi.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.passportClient, err = passapi.NewClient(c.PassClient); err != nil {
		panic(err)
	}
	if s.silverClient, err = silverapi.NewClient(c.SilverClient); err != nil {
		panic(err)
	}
	if s.resourceClient, err = resourceapi.NewClient(c.ResourceClient); err != nil {
		panic(err)
	}
	s.loadActivityAppstore()
	s.createCron()
	return
}

func (s *Service) createCron() {
	var err error
	if err = s.cron.AddFunc("@every 10s", s.loadActivityAppstore); err != nil {
		panic(err)
	}
	s.cron.Start()
}

func (s *Service) AppStoreState(ctx context.Context, arg *appstoremdl.AppStoreStateArg, ua string) (state int64, err error) {
	var (
		ok               bool
		activityAppstore *appstoremdl.ActivityAppstore
		UserTelHashReply *passapi.UserTelHashReply
	)

	if s.c.ModelNameOpen {
		// check ua是否包含model_name
		if !strings.Contains(ua, arg.ModelName) {
			log.Error("AppStoreState ua:%s not contain ModelName: %s", ua, arg.ModelName)
			err = ecode.ActivityAppstoreModelNameValid
			return
		}
	}

	// check ModelName 是非可参加活动
	if activityAppstore, ok = s.activityAppstore[arg.ModelName]; !ok {
		log.Error("AppStoreState MdoelName:%d not exist", arg.ModelName)
		err = ecode.ActivityAppstoreModelNameValid
		return
	}

	if UserTelHashReply, err = s.passportClient.UserTelHash(ctx, &passapi.UserTelHashReq{Mid: arg.MID}); err != nil {
		log.Error("AppStoreState s.passportClient.UserTelHash mid(%d) error(%v)", arg.MID, err)
		if xecode.EqualError(xecode.Int(appstoremdl.PassportNotFoundUserByTel), err) {
			err = xecode.MobileNoVerfiy
		}
		return
	}
	if UserTelHashReply == nil || UserTelHashReply.TelHash == "" {
		err = xecode.MobileNoVerfiy
		return
	}
	log.Info("AppStoreState UserTelHash mid: %d TelHash: %s", arg.MID, UserTelHashReply.TelHash)
	state = appstoremdl.StateUserNotReceived
	if ok, _, _, err = s.checkIsReceived(ctx, activityAppstore.BatchToken, arg.MID, UserTelHashReply.TelHash, arg.Fingerprint, arg.LocalFingerprint, arg.Buvid); err != nil {
		return
	}
	if ok {
		state = appstoremdl.StateUserIsReceived
	}
	return
}

func (s *Service) APPStoreReceive(ctx context.Context, arg *appstoremdl.AppStoreReceiveArg, ua string, path string, referer string) (err error) {
	var (
		ok               bool
		isReceived       bool
		matchLabel       string
		matchKind        int64
		activityAppstore *appstoremdl.ActivityAppstore
	)
	log.Warn("APPStoreReceive arg:+v ua: %s path: %s referer: %s", arg, ua, path, referer)
	if s.c.ModelNameOpen {
		// check ua是否包含model_name
		if !strings.Contains(ua, arg.ModelName) {
			log.Error("APPStoreReceive ua:%s not contain ModelName: %s", ua, arg.ModelName)
			err = ecode.ActivityAppstoreModelNameValid
			return
		}
	}

	// check ModelName 是非可参加活动
	if activityAppstore, ok = s.activityAppstore[arg.ModelName]; !ok {
		log.Error("APPStoreReceive MdoelName:%d not exist", arg.ModelName)
		err = ecode.ActivityAppstoreModelNameValid
		return
	}
	// check 活动时间
	if err = checkTime(activityAppstore.StartTime, activityAppstore.EndTime); err != nil {
		return
	}
	// 已领完
	if activityAppstore.State == appstoremdl.StateAppstoreEnd {
		err = ecode.ActivityAppstoreVipBatchNotEnoughErr
		return
	}
	var (
		UserTelHashReply *passapi.UserTelHashReply
	)
	if UserTelHashReply, err = s.passportClient.UserTelHash(ctx, &passapi.UserTelHashReq{Mid: arg.MID}); err != nil {
		log.Error("s.passportClient.UserTelHash mid(%d) error(%v)", arg.MID, err)
		if xecode.EqualError(xecode.Int(appstoremdl.PassportNotFoundUserByTel), err) {
			err = xecode.MobileNoVerfiy
		}
		return
	}
	if UserTelHashReply == nil || UserTelHashReply.TelHash == "" {
		err = xecode.MobileNoVerfiy
		return
	}
	log.Info("APPStoreReceive UserTelHash mid: %d TelHash: %s", arg.MID, UserTelHashReply.TelHash)
	if isReceived, matchLabel, matchKind, err = s.checkIsReceived(ctx, activityAppstore.BatchToken, arg.MID, UserTelHashReply.TelHash, arg.Fingerprint, arg.LocalFingerprint, arg.LocalFingerprint); err != nil {
		return
	}
	if isReceived {
		err = ecode.ActivityAppstoreIsReceived
		return
	}
	// 风控判断
	riskArg := &silverapi.RiskInfoReq{
		Mid:          arg.MID,
		StrategyName: []string{appstoremdl.RiskName},
		Ip:           metadata.String(ctx, metadata.RemoteIP),
		Api:          path,
		Ua:           ua,
		Referer:      referer,
	}
	reply, err := s.silverClient.RiskInfo(ctx, riskArg, warden.WithTimeoutCallOption(time.Millisecond*50))
	if err != nil {
		log.Warn("APPStoreReceive fail to query user risk info. arg=%+v, err=%+v", arg, err)
		err = nil
	}

	if reply != nil {
		if hit, ok := reply.Infos[appstoremdl.RiskName]; ok {
			if hit.Level > s.c.StrategyLevel {
				log.Error("APPStoreReceive hit.Level: %d > s.c.StrategyLevel: %d", hit.Level, s.c.StrategyLevel)
				err = ecode.ActivityDecideRiskErr
				return
			}
		}
	}

	// 是否封禁判断
	var accountReply *accountapi.ProfileReply
	if accountReply, err = s.accountClient.Profile3(ctx, &accountapi.MidReq{Mid: arg.MID}); err != nil {
		log.Error("APPStoreReceive s.accountClient.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", arg.MID, err)
		return
	}
	if accountReply == nil || accountReply.Profile == nil {
		log.Error("APPStoreReceive s.accountClient.Profile3(c,&accmdl.ArgMid{Mid:%d}) accountReply: %+v error(%v)", arg.MID, accountReply, err)
	} else {
		// 禁封用户
		if accountReply.Profile.Silence == appstoremdl.SilenceForbid {
			err = ecode.ActivityDecideRiskErr
			return
		}
	}

	defer func() {
		s.cache.Do(ctx, func(ctx context.Context) {
			s.dao.DelCacheAppstoreMIDIsRecieved(ctx, activityAppstore.BatchToken, arg.MID)
			s.dao.DelCacheAppstoreTelIsRecieved(ctx, activityAppstore.BatchToken, UserTelHashReply.TelHash)
			s.dao.DelCacheAppstoreIsRecieved(ctx, activityAppstore.BatchToken, matchLabel, matchKind)
		})
	}()

	var orderNo = s.orderID()
	resourceReq := &resourceapi.ResourceUseReq{
		BatchToken: activityAppstore.BatchToken,
		Mid:        arg.MID,
		OrderNo:    orderNo,
		Remark:     "主站活动",
		Appkey:     activityAppstore.Appkey,
	}
	// 调用发放接口
	if _, err = s.resourceClient.ResourceUse(ctx, resourceReq); err != nil {
		log.Error("APPStoreReceive ResourceUse mid: %d BatchToken: %s err: %+v", arg.MID, activityAppstore.BatchToken, err)
		if xecode.EqualError(xecode.Int(appstoremdl.VipBatchNotEnoughErr), err) {
			err = ecode.ActivityAppstoreVipBatchNotEnoughErr
			if _, err1 := s.dao.UpdateAppstoreState(ctx, activityAppstore.BatchToken); err1 != nil {
				log.Error("APPStoreReceive UpdateAppstoreState BatchToken: %s err1: %+v", activityAppstore.BatchToken, err1)
				return
			}
		}
		return
	}
	log.Info("APPStoreReceive success modelName: %s mid: %d telHash: %s fingerprint: %s localFingerprint: %s buvid: %s build: %d", arg.ModelName, arg.MID, UserTelHashReply.TelHash, arg.Fingerprint, arg.LocalFingerprint, arg.Buvid)
	argA := &appstoremdl.ActivityAppstoreReceived{
		Mid:              arg.MID,
		TelHash:          UserTelHashReply.TelHash,
		BatchToken:       activityAppstore.BatchToken,
		Fingerprint:      arg.Fingerprint,
		LocalFingerprint: arg.LocalFingerprint,
		Buvid:            arg.Buvid,
		MatchLabel:       matchLabel,
		MatchKind:        matchKind,
		Build:            arg.Build,
		OrderNo:          orderNo,
		State:            appstoremdl.StateUserIsReceived,
		UserIP:           []byte(metadata.String(ctx, metadata.RemoteIP)),
	}
	if _, err = s.dao.AddAppstoreReceived(ctx, argA); err != nil {
		return
	}
	return
}

// orderID get order id
func (s *Service) orderID() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%05d", s.rnd.Int63n(99999)))
	b.WriteString(fmt.Sprintf("%03d", time.Now().UnixNano()/1e6%1000))
	b.WriteString(time.Now().Format("060102150405"))
	return b.String()
}

func checkTime(startTime time.Time, endTime time.Time) (err error) {
	nowTs := time.Now()
	if nowTs.Before(startTime) {
		err = ecode.ActivityAppstoreNotStart
		return
	}
	if nowTs.After(endTime) {
		err = ecode.ActivityAppstoreEnd
		return
	}
	return
}

// checkIsReceived 是否已领取过 true-已领取 false-未领取
func (s *Service) checkIsReceived(ctx context.Context, batchToken string, mid int64, telHash string, fingerprint string, localFingerprint string, buvid string) (ok bool, matchLabel string, matchKind int64, err error) {
	var (
		res int64
	)
	if mid <= 0 {
		err = xecode.NoLogin
		return
	}
	// mid是否已领取
	if res, err = s.dao.AppstoreMIDIsRecieved(ctx, batchToken, mid); err != nil {
		return
	}
	if res > 0 {
		log.Warn("checkIsReceived AppstoreMIDIsRecieved batchToken: %s mid: %d", batchToken, mid)
		ok = true
		return
	}
	// 手机号是否已领取
	if res, err = s.dao.AppstoreTelIsRecieved(ctx, batchToken, telHash); err != nil {
		return
	}
	if res > 0 {
		log.Warn("checkIsReceived AppstoreTelIsRecieved batchToken: %s mid: %d telHash: %s", batchToken, mid, telHash)
		ok = true
		return
	}

	matchLabel, matchKind = lableMatch(fingerprint, localFingerprint, buvid)
	if matchLabel == "" {
		log.Warn("checkIsReceived matchLabel not match mid: %d", mid)
		err = ecode.ActivityDecideRiskErr
		ok = true
		return
	}
	if res, err = s.dao.AppstoreIsRecieved(ctx, batchToken, matchLabel, matchKind); err != nil {
		return
	}
	if res > 0 {
		log.Warn("checkIsReceived AppstoreTelIsRecieved batchToken: %s matchLabel: %s matchKind: %d mid: %d", batchToken, matchLabel, matchKind, mid)
		ok = true
		return
	}
	return
}

func lableMatch(fingerprint string, localFingerprint string, buvid string) (matchLabel string, matchKind int64) {
	if fingerprint != "" {
		matchLabel = fingerprint
		matchKind = appstoremdl.MatchkindFingerprint
		return
	}
	if localFingerprint != "" {
		matchLabel = localFingerprint
		matchKind = appstoremdl.MatchkindLocalFingerprint
		return
	}
	if buvid != "" {
		matchLabel = buvid
		matchKind = appstoremdl.MatchkindLocalBuvid
		return
	}
	return
}

func (s *Service) loadActivityAppstore() {
	var (
		err   error
		datas []*appstoremdl.ActivityAppstore
	)
	if datas, err = s.dao.RawAppstoreAll(context.Background()); err != nil {
		return
	}
	if datas == nil {
		log.Warn("loadActivityAppstore datas is nil")
		return
	}
	temps := make(map[string]*appstoremdl.ActivityAppstore, len(datas))
	for _, data := range datas {
		temps[data.ModelName] = data
		log.Warn("loadActivityAppstore %+v", data)
	}
	s.activityAppstore = temps
	return
}
