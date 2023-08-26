package archivehonor

import (
	"context"

	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"
	archviehonorapi "go-gateway/app/app-svr/archive-honor/service/api"

	"go-common/library/ecode"

	"github.com/pkg/errors"
)

var _ dependency.ArchiveHonorDependency = &Impl{}

type Impl struct {
	Origin dependency.ArchiveHonorDependency

	Reply struct {
		ArchiveHonors map[int64]*archviehonorapi.HonorReply
	}
}

func (impl *Impl) Honors(ctx context.Context, aid, _ int64, _, _ string) ([]*archviehonorapi.Honor, error) {
	v, ok := impl.Reply.ArchiveHonors[aid]
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "aid: %d", aid)
	}
	return v.Honor, nil
}
