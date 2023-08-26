package bangumi

import (
	"context"

	pgcmedia "git.bilibili.co/bapis/bapis-go/pgc/servant/media"
	pgcfollow "git.bilibili.co/bapis/bapis-go/pgc/service/follow/media"
	pgcreview "git.bilibili.co/bapis/bapis-go/pgc/service/review"
)

func (d *Dao) GetMediaBizInfoByMediaBizId(c context.Context, mediaBizId int64) (*pgcmedia.MediaBizInfoGetReply, error) {
	return d.pgcMediaClient.GetMediaBizInfoByMediaBizId(c, &pgcmedia.MediaBizInfoGetReq{MediaBizId: mediaBizId})
}

func (d *Dao) ReviewInfo(c context.Context, mediaBizId int64) (*pgcreview.ReviewInfoReply, error) {
	return d.pgcReviewClient.ReviewInfo(c, &pgcreview.ReviewInfoReq{MediaId: mediaBizId})
}

func (d *Dao) AllowReview(c context.Context, mediaBizId int32) (*pgcreview.AllowReviewReply, error) {
	return d.pgcReviewClient.AllowReview(c, &pgcreview.AllowReviewReq{MediaId: mediaBizId})
}

func (d *Dao) MediaStatus(c context.Context, mediaBizId int32, mid int64) (*pgcreview.MediaStatusReply, error) {
	return d.pgcReviewUserClient.MediaStatus(c, &pgcreview.MediaStatusReq{MediaId: mediaBizId, Mid: mid})
}

func (d *Dao) StatusByMid(c context.Context, mid, mediaBizId int64) (*pgcfollow.MediaFollowStatusByMidReply, error) {
	return d.pgcFollowClient.StatusByMid(c, &pgcfollow.MediaFollowStatusByMidReq{Mid: mid, MediaId: []int32{int32(mediaBizId)}})
}

func (d *Dao) AddMediaFollow(c context.Context, mid, mediaBizId int64) error {
	_, err := d.pgcFollowClient.AddMediaFollow(c, &pgcfollow.MediaFollowReq{Mid: mid, MediaId: int32(mediaBizId), FollowStatus: 1})
	return err
}
