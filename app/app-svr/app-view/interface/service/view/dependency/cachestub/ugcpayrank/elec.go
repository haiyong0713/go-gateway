package ugcpayrank

import (
	"context"

	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"

	"go-common/library/ecode"

	elecapi "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	"github.com/pkg/errors"
)

var _ dependency.UgcpayRankDependency = &Impl{}

type Impl struct {
	Origin dependency.UgcpayRankDependency

	Reply struct {
		RankElec *elecapi.RankElecUPListResp
	}
}

func (impl *Impl) RankElecMonthUP(ctx context.Context, upmid, _ int64, _, _, _ string) (*elecapi.RankElecUPResp, error) {
	v, ok := impl.Reply.RankElec.GetMap()[upmid]
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "upmid: %d", upmid)
	}
	return &elecapi.RankElecUPResp{UP: v}, nil
}

func (impl *Impl) UPRankWithPanelByUPMid(ctx context.Context, mid, upmid, build int64, mobiApp, platform, device string) (*elecapi.UPRankWithPanelReply, error) {
	return impl.Origin.UPRankWithPanelByUPMid(ctx, mid, upmid, build, mobiApp, platform, device)
}
