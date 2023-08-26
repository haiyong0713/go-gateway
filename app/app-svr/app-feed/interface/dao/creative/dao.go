package creative

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	materialgrpc "git.bilibili.co/bapis/bapis-go/material/interface"
	opensourcegrpc "git.bilibili.co/bapis/bapis-go/platform/open-course/interface"
	vogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

type Dao struct {
	voClient         vogrpc.VideoUpOpenClient
	materialClient   materialgrpc.MaterialClient
	opensourceClient opensourcegrpc.OpenCourseInterfaceV1Client
}

func New(c *conf.Config) *Dao {
	d := &Dao{}
	var err error
	if d.voClient, err = vogrpc.NewClient(c.VideoOpenClient); err != nil {
		panic(err)
	}
	if d.materialClient, err = materialgrpc.NewClient(c.MaterialClient); err != nil {
		panic(err)
	}
	if d.opensourceClient, err = opensourcegrpc.NewClientOpenCourseInterfaceV1(c.OpenCourseClient); err != nil {
		panic(err)
	}
	return d
}

// Argument .
func (d *Dao) Arguments(ctx context.Context, aids []int64) (map[int64]*vogrpc.Argument, error) {
	arguRly, err := d.voClient.MultiArchiveArgument(ctx, &vogrpc.MultiArchiveArgumentReq{Aids: aids})
	if err != nil {
		return nil, err
	}
	return arguRly.GetArguments(), nil
}

func (d *Dao) StoryTagList(ctx context.Context, arg []*materialgrpc.StoryReq) (map[string]*materialgrpc.StoryRes, error) {
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

func (d *Dao) OpenCoursePegasusMark(ctx context.Context, aids []int64) (map[int64]bool, error) {
	result, err := d.opensourceClient.BatchGetAidPegasusMark(ctx, &opensourcegrpc.BatchGetPegasusMarkReq{
		Aids: aids,
	})
	if err != nil {
		return nil, err
	}
	return result.AidPegasusMarkMap, nil
}
