package archive

import (
	"context"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	gwecode "go-gateway/app/app-svr/app-card/ecode"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/archive/service/api"

	creativerpc "git.bilibili.co/bapis/bapis-go/creative/open/service"

	"github.com/pkg/errors"
)

const (
	_max = 50
)

type DescriptionV2Request struct {
	Aid int64
}

type DescriptionV2Reply struct {
	Desc   string
	DescV2 []*api.DescV2
	Mids   []int64
}

// View3 view archive with pages pb.
func (d *Dao) View3(c context.Context, aid, mid int64, mobiApp, device, platform string) (*api.ViewReply, error) {
	arg := &api.ViewRequest{Aid: aid, Mid: mid, MobiApp: mobiApp, Device: device, Platform: platform}
	reply, err := d.arcGRPC.View(c, arg)
	if err != nil {
		log.Error("%+v", err)
		if !ecode.EqualError(ecode.NothingFound, err) {
			return nil, gwecode.AppViewForRetry
		}
		return nil, err
	}
	return reply, nil
}

// Description get archive description by aid.
func (d *Dao) Description(c context.Context, aid int64) (desc string, err error) {
	arg := &api.DescriptionRequest{Aid: aid}
	reply, err := d.arcGRPC.Description(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	desc = reply.Desc
	return
}

func (d *Dao) DescriptionsV2(ctx context.Context, req []*DescriptionV2Request) (map[int64]*DescriptionV2Reply, error) {
	aids := []int64{}
	for _, i := range req {
		aids = append(aids, i.Aid)
	}
	reply, err := d.arcGRPC.Descriptions(ctx, &api.DescriptionsRequest{Aids: aids})
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", aids)
	}

	out := map[int64]*DescriptionV2Reply{}
	for _, i := range req {
		desc, ok := reply.Description[i.Aid]
		if !ok {
			log.Warn("No extra description on: %d", i.Aid)
			continue
		}
		descReply := &DescriptionV2Reply{
			Desc:   desc.Desc,
			DescV2: desc.DescV2Parse,
		}
		if len(descReply.DescV2) == 0 {
			if desc.Desc != "" {
				descReply.DescV2 = append(descReply.DescV2, &api.DescV2{
					RawText: desc.Desc,
					Type:    api.DescType_DescTypeText,
				})
			}
		}
		for _, v := range descReply.DescV2 {
			if v == nil {
				continue
			}
			if viewApi.DescType(v.Type) != viewApi.DescType_DescTypeAt {
				continue
			}
			descReply.Mids = append(descReply.Mids, v.BizId)
		}
		out[i.Aid] = descReply
	}
	return out, nil
}

func (d *Dao) DescriptionV2(c context.Context, aid int64) (desc string, descV2 []*api.DescV2, mids []int64, err error) {
	arg := &api.DescriptionRequest{Aid: aid}
	reply, err := d.arcGRPC.Description(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return "", nil, nil, err
	}
	desc = reply.Desc
	descV2 = reply.DescV2Parse
	if len(descV2) == 0 {
		if desc != "" {
			descV2 = append(descV2, &api.DescV2{
				RawText: desc,
				Type:    api.DescType_DescTypeText,
			})
		}
	}
	for _, v := range descV2 {
		if v == nil {
			continue
		}
		if viewApi.DescType(v.Type) != viewApi.DescType_DescTypeAt {
			continue
		}
		mids = append(mids, v.BizId)
	}
	return desc, descV2, mids, nil
}

// Argument .
func (d *Dao) Argument(c context.Context, aid int64) (argueMsg string, err error) {
	var arguRly *creativerpc.ArgumentReply
	if arguRly, err = d.creativeClient.ArchiveArgument(c, &creativerpc.ArgumentRequest{Aid: aid}); err != nil {
		err = errors.Wrapf(err, "Argument:%d", aid)
		return
	}
	argueMsg = arguRly.ArgueMsg
	return
}

// Argument .
func (d *Dao) UpLikeImgCreative(c context.Context, mid int64, avid int64) (*viewApi.UpLikeImg, error) {
	req := creativerpc.UpLikeImgReq{
		Mid:  mid,
		Avid: avid,
	}
	reply, err := d.creativeClient.UpLikeImg(c, &req)
	if err != nil {
		err = errors.Wrapf(err, " d.creativeClient.UpLikeImg is err:%d", mid)
		return nil, err
	}
	if reply == nil {
		return nil, ecode.NothingFound
	}
	return &viewApi.UpLikeImg{
		PreImg:  reply.PreImg,
		SucImg:  reply.SucImg,
		Content: reply.Content,
		Type:    reply.Type,
	}, nil
}

// Views is
func (d *Dao) Views(c context.Context, aids []int64) (map[int64]*api.ViewReply, error) {
	if len(aids) == 0 {
		return nil, errors.New("empty aids")
	}
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	views := make(map[int64]*api.ViewReply)
	for i := 0; i < len(aids); i += _max {
		var partAids []int64
		if i+_max > len(aids) {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_max]
		}
		g.Go(func(ctx context.Context) (err error) {
			var res *api.ViewsReply
			arg := &api.ViewsRequest{Aids: partAids}
			if res, err = d.arcGRPC.Views(ctx, arg); err != nil {
				return err
			}
			mu.Lock()
			for aid, v := range res.GetViews() {
				views[aid] = v
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return views, nil
}

// redirect
func (d *Dao) ArcRedirectUrl(c context.Context, aid int64) (*api.RedirectPolicy, error) {
	req := &api.ArcsRedirectPolicyRequest{
		Aids: []int64{aid},
	}
	res, err := d.arcGRPC.ArcsRedirectPolicy(c, req)
	if err != nil {
		return nil, err
	}
	redirects := res.GetRedirectPolicy()
	v, ok := redirects[aid]
	if !ok {
		return nil, ecode.NothingFound
	}
	return v, nil
}

// batch redirect
func (d *Dao) BatchArcRedirectUrls(c context.Context, req *api.ArcsRedirectPolicyRequest) (map[int64]*api.RedirectPolicy, error) {
	res, err := d.arcGRPC.ArcsRedirectPolicy(c, req)
	if err != nil {
		return nil, err
	}
	return res.GetRedirectPolicy(), nil
}
