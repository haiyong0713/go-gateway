package component

import (
	"go-common/library/database/elastic"
	"go-gateway/app/web-svr/activity/interface/conf"
)

var (
	EsClient *elastic.Elastic
)

func initES(conf *conf.Config) {
	EsClient = elastic.NewElastic(conf.Elastic)
}
