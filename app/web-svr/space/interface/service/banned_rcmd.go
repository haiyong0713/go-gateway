package service

import (
	"github.com/golang/protobuf/ptypes/empty"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
)
import "context"

// blacklist space blacklist
func (s *Service) UpRcmdBlackList(ctx context.Context, _ *empty.Empty) (rep *pb.UpRcmdBlackListReply, err error) {
	var bannedMids []int64
	bannedMids, err = s.dao.GetBannedRcmdMids(ctx)
	if err != nil {
		bannedMids = s.upRcmdBlackList
	} else {
		s.upRcmdBlackList = bannedMids
	}
	rep = new(pb.UpRcmdBlackListReply)
	rep.BannedMids = bannedMids
	return rep, nil
}
