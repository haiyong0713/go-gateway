package copyright

import (
	"context"
	"fmt"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/copyright-manage/interface"
)

type Dao struct {
	copyrightClient api.CopyrightManageClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.copyrightClient, err = api.NewClient(c.CopyrightClient); err != nil {
		panic(fmt.Sprintf("copyright NewClient not found err(%v)", err))
	}
	return
}

// 是否后台播放
func (d *Dao) GetArcsBanPlay(c context.Context, aids []int64) (map[int64]api.BanPlayEnum, error) {
	req := &api.AidsReq{
		Aids:   aids,
		Option: api.BanOption_BanBackend,
	}
	res, err := d.copyrightClient.GetArcsBanPlay(c, req)
	if err != nil {
		log.Error("d.GetArcsBanPlay err:%+v", err)
		return nil, err
	}
	return res.BanPlay, nil
}

func (d *Dao) GetArcBanPlay(ctx context.Context, aid int64) (bool, error) {
	req := &api.AidReq{
		Aid: aid,
	}
	res, err := d.copyrightClient.GetArcBanPlay(ctx, req)
	if err != nil {
		log.Error("d.GetArcBanPlay err:%+v", err)
		return false, err
	}
	for _, v := range res.GetBanPlay() {
		if v.Key == "ban_listen" && v.Value == 1 {
			return true, nil
		}
	}
	return false, nil
}
