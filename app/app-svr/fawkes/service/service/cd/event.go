package cd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"go-common/library/ecode"

	databusV2 "go-common/library/queue/databus.v2"
	"go-common/library/sync/errgroup.v2"
	"k8s.io/apimachinery/pkg/util/sets"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"go-gateway/app/app-svr/fawkes/service/conf"
)

const (
	// SteadyPackAutoGenChannelPack 标记稳定版本时自动构建渠道包事件
	SteadyPackAutoGenChannelPack = "inner.steady.state.auto.gen.channel.pack"

	// PackGenerateUpdateEvent 渠道包上传后自动上传CDN事件
	PackGenerateUpdateEvent = "inner.pack.generate.update"

	// PackGreyPushEvent 推送构建包灰度信息事件
	PackGreyPushEvent = "inner.pack.grey.push"
)

type PackGenerateArgs struct {
	Ctx           context.Context
	AppKey        string
	ChannelStatus int
	ChannelPacks  []*cdmdl.ChannelFileInfo
}

type AutoGenChannelPackArgs struct {
	Ctx      context.Context
	Operator string
	AppKey   string
	BuildId  int64
}

type Groups []*appmdl.GroupChannels

type GenFunc func(c context.Context, appKey string, channels string, userName string, buildID int64) (err error)

func (g Groups) Len() int { return len(g) }
func (g Groups) Less(i, j int) bool {
	var iv, jv int8
	if g[i] != nil {
		iv = g[i].Group.Priority
	}
	if g[j] != nil {
		jv = g[j].Group.Priority
	}
	return iv > jv
}
func (g Groups) Swap(i, j int) { g[i], g[j] = g[j], g[i] }

const _chunk = 50

// packGenerateUpdateAction 渠道包自动上传CDN
func (s *Service) packGenerateUpdateAction(args PackGenerateArgs) (err error) {
	ctx := args.Ctx
	whiteSet := sets.NewString(conf.Conf.Switch.PackAutoUploadCDN.WhiteList...)
	if !whiteSet.Has(args.AppKey) {
		log.Warnc(ctx, "PackGenerateUpdateAction conf.Conf.Switch.Upload.WhiteList has no appKey[%s], switch off", args.AppKey)
		return
	}
	channelFiles, _ := json.Marshal(args.ChannelPacks)
	log.Warnc(ctx, "PackGenerateUpdateAction appKey: %s, status: %d, fileInfo: %s", args.AppKey, args.ChannelStatus, channelFiles)
	if cdmdl.GenerateSuccess != args.ChannelStatus {
		log.Warnc(ctx, "PackGenerateUpdateAction appKey[%s] PackGenerateUpdate status[%d] 不需要上传文件", args.AppKey, args.ChannelStatus)
		return
	}
	var (
		packGenerateIds     []int64
		chList              []*appmdl.Channel
		generateInfoMap     map[int64]*cdmdl.Generate
		channelMap          = map[int64]*appmdl.Channel{}
		generateSuccessList []*cdmdl.Generate
	)
	for _, v := range args.ChannelPacks {
		packGenerateIds = append(packGenerateIds, v.ID)
	}
	if generateInfoMap, err = s.fkDao.GenerateListByIds(context.Background(), packGenerateIds); err != nil {
		log.Errorc(ctx, "PackGenerateUpdateAction packGenerateIds[%d] GenerateListByIds error: %#v", packGenerateIds, err)
		return
	}
	if chList, err = s.fkDao.AppChannelList(context.Background(), args.AppKey, "", "", "", -1, -1, -1); err != nil {
		log.Errorc(ctx, "PackGenerateUpdateAction app[%s] get channel list error: %#v", args.AppKey, err)
		return
	}
	for _, channel := range chList {
		channelMap[channel.ID] = channel
	}
	for _, v := range generateInfoMap {
		if info, ok := channelMap[v.ChannelID]; ok {
			if info.Group != nil && v.Status == cdmdl.GenerateSuccess && info.Group.AutoPushCdn == 1 {
				generateSuccessList = append(generateSuccessList, v)
			}
		}
	}
	for _, v := range generateSuccessList {
		appKey, id := v.AppKey, v.ID
		if err = s.channelPackUploadCdnWorker.SyncDo(ctx, func(ctx context.Context) {
			if err = s.AppCDGenerateUpload(context.Background(), appKey, model.SystemAuto, id); err != nil {
				log.Errorc(ctx, "PackGenerateUpdateAction pack_generate id[%d] upload error: %#v", v.ID, err)
			}
		}); err != nil {
			err = ecode.Errorf(ecode.ServerErr, "uploadCdnWorker Do error:%v", err)
			log.Errorc(ctx, err.Error())
			return
		}
	}
	return
}

// autoGenChannelPack 自动生成渠道包
func (s *Service) autoGenChannelPack(args AutoGenChannelPackArgs) (err error) {
	var (
		pack                 *cdmdl.Pack
		autoGenGroupChannels []*appmdl.Channel
		groupMap             map[int64]*appmdl.ChannelGroupInfo
		app                  *appmdl.APP
		groupChannels        []*appmdl.GroupChannels
	)
	var (
		ctx     = args.Ctx
		appKey  = args.AppKey
		buildId = args.BuildId
		op      = args.Operator
	)
	log.Warnc(ctx, "%s", "自动构建流程开始")
	if app, err = s.fkDao.AppPass(ctx, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if pack, err = s.fkDao.PackByBuild(ctx, appKey, "prod", buildId); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if pack.SteadyState != 1 {
		log.Errorc(ctx, "appKey:%s, buildId: %d 非稳定版本", appKey, buildId)
		return
	}
	// 找到所有需要自动生成的渠道
	if autoGenGroupChannels, groupMap, err = s.getAutoGenGroupChannels(ctx, appKey); err != nil {
		log.Errorc(ctx, "appKey:%s,获取自动构建渠道失败 error:%v", appKey, err)
		return
	}
	// 按照渠道组分组 并按照优先级排序
	groupChannels = groupByChannelGroup(autoGenGroupChannels, groupMap)
	for _, g := range groupChannels {
		// 分渠道执行
		if err = s.TriggerGitPipeline(ctx, appKey, g, op, buildId, s.AppCDGenerateAddGit); err != nil {
			log.Errorc(ctx, "渠道组%s 提交构建任务，存在部分任务失败， err：%v", g.Group.Name, err)
			continue
		}
		// 一个渠道全部完成请求后， 通知监控进程检查打包完成情况
		g := g
		go func() {
			deadline, cancelFunc := context.WithDeadline(ctx, time.Now().Add(2*time.Hour))
			defer func() {
				cancelFunc()
				if err := recover(); err != nil {
					log.Errorc(ctx, "%v", err)
					return
				}
			}()
			err = s.MonitorGenerateState(deadline, app, g, buildId)
			if err != nil {
				log.Errorc(ctx, "%v", err)
				return
			}
		}()
	}
	return
}

// MonitorGenerateState 监控生成状态
func (s *Service) MonitorGenerateState(ctx context.Context, app *appmdl.APP, g *appmdl.GroupChannels, buildId int64) (err error) {
	var buildSuccess, uploadSuccess bool
	log.Infoc(ctx, "开始监听渠道%s的构建状态", g.Group.Name)
	s.channelPackStepNotify(ctx, app, buildId, g.Group, cdmdl.GenerateRunning)
	// 构建阶段
	if buildSuccess, err = s.stepFinished(ctx, app, buildId, g, cdmdl.GenerateSuccess); err != nil {
		log.Errorc(ctx, "自动构建渠道包失败，app:%s, buildId:%d err:%v", app.AppKey, buildId, err)
		return
	}
	if buildSuccess {
		s.channelPackStepNotify(ctx, app, buildId, g.Group, cdmdl.GenerateSuccess)
	} else {
		log.Errorc(ctx, "自动构建渠道包失败，err: %v", err)
	}
	if g.Group.AutoPushCdn == 1 {
		// 上传阶段
		log.Infoc(ctx, "开始监听渠道%s的自动上传CDN状态", g.Group.Name)
		if uploadSuccess, err = s.stepFinished(ctx, app, buildId, g, cdmdl.GenerateUpload); err != nil {
			log.Warnc(ctx, "自动构建渠道包失败，%v", err)
			return
		}
		if uploadSuccess {
			s.channelPackStepNotify(ctx, app, buildId, g.Group, cdmdl.GenerateUpload)
		} else {
			log.Errorc(ctx, "自动构建渠道包失败，upload CDN error, err: %v", err)
		}
	}
	return
}

func (s *Service) TriggerGitPipeline(ctx context.Context, appKey string, g *appmdl.GroupChannels, op string, buildId int64, genChannelPackHandler GenFunc) (err error) {
	eg := errgroup.WithContext(ctx)
	rand.Seed(time.Now().UnixNano())
	for _, c := range splitChannels(g.Channels, _chunk) {
		c := c
		eg.Go(func(ctx context.Context) error {
			// 随机睡眠一段时间 错开对git的请求 避免git超时
			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			channelBytes, _ := json.Marshal(genChannelArg(c))
			err := genChannelPackHandler(ctx, appKey, string(channelBytes), op, buildId)
			if err != nil {
				return err
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	return
}

// 根据status判断该步骤是否结束 未结束则堵塞直至超时
func (s *Service) stepFinished(ctx context.Context, app *appmdl.APP, glJobId int64, g *appmdl.GroupChannels, status int8) (success bool, err error) {
	var (
		generates []*cdmdl.Generate
		finished  = false
	)
loop:
	for {
		select {
		case <-time.After(15 * time.Second):
			log.Warnc(ctx, "开始轮询, %v, %v", app.AppKey, glJobId)
			if generates, err = s.fkDao.GenerateList(ctx, app.AppKey, glJobId); err != nil {
				log.Errorc(ctx, "%v", err)
				finished = false
				break loop
			}
			if len(generates) == 0 {
				log.Warnc(ctx, "%s", "没有找到构建中的渠道包")
				continue loop
			}
			for _, v := range generates {
				// 该分组下存在未构建完成的渠道包
				if _, ok := g.ChannelMap[v.ChannelID]; ok {
					if v.Status < status {
						log.Warnc(ctx, "%s", "还未完成")
						continue loop
					}
				}
			}
			finished = true
			log.Warnc(ctx, "%s", "任务完成")
			break loop
		case <-ctx.Done():
			log.Warnc(ctx, "%s", "任务超时退出")
			err = ecode.Error(ecode.Deadline, "任务超时退出")
			finished = false
			break loop
		}
	}
	return finished, err
}

func groupByChannelGroup(channels []*appmdl.Channel, groupMap map[int64]*appmdl.ChannelGroupInfo) (g Groups) {
	for groupId, group := range groupMap {
		groupC := &appmdl.GroupChannels{}
		groupC.ChannelMap = make(map[int64]*appmdl.Channel)
		groupC.Group = group
		for _, c := range channels {
			if c.Group.ID == groupId {
				groupC.Channels = append(groupC.Channels, c)
				groupC.ChannelMap[c.ID] = c
			}
		}
		g = append(g, groupC)
	}
	sort.Sort(g)
	return
}

// 获取app下所有需要自动构建的渠道
func (s *Service) getAutoGenGroupChannels(ctx context.Context, appKey string) (autoGenGroupChannel []*appmdl.Channel, groupMap map[int64]*appmdl.ChannelGroupInfo, err error) {
	autoGenGroupChannel = []*appmdl.Channel{}
	groupMap = make(map[int64]*appmdl.ChannelGroupInfo)
	// app下所有渠道
	appChannelList, err := s.fkDao.AppChannelList(ctx, appKey, "", "", "", -1, -1, -1)
	if err != nil {
		return
	}
	for _, ac := range appChannelList {
		if ac.Group != nil && ac.Group.IsAutoGen == 1 {
			autoGenGroupChannel = append(autoGenGroupChannel, ac)
			if _, ok := groupMap[ac.Group.ID]; !ok {
				groupMap[ac.Group.ID] = ac.Group
			}
		}
	}
	if len(autoGenGroupChannel) == 0 {
		log.Infoc(ctx, "没有需要自动构建的渠道组")
		return
	}
	return
}

func (s *Service) channelPackStepNotify(ctx context.Context, appInfo *appmdl.APP, buildId int64, group *appmdl.ChannelGroupInfo, step int8) {
	title := "渠道包构建通知"
	content := fmt.Sprintf("%s(%s)\t构建号:%d\t渠道组:%s\t\n", appInfo.Name, appInfo.AppKey, buildId, group.Name)
	switch step {
	case cdmdl.GenerateRunning:
		content = content + "开始自动构建。。。"
	case cdmdl.GenerateSuccess:
		content = content + "构建完成。。。"
	case cdmdl.GenerateUpload:
		content = content + "上传CDN完成。。。"
	}
	log.Infoc(ctx, "%s", content)
	link := fmt.Sprintf("%s/#/cd/generatelist?app_key=%s&env=prod&build_id=%d", conf.Conf.Host.Fawkes, appInfo.AppKey, buildId)
	receiverArr := make([]string, 0)
	receiverArr = append(receiverArr, strings.Split(group.QaOwner, ",")...)
	receiverArr = append(receiverArr, strings.Split(group.MarketOwner, ",")...)
	receiverArr = append(receiverArr, conf.Conf.AlarmReceiver.ChannelPackAutoBuildReceiver...)
	receiver := strings.Join(receiverArr, "|")
	wechatBot := conf.Conf.Comet.FawkesAppID
	err := s.fkDao.WechatCardMessageNotify(title, content, link, "", receiver, wechatBot)
	if err != nil {
		return
	}
}

func splitChannels(channels []*appmdl.Channel, num int64) [][]*appmdl.Channel {
	max := int64(len(channels))
	//判断数组大小是否小于等于指定分割大小的值，是则把原数组放入二维数组返回
	if max <= num {
		return [][]*appmdl.Channel{channels}
	}
	//获取应该数组分割为多少份
	var quantity int64
	if max%num == 0 {
		quantity = max / num
	} else {
		quantity = (max / num) + 1
	}
	var segments = make([][]*appmdl.Channel, 0)
	//声明分割数组的截止下标
	var start, end, i int64
	for i = 1; i <= quantity; i++ {
		end = i * num
		if i != quantity {
			segments = append(segments, channels[start:end])
		} else {
			segments = append(segments, channels[start:])
		}
		start = i * num
	}
	return segments
}

func genChannelArg(v []*appmdl.Channel) (channelList []*cdmdl.ChannelGeneParam) {
	channelList = []*cdmdl.ChannelGeneParam{}
	for _, v := range v {
		channelList = append(channelList, &cdmdl.ChannelGeneParam{Channel: v.Code, ChannelID: v.ID})
	}
	return
}

type PackGreyArgs struct {
	Context   context.Context
	AppKey    string
	Env       string
	VersionId int64
	Operator  string
}

// AddPackGreyHistory pack灰度信息记录&&推送databus
func (s *Service) AddPackGreyHistory(args PackGreyArgs) (err error) {
	log.Infoc(args.Context, "AddPackGreyHistory %+v", args)
	if args.Env != "prod" {
		return
	}
	var app *appmdl.APP
	if app, err = s.fkDao.AppPass(args.Context, args.AppKey); err != nil {
		log.Errorc(args.Context, "%v", err)
		return
	}
	var (
		buildPacks  []*cdmdl.Pack
		version     *model.Version
		glJobIds    []int64
		lastHistory map[int64]*cdmdl.PackGreyHistory
		flowsCfg    map[int64]*cdmdl.FlowConfig
		filterCfg   map[int64]*cdmdl.FilterConfig
	)
	if version, err = s.fkDao.PackVersionByID(args.Context, args.AppKey, args.VersionId); err != nil {
		log.Errorc(args.Context, "PackVersionByID %v", err)
		return
	}
	if buildPacks, err = s.fkDao.PackByVersion(args.Context, args.AppKey, args.Env, args.VersionId); err != nil {
		log.Errorc(args.Context, "PackByVersion %v", err)
		return
	}
	if version == nil || buildPacks == nil {
		return
	}
	if app.Platform == "ios" && buildPacks[0].PackType != int8(cimdl.Publish) {
		log.Infoc(args.Context, "ios pkg type %v", buildPacks[0].PackType)
		return
	}
	for _, pack := range buildPacks {
		glJobIds = append(glJobIds, pack.BuildID)
	}
	if flowsCfg, err = s.fkDao.PackFlowConfig(args.Context, args.AppKey, args.Env, glJobIds); err != nil {
		log.Errorc(args.Context, "PackFlowConfig %v", err)
		return
	}
	// 得到每个构件包的最新历史灰度信息
	if lastHistory, err = s.fkDao.LastPackGreyHistory(args.Context, args.AppKey, glJobIds); err != nil {
		log.Errorc(args.Context, "LastPackGreyHistory %v", err)
		return
	}
	if filterCfg, err = s.fkDao.PackFilterConfig(args.Context, args.AppKey, args.Env, glJobIds); err != nil {
		log.Errorc(args.Context, "PackFilterConfig %v", err)
		return
	}
	var packGreyDataList []*cdmdl.PackGreyData
	if app.Platform == "ios" {
		log.Warnc(args.Context, "ios grey pack record start")
		nilTime := time.Time{}
		nowTime := time.Now()
		for index, glJobId := range glJobIds {
			startTime := time.Unix(buildPacks[index].CTime, 0)
			finishTime := startTime.AddDate(0, 0, 3)
			packGreyData := &cdmdl.PackGreyData{
				AppKey:          args.AppKey,
				DatacenterAppId: app.DataCenterAppID,
				Platform:        app.Platform,
				MobiApp:         app.MobiApp,
				Version:         version.Version,
				VersionCode:     version.VersionCode,
				GlJobID:         glJobId,
				Config:          filterCfg[glJobId],
				GreyStartTime:   startTime,
				GreyFinishTime:  finishTime,
				GreyCloseTime:   nilTime,
				OperateTime:     nowTime,
			}
			packGreyDataList = append(packGreyDataList, packGreyData)
		}
		log.Warnc(args.Context, "ios grey pack record end")
	} else if app.Platform == "android" {
		log.Warnc(args.Context, "android grey pack record start")
		for _, glJobId := range glJobIds {
			// 没有流量配置
			if flowsCfg[glJobId] == nil || flowsCfg[glJobId].Flow == cdmdl.PackFlowZero {
				continue
			}
			// 流量配置无变化
			if lastHistory[glJobId] != nil && flowsCfg[glJobId].Flow == lastHistory[glJobId].Flow && lastHistory[glJobId].IsUpgrade == version.IsUpgrade {
				continue
			}
			// 得到灰度数据
			packGreyData := s.GeneratePackGreyData(args.AppKey, app.Platform, app.MobiApp, flowsCfg[glJobId].Flow, version.Version, version.IsUpgrade, app.DataCenterAppID, glJobId, version.VersionCode, time.Unix(flowsCfg[glJobId].MTime, 0), time.Unix(version.MTime, 0), lastHistory[glJobId], filterCfg[glJobId])
			if packGreyData != nil {
				packGreyDataList = append(packGreyDataList, packGreyData)
			}
		}
		log.Warnc(args.Context, "android grey pack record end")
	}
	log.Infoc(args.Context, "AddPackGreyHistory packGreyDataList len %d", len(packGreyDataList))
	if len(packGreyDataList) > 0 {
		// insert db
		if err = s.fkDao.AddPackGreyHistory(args.Context, args.Operator, packGreyDataList); err != nil {
			log.Errorc(args.Context, "AddPackGreyHistory %v", err)
			return
		}
		//push databus
		if err = s.PubPackGreyData(args.Context, packGreyDataList); err != nil {
			log.Errorc(args.Context, "PubPackGreyData %v", err)
			return
		}
		log.Warnc(args.Context, "grey pack push end")
	}
	return
}

func (s *Service) GeneratePackGreyData(appKey, platform, mobiApp, flow, version string, isUpgrade int8, DatacenterAppId, glJobId, versionCode int64, flowTime, switchUpgradeTime time.Time, lastGreyHistory *cdmdl.PackGreyHistory, filterCfg *cdmdl.FilterConfig) (res *cdmdl.PackGreyData) {
	var (
		startTime, finishTime, closeTime time.Time
		nilTime                          = time.Time{}
		hasGreyData                      bool
	)
	// switchUpgradeTime 决定 startTime 和 closeTime, flowTime决定 startTime 和 finishTime
	if switchUpgradeTime.Unix() > flowTime.Unix() {
		flowTime = switchUpgradeTime
	}
	// 生效开关
	if isUpgrade == cdmdl.Upgrade {
		closeTime = nilTime
		// 构建包是否记录过
		if lastGreyHistory == nil || lastGreyHistory.IsUpgrade == cdmdl.NotUpgrade {
			startTime = switchUpgradeTime
		} else {
			// 已存在灰度信息是否已全量
			if lastGreyHistory.Flow == cdmdl.PackFlowFull {
				startTime = flowTime
			} else {
				startTime = lastGreyHistory.GreyStartTime
			}
		}
		// 灰度是否结束
		if flow != cdmdl.PackFlowFull {
			finishTime = nilTime
		} else {
			finishTime = flowTime
		}
		hasGreyData = true
	} else {
		// 灰度关闭   区分 测试环境->正式环境 || 只配置流量
		if lastGreyHistory != nil && lastGreyHistory.GreyCloseTime == nilTime {
			startTime = lastGreyHistory.GreyStartTime
			finishTime = lastGreyHistory.GreyFinishTime
			closeTime = switchUpgradeTime
			hasGreyData = true
		}
	}
	if hasGreyData {
		res = &cdmdl.PackGreyData{
			AppKey:          appKey,
			DatacenterAppId: DatacenterAppId,
			Platform:        platform,
			MobiApp:         mobiApp,
			Version:         version,
			VersionCode:     versionCode,
			GlJobID:         glJobId,
			IsUpgrade:       isUpgrade,
			Flow:            flow,
			Config:          filterCfg,
			GreyStartTime:   startTime,
			GreyFinishTime:  finishTime,
			GreyCloseTime:   closeTime,
			OperateTime:     time.Now(),
		}
	}
	return
}

//nolint:bilirailguncheck
func (s *Service) PubPackGreyData(c context.Context, packGreyData []*cdmdl.PackGreyData) (err error) {
	var packGreyDataProducer databusV2.Producer
	pubKey := fmt.Sprintf("pack-grey-data-%d", time.Now().Unix())
	bs, _ := json.Marshal(packGreyData)
	packGreyDataProducer = s.fkDao.NewProducer(c, s.c.Databus.Topics.PackGreyDataPub.Group, s.c.Databus.Topics.PackGreyDataPub.Name)
	err = packGreyDataProducer.Send(c, pubKey, bs)
	return
}
