package service

import (
	"context"

	"go-common/library/ecode"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/steins-gate/admin/internal/model"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

func (s *Service) NodeInfoAudit(c context.Context, GraphVersion, nodeID int64) (res *model.NodeInfo, err error) {
	var (
		nodeInfo    *api.GraphNode
		showChoices []*model.Choice
		oriChoices  []*api.GraphEdge
	)
	eg := errgroup.WithContext(c) // 获取选项
	eg.Go(func(c context.Context) (err error) {
		nodeInfo, err = s.dao.Node(c, nodeID)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		oriChoices, err = s.dao.EdgeByNode(c, nodeID)
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if nodeInfo.GraphId != GraphVersion { // if node not match graphid, return 400
		err = ecode.RequestErr
		return
	}
	showChoices = model.BuildChoices(oriChoices)
	res = model.BuildReply(nodeInfo, showChoices)
	return

}
