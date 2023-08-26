package newyear2021

import (
	"context"
	"fmt"
	innerLog "log"
	"os"
	"sync/atomic"
	"time"

	"go-common/library/log"
	http "go-common/library/net/http/blademaster"

	"github.com/Shopify/sarama"
	"github.com/fsnotify/fsnotify"

	"go-gateway/app/web-svr/activity/interface/conf"
	dao "go-gateway/app/web-svr/activity/interface/dao/newyear2021"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
	"go-gateway/app/web-svr/activity/interface/tool"

	databusv1 "go-common/library/queue/databus"
	"go-common/library/queue/databus.v2"
)

const (
	filenameOfAndroidScore  = "%v/android.json"
	filenameOfIosScore      = "%v/ios.json"
	filenameOfBlacklist     = "%v/black_list.json"
	filenameOfExamBank      = "%v/exam_bank.json"
	filenameOfAR            = "%v/ar.json"
	filenameOfActivityID    = "%v/quota_activity_id.json"
	filenameOfBnjActivityID = "%v/bnj_activity_id.json"
	filenameOfBnjStrategy   = "%v/bnj_strategy.json"
)

var (
	ARRewardClient    databus.Client
	LiveLotteryClient databus.Client

	ARRewardProducer    databus.Producer
	LiveLotteryProducer databus.Producer
	ExamProducer        sarama.SyncProducer
	ARDeviceProducer    sarama.SyncProducer

	BnjARConfig *model.ARConfig
	// 需要走拜年纪2021特殊处理的抽奖次数map
	QuotaActivityIDMap  map[string]int64
	BnjReserveInfo      *model.ReserveInPublicize
	BnjARDeviceProducer *model.ExamProducer
	BnjExamProducer     *model.ExamProducer
	BnjStrategyInfo     *model.BnjStrategy
)

// Service ...
type Service struct {
	c              *conf.Config
	dao            *dao.Dao
	config         *atomic.Value
	httpClient     *http.Client
	actPlatDatabus *databusv1.Databus
}

func init() {
	BnjStrategyInfo = new(model.BnjStrategy)
	BnjARConfig = new(model.ARConfig)
	QuotaActivityIDMap = make(map[string]int64, 0)
	BnjReserveInfo = new(model.ReserveInPublicize)
	BnjARDeviceProducer = new(model.ExamProducer)
	{
		BnjARDeviceProducer.Addresses = make([]string, 0)
	}
	BnjExamProducer = new(model.ExamProducer)
	{
		BnjExamProducer.Addresses = make([]string, 0)
	}
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:              c,
		dao:            dao.New(c),
		config:         &atomic.Value{},
		httpClient:     http.NewClient(c.HTTPClient),
		actPlatDatabus: databusv1.New(c.DataBus.ActPlatPub),
	}
	ctx := context.Background()
	_, sc, err := s.dao.GetLatestConf(ctx)
	if err != nil {
		log.Errorc(ctx, "New Bnj2021Service error: get GetLatestConf fail: %v", err)
	}
	if tmpErr := s.InitProducer(c); tmpErr != nil {
		log.Errorc(ctx, "init new year 2021 producer failed, err: %v", tmpErr)
		panic(tmpErr)
	}

	s.config.Store(sc)
	go s.updateConfLoop()

	_ = UpdateScore2CouponRelations(context.Background())
	go UpdateCurrentDateStr(ctx)
	go ASyncResetScore2CouponRelations(ctx)

	// watch 拜年纪配置变更
	updateExamBankByFilename(genWatchedFilename(filenameOfExamBank))
	updateDeviceScoreMapByFilename(genWatchedFilename(filenameOfAndroidScore))
	updateDeviceScoreMapByFilename(genWatchedFilename(filenameOfIosScore))
	updateBlackListByFilename(genWatchedFilename(filenameOfBlacklist))
	updateARConfigurationByFilename(genWatchedFilename(filenameOfAR))
	updateQuotaActivityIDByFilename(genWatchedFilename(filenameOfActivityID))
	updateBnjActivityIDByFilename(genWatchedFilename(filenameOfBnjActivityID))
	updateBnjStrategyByFilename(genWatchedFilename(filenameOfBnjStrategy))
	batchRegisterFileWatch()

	go s.ASyncUpdateBnjReserveCount()
	go AsyncResetWebViewData4PC()

	return s
}

func genWatchedFilename(old string) (new string) {
	if dir := os.Getenv("CONF_PATH"); dir != "" {
		new = fmt.Sprintf(old, dir)
	}

	return
}

func batchRegisterFileWatch() {
	list := []string{
		filenameOfAR,
		filenameOfAndroidScore,
		filenameOfIosScore,
		filenameOfExamBank,
		filenameOfBlacklist,
		filenameOfActivityID,
		filenameOfBnjActivityID,
		filenameOfBnjStrategy,
	}
	for _, v := range list {
		if filename := genWatchedFilename(v); filename != "" {
			if err := tool.RegisterWatchHandlerV1(filename, UpdateConfigurationByFilename); err != nil {
				log.Error("RegisterWatchHandlerV1 failed, , filename: %v, err: %v", v, err)
			}
		}
	}
}

func UpdateConfigurationByFilename(ctx context.Context, event fsnotify.Event) {
	if event.Op.String() != fsnotify.Write.String() {
		return
	}

	switch event.Name {
	case genWatchedFilename(filenameOfAndroidScore), genWatchedFilename(filenameOfIosScore):
		updateDeviceScoreMapByFilename(event.Name)
	case genWatchedFilename(filenameOfExamBank):
		updateExamBankByFilename(event.Name)
	case genWatchedFilename(filenameOfBlacklist):
		updateBlackListByFilename(event.Name)
	case genWatchedFilename(filenameOfAR):
		updateARConfigurationByFilename(event.Name)
	case genWatchedFilename(filenameOfActivityID):
		updateQuotaActivityIDByFilename(event.Name)
	case genWatchedFilename(filenameOfBnjActivityID):
		updateBnjActivityIDByFilename(event.Name)
	case genWatchedFilename(filenameOfBnjStrategy):
		updateBnjStrategyByFilename(event.Name)
	}
}

func (s *Service) InitProducer(c *conf.Config) (err error) {
	sarama.Logger = innerLog.New(os.Stdout, "[sarama] ", innerLog.LstdFlags)
	if len(c.ExamProducer.Addresses) > 0 {
		BnjExamProducer = c.ExamProducer.DeepCopy()
		tmpCfg := sarama.NewConfig()
		{
			tmpCfg.Net.WriteTimeout = 250 * time.Millisecond
			tmpCfg.Producer.Return.Errors = true
			tmpCfg.Producer.Return.Successes = true
			tmpCfg.Producer.Retry.Max = 0
			tmpCfg.Producer.Timeout = 100 * time.Millisecond
		}
		ExamProducer, err = sarama.NewSyncProducer(BnjExamProducer.Addresses, tmpCfg)
		if err != nil {
			panic(err)
		}
	}

	if len(c.ARDeviceProducer.Addresses) > 0 {
		BnjARDeviceProducer = c.ARDeviceProducer.DeepCopy()
		tmpCfg := sarama.NewConfig()
		{
			tmpCfg.Net.WriteTimeout = 150 * time.Millisecond
			tmpCfg.Producer.Return.Errors = true
			tmpCfg.Producer.Return.Successes = true
			tmpCfg.Producer.Retry.Max = 0
			tmpCfg.Producer.Timeout = 100 * time.Millisecond
		}
		ARDeviceProducer, err = sarama.NewSyncProducer(BnjARDeviceProducer.Addresses, tmpCfg)
		if err != nil {
			panic(err)
		}
	}

	ARRewardClient, err = databus.NewClient(
		context.Background(),
		c.Bnj2021ARPub.Target,
		databus.WithAppID(c.Bnj2021ARPub.AppID),
		databus.WithToken(c.Bnj2021ARPub.Token))
	if err != nil {
		return
	}

	ARRewardProducer, err = ARRewardClient.NewProducer(c.Bnj2021ARPub.Topic)
	if err != nil {
		return
	}

	LiveLotteryClient, err = databus.NewClient(
		context.Background(),
		c.Bnj2021LiveLotteryRec.Target,
		databus.WithAppID(c.Bnj2021LiveLotteryRec.AppID),
		databus.WithToken(c.Bnj2021LiveLotteryRec.Token))
	if err != nil {
		return
	}

	LiveLotteryProducer, err = LiveLotteryClient.NewProducer(c.Bnj2021LiveLotteryRec.Topic)
	if err != nil {
		return
	}

	return
}

func (s *Service) GetConf() *model.Config {
	return s.config.Load().(*model.Config)
}

func (s *Service) updateConfLoop() {
	ctx := context.Background()
	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		_, c, err := s.dao.GetLatestConf(ctx)
		if err != nil {
			continue
		}
		s.config.Store(c)
	}
}

func (s *Service) GetConfFromDB(ctx context.Context) (int64, *model.Config, error) {
	return s.dao.GetLatestConf(ctx)
}

func (s *Service) UpdateConfInDB(ctx context.Context, config *model.Config) error {
	return s.dao.UpdateConf(ctx, config)
}

func (s *Service) DeleteConfInDB(ctx context.Context, version int64) error {
	return s.dao.DeleteConf(ctx, version)
}
