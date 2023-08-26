package dao

import (
	"context"

	pgcfollowgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
)

type pgcfollowDao struct {
	client pgcfollowgrpc.FollowClient
}

func (d *pgcfollowDao) AddFollow(c context.Context, req *pgcfollowgrpc.FollowReq) error {
	if _, err := d.client.AddFollow(c, req); err != nil {
		return err
	}
	return nil
}

func (d *pgcfollowDao) DeleteFollow(c context.Context, req *pgcfollowgrpc.FollowReq) error {
	if _, err := d.client.DeleteFollow(c, req); err != nil {
		return err
	}
	return nil
}

func (d *pgcfollowDao) StatusByMid(c context.Context, req *pgcfollowgrpc.FollowStatusByMidReq) (*pgcfollowgrpc.FollowStatusByMidReply, error) {
	rly, err := d.client.StatusByMid(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
