package model

import (
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"

	accSvc "git.bilibili.co/bapis/bapis-go/account/service"
)

type MemberInfo struct {
	*accSvc.Info
	*accSvc.ProfileStatReply
}

func (mi MemberInfo) ToV1FavFolderAuthor() *v1.FavFolderAuthor {
	if mi.Info == nil {
		return nil
	}
	return &v1.FavFolderAuthor{Mid: mi.Mid, Name: mi.Name}
}

func (mi MemberInfo) ToV1MedialistUpInfo() *v1.MedialistUpInfo {
	if mi.ProfileStatReply == nil {
		return nil
	}
	return &v1.MedialistUpInfo{
		Mid: mi.Profile.GetMid(), Avatar: mi.Profile.GetFace(), Name: mi.Profile.GetName(),
		Fans: mi.ProfileStatReply.GetFollower(),
	}
}
