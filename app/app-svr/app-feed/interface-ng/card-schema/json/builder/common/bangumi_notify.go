package jsoncommon

import (
	"strings"

	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
)

type BangumiNotify struct{}

func (BangumiNotify) ConstructRemindCover(in *bangumi.Remind) string {
	if len(in.List) <= 0 {
		return ""
	}
	cover := in.List[0].SquareCover
	if cover == "" {
		cover = in.List[0].Cover
	}
	return cover
}

func (BangumiNotify) ConstructRemindURI(in *bangumi.Remind) string {
	uriStr := in.List[0].Uri
	uri := strings.Split(uriStr, "?")
	if len(uri) == 1 {
		uriStr = uriStr + "?from=21"
		return uriStr
	}
	if len(uri) > 1 {
		uriStr = uriStr + "&from=21"
		return uriStr
	}
	return uriStr
}
