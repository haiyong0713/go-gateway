package dao

import (
	"context"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

type relationDao struct {
	client relationgrpc.RelationClient
}

func (d *relationDao) Relations(c context.Context, req *relationgrpc.RelationsReq) (map[int64]*relationgrpc.FollowingReply, error) {
	rly, err := d.client.Relations(c, req)
	if err != nil {
		return nil, err
	}
	return rly.GetFollowingMap(), nil
}

func (d *relationDao) AddFollowing(c context.Context, req *relationgrpc.FollowingReq) error {
	_, err := d.client.AddFollowing(c, req)
	return err
}

func (d *relationDao) DelFollowing(c context.Context, req *relationgrpc.FollowingReq) error {
	_, err := d.client.DelFollowing(c, req)
	return err
}
