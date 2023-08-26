package answer

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	api "git.bilibili.co/bapis/bapis-go/community/interface/answer"

	"github.com/pkg/errors"
)

// Dao is answer dao.
type Dao struct {
	answerClient api.AnswerClient
}

// New answerClient.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.answerClient, err = api.NewClient(c.AnswerGRPC); err != nil {
		panic(fmt.Sprintf("answerClient NewClient error(%v)", err))
	}
	return
}

// AnswerStatus get AnswerStatus
func (d *Dao) AnswerStatus(c context.Context, mid int64, mobiApp, source string) (res *api.AnswerStatus, err error) {
	var (
		req   = &api.StatusReq{Mid: mid, MobiApp: mobiApp, Source: source}
		reply *api.StatusReply
	)
	if reply, err = d.answerClient.Status(c, req); err != nil {
		err = errors.Wrapf(err, "%v", req)
		return
	}
	if reply != nil {
		res = reply.Status
	}
	return
}

func (d *Dao) SeniorGate(ctx context.Context, mid, build int64, mobiApp, device string) (*api.SeniorGateResp, error) {
	return d.answerClient.SeniorGate(ctx, &api.SeniorGateReq{Mid: mid, Build: build, MobiApp: mobiApp, Device: device})
}
