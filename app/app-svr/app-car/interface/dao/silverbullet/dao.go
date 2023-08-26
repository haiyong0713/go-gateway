package silverbullet

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/view"

	api "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
)

// Dao is gaia dao
type Dao struct {
	// grpc
	client api.GaiaClient
}

// New is
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.client, err = api.NewClient(c.GaiaClient); err != nil {
		panic(fmt.Sprintf("Gaia Client not found err(%v)", err))
	}
	return
}

// Honors is
func (d *Dao) RuleCheck(c context.Context, silverEventCtx *view.SilverEventCtx, scene string) bool {
	eventCtx, err := json.Marshal(silverEventCtx)
	if err != nil {
		log.Error("json.Marshal err(%+v) v(%+v)", err, silverEventCtx)
		return false
	}
	req := &api.RuleCheckReq{
		Scene:    scene,
		EventCtx: string(eventCtx),
		EventTs:  time.Now().Unix(),
	}
	rep, err := d.client.RuleCheck(c, req)
	if err != nil {
		log.Error("d.client.RuleCheck err(%+v)", err)
		return false
	}
	for _, v := range rep.GetDecisions() {
		if strings.Contains(v, "reject") {
			log.Warn("RuleCheck reject req(%+v)", req)
			return true
		}
	}
	return false
}
