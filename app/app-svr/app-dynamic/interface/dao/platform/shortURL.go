package platform

import (
	"context"

	"go-common/library/log"

	grpcShortURL "git.bilibili.co/bapis/bapis-go/platform/interface/shorturl"
)

func (d *Dao) ShortUrls(c context.Context, urls []string) (map[string]string, error) {
	var args = &grpcShortURL.ShortUrlsReq{ShortUrls: urls}
	resTmp, err := d.grpcClientShortURL.ShortUrls(c, args)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var res = make(map[string]string)
	for _, reTmp := range resTmp.GetDetails() {
		if resTmp == nil {
			continue
		}
		res[reTmp.ShortUrl] = reTmp.OriginUrl
	}
	return res, nil
}
