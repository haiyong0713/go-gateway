package reply

import (
	"context"
	"github.com/google/martian/log"
	"go-gateway/app/web-svr/native-page/interface/conf"

	replygrpc "git.bilibili.co/bapis/bapis-go/community/interface/reply"
)

type Dao struct {
	replyClient replygrpc.ReplyInterfaceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.replyClient, err = replygrpc.NewClient(c.ReplyClient); err != nil {
		panic(err)
	}
	return
}

// UpdateActivityState .
func (d *Dao) UpdateActivityState(c context.Context, pid, state int64) error {
	if _, err := d.replyClient.UpdateActivityState(c, &replygrpc.UpdateActivityStateReq{ActivityId: pid, State: state}); err != nil {
		log.Errorf("d.replyClient.UpdateActivityState(%d,%d),error(%v)", pid, state, err)
		return err
	}
	return nil
}
