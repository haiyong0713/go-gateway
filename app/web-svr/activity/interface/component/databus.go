package component

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go-gateway/app/web-svr/activity/tools/lib/initialize"

	"go-common/library/queue/databus"
	databusV2 "go-common/library/queue/databus.v2"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/tool"
)

type RiskInfo struct {
	Scene     string `json:"scene"`
	TraceID   string `json:"trace_id"`
	Timestamp int64  `json:"ts"`
	Extra     string `json:"event_ctx"`
}

const (
	bizNameOfRiskManagementReport = "risk_management_report"

	RiskManagementScene4ARGame = "newyear_game_end"
	RiskManagementScene4Exam   = "newyear_answer"
)

var (
	GaiaRiskProducer         *databus.Databus
	ActAuditMaterialProducer *databus.Databus
	ActGuessProducer         *databus.Databus
	EnablePublishLog         bool
	DatabusV2ActivityClient  databusV2.Client
	DatabusV2JobClient       databusV2.Client
	AsyncReserveProducer     databusV2.Producer
	ActPlatProducer          databusV2.Producer
	UpActReserveProducer     databusV2.Producer
	CardsComposeProducer     databusV2.Producer
	StockServerSyncProducer  *databus.Databus
)

func InitProducer(cfg *conf.Config) (err error) {
	GaiaRiskProducer = initialize.NewDatabusV1(cfg.GaiaRiskPub)
	ActGuessProducer = initialize.NewDatabusV1(cfg.ActGuessPub)
	ActAuditMaterialProducer = initialize.NewDatabusV1(cfg.ManuScriptAuditPub)
	AsyncReserveProducer = initialize.NewProducer(DatabusV2ActivityClient, cfg.AsyncReserveConfig.Topic)
	ActPlatProducer = initialize.NewProducer(DatabusV2ActivityClient, cfg.ActPlatConfig.Topic)
	UpActReserveProducer = initialize.NewProducer(DatabusV2JobClient, cfg.UpActReserveConfig.Topic)
	CardsComposeProducer = initialize.NewProducer(DatabusV2ActivityClient, cfg.CardsComposeConfig.Topic)
	StockServerSyncProducer = initialize.NewDatabusV1(cfg.StockServerSyncPubConfig)
	return
}

func Report2RiskManagement(ctx context.Context, scene, tranceID string, eventCtx interface{}) (err error) {
	if GaiaRiskProducer == nil {
		err = errors.New("GaiaRiskProducer is not initialized")

		return
	}

	var extraStr string
	if d, ok := eventCtx.(string); ok {
		extraStr = d
	} else {
		var bs []byte
		bs, err = json.Marshal(eventCtx)
		if err != nil {
			return
		}

		extraStr = string(bs)
	}

	now := time.Now()
	if tranceID == "" {
		tranceID = fmt.Sprintf("%v_%v", scene, now.UnixNano())
	}

	info := new(RiskInfo)
	{
		info.Scene = scene
		info.TraceID = tranceID
		info.Timestamp = now.Unix()
		info.Extra = extraStr
	}

	if err := GaiaRiskProducer.Send(ctx, info.TraceID, info); err != nil {
		tool.IncrCommonBizStatus(bizNameOfRiskManagementReport, tool.StatusOfFailed)
	}

	return
}

func InitClient() (err error) {
	DatabusV2ActivityClient, err = databusV2.NewClient(
		context.Background(),
		conf.Conf.MainWebSvrActivity.Target,
		databusV2.WithAppID(conf.Conf.MainWebSvrActivity.AppID),
		databusV2.WithToken(conf.Conf.MainWebSvrActivity.Token),
	)
	DatabusV2JobClient, err = databusV2.NewClient(
		context.Background(),
		conf.Conf.MainWebSvrJob.Target,
		databusV2.WithAppID(conf.Conf.MainWebSvrJob.AppID),
		databusV2.WithToken(conf.Conf.MainWebSvrJob.Token),
	)
	return
}
