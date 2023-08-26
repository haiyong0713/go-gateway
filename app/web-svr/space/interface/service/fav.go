package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"
	"go-gateway/pkg/idsafe/bvid"

	noteapi "git.bilibili.co/bapis/bapis-go/app/note/service"
	api "git.bilibili.co/bapis/bapis-go/cheese/service/auth"
	favapi "git.bilibili.co/bapis/bapis-go/community/service/favorite"

	"go-common/library/sync/errgroup.v2"
)

const (
	_typeFavAlbum   = 2
	_typeFavArchive = 2
	//_typeFavOldPlaylist = 5
	_typeFavPlaylist  = 6
	_typeFavTopic     = 4
	_typeFavArticle   = 1
	_typeFavTopicList = 27
)

var _emptyArcFavFolder = make([]*model.VideoFolder, 0)

// playlist,topic,article,pugv
var _userFavTypes = []int32{_typeFavPlaylist, _typeFavTopic, _typeFavArticle, _typeFavTopicList}

// FavNav get fav info.
func (s *Service) FavNav(c context.Context, mid int64, vmid int64) (res *model.FavNav, err error) {
	group := errgroup.WithContext(c)
	res = new(model.FavNav)
	if mid == vmid || s.privacyCheck(c, vmid, model.PcyFavVideo) == nil {
		// arc
		group.Go(func(ctx context.Context) error {
			if data, e := s.favClient.CntUserFolders(ctx, &favapi.CntUserFoldersReq{Typ: _typeFavArchive, Mid: mid, Vmid: vmid}); e != nil {
				log.Error("s.favClient.UserFavs %d error(%v)", vmid, e)
			} else if data != nil {
				res.Arc = int64(data.Count)
			}
			return nil
		})
	}
	// playlist,topic,article,pugv
	group.Go(func(ctx context.Context) error {
		if data, e := s.favClient.UserFavs(ctx, &favapi.UserFavsReq{Mid: vmid, Types: _userFavTypes}); e != nil {
			log.Error("s.favClient.UserFavs %d error(%v)", vmid, e)
		} else if data != nil {
			res.Playlist = data.Favs[_typeFavPlaylist]
			res.Topic = data.Favs[_typeFavTopic]
			res.Article = data.Favs[_typeFavArticle]
			res.TopicList = data.Favs[_typeFavTopicList]
		}
		return nil
	})
	// pugv
	group.Go(func(ctx context.Context) error {
		if data, e := s.pugvAuthClient.FavoriteCount(ctx, &api.FavoriteCountReq{Mid: vmid}); e != nil {
			log.Error("s.favClient.UserFavs %d error(%v)", vmid, e)
		} else if data != nil {
			res.Pugv = int64(data.Total)
		}
		return nil
	})
	// album
	group.Go(func(ctx context.Context) error {
		if albumCount, e := s.dao.LiveFavCount(ctx, vmid, _typeFavAlbum); e != nil {
			log.Error("s.dao.LiveFavCount(%d,%d) error(%v)", vmid, _typeFavAlbum, e)
		} else if albumCount > 0 {
			res.Album = albumCount
		}
		return nil
	})
	// note
	group.Go(func(ctx context.Context) error {
		noteCount, e := s.noteClient.NoteCount(ctx, &noteapi.NoteCountReq{Mid: vmid})
		if e != nil {
			log.Error("s.noteClient.NoteCount(%d) error(%v)", vmid, e)
			return nil
		}
		if noteCount == nil {
			log.Error("s.noteClient.NoteCount(%d) empty result", vmid)
			return nil
		}
		if noteCount.NoteCount > 0 {
			res.Note = noteCount.NoteCount
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	res.Archive = _emptyArcFavFolder
	return
}

// FavArchive get favorite archive.
func (s *Service) FavArchive(c context.Context, mid int64, arg *model.FavArcArg) (res *model.SearchArchive, err error) {
	if mid != arg.Vmid {
		if err = s.privacyCheck(c, arg.Vmid, model.PcyFavVideo); err != nil {
			return
		}
	}
	return s.dao.FavArchive(c, mid, arg)
}

// nolint:gomnd
func (s *Service) FavSeasonList(ctx context.Context, seasonID int64) (*model.FavSeasonList, error) {
	reply, err := s.dao.SeasonView(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	season := reply.GetSeason()
	acc, err := s.dao.AccountInfo(ctx, season.GetMid())
	if err != nil {
		log.Error("%+v", err)
	}
	info := &model.FavInfo{
		ID: season.GetID(),
		// 是否基础合集
		// 0:精选合集 1:基础合集
		SeasonType: (season.Attribute >> 2) & int64(1),
		Title:      season.GetTitle(),
		Cover:      season.GetCover(),
		Upper: &model.FavUpper{
			Mid:  season.GetMid(),
			Name: acc.GetName(),
		},
		CntInfo: &model.FavCntInfo{
			Play: int64(season.GetStat().View),
		},
		MediaCount: season.GetEpCount(),
	}
	var medias []*model.FavMedia
	for _, sections := range reply.GetSections() {
		for _, val := range sections.GetEpisodes() {
			bid, _ := bvid.AvToBv(val.GetAid())
			media := &model.FavMedia{
				ID:       val.GetAid(),
				Title:    val.GetTitle(),
				Cover:    val.GetArc().GetPic(),
				Duration: val.GetPage().GetDuration(),
				Pubtime:  int64(val.GetArc().GetPubDate()),
				Bvid:     bid,
				Upper: &model.FavUpper{
					Mid:  val.GetArc().GetAuthor().GetMid(),
					Name: val.GetArc().GetAuthor().GetName(),
				},
				CntInfo: &model.FavCntInfo{
					Collect: int64(val.GetArc().GetStat().GetFav()),
					Play:    int64(val.GetArc().GetStat().GetView()),
				},
			}
			medias = append(medias, media)
		}
	}
	return &model.FavSeasonList{
		Info:   info,
		Medias: medias,
	}, nil
}
