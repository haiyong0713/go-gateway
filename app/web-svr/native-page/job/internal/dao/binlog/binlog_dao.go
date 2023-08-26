package binlog

import (
	"context"
	"encoding/json"
	"go-common/library/cache/credis"
	"go-common/library/railgun.v2/message"
	"go-common/library/railgun.v2/processor/single"
	"math/rand"
	"time"

	"go-gateway/app/web-svr/native-page/job/internal/model"

	"github.com/pkg/errors"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Config struct {
	Databus                  *databus.Config
	Railgun                  *railgun.SingleConfig
	WhiteListByMidExpire     int64
	WhiteListByMidNullExpire int64
}

type Dao struct {
	cfg   *Config
	redis credis.Redis
}

func NewDao(cfg *Config, r credis.Redis) *Dao {
	return &Dao{
		cfg:   cfg,
		redis: r,
	}
}

type Processor interface {
	ParseMsg(context.Context, json.RawMessage) (interface{}, error)
	HandleInsert(context.Context, json.RawMessage) error
	HandleUpdate(c context.Context, new json.RawMessage, old json.RawMessage) error
	HandleDelete(context.Context, json.RawMessage) error
}

func process(c context.Context, processor Processor, msg *model.BinlogMsg) error {
	if msg == nil {
		return errors.Errorf("msg is empty")
	}
	switch msg.Action {
	case model.ActionInsert:
		return processor.HandleInsert(c, msg.New)
	case model.ActionUpdate:
		return processor.HandleUpdate(c, msg.New, msg.Old)
	case model.ActionDelete:
		return processor.HandleDelete(c, msg.New)
	default:
		log.Errorc(c, "Unexpected action=%+v", msg.Action)
		return errors.Errorf("Unexpected action=%+v", msg.Action)
	}
}

func (d *Dao) DoBinlog(c context.Context, item interface{}, extra *single.Extra) message.Policy {
	binlogMsg := item.(*model.BinlogMsg)
	if binlogMsg == nil {
		return message.Ignore
	}
	if _, ok := model.Tables[binlogMsg.Table]; !ok {
		return message.Ignore
	}
	log.Info("process-binlog, action=%+v table=%+v new=%+v old=%+v extra(%+v)", binlogMsg.Action, binlogMsg.Table, string(binlogMsg.New), string(binlogMsg.Old), extra)
	processor, err := d.buildProcessor(binlogMsg.Table)
	if err != nil {
		return message.Ignore
	}
	_ = process(c, processor, binlogMsg)
	return message.Success
}

func (d *Dao) buildProcessor(table string) (Processor, error) {
	switch table {
	case model.TableWhiteList:
		return &WhiteListProcessor{dao: d}, nil
	case model.TableNatUserSpace:
		return &UserSpaceProcessor{dao: d}, nil
	default:
		log.Error("Unexpected table=%+v", table)
		return nil, errors.Errorf("Unexpected table=%+v", table)
	}
}

func UnpackBinlog(msg message.Message) (*single.UnpackMessage, error) {
	binlogMsg := &model.BinlogMsg{}
	if err := json.Unmarshal(msg.Payload(), binlogMsg); err != nil {
		log.Error("Fail to unmarshal binlog msg, msg=%+v error=%+v", string(msg.Payload()), err)
		return nil, err
	}
	return &single.UnpackMessage{
		Group: int64(rand.Intn(256)),
		Item:  binlogMsg,
	}, nil
}
