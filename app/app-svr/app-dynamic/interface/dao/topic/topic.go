package topic

import (
	"context"

	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	topicmdl "go-gateway/app/app-svr/app-dynamic/interface/model/topic"

	"go-common/library/sync/errgroup.v2"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	topicV2 "git.bilibili.co/bapis/bapis-go/topic/service"
)

// 老版本动态话题
func (d *Dao) OldRcmdActList(c context.Context, mid int64, buvid string, req *topicmdl.SquareReq) ([]*dyntopicgrpc.HotListDetail, error) {
	resTmp, err := d.dynTopic.RcmdActList(c, &dyntopicgrpc.RcmdActListReq{
		Uid: mid,
		MetaData: &dyncommongrpc.CmnMetaData{
			Build:     req.Build,
			Platform:  req.Platform,
			MobiApp:   req.MobiApp,
			Device:    req.Device,
			FromSpmid: req.FromSpmid,
			Version:   req.Version,
			Buvid:     buvid,
		}})
	if err != nil {
		return nil, err
	}
	return resTmp.GetHotList(), nil
}

// 老版本动态话题
func (d *Dao) OldHotList(c context.Context, mid int64, buvid string, req *topicmdl.HotListReq) (*dyntopicgrpc.HotListRsp, error) {
	res, err := d.dynTopic.HotList(c, &dyntopicgrpc.HotListReq{
		Uid:         mid,
		HotListType: req.HotListType,
		MetaData: &dyncommongrpc.CmnMetaData{
			Build:     req.Build,
			Platform:  req.Platform,
			MobiApp:   req.MobiApp,
			Device:    req.Device,
			FromSpmid: req.FromSpmid,
			Version:   req.Version,
			Buvid:     buvid,
		},
		Offset:   req.Offset,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 老版本动态话题
func (d *Dao) OldTopicSearch(c context.Context, keyword string, general *mdlv2.GeneralParam) (*mdlv2.OldTopicSearchResImpl, error) {
	req := &dyntopicgrpc.TopicSearchReq{
		Mid: general.Mid,
		MetaData: &dyncommongrpc.CmnMetaData{
			Build:    general.GetBuildStr(),
			Platform: general.GetPlatform(),
			MobiApp:  general.GetMobiApp(),
			Device:   general.GetDevice(),
			Buvid:    general.GetBuvid(),
		},
		Word: keyword,
	}
	reply, err := d.dynTopic.TopicSearch(c, req)
	if err != nil {
		return nil, err
	}
	return &mdlv2.OldTopicSearchResImpl{Res: reply.GetTopics()}, nil
}

func (d *Dao) TopicSearchV2(ctx context.Context, keyword string, general *mdlv2.GeneralParam) (*mdlv2.NewTopicSearchResImpl, error) {
	req := &topicV2.VertSearchTopicInfoReq{
		Uid:     general.Mid,
		KeyWord: keyword,
	}
	res, err := d.topicV2.VertSearchTopicInfoV2(ctx, req)
	if err != nil {
		return nil, err
	}
	return &mdlv2.NewTopicSearchResImpl{Res: res}, nil
}

const defaultRcmdTopicNum = 9

func (d *Dao) RcmdNewTopics(ctx context.Context, general *mdlv2.GeneralParam) (*mdlv2.NewTopicSquareImpl, error) {
	req := &topicV2.RcmdNewTopicsReq{
		Uid:      general.Mid,
		PageSize: defaultRcmdTopicNum,
		MetaData: general.ToTopicCmnMetaData(),
	}
	resp, err := d.topicV2.RcmdNewTopics(ctx, req)
	if err != nil {
		return nil, err
	}
	return &mdlv2.NewTopicSquareImpl{Resp: resp}, nil
}

func (d *Dao) NewTopicSetDetails(ctx context.Context, idm map[int64]int64, general *mdlv2.GeneralParam) (ret map[int64]*mdlv2.NewTopicSetDetail, err error) {
	eg := errgroup.WithContext(ctx)
	cmnMeta := general.ToTopicCmnMetaData()
	ret = make(map[int64]*mdlv2.NewTopicSetDetail)

	for pid, sid := range idm {
		pushid := pid
		setid := sid
		res := new(mdlv2.NewTopicSetDetail)
		eg.Go(func(ctx context.Context) error {
			resp, err := d.topicV2.TopicSetInfo(ctx, &topicV2.TopicSetInfoReq{
				SetId: setid, Uid: general.Mid, Metadata: cmnMeta,
			})
			res.FromSetInfo(resp)
			return err
		})
		eg.Go(func(ctx context.Context) error {
			resp, err := d.topicV2.SetExposureTopics(ctx, &topicV2.SetExposureTopicsReq{
				TopicSetId: setid, TopicSetPushId: pushid,
			})
			res.FromTopicList(resp)
			return err
		})
		ret[pushid] = res
	}

	return ret, eg.Wait()
}

func (d *Dao) LegacyTopicFeed(ctx context.Context, req *dyntopicgrpc.ListDynsV2Req) (*dyntopicgrpc.ListDynsV2Rsp, error) {
	return d.dynTopic.ListDynsV2(ctx, req)
}
