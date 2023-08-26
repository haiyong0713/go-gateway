package creative_spark

import (
	"context"
	"fmt"
	"go-gateway/app/app-svr/app-view/interface/conf"

	"github.com/pkg/errors"

	api "git.bilibili.co/bapis/bapis-go/creative/spark/service"
)

type Dao struct {
	creativeSpark api.CreativeSparkClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.creativeSpark, err = api.NewClientCreativeSpark(c.CreativeSparkClient); err != nil {
		panic(fmt.Sprintf("creativeSpark NewClient not found err(%v)", err))
	}
	return
}

func (d *Dao) GetInspirationTopics(ctx context.Context) ([]*api.InspirationTopic, error) {
	res, err := d.creativeSpark.InspirationTopics(ctx, &api.Empty{})
	if err != nil {
		return nil, errors.Wrapf(err, " d.creativeSpark.InspirationTopics is err %+v", err)
	}
	return res.List, nil
}
