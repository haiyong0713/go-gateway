package act

import (
	"fmt"

	"go-gateway/app/app-svr/app-show/interface/conf"
	natrpc "go-gateway/app/web-svr/native-page/interface/api"

	actrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	media "git.bilibili.co/bapis/bapis-go/pgc/service/media"
)

// Dao is activity dao.
type Dao struct {
	actRPC actrpc.ActivityClient
	//新代码库rpc
	natRPC          natrpc.NaPageClient
	ClickSpecialTip *conf.ClickSpecialTip
	charGRPC        media.CharacterClient
}

// New new a activity dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		ClickSpecialTip: c.ClickSpecialTip,
	}
	var err error
	if d.actRPC, err = actrpc.NewClient(c.ActivityGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.natRPC, err = natrpc.NewClient(c.ActivityGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.charGRPC, err = media.NewClientCharacter(c.CharGRPC); err != nil {
		panic(fmt.Sprintf("Fail to new characterClient, error=%+v", err))
	}
	return d
}
