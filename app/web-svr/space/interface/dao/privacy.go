package dao

import (
	"context"

	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	liveusergrpc "git.bilibili.co/bapis/bapis-go/live/xuserex/v1"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	spacegrpc "git.bilibili.co/bapis/bapis-go/space/service/v1"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

// FollowingPrivacy get following privacy from account
func (d *Dao) FollowingPrivacy(c context.Context, mid int64) (*relationgrpc.PrivacySettingReply, error) {
	followingSetting, err := d.relGRPC.PrivacySetting(c, &relationgrpc.MidReq{Mid: mid})
	if err != nil {
		return nil, err
	}
	return followingSetting, nil
}

// UpdateFollowingPrivacy update following type privacy setting
func (d *Dao) UpdateFollowingPrivacy(c context.Context, mid int64, state bool) error {
	if _, err := d.relGRPC.UpdatePrivacySetting(c, &relationgrpc.UpdatePrivacySettingReq{
		Mid:        mid,
		SwitchType: relationgrpc.PrivacySwitchType_Following,
		Switch:     state,
	}); err != nil {
		return err
	}
	return nil
}

// LiveUserMedalStatusPrivacy get all live user privacy
func (d *Dao) LiveUserMedalStatusPrivacy(c context.Context, mid int64) (*liveusergrpc.UserPlugsRes, error) {
	arg := &liveusergrpc.UserPlugsReq{
		Uid: mid,
	}
	liveMedalStatus, err := d.liveUserGRPC.ConfigPlugs(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", arg)
	}
	return liveMedalStatus, nil
}

// UpdateLiveUserMedalStatusPrivacy update live user privacy
func (d *Dao) UpdateLiveUserMedalStatusPrivacy(c context.Context, mid int64, key string, status int64) error {
	arg := &liveusergrpc.EditPlugsReq{
		Uid:    mid,
		Key:    key,
		Status: status,
	}
	if _, err := d.liveUserGRPC.EditPlugs(c, arg); err != nil {
		return errors.Wrapf(err, "%+v", arg)
	}
	return nil
}

func (d *Dao) AccMemberPrivacySetting(ctx context.Context, args *membergrpc.MidReq) (*membergrpc.PrivacySettingReply, error) {
	return d.memberGRPC.PrivacySetting(ctx, args)
}

func (d *Dao) UpdateAccMemberPrivacySetting(ctx context.Context, args *membergrpc.UpdatePrivacySettingReq) (*membergrpc.EmptyStruct, error) {
	return d.memberGRPC.UpdatePrivacySetting(ctx, args)
}

func (d *Dao) HasNFT(ctx context.Context, args *gallerygrpc.MidReq) (*gallerygrpc.OwnerReply, error) {
	return d.galleryClient.HasNFT(ctx, args)
}

func (d *Dao) NftGalleryPrivacySetting(ctx context.Context, args *gallerygrpc.MidReq) (*gallerygrpc.PrivacySettingReply, error) {
	return d.galleryClient.GetPrivacySetting(ctx, args)
}

func (d *Dao) UpdateNftGalleryPrivacySetting(ctx context.Context, args *gallerygrpc.UpdatePrivacySettingReq) (*empty.Empty, error) {
	return d.galleryClient.UpdatePrivacySetting(ctx, args)
}

func (d *Dao) PrivacySetting(ctx context.Context, mid int64, option []spacegrpc.PrivacyOption) (map[string]*spacegrpc.PrivacySetting, error) {
	in := &spacegrpc.PrivacySettingReq{Mid: mid, PrivacyOption: option}
	reply, err := d.spaceClient.PrivacySetting(ctx, in)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", in)
	}
	return reply.GetSettings(), nil
}

func (d *Dao) UpdatePrivacySetting(ctx context.Context, mid int64, settings []*spacegrpc.PrivacySetting) error {
	in := &spacegrpc.UpdatePrivacysReq{Mid: mid, Settings: settings}
	if _, err := d.spaceClient.UpdatePrivacySetting(ctx, in); err != nil {
		return errors.Wrapf(err, "%+v", in)
	}
	return nil
}
