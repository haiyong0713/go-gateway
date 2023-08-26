package dynamicV2

import (
	relagrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	dynCheeseGrpc "git.bilibili.co/bapis/bapis-go/cheese/service/dynamic"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	pgcFollowGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
)

func GetAttentionsParams(mid int64, following *relagrpc.FollowingsReply, pgc *pgcFollowGrpc.MyRelationsReply, cheese *dynCheeseGrpc.MyPaidReply, ugcSeason *favgrpc.BatchFavsReply, batchListFavorite []int64) *dyncommongrpc.AttentionInfo {
	ret := &dyncommongrpc.AttentionInfo{}
	if following != nil {
		for _, item := range following.FollowingList {
			ret.AttentionList = append(ret.AttentionList, &dyncommongrpc.Attention{
				Uid:       item.Mid,
				UidType:   1,
				IsSpecial: Int32ToBool(item.Special),
			})
		}
	}
	if pgc != nil {
		for _, item := range pgc.Relations {
			ret.AttentionList = append(ret.AttentionList, &dyncommongrpc.Attention{
				Uid:     int64(item.SeasonId),
				UidType: 2,
			})
		}
	}
	if cheese != nil {
		for _, item := range cheese.SeasonIds {
			ret.AttentionList = append(ret.AttentionList, &dyncommongrpc.Attention{
				Uid:     int64(item),
				UidType: 4,
			})
		}
	}
	if ugcSeason != nil {
		for _, item := range ugcSeason.GetOids() {
			ret.AttentionList = append(ret.AttentionList, &dyncommongrpc.Attention{
				Uid:     item,
				UidType: 5,
			})
		}
	}
	for _, v := range batchListFavorite {
		ret.AttentionList = append(ret.AttentionList, &dyncommongrpc.Attention{
			Uid:     v,
			UidType: 6,
		})
	}
	// 赋自己
	ret.AttentionList = append(ret.AttentionList, &dyncommongrpc.Attention{
		Uid:       mid,
		UidType:   1,
		IsSpecial: false,
	})
	return ret
}
