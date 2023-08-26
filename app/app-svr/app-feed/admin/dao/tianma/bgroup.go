package tianma

import (
	"context"

	bgroupGRPC "git.bilibili.co/bapis/bapis-go/platform/service/bgroup"
)

// AddBGroup 新建人群包
func (d *Dao) AddBGroup(ctx context.Context, req *bgroupGRPC.AddBGroupReq) (*bgroupGRPC.AddBGroupResp, error) {
	reply, err := d.bgroupClient.AddBGroup(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// UpdateBGroup 更新人群包
func (d *Dao) UpdateBGroup(ctx context.Context, req *bgroupGRPC.UpdateBGroupReq) (*bgroupGRPC.UpdateBGroupResp, error) {
	reply, err := d.bgroupClient.UpdateBGroup(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// MidBGroups 根据Mid查询人群包
func (d *Dao) MidBGroups(ctx context.Context, req *bgroupGRPC.MidBGroupsReq) (*bgroupGRPC.MidBGroupsResp, error) {
	reply, err := d.bgroupClient.MidBGroups(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// BGroups 查询人群包
func (d *Dao) BGroup(ctx context.Context, req *bgroupGRPC.BGroupReq) (*bgroupGRPC.BGroupResp, error) {
	reply, err := d.bgroupClient.BGroup(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
