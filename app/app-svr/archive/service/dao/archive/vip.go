package archive

import (
	"context"

	vipinfoAPI "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"
)

// VipInfo .
func (d *Dao) VipInfo(c context.Context, mid int64, buvid string, withControl bool) (*vipinfoAPI.InfoReply, error) {
	return d.vipClient.Info(c, &vipinfoAPI.InfoReq{Mid: mid, Buvid: buvid, WithControl: withControl})
}
