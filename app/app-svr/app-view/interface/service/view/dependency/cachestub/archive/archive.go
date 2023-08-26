package archive

import (
	"context"
	viewapi "go-gateway/app/app-svr/app-view/interface/api/view"
	archivedao "go-gateway/app/app-svr/app-view/interface/dao/archive"
	"go-gateway/app/app-svr/app-view/interface/model/view"
	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"
	archive "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/ecode"

	"github.com/pkg/errors"
)

var _ dependency.ArchiveDependency = &Impl{}

type Impl struct {
	Origin dependency.ArchiveDependency

	Reply struct {
		Views          map[int64]*archive.ViewReply
		DescriptionsV2 map[int64]*archivedao.DescriptionV2Reply
		Arguments      map[int64]string
		RedirectPolicy map[int64]*archive.RedirectPolicy
	}
}

func (impl *Impl) ArcsPlayer(ctx context.Context, arcsPlayAv []*archive.PlayAv) (map[int64]*archive.ArcPlayer, error) {
	return impl.Origin.ArcsPlayer(ctx, arcsPlayAv)
}

func (impl *Impl) NewRelateAids(ctx context.Context, aid, mid, zoneID int64, build, parentMode, autoplay, isAct int, buvid, sourcePage, trackid, cmd, tabid string, plat int8, pageVersion, fromSpmid string) (res *view.RelateRes, returnCode string, err error) {
	return impl.Origin.NewRelateAids(ctx, aid, mid, zoneID, build, parentMode, autoplay, isAct, buvid, sourcePage, trackid, cmd, tabid, plat, pageVersion, fromSpmid)
}

func (impl *Impl) Archives(ctx context.Context, aids []int64, mid int64, mobiApp, device string) (map[int64]*archive.Arc, error) {
	return impl.Origin.Archives(ctx, aids, mid, mobiApp, device)
}

func (impl *Impl) UpLikeImgCreative(ctx context.Context, mid int64, avid int64) (*viewapi.UpLikeImg, error) {
	return impl.Origin.UpLikeImgCreative(ctx, mid, avid)
}

func (impl *Impl) ArcRedirectUrl(ctx context.Context, aid int64) (*archive.RedirectPolicy, error) {
	v, ok := impl.Reply.RedirectPolicy[aid]
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "aid: %d", aid)
	}
	return v, nil
}

func (impl *Impl) Argument(ctx context.Context, aid int64) (string, error) {
	v, ok := impl.Reply.Arguments[aid]
	if !ok {
		return "", errors.Wrapf(ecode.NothingFound, "aid: %d", aid)
	}
	return v, nil
}

func (impl *Impl) View3(ctx context.Context, aid, mid int64, mobiApp, device, platform string) (*archive.ViewReply, error) {
	return impl.Origin.View3(ctx, aid, mid, mobiApp, device, platform)
}

func (impl *Impl) DescriptionV2(ctx context.Context, aid int64) (desc string, descV2 []*archive.DescV2, mids []int64, err error) {
	reply, ok := impl.Reply.DescriptionsV2[aid]
	if !ok {
		return "", nil, nil, errors.Wrapf(ecode.NothingFound, "aid: %d", aid)
	}
	return reply.Desc, reply.DescV2, reply.Mids, nil
}
