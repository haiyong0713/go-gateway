package history

import (
	"context"
	"encoding/json"

	"go-common/library/queue/databus.v2"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
)

func NewProducer(ctx context.Context, conf *conf.DataBusV2Conf) (databus.Client, databus.Producer, error) {
	client, err := databus.NewClient(ctx,
		"discovery://default/infra.databus.v2",
		databus.WithAppID(conf.AppId),
		databus.WithToken(conf.Token))
	if err != nil {
		return nil, nil, err
	}
	producer, err := client.NewProducer(conf.Topic)
	if err != nil {
		return nil, nil, err
	}
	return client, producer, nil
}

func (d *Dao) ReportToAI(c context.Context, req *fm_v2.HistoryReportFm) error { //nolint:bilirailguncheck
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = d.fmReportProducer.Send(c, req.Buvid, bytes)
	if err != nil {
		return err
	}
	return nil
}
