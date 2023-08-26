package tag

import (
	"context"
	"time"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-gateway/app/app-svr/app-channel/interface/conf"
)

// Dao tag
type Dao struct {
	c       *conf.Config
	tagGRPC taggrpc.TagRPCClient
}

// New tag
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.tagGRPC, err = taggrpc.NewClient(c.TagGRPC); err != nil {
		panic(err)
	}
	return
}

// ChannelDetail channel info by id or name
func (d *Dao) ChannelDetail(c context.Context, mid, tid int64, tname string, isOversea int32) (t *taggrpc.ChannelReply, err error) {
	arg := &taggrpc.ChannelReq{Mid: mid, Tid: tid, Tname: tname, From: isOversea}
	t, err = d.tagGRPC.Channel(c, arg)
	return
}

// Tags by tag ids
func (d *Dao) Tags(ctx context.Context, mid int64, tids []int64) (map[int64]*taggrpc.Tag, error) {
	res, err := d.tagGRPC.Tags(ctx, &taggrpc.TagsReq{Mid: mid, Tids: tids})
	if err != nil {
		return nil, err
	}
	return res.Tags, nil
}

// Resources channel resources aids
func (d *Dao) Resources(c context.Context, plat int8, id, mid int64, name, buvid string, build, requestCnt, loginEvent, displayID, teenagersMode int) (res *taggrpc.ChannelResourcesReply, err error) {
	arg := &taggrpc.ChannelResourcesReq{
		Tid:        id,
		Mid:        mid,
		Plat:       int32(plat),
		Build:      int32(build),
		LoginEvent: int32(loginEvent),
		RequestCnt: int32(requestCnt),
		DisplayId:  int32(displayID),
		Type:       3,
		Tname:      name,
		Buvid:      buvid,
		From:       0,
	}
	if res, err = d.tagGRPC.ChannelResources(c, arg); err != nil {
		return
	}
	return
}

// SubscribeUpdate subscribe update
func (d *Dao) SubscribeUpdate(c context.Context, mid int64, tids string) (err error) {
	arg := &taggrpc.UpdateCustomSortReq{Tids: tids, Mid: mid, Type: 1}
	_, err = d.tagGRPC.UpdateCustomSort(c, arg)
	return
}

// SubscribeAdd subscribe add
func (d *Dao) SubscribeAdd(c context.Context, mid, tagID int64, now time.Time) (err error) {
	arg := &taggrpc.AddSubReq{Tids: []int64{tagID}, Mid: mid}
	_, err = d.tagGRPC.AddSub(c, arg)
	return
}

// SubscribeCancel subscribe add
func (d *Dao) SubscribeCancel(c context.Context, mid, tagID int64, now time.Time) (err error) {
	arg := &taggrpc.CancelSubReq{Tid: tagID, Mid: mid}
	_, err = d.tagGRPC.CancelSub(c, arg)
	return
}

// Recommend func
func (d *Dao) Recommend(c context.Context, mid int64, isOversea int32) ([]*taggrpc.Channel, error) {
	arg := &taggrpc.ChannelRecommendReq{Mid: mid, From: isOversea}
	res, err := d.tagGRPC.ChannelRecommend(c, arg)
	if err != nil {
		return nil, err
	}
	return res.GetChannels(), nil
}

// ListByCategory 分类下的频道
func (d *Dao) ListByCategory(c context.Context, id, mid int64, isOversea int32) ([]*taggrpc.Channel, error) {
	arg := &taggrpc.ChanneListReq{Id: id, Mid: mid, From: isOversea}
	res, err := d.tagGRPC.ChanneList(c, arg)
	if err != nil {
		return nil, err
	}
	return res.GetChannels(), nil
}

// Subscribe 已订阅频道
func (d *Dao) Subscribe(c context.Context, mid int64) (customSort *taggrpc.CustomSortChannelReply, err error) {
	arg := &taggrpc.CustomSortChannelReq{Type: 1, Mid: mid, Pn: 1, Ps: 400, Order: -1}
	customSort, err = d.tagGRPC.CustomSortChannel(c, arg)
	return
}

// Discover 频道tab页的3个发现频道
func (d *Dao) Discover(c context.Context, mid int64, isOversea int32) ([]*taggrpc.Channel, error) {
	arg := &taggrpc.ChannelDiscoveryReq{Mid: mid, From: isOversea}
	discover, err := d.tagGRPC.ChannelDiscovery(c, arg)
	if err != nil {
		return nil, err
	}
	return discover.GetChannels(), nil
}

// Category channel category
func (d *Dao) Category(c context.Context, isOversea int32) ([]*taggrpc.ChannelCategory, error) {
	category, err := d.tagGRPC.ChannelCategory(c, &taggrpc.ChannelCategoryReq{From: isOversea})
	if err != nil {
		return nil, err
	}
	return category.GetCategories(), nil
}

// Square 频道广场页推荐频道+稿件
func (d *Dao) Square(c context.Context, mid int64, tagNum, oidNum, build int, loginEvent int32, plat int8, buvid string, isOversea int32) (square *taggrpc.ChannelSquareReply, err error) {
	arg := &taggrpc.ChannelSquareReq{
		Mid:            mid,
		TagNumber:      int32(tagNum),
		ResourceNumber: int32(oidNum),
		Type:           3,
		Buvid:          buvid,
		Build:          int32(build),
		Plat:           int32(plat),
		LoginEvent:     loginEvent,
		DisplayId:      1,
		From:           isOversea,
	}
	square, err = d.tagGRPC.ChannelSquare(c, arg)
	return
}
