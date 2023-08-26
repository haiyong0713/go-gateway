package model

const (
	// 用户关系：FollowingReply.Attribute
	RelationPrivateFollow = 1   //悄悄关注
	RelationFollow        = 2   //关注
	RelationFriend        = 6   //好友
	RelationBlocked       = 128 //拉黑
	// 认证角色：OfficialInfo.Role
	RoleUp       = 1 //UP主认证
	RoleVertical = 7 //垂直领域认证
)
