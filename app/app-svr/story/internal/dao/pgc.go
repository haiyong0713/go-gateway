package dao

import (
	"context"

	"go-common/library/net/metadata"
	arcmid "go-gateway/app/app-svr/archive/middleware"

	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcstory "git.bilibili.co/bapis/bapis-go/pgc/service/card/story"
	pgcClient "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
)

func (d *dao) InlineCards(c context.Context, epIDs []int32, mobiApp, platform, device string, build int, mid int64, needHe bool, buvid string, heInlineReq []*pgcinline.HeInlineReq) (map[int32]*pgcinline.EpisodeCard, error) {
	batchArg, _ := arcmid.FromContext(c)
	arg := &pgcinline.EpReq{
		EpIds: epIDs,
		User: &pgcinline.UserReq{
			Mid:      mid,
			MobiApp:  mobiApp,
			Device:   device,
			Platform: platform,
			Ip:       metadata.String(c, metadata.RemoteIP),
			Fnver:    uint32(batchArg.Fnver),
			Fnval:    uint32(batchArg.Fnval),
			Qn:       uint32(batchArg.Qn),
			Build:    int32(build),
			Fourk:    int32(batchArg.Fourk),
			NetType:  pgccard.NetworkType(batchArg.NetType),
			TfType:   pgccard.TFType(batchArg.TfType),
			Buvid:    buvid,
		},
		SceneControl: &pgcinline.SceneControl{
			WasStory: true,
		},
		CustomizeReq: &pgcinline.CustomizeReq{
			NeedShareCount: true,
			NeedHe:         needHe,
		},
		HeInlineReq: heInlineReq,
	}
	info, err := d.pgcinlineClient.EpCard(c, arg)
	if err != nil {
		return nil, err
	}
	return info.Infos, nil
}

func (d *dao) OgvPlaylist(ctx context.Context, arg *pgcstory.StoryPlayListReq) (*pgcstory.StoryPlayListReply, error) {
	resp, err := d.pgcStoryClient.QueryPlayList(ctx, arg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *dao) StatusByMid(c context.Context, mid int64, SeasonIDs []int32) (map[int32]*pgcClient.FollowStatusProto, error) {
	rly, err := d.pgcFollowClient.StatusByMid(c, &pgcClient.FollowStatusByMidReq{Mid: mid, SeasonId: SeasonIDs})
	if err != nil {
		return nil, err
	}
	return rly.GetResult(), nil
}
