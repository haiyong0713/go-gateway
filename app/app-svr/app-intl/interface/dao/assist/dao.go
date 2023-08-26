package assist

import (
	"context"
	"fmt"

	assistApi "git.bilibili.co/bapis/bapis-go/assist/service"
	"go-gateway/app/app-svr/app-intl/interface/conf"

	"github.com/pkg/errors"
)

// Dao is assist dao
type Dao struct {
	assistGRPC assistApi.AssistClient
}

// New initial assist dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.assistGRPC, err = assistApi.NewClient(c.AssistClient)
	if err != nil {
		panic(fmt.Sprintf("assist NewClient error(%v)", err))
	}
	return
}

// Assist get assists data from api.
func (d *Dao) Assist(c context.Context, upMid int64) (asss []int64, err error) {
	var (
		arg     = &assistApi.AssistIDsReq{Mid: upMid}
		assists *assistApi.AssistIDsReply
	)
	if assists, err = d.assistGRPC.AssistIDs(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if assists != nil {
		asss = assists.AssistMids
	}
	return
}
