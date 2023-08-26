package topic

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/topic/service"
)

type Dao struct {
	topicClient api.TopicClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.topicClient, err = api.NewClient(c.TopicClient); err != nil {
		panic(fmt.Sprintf("topic NewClient not found err(%v)", err))
	}
	return
}

func (d *Dao) GetBatchResTopicByType(c context.Context, req *api.BatchResTopicByTypeReq) (map[int64]*api.TopicIconInfo, error) {
	res, err := d.topicClient.BatchResTopicByType(c, req)
	if err != nil {
		return nil, err
	}
	return res.GetResTopics(), nil
}
