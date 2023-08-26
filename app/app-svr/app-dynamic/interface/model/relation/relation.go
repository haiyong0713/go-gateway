package relation

import (
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

// 互相关注关系转换
func RelationChange(upMid int64, relations map[int64]*relationgrpc.InterrelationReply) (r *api.Relation) {
	const (
		// state使用
		_statenofollow      = 1
		_statefollow        = 2
		_statefollowed      = 3
		_statemutualConcern = 4
		_specialFollow      = 5
		// 关注关系
		_follow = 1
	)
	r = &api.Relation{
		Status: _statenofollow,
		Title:  "未关注",
	}
	rel, ok := relations[upMid]
	if !ok {
		return
	}
	switch rel.Attribute {
	// nolint:gomnd
	case 2, 6: // 用户关注UP主
		r.Status = _statefollow
		r.IsFollow = _follow
		r.Title = "已关注"
	}
	if rel.IsFollowed { // UP主关注用户
		r.Status = _statefollowed
		r.IsFollowed = _follow
		r.Title = "被关注"
	}
	if r.IsFollow == _follow && r.IsFollowed == _follow { // 用户和UP主互相关注
		r.Status = _statemutualConcern
		r.Title = "互相关注"
	}
	if rel.Special == 1 {
		r.Status = _specialFollow
		r.Title = "特别关注"
	}
	return
}
