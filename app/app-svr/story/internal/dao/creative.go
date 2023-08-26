package dao

import (
	"context"
	"fmt"

	materialgrpc "git.bilibili.co/bapis/bapis-go/material/interface"
	vogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

// Argument .
func (d *dao) Arguments(ctx context.Context, aids []int64) (map[int64]*vogrpc.Argument, error) {
	arguRly, err := d.voClient.MultiArchiveArgument(ctx, &vogrpc.MultiArchiveArgumentReq{Aids: aids})
	if err != nil {
		return nil, err
	}
	return arguRly.GetArguments(), nil
}

func (d *dao) StoryTagList(ctx context.Context, arg []*materialgrpc.StoryReq) (map[string]*materialgrpc.StoryRes, error) {
	result, err := d.materialClient.GetStoryInfo(ctx, &materialgrpc.StoryTagReq{
		StoryReq: arg,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]*materialgrpc.StoryRes, len(result.StoryRes))
	for _, v := range result.StoryRes {
		out[fmt.Sprintf("%d:%d", v.Avid, v.Type)] = v
	}
	return out, nil
}
