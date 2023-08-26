package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
)

const _tagHotsURI = "/x/internal/tag/hots"

func (d *dao) TagHots(ctx context.Context, rid int64) ([]int64, error) {
	params := url.Values{}
	params.Set("rid", strconv.FormatInt(rid, 10))
	var res struct {
		Code int `json:"code"`
		Data []*struct {
			Rid  int64 `json:"rid"`
			Tags []*struct {
				TagID int64 `json:"tag_id"`
			} `json:"tags"`
		} `json:"data"`
	}
	if err := d.httpR.Get(ctx, d.tagHotsURL, "", params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, ecode.Int(res.Code)
	}
	var tagIDs []int64
	for _, v := range res.Data {
		if v != nil && v.Rid == rid {
			for _, tagItem := range v.Tags {
				if tagItem != nil {
					tagIDs = append(tagIDs, tagItem.TagID)
				}
			}
		}
	}
	return tagIDs, nil
}
