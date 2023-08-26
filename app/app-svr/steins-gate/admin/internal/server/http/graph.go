package http

import (
	bm "go-common/library/net/http/blademaster"
)

func graphShow(c *bm.Context) {
	v := new(struct {
		GraphID int64 `form:"graph_id" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(svc.GraphShow(c, v.GraphID))
}

func nodeInfoAudit(c *bm.Context) {
	v := new(struct {
		GraphVersion int64 `form:"graph_version" validate:"required"`
		NodeID       int64 `form:"node_id" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(svc.NodeInfoAudit(c, v.GraphVersion, v.NodeID))
}

func edgeInfoV2Audit(c *bm.Context) {
	v := new(struct {
		GraphVersion int64 `form:"graph_version" validate:"required"`
		EdgeId       int64 `form:"edge_id"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	// 由于下发给前端的是node，需要转下
	edgeId, err := svc.GetEdgeIdByNode(c, v.EdgeId)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	v.EdgeId = edgeId
	c.JSON(svc.EdgeInfoV2Audit(c, v.GraphVersion, v.EdgeId))

}
