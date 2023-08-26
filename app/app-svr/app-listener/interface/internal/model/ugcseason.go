package model

import (
	ffavSvc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	ugcSeasonSvc "git.bilibili.co/bapis/bapis-go/ugc-season/service"
)

type UgcSeasonDetail struct {
	Season   *ugcSeasonSvc.Season
	Sections []*ugcSeasonSvc.Section
}

func (us UgcSeasonDetail) ToModelFavItemDetails(fmeta FavFolderMeta) (ret []FavItemDetail) {
	var idx int32
	for _, sec := range us.Sections {
		for _, ep := range sec.GetEpisodes() {
			ret = append(ret, FavItemDetail{
				Item: &ffavSvc.ModelFavorite{
					Oid:   ep.Aid,
					Type:  FavTypeVideo,
					Mid:   fmeta.Mid,
					Ctime: ep.GetArc().GetPubDate().Time().Unix(),
					Index: idx,
				},
				AuthorMid:  ep.GetArc().GetAuthor().GetMid(),
				AuthorName: ep.GetArc().GetAuthor().GetName(),
				ViewCnt:    ep.GetArc().GetStat().GetView(),
				ReplyCnt:   ep.GetArc().GetStat().GetReply(),
				Title:      ep.Title,
				Cover:      ep.GetArc().GetPic(),
			})
			idx++
		}
	}
	return
}
