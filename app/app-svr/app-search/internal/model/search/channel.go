package search

import (
	mediagrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/media"
	reviewgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/review"
)

type OgvChannelMaterial struct {
	BizId        int64
	BizType      int64
	MediaBizInfo *mediagrpc.MediaBizInfoGetReply
	ReviewInfo   *reviewgrpc.ReviewInfoReply
	AllowReview  *reviewgrpc.AllowReviewReply
}
