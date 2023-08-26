package newyear2021

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"go-gateway/app/web-svr/activity/interface/client"
	rewardModel "go-gateway/app/web-svr/activity/interface/model/rewards"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"

	xecode "go-gateway/app/web-svr/activity/ecode"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	dao "go-gateway/app/web-svr/activity/interface/dao/newyear2021"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
	likeSvr "go-gateway/app/web-svr/activity/interface/service/like"
	"go-gateway/app/web-svr/activity/interface/tool"

	"github.com/Shopify/sarama"
)

const (
	pubKey4ARReward      = "bnj_AR_reward_%v_%v_%v"
	bizNameOfARPub       = "bnj_2021_AR_pub"
	bizNameOfARPubBackup = "bnj_2021_AR_pub_backup"
	bizNameOfARDevicePub = "bnj2021_AR_device_pub"

	adaptLevelOfHigh          = "high"
	adaptLevelOfMiddle        = "middle"
	adaptLevelOfMiddleUnknown = "middle_unknown"
	adaptLevelOfLow           = "low"

	cacheKey4BackupOfARExchange = "bnj2021_AR_exchange_%02d"
)

var (
	currentDateStr        string
	maxCommitTimes        int64
	bnjReserveCount       int64
	bnjARUV               int64
	score2CouponRelations []*model.Score2Coupon
	arSetting             *model.ARSetting

	blackList4Android      *model.ARBlackList
	blackList4Ios          *model.ARBlackList
	deviceScoreMap4Android map[string]int64
	deviceScoreMap4Ios     map[string]int64
	appRedirect            *model.AppRedirect

	resourceOfPC map[string]interface{}
)

func init() {
	resourceOfPC = make(map[string]interface{}, 0)
	appRedirect = new(model.AppRedirect)
	blackList4Android = new(model.ARBlackList)
	{
		blackList4Android.ModelMap = make(map[string]int64, 0)
		blackList4Android.VersionMap = make(map[string]int64, 0)
		blackList4Android.HighScore = 9000
		blackList4Android.MiddleScore = 8000
		blackList4Android.MemoryRuleList = make([]*model.MemoryRule, 0)
		blackList4Android.VersionRuleList = make([]*model.VersionRule, 0)
	}
	blackList4Ios = new(model.ARBlackList)
	{
		blackList4Ios.ModelMap = make(map[string]int64, 0)
		blackList4Ios.VersionMap = make(map[string]int64, 0)
		blackList4Ios.HighScore = 9000
		blackList4Ios.MiddleScore = 8500
		blackList4Ios.MemoryRuleList = make([]*model.MemoryRule, 0)
		blackList4Ios.VersionRuleList = make([]*model.VersionRule, 0)
	}
	deviceScoreMap4Android = make(map[string]int64, 0)
	deviceScoreMap4Ios = make(map[string]int64, 0)

	score2CouponRelations = make([]*model.Score2Coupon, 0)
	currentDateStr = time.Now().Format("20060102")
	maxCommitTimes = 3
	arSetting = new(model.ARSetting)
}

func AsyncResetWebViewData4PC() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if d, err := dao.FetchWebViewResource4PC(context.Background()); err == nil {
				resourceOfPC = d
			}
		}
	}
}

func (s *Service) ASyncUpdateBnjReserveCount() {
	if BnjReserveInfo.ActivityID == 0 {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sc := s.GetConf()
			updateBnjReserveCount(BnjReserveInfo.ActivityID)
			var (
				counterErr error
				counter    int64
			)
			counter, counterErr = client.GetActPlatformCounterTotal(
				context.Background(),
				88888888,
				sc.ActPlatLotteryCounterName,
				sc.ActPlatActId)
			if counterErr == nil {
				bnjARUV = counter
			}
		}
	}
}

func updateBnjReserveCount(reserveID int64) {
	m, err := likeSvr.CommonLikeService.GetActSubjectsReserveIDsFollowTotalByOptimization(
		context.Background(), []int64{reserveID})
	if err != nil || m == nil || len(m) == 0 {
		return
	}

	if d, ok := m[reserveID]; ok {
		bnjReserveCount = d
	}
}

func updateBlackListByFilename(filename string) {
	if filename == "" {
		return
	}

	if bs, err := readByteSliceByFilename(filename); err == nil {
		list := new(model.AllARBlackList)
		if err := json.Unmarshal(bs, &list); err == nil {
			blackList4Android = list.Android
			blackList4Ios = list.Ios
			appRedirect = list.Redirect
		}
	}
}

func readByteSliceByFilename(filename string) (bs []byte, err error) {
	var f *os.File
	bs = make([]byte, 0)
	f, err = os.Open(filename)
	if err != nil {
		return
	}

	defer func() {
		_ = f.Close()
	}()

	bs, err = ioutil.ReadAll(f)

	return
}

func updateBnjStrategyByFilename(filename string) {
	if filename == "" {
		return
	}

	var (
		bs  []byte
		err error
	)

	bs, err = readByteSliceByFilename(filename)
	if err == nil {
		cfg := new(model.BnjStrategy)
		if tmpErr := json.Unmarshal(bs, &cfg); tmpErr == nil {
			BnjStrategyInfo = cfg
		}
	}
}

func updateBnjActivityIDByFilename(filename string) {
	if filename == "" {
		return
	}

	var (
		bs  []byte
		err error
	)

	bs, err = readByteSliceByFilename(filename)
	if err == nil {
		cfg := new(model.ReserveInPublicize)
		if tmpErr := json.Unmarshal(bs, &cfg); tmpErr == nil {
			BnjReserveInfo = cfg
		}
	}
}

func updateQuotaActivityIDByFilename(filename string) {
	if filename == "" {
		return
	}

	var (
		bs  []byte
		err error
	)

	bs, err = readByteSliceByFilename(filename)
	if err == nil {
		cfg := make(map[string]int64, 0)
		if tmpErr := json.Unmarshal(bs, &cfg); tmpErr == nil {
			QuotaActivityIDMap = cfg
		}
	}
}

func updateARConfigurationByFilename(filename string) {
	if filename == "" {
		return
	}

	var (
		bs  []byte
		err error
	)

	bs, err = readByteSliceByFilename(filename)
	if err == nil {
		cfg := new(model.ARConfig)
		if tmpErr := json.Unmarshal(bs, &cfg); tmpErr == nil {
			BnjARConfig = cfg
			maxCommitTimes = BnjARConfig.AR.DayGameTimes
		}
	}
}

func updateDeviceScoreMapByFilename(filename string) {
	if filename == "" {
		return
	}

	m, err := readDeviceScoreMapByFileName(filename)
	if err != nil {
		return
	}

	tmp := make(map[string]int64, 0)
	for k, v := range m {
		tmp[strings.ToLower(k)] = v
	}

	switch filename {
	case genWatchedFilename(filenameOfIosScore):
		deviceScoreMap4Ios = tmp
	case genWatchedFilename(filenameOfAndroidScore):
		deviceScoreMap4Android = tmp
	}
}

func readDeviceScoreMapByFileName(filename string) (m map[string]int64, err error) {
	m = make(map[string]int64, 0)
	var bs []byte
	bs, err = readByteSliceByFilename(filename)
	if err == nil {
		err = json.Unmarshal(bs, &m)
	}

	return
}

func UpdateCurrentDateStr(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			currentDateStr = time.Now().Format("20060102")
		case <-ctx.Done():
			return
		}
	}
}

func UpdateScore2CouponRelations(ctx context.Context) (err error) {
	score2CouponRelations, err = dao.FetchARExchangeRuleList(context.Background())

	return
}

func ASyncResetScore2CouponRelations(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			_ = UpdateScore2CouponRelations(context.Background())
		case <-ctx.Done():
			return
		}
	}
}

func calculateScore2CouponByScore(score int64) (info model.Score2Coupon) {
	info = model.Score2Coupon{
		Score:  score,
		Coupon: 0,
	}

	for _, v := range score2CouponRelations {
		if score >= v.Score {
			info.Coupon = v.Coupon

			continue
		}

		break
	}

	return
}

func (s *Service) ARConfiguration(optType int64) (resp interface{}) {
	switch optType {
	case 1:
		resp = blackList4Android
	case 2:
		resp = blackList4Ios
	case 3:
		resp = deviceScoreMap4Android
	case 4:
		resp = deviceScoreMap4Ios
	case 5:
		resp = examBank
	}

	return
}

func fetchQuotaByMID(ctx context.Context, mid int64) (quota int64, err error) {
	var usedTimes int64
	usedTimes, err = dao.FetchGameCommitTimes(ctx, mid, currentDateStr)
	if err != nil {
		err = ecode.ServerErr

		return
	}

	quota = maxCommitTimes - usedTimes

	return
}

func fetchQuotaAndRequestIDByMID(ctx context.Context, mid int64) (quota int64, requestID string, err error) {
	var usedTimes int64
	usedTimes, err = dao.FetchGameCommitTimes(ctx, mid, currentDateStr)
	if err != nil {
		err = ecode.ServerErr

		return
	}

	quota = maxCommitTimes - usedTimes
	if quota <= 0 {
		quota = 0

		return
	}

	requestID, err = dao.GenGameCommitRequestID(ctx, mid, currentDateStr)
	if err != nil {
		err = ecode.ServerErr
	}

	return
}

func usedTimes4ARCommit(ctx context.Context, mid int64) (usedTimes int64, err error) {
	usedTimes, err = dao.FetchGameCommitTimes(ctx, mid, currentDateStr)
	if err != nil {
		err = ecode.ServerErr
	}

	return
}

func canCommitByUsedTimes(usedTimes int64) (can bool) {
	if usedTimes < maxCommitTimes {
		can = true
	}

	return
}

func (s *Service) ARQuota(ctx context.Context, mid int64) (quota int64, err error) {
	quota, err = fetchQuotaByMID(ctx, mid)

	return
}

func (s *Service) ARConfirm(ctx context.Context, mid int64) (resp *model.ARConfirm, err error) {
	resp = new(model.ARConfirm)
	{
		resp.Confirm = dao.ARConfirmInH5(ctx, mid)
		resp.Message = BnjARConfig.ConfirmMessage
	}

	return
}

func (s *Service) ARPreExchange(ctx context.Context, mid, score int64) (resp *model.GamePreCommitResp, err error) {
	resp = new(model.GamePreCommitResp)
	{
		resp.Reward = calculateScore2CouponByScore(score)
		resp.Quota, resp.RequestID, err = fetchQuotaAndRequestIDByMID(ctx, mid)
	}

	sc := s.GetConf()
	_ = dao.ResetLastARReward(ctx, mid, resp.Reward)
	msg := &rewardModel.ActPlatActivityPoints{
		Points:    1,
		Timestamp: time.Now().Unix(),
		Mid:       mid,
		Source:    408933983,
		Activity:  sc.ActPlatActId,
		Business:  sc.ActPlatGameCounterName,
		Extra:     "",
	}
	errDataBus := s.actPlatDatabus.Send(ctx, fmt.Sprintf("%v-%v", mid, time.Now().Unix()), msg)
	if errDataBus != nil { //do not return error here
		log.Errorc(ctx, "s.ARPreExchange send actPlatDatabus error: %v", err)
	}
	return
}

func LastARRewardByMID(ctx context.Context, mid int64) *model.Score2Coupon {
	reward, _ := dao.GetLastARReward(ctx, mid)

	return reward
}

func LastDrawAwardByMID(ctx context.Context, mid int64) (name string, err error) {
	name, err = dao.GetLastDrawAward(ctx, mid)

	return
}

func (s *Service) ARExchange(ctx context.Context, mid int64, score *model.GameScore,
	report *model.RiskManagementReportInfoOfGame) (resp *model.GameCommitResp, err error) {
	var (
		isRequestIDValid bool
		usedTimes        int64
	)

	if score.GameType == 1 {
		_ = dao.UpdateUserLastARIdentity(ctx, mid)
	}

	resp = new(model.GameCommitResp)
	isRequestIDValid, err = dao.IsRequestIDValid(ctx, mid, score.RequestID)
	if err == redis.ErrNil {
		err = xecode.BNJInvalidRequestID

		return
	}

	if err != nil {
		err = ecode.ServerErr

		return
	}

	if !isRequestIDValid {
		err = xecode.BNJInvalidRequestID

		return
	}

	usedTimes, err = usedTimes4ARCommit(ctx, mid)
	if err != nil {
		return
	}

	if !canCommitByUsedTimes(usedTimes) {
		err = xecode.BNJNoEnoughCommitTimes

		return
	}

	{
		resp.Reward = calculateScore2CouponByScore(score.Score)
		resp.Quota = maxCommitTimes - usedTimes - 1
	}

	pubKey := fmt.Sprintf(pubKey4ARReward, mid, currentDateStr, usedTimes+1)
	exchange := new(v1.BNJ2021ARExchangeReq)
	{
		exchange.Mid = mid
		exchange.Score = score.Score
		exchange.Coupon = resp.Reward.Coupon
		exchange.DateStr = currentDateStr
	}
	bs, _ := json.Marshal(exchange)
	if ARRewardProducer != nil || !BnjStrategyInfo.BackupPub {
		if pubErr := ARRewardProducer.Send(ctx, pubKey, bs); pubErr != nil {
			tool.IncrCommonBizStatus(bizNameOfARPub, tool.StatusOfFailed)
			log.Errorc(ctx, "BNJ_AR pub reward failed, err: %v", err)
			if pubErr := pubExchangeIntoBackup(ctx, string(bs)); pubErr != nil {
				err = ecode.ServerErr
			}
		}
	} else {
		err = pubExchangeIntoBackup(ctx, string(bs))
	}

	{
		report.Count = usedTimes + 1
		report.GameType = score.GameType
		report.EndTime = time.Now().Unix()
		report.Coupon = exchange.Coupon
		report.Duration = arSetting.GameDuration
	}
	_ = component.Report2RiskManagement(ctx, component.RiskManagementScene4ARGame, "", report)

	return
}

func genBackupRandSuffix() int64 {
	rand.Seed(time.Now().UnixNano())

	return rand.Int63n(100)
}

func cacheKey4ARExchange() string {
	suffix := genBackupRandSuffix()

	return fmt.Sprintf(cacheKey4BackupOfARExchange, suffix)
}

func pubExchangeIntoBackup(ctx context.Context, info string) (err error) {
	cacheKey := cacheKey4ARExchange()
	_, err = component.BackUpMQ.Do(ctx, "LPUSH", cacheKey, info)
	if err != nil {
		log.Errorc(ctx, "BNJ_AR pubExchangeIntoBackup failed, err: %v", err)
		err = ecode.ServerErr
	}

	return
}

func IncrARCoupon(ctx context.Context, req *v1.BNJ2021ARCouponReq) (err error) {
	err = dao.IncrARCoupon(ctx, req.Mid, req.Coupon)

	return
}

func InsertARGameLog(ctx context.Context, req *v1.BNJ2021ARExchangeReq) (reply *v1.BNJ2021ARExchangeReply, err error) {
	var usedTimes int64
	reply = new(v1.BNJ2021ARExchangeReply)
	usedTimes, err = usedTimes4ARCommit(ctx, req.Mid)
	if err != nil {
		return
	}

	if !canCommitByUsedTimes(usedTimes) {
		err = xecode.BNJNoEnoughCommitTimes

		return
	}

	reward := calculateScore2CouponByScore(req.Score)
	tmpLog := model.NewScore2Coupon(
		req.Mid,
		req.Score,
		reward.Coupon,
		usedTimes+1,
		req.DateStr)
	if cacheErr := dao.UpsertARCoupon(ctx, tmpLog); cacheErr != nil {
		err = ecode.ServerErr
	}

	return
}

func ARDrawQuota(ctx context.Context, mid int64, activityID string) (isAR bool, quota int64) {
	if _, ok := QuotaActivityIDMap[activityID]; ok {
		isAR = true
		if d, err := dao.FetchUserCoupon(ctx, mid); err == nil {
			quota = d.ND
		}
	}

	return
}

func (s *Service) ARProfile(ctx context.Context, mid int64) (info *model.Profile, err error) {
	info = new(model.Profile)
	{
		info.Score, err = dao.FetchUserGameScore(ctx, mid)
		info.Coupon, err = dao.FetchUserCoupon(ctx, mid)
		info.Exp = 0 // TODO
	}

	return
}

func (s *Service) ARSetting(ctx context.Context) (cfg *model.ARConfig, err error) {
	cfg = BnjARConfig
	if cfg.Timer == 0 {
		cfg.Timer = 180
	}

	return
}

func GenAppTaskGameRedirect4Gateway(ctx context.Context, req *v1.AppJumpReq) (reply *v1.AppJumpReply) {
	reply = GenAppRedirect4Gateway(ctx, req)
	reply.JumpUrl = fmt.Sprintf("%v%v", appRedirect.TaskGame, url.QueryEscape(reply.JumpUrl))

	return
}

func GenAppRedirect4Gateway(ctx context.Context, req *v1.AppJumpReq) (reply *v1.AppJumpReply) {
	reply = new(v1.AppJumpReply)
	{
		reply.JumpUrl = appRedirect.UnSupportAppH5
	}

	info, err := model.ParseUserAgent2UserAppInfo(req.UserAgent)
	if err != nil {
		return
	}

	report := info.Copy2ARDeviceReportInfo(0)
	defer func() {
		pubArDeviceInfoIntoKafka(report)
	}()

	if isUnSupportApp(info) {
		report.SceneID = model.ARDeviceReportSceneOfUnSupportApp

		return
	}

	if isUnSupportBuild(info) {
		reply.JumpUrl = appRedirect.UnSupportBuildH5
		report.SceneID = model.ARDeviceReportSceneOfUnSupportBuild

		return
	}

	reply.JumpUrl = appRedirect.GameH5
	if isAppInfoInBlacklist(info) {
		report.SceneID = model.ARDeviceReportSceneOfBlacklist

		return
	}

	if d := calculateAppLevelByScore(info, req.Memory, report); d != adaptLevelOfLow {
		reply.JumpUrl = appRedirect.ARScheme
	}

	return
}

func isUnSupportBuild(info *model.UserAppInfo) bool {
	b := true
	switch info.Os {
	case model.Os4Android:
		if info.Build >= blackList4Android.SupportBuild {
			b = false
		}
	case model.Os4Ios:
		if info.Build >= blackList4Ios.SupportBuild {
			b = false
		}
	default:
		// TODO
	}

	return b
}

func isUnSupportApp(info *model.UserAppInfo) bool {
	return info.MobiApp != model.MobileApp4Android && info.MobiApp != model.MobileApp4IPhone
}

func (s *Service) ARAdaptLevel(ctx context.Context, ua string, memory int64) (m map[string]interface{}, err error) {
	m = make(map[string]interface{}, 0)
	m["level"] = adaptLevelOfLow
	if level, tmpErr := calculateAdaptLevel(ua, memory); tmpErr == nil {
		m["level"] = level
	}

	return
}

func calculateAdaptLevel(ua string, memory int64) (level string, err error) {
	level = adaptLevelOfLow
	info := new(model.UserAppInfo)
	info, err = model.ParseUserAgent2UserAppInfo(ua)
	if err != nil {
		return
	}

	report := info.Copy2ARDeviceReportInfo(0)
	defer func() {
		pubArDeviceInfoIntoKafka(report)
	}()

	if isAppInfoInBlacklist(info) {
		report.SceneID = model.ARDeviceReportSceneOfBlacklist

		return
	}

	level = calculateAppLevelByScore(info, memory, report)

	return
}

func pubArDeviceInfoIntoKafka(report *model.ARDeviceReportInfo) {
	if BnjARDeviceProducer.Topic == "" {
		return
	}

	bs, _ := json.Marshal(report)
	msg := &sarama.ProducerMessage{
		Topic:    BnjARDeviceProducer.Topic,
		Value:    sarama.ByteEncoder(bs),
		Metadata: "bnj2021_AR_device",
	}

	_, _, pubErr := ARDeviceProducer.SendMessage(msg)
	if pubErr != nil {
		log.Error("AR_device_info pub failed, err: ", pubErr)
		tool.IncrCommonBizStatus(bizNameOfARDevicePub, tool.StatusOfFailed)
	}
}

func calculateAppLevelByScore(info *model.UserAppInfo, memory int64, report *model.ARDeviceReportInfo) (level string) {
	level = adaptLevelOfLow
	switch info.Os {
	case model.Os4Ios:
		level = genAppLevelByModelAndMap(info, deviceScoreMap4Ios, blackList4Ios, memory, report)
	case model.Os4Android:
		level = genAppLevelByModelAndMap(info, deviceScoreMap4Android, blackList4Android, memory, report)
	}

	return
}

func genAppLevelByScore(info *model.UserAppInfo, m map[string]int64,
	list *model.ARBlackList) (level string, score int64) {
	var ok bool
	level = adaptLevelOfLow
	score, ok = m[info.Model]
	if !ok {
		level = adaptLevelOfMiddleUnknown

		return
	}

	if score >= list.HighScore {
		level = adaptLevelOfHigh
	} else if score >= list.MiddleScore {
		level = adaptLevelOfMiddle
	}

	return
}

func genAppLevelByMemory(list *model.ARBlackList, memory int64) (level string) {
	if len(list.MemoryRuleList) == 0 {
		return
	}

	level = adaptLevelOfLow
	for _, v := range list.MemoryRuleList {
		if memory < v.Threshold {
			break
		}

		level = v.Level
	}

	return
}

func genAppLevelByVersion(info *model.UserAppInfo, list *model.ARBlackList, score int64) (level string) {
	// if it is un_matched device, do not apply version rule
	if score == 0 {
		return
	}

	for _, rule := range list.VersionRuleList {
		switch rule.BizType {
		case model.VersionRuleBizType4First:
			if score <= rule.Score && info.OsVersion > rule.Version {
				level = rule.Level

				break
			}
		}

		if level != "" {
			break
		}
	}

	return
}

func genAppLevelByModelAndMap(info *model.UserAppInfo, m map[string]int64,
	list *model.ARBlackList, memory int64, report *model.ARDeviceReportInfo) (level string) {
	var score int64
	level, score = genAppLevelByScore(info, m, list)
	report.Score = score
	if level != adaptLevelOfLow {
		level4Memory := genAppLevelByMemory(list, memory)
		switch level4Memory {
		case adaptLevelOfLow, adaptLevelOfMiddle:
			level = level4Memory
		}

		if level != adaptLevelOfLow {
			level4Version := genAppLevelByVersion(info, list, score)
			switch level4Version {
			case adaptLevelOfLow, adaptLevelOfMiddle:
				level = level4Version
			}

			switch level {
			case adaptLevelOfLow:
				report.SceneID = model.ARDeviceReportSceneOfVersionRuleLow
			case adaptLevelOfMiddle:
				report.SceneID = model.ARDeviceReportSceneOfVersionRuleMiddle
			case adaptLevelOfHigh:
				report.SceneID = model.ARDeviceReportSceneOfVersionRuleHigh
			case adaptLevelOfMiddleUnknown:
				report.SceneID = model.ARDeviceReportSceneOfUnknownMiddle
				level = adaptLevelOfMiddle
			}
		} else {
			report.SceneID = model.ARDeviceReportSceneOfMemoryLow
		}
	} else {
		report.SceneID = model.ARDeviceReportSceneOfScoreLow
	}

	return
}

func isAppInfoInBlacklist(info *model.UserAppInfo) (in bool) {
	appModel := info.Model
	appOsVersion := strconv.FormatFloat(info.OsVersion, 'E', -1, 64)

	switch info.Os {
	case model.Os4Ios:
		if blackList4Ios.SupportVersion > 0 && info.OsVersion < blackList4Ios.SupportVersion {
			in = true

			return
		}

		_, in = blackList4Ios.ModelMap[appModel]
		if !in {
			_, in = blackList4Ios.VersionMap[appOsVersion]
		}
	case model.Os4Android:
		if blackList4Android.SupportVersion > 0 && info.OsVersion < blackList4Android.SupportVersion {
			in = true

			return
		}

		_, in = blackList4Android.ModelMap[appModel]
		if !in {
			_, in = blackList4Android.VersionMap[appOsVersion]
		}
	}

	return
}

func (s *Service) ReserveStatus(ctx context.Context, mid int64) (m map[string]interface{}, err error) {
	m = make(map[string]interface{}, 0)
	{
		m["reserved"] = false
	}

	if mid == 0 {
		return
	}

	if d, tmpErr := like.CommonDao.ReserveOnly(ctx, BnjReserveInfo.ActivityID, mid); tmpErr == nil && d != nil {
		if d.Mtime > 0 {
			m["reserved"] = true
		}
	}

	return
}

func (s *Service) PublicizeAggregation(ctx context.Context, mid int64) (info *model.PublicizeAggregation, err error) {
	info = new(model.PublicizeAggregation)
	{
		info.AR = new(model.ARInPublicize)
		info.Reserve = new(model.ReserveInPublicize)
		info.Reserve.ActivityID = BnjReserveInfo.ActivityComponentID
		info.Reserve.Total = bnjReserveCount
		info.PCResource = resourceOfPC
	}

	if mid > 0 {
		info.Reserve.IsLogin = 1
		if d, tmpErr := like.CommonDao.ReserveOnly(ctx, BnjReserveInfo.ActivityID, mid); tmpErr == nil && d != nil {
			if d.State == 1 {
				info.Reserve.Reserved = 1
			}
		}
	}

	info.AR.DrawUV = bnjARUV
	var coupon *model.UserCoupon
	coupon, err = dao.FetchUserCoupon(ctx, mid)
	if err == nil {
		info.AR.Coupon = coupon.ND
	}

	return
}
