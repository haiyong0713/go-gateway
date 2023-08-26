package playurl

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-intl/interface/conf"
	"go-gateway/app/app-svr/app-intl/interface/model/player"
	"go-gateway/app/app-svr/app-player/interface/model"
	"go-gateway/app/app-svr/playurl/service/api"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"

	"github.com/pkg/errors"
)

var (
	_fhMobiAppMap = map[string]struct{}{
		"android":    {},
		"android_tv": {},
		"android_G":  {},
		"android_i":  {},
		"iphone":     {},
		"ipad":       {},
		"white":      {},
	}
)

// Dao is player dao.
type Dao struct {
	// rpc
	playURLRPC   api.PlayURLClient
	playURLRPCV2 v2.PlayURLClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.playURLRPC, err = api.NewClient(c.PlayURLClient)
	if err != nil {
		panic(fmt.Sprintf("player NewClient error(%v)", err))
	}
	d.playURLRPCV2, err = v2.NewClient(c.PlayURLClient)
	if err != nil {
		panic(fmt.Sprintf("player v2 NewClient error(%v)", err))
	}
	return
}

// PlayURL is
func (d *Dao) PlayURL(c context.Context, params *player.Param, mid int64) (player *api.PlayURLReply, err error) {
	req := &api.PlayURLReq{
		Aid:       params.AID,
		Cid:       params.CID,
		Qn:        params.Qn,
		Npcybs:    params.Npcybs,
		Platform:  params.Platform,
		Fnver:     params.Fnver,
		Fnval:     params.Fnval,
		Session:   params.Session,
		Build:     params.Build,
		ForceHost: params.ForceHost,
		Buvid:     params.Buvid,
		Mid:       mid,
		Fourk:     params.Fourk,
		Device:    params.Device,
		MobiApp:   params.MobiApp,
		Dl:        params.Dl,
	}
	if params.Dl == 1 || params.Npcybs == 1 {
		req.ForceHost = 2 //离线下载默认https
	}
	if player, err = d.playURLRPC.PlayURL(c, req); err != nil {
		err = errors.Wrapf(err, "%v", req)
		return
	}
	return
}

// PlayURLV2 is
func (d *Dao) PlayURLV2(c context.Context, params *player.Param, mid int64) (player *v2.PlayURLReply, err error) {
	var dl uint32
	if params.Dl == 1 {
		dl = model.DlDash
	} else if params.Npcybs == 1 {
		dl = model.DlFlv
	}
	// 以下参数转换见视频云tapd地址https://www.tapd.cn/20095661/prong/stories/view/1120095661001131850
	fh := int32(1)            //force_host默认是1
	if params.ForceHost > 0 { //客户端有值即透传
		fh = params.ForceHost
	} else if _, ok := _fhMobiAppMap[params.MobiApp]; ok { //未传值判断platform
		fh = 0
	}
	req := &v2.PlayURLReq{
		Aid:       params.AID,
		Cid:       params.CID,
		Qn:        params.Qn,
		Platform:  params.Platform,
		Fnver:     params.Fnver,
		Fnval:     params.Fnval,
		ForceHost: fh,
		Mid:       mid,
		Fourk:     params.Fourk == 1,
		Device:    params.Device,
		MobiApp:   params.MobiApp,
		Download:  dl,
		BackupNum: 2, //客户端请求默认2个
	}
	if params.Dl == 1 || params.Npcybs == 1 {
		req.ForceHost = 2 //离线下载默认https
	}
	if player, err = d.playURLRPCV2.PlayURL(c, req); err != nil {
		err = errors.Wrapf(err, "%v", req)
		return
	}
	return
}
