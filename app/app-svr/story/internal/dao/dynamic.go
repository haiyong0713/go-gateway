package dao

import (
	"context"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

func (d *dao) DynamicGeneralStory(ctx context.Context, param *dyngrpc.GeneralStoryReq) (*dyngrpc.GeneralStoryRsp, error) {
	reply, err := d.dynamicClient.GeneralStory(ctx, param)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *dao) DynamicSpaceStory(ctx context.Context, param *dyngrpc.SpaceStoryReq) (*dyngrpc.SpaceStoryRsp, error) {
	reply, err := d.dynamicClient.SpaceStory(ctx, param)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *dao) DynamicInsert(ctx context.Context, param *dyngrpc.InsertedStoryReq) (*dyngrpc.InsertedStoryRsp, error) {
	reply, err := d.dynamicClient.InsertedStory(ctx, param)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
