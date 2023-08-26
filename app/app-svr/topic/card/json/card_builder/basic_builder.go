package cardbuilder

import (
	"strconv"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
)

func ConstructDynCardBasic(dynBriefs *dynmdlV2.Dynamic) *jsonwebcard.Basic {
	return &jsonwebcard.Basic{
		RidStr:       strconv.FormatInt(dynBriefs.Rid, 10),
		LikeShowIcon: constructLikeShowIcon(dynBriefs.Extend),
	}
}

func constructLikeShowIcon(extend *dynmdlV2.Extend) *jsonwebcard.LikeIcon {
	if extend == nil || extend.LikeIcon == nil {
		return nil
	}
	return &jsonwebcard.LikeIcon{
		NewIconId: extend.LikeIcon.NewIconID,
		StartUrl:  extend.LikeIcon.Begin,
		ActionUrl: extend.LikeIcon.Proc,
		EndUrl:    extend.LikeIcon.End,
	}
}
