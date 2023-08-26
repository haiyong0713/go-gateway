package danmu

import (
	"context"

	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"

	dmApi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
)

var _ dependency.DanmuDependency = &Impl{}

type Impl struct {
	Origin dependency.DanmuDependency

	Reply struct {
		ArchiveSubjectInfos map[int64]*dmApi.SubjectInfo
	}
}

func (impl *Impl) SubjectInfos(ctx context.Context, _ int32, _ int8, _ ...int64) (map[int64]*dmApi.SubjectInfo, error) {
	return impl.Reply.ArchiveSubjectInfos, nil
}
