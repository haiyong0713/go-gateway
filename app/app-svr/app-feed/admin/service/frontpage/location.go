package frontpage

import (
	"context"
	locadmingrpc "git.bilibili.co/bapis/bapis-go/platform/admin/location"
	"go-common/library/log"
)

func (s *Service) GetAllPolicyGroups(ctx context.Context) (res []*locadmingrpc.PolicyGroupInfo, err error) {
	if res, err = s.locationDAO.ListGroup(ctx, locadmingrpc.FRONTPAGE, 1, 99); err != nil {
		log.Error("Service: GetAllPolicyGroups ListGroup error %v", err)
	}
	return
}
