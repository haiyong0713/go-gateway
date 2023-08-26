package favorite

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	channelApi "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	"go-common/library/log"
	"go-common/library/sync/errgroup"

	fav "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	dyntopicapi "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	channelModel "go-gateway/app/app-svr/app-channel/interface/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/favorite"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

const (
	_av           = "av"            //视频（ipad没有播单还是视频）
	_playlist     = "playlist"      // 播单
	_bangumi      = "bangumi"       // 追番
	_cinema       = "cinema"        // 追剧
	_comic        = "comic"         // 追漫
	_topic        = "topic"         // 话题
	_specialTopic = "special_topic" // 专题
	_article      = "article"       // 专栏
	_note         = "note"          // 笔记
	_menu         = "menu"          // 歌单
	_pgcMenu      = "pgc_menu"      // 专辑
	_albums       = "albums"        // 相簿
	_product      = "product"       // 商品
	_workshop     = "workshop"      //工房
	_checkin      = "checkin"       //打卡
	_ticket       = "ticket"        // 展演
	_channel      = "channel"       // 频道
	_favorite     = "favorite"
	_cheese       = "cheese" // pugv付费
	_cheeseIPad   = "cheese_ipad"
	_ogvFilm      = "ogv_film"   //二级标题-ogv片单
	_topicAct     = "topic_act"  //二级标题-话题
	_topicList    = "topic_list" //话题列表
	_tpCheese     = 17
	_tpTopic      = 4  //老的话题
	_toOgvFilm    = 18 //片单
	_tpNewTopic   = 27 //新话题
)

var secondTabMap = map[string]*favorite.TabItem{
	_ogvFilm:  {Name: "片单", Uri: "bilibili://pgc/favorite/playlist", Tab: _ogvFilm},
	_topicAct: {Name: "活动", Uri: "bilibili://main/favorite/activity", Tab: _topicAct},
}

var secondTabArr = []string{_ogvFilm, _topicAct}

var tabMap = map[string]*favorite.TabItem{
	_av:           {Name: "视频", Uri: "bilibili://main/favorite/video", Tab: _favorite}, //ipad还是老的
	_playlist:     {Name: "视频", Uri: "bilibili://main/favorite/playlist", Tab: _favorite},
	_bangumi:      {Name: "追番", Uri: "bilibili://pgc/favorite/bangumi", Tab: _bangumi},
	_cinema:       {Name: "追剧", Uri: "bilibili://pgc/favorite/cinema", Tab: _cinema},
	_cheese:       {Name: "课程", Uri: "bilibili://main/favorite/cheese", Tab: _cheese},
	_cheeseIPad:   {Name: "课程", Uri: "bilibili://main/favorite/cheese/pad", Tab: _cheese},
	_comic:        {Name: "追漫", Uri: "bilibili://comic/favorite/list", Tab: _comic},
	_topic:        {Name: "话题", Uri: "bilibili://main/favorite/topic", Tab: _topic},
	_specialTopic: {Name: "专题", Uri: "bilibili://main/favorite/special_topic", Tab: _specialTopic},
	_article:      {Name: "专栏", Uri: "bilibili://column/favorite/article", Tab: _article},
	_note:         {Name: "笔记", Uri: "bilibili://main/favorite/notes", Tab: _note},
	_menu:         {Name: "歌单", Uri: "bilibili://music/favorite/menu", Tab: _menu},
	_pgcMenu:      {Name: "专辑", Uri: "bilibili://music/favorite/album", Tab: _pgcMenu},
	_albums:       {Name: "相簿", Uri: "bilibili://pictureshow/favorite", Tab: _albums},
	_product:      {Name: "商品", Uri: "bilibili://mall/favorite/goods", Tab: _product},
	_ticket:       {Name: "展演", Uri: "bilibili://mall/favorite/ticket", Tab: _ticket},
	_channel:      {Name: "频道", Uri: "bilibili://main/favorite/channel", Tab: _channel},
	_topicList:    {Name: "话题", Uri: "bilibili://main/favorite/topic_list", Tab: _topicList},
	_workshop:     {Name: "工房", Uri: "bilibili://mall/favorite/workshop", Tab: _workshop},
	_checkin:      {Name: "打卡", Uri: "bilibili://main/favorite/checkin", Tab: _checkin},
}

var tabArr = []string{_av, _playlist, _bangumi, _cinema, _cheese, _cheeseIPad, _topicList, _channel, _comic, _topic, _specialTopic, _article, _note, _menu, _pgcMenu, _albums, _product, _workshop, _ticket, _checkin}

// Folder get my favorite.
//
//nolint:gocognit
func (s *Service) Folder(c context.Context, accessKey, actionKey, device, mobiApp, platform string, build int, aid, vmid, mid int64) (rs *favorite.MyFavorite, err error) {
	var pn, ps int = 1, 5
	rs = &favorite.MyFavorite{
		Tab: &favorite.Tab{
			Fav: true,
		},
	}
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		var (
			mediaList bool
			folders   []*favorite.Folder
		)
		plat := model.Plat(mobiApp, device)
		// 双端版本号限制，符合此条件显示为“默认收藏夹”：
		// iPhone <5.36.1(8300) 或iPhone>5.36.1(8300)
		// Android <5360001或Android>5361000
		// 双端版本号限制，符合此条件显示为“默认播单”：
		// iPhone=5.36.1(8300)
		// 5360001 <=Android <=5361000
		if (plat == model.PlatIPhone && build == 8300) || (plat == model.PlatAndroid && build >= 5360001 && build <= 5361000) {
			mediaList = true
		}
		if folders, err = s.favDao.Folders(ctx, mid, vmid, mobiApp, build, mediaList); err != nil {
			log.Error("%+v", err)
			return
		}
		if len(folders) != 0 {
			rs.Favorite = &favorite.FavList{
				Count: len(folders),
				Items: make([]*favorite.FavItem, 0, len(folders)),
			}
			for _, v := range folders {
				fi := &favorite.FavItem{}
				fi.FromFav(v)
				rs.Favorite.Items = append(rs.Favorite.Items, fi)
			}
		}
		return
	})
	g.Go(func() (err error) {
		var topic *fav.UserFolderReply
		if topic, err = s.topicDao.UserFolder(ctx, mid, 4); err != nil {
			log.Error("%+v", err)
			err = nil
			return
		}
		if topic != nil && topic.Res != nil && topic.Res.Count > 0 {
			rs.Tab.Topic = true
		}
		return
	})
	g.Go(func() error {
		article := s.Article(ctx, mid, pn, ps)
		if article != nil && article.Count > 0 {
			rs.Tab.Article = true
		}
		return nil
	})
	g.Go(func() error {
		clips := s.Clips(ctx, mid, accessKey, actionKey, device, mobiApp, platform, build, pn, ps)
		if clips != nil && clips.PageInfo != nil && clips.Count > 0 {
			rs.Tab.Clips = true
		}
		return nil
	})
	g.Go(func() error {
		albums := s.Albums(ctx, mid, accessKey, actionKey, device, mobiApp, platform, build, pn, ps)
		if albums != nil && albums.PageInfo != nil && albums.Count > 0 {
			rs.Tab.Albums = true
		}
		return nil
	})
	g.Go(func() error {
		specil := s.Specil(ctx, accessKey, actionKey, device, mobiApp, platform, build, pn, ps)
		if specil != nil && specil.Count > 0 {
			rs.Tab.Specil = true
		}
		return nil
	})
	g.Go(func() (err error) {
		if mid <= 0 {
			return nil
		}
		var cinemaFav int
		if _, cinemaFav, err = s.bangumiDao.FavDisplay(ctx, mid); err != nil {
			log.Error("%+v", err)
			err = nil
			return
		}
		rs.Tab.Cinema = cinemaFav == 1
		return
	})
	g.Go(func() (err error) {
		fav, err := s.audioDao.Fav(ctx, mid)
		if err != nil {
			log.Error("%+v", err)
			err = nil
			return
		}
		if fav != nil {
			rs.Tab.Menu = fav.Menu
			rs.Tab.PGCMenu = fav.PGCMenu
			rs.Tab.Audios = fav.Song
		}
		return
	})
	g.Go(func() (err error) {
		var ticket int32
		if ticket, err = s.ticketDao.FavCount(ctx, mid); err != nil {
			log.Error("%+v", err)
			err = nil
			return
		}
		if ticket > 0 {
			rs.Tab.Ticket = true
		}
		return
	})
	g.Go(func() (err error) {
		var product int32
		if product, err = s.mallDao.FavCount(ctx, mid); err != nil {
			log.Error("%+v", err)
			err = nil
			return
		}
		if product > 0 {
			rs.Tab.Product = true
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

func (s *Service) FolderVideo(c context.Context, accessKey, actionKey, device, mobiApp, platform, keyword, order string, build, tid, pn, ps int, mid, fid, vmid int64) (folder *favorite.FavideoList) {
	video, err := s.favDao.FolderVideo(c, accessKey, actionKey, device, mobiApp, platform, keyword, order, build, tid, pn, ps, mid, fid, vmid)
	if err != nil {
		folder = &favorite.FavideoList{Items: []*favorite.FavideoItem{}}
		log.Error("%+v", err)
		return
	}
	folder = &favorite.FavideoList{
		//nolint:staticcheck
		Count: video.Total,
		//nolint:staticcheck
		Items: make([]*favorite.FavideoItem, 0, len(video.Archives)),
	}
	//nolint:staticcheck
	if video != nil {
		for _, v := range video.Archives {
			fi := &favorite.FavideoItem{}
			fi.FromFavideo(v)
			folder.Items = append(folder.Items, fi)
		}
	}
	return
}

func (s *Service) Topic(c context.Context, accessKey, actionKey, device, mobiApp, platform string, build, ps, pn int, mid int64) (topic *favorite.TopicList) {
	topics, err := s.topicDao.Topic(c, accessKey, actionKey, device, mobiApp, platform, build, ps, pn, mid)
	if err != nil {
		topic = &favorite.TopicList{Items: []*favorite.TopicItem{}}
		log.Error("%+v", err)
		return
	}
	topic = &favorite.TopicList{
		//nolint:staticcheck
		Count: topics.Total,
		//nolint:staticcheck
		Items: make([]*favorite.TopicItem, 0, len(topics.Lists)),
	}
	//nolint:staticcheck
	if topics != nil {
		for _, v := range topics.Lists {
			fi := &favorite.TopicItem{}
			fi.FromTopic(v)
			topic.Items = append(topic.Items, fi)
		}
	}
	return
}

func (s *Service) Article(c context.Context, mid int64, pn, ps int) (article *favorite.ArticleList) {
	articleTmp, err := s.artDao.Favorites(c, mid, pn, ps)
	if err != nil {
		article = &favorite.ArticleList{Items: []*favorite.ArticleItem{}}
		log.Error("%+v", err)
		return
	}
	article = &favorite.ArticleList{
		Count: len(articleTmp),
		Items: make([]*favorite.ArticleItem, 0, len(articleTmp)),
	}
	if len(articleTmp) != 0 {
		for _, v := range articleTmp {
			fi := &favorite.ArticleItem{}
			fi.FromArticle(c, v)
			article.Items = append(article.Items, fi)
		}
	}
	return
}

// Clips
func (s *Service) Clips(c context.Context, mid int64, accessKey, actionKey, device, mobiApp, platform string, build, pn, ps int) (clips *favorite.ClipsList) {
	clipsTmp, err := s.bplusDao.FavClips(c, mid, accessKey, actionKey, device, mobiApp, platform, build, pn, ps)
	if err != nil {
		clips = &favorite.ClipsList{Items: []*favorite.ClipsItem{}}
		log.Error("%+v", err)
		return
	}
	clips = &favorite.ClipsList{
		//nolint:staticcheck
		PageInfo: clipsTmp.PageInfo,
		//nolint:staticcheck
		Items: make([]*favorite.ClipsItem, 0, len(clipsTmp.List)),
	}
	//nolint:staticcheck
	if clipsTmp != nil {
		for _, v := range clipsTmp.List {
			fi := &favorite.ClipsItem{}
			fi.FromClips(v)
			clips.Items = append(clips.Items, fi)
		}
	}
	return
}

func (s *Service) Albums(c context.Context, mid int64, accessKey, actionKey, device, mobiApp, platform string, build, pn, ps int) (albums *favorite.AlbumsList) {
	albumsTmp, err := s.bplusDao.FavAlbums(c, mid, accessKey, actionKey, device, mobiApp, platform, build, pn, ps)
	if err != nil {
		albums = &favorite.AlbumsList{Items: []*favorite.AlbumItem{}}
		log.Error("%+v", err)
		return
	}
	albums = &favorite.AlbumsList{
		//nolint:staticcheck
		PageInfo: albumsTmp.PageInfo,
		//nolint:staticcheck
		Items: make([]*favorite.AlbumItem, 0, len(albumsTmp.List)),
	}
	//nolint:staticcheck
	if albumsTmp != nil {
		for _, v := range albumsTmp.List {
			fi := &favorite.AlbumItem{}
			fi.FromAlbum(v)
			albums.Items = append(albums.Items, fi)
		}
	}
	return
}

func (s *Service) Specil(c context.Context, accessKey, actionKey, device, mobiApp, platform string, build, pn, ps int) (specil *favorite.SpList) {
	specilTmp, err := s.spDao.Specil(c, accessKey, actionKey, device, mobiApp, platform, build, pn, ps)
	if err != nil {
		specil = &favorite.SpList{Items: []*favorite.SpItem{}}
		log.Error("%+v", err)
		return
	}
	specil = &favorite.SpList{
		//nolint:staticcheck
		Count: len(specilTmp.Items),
		//nolint:staticcheck
		Items: make([]*favorite.SpItem, 0, len(specilTmp.Items)),
	}
	//nolint:staticcheck
	if specilTmp != nil {
		for _, v := range specilTmp.Items {
			fi := &favorite.SpItem{}
			fi.FromSp(v)
			specil.Items = append(specil.Items, fi)
		}
	}
	return
}

func (s *Service) Audio(c context.Context, accessKey string, mid int64, pn, ps int) (audio *favorite.AudioList) {
	audioTmp, err := s.audioDao.FavAudio(c, accessKey, mid, pn, ps)
	if err != nil {
		audio = &favorite.AudioList{Items: []*favorite.AudioItem{}}
		log.Error("%+v", err)
		return
	}
	audio = &favorite.AudioList{
		Count: len(audioTmp),
		Items: make([]*favorite.AudioItem, 0, len(audioTmp)),
	}
	for _, v := range audioTmp {
		fi := &favorite.AudioItem{}
		fi.FromAudio(v)
		audio.Items = append(audio.Items, fi)
	}
	return
}

// SecondTab .
func (s *Service) SecondTab(c context.Context, tab string, mid int64) (rly *favorite.SecondReply) {
	var (
		//nolint:ineffassign
		userFavs   = make(map[int32]int64)
		err        error
		tabDisplay []string
	)
	rly = &favorite.SecondReply{}
	if tab != _specialTopic {
		return
	}
	types := []int32{_tpTopic, _toOgvFilm}
	if userFavs, err = s.favDao.UserFavs(c, types, mid); err != nil {
		log.Error("s.favDao.UserFavs err(%+v)", err)
		return
	}
	if topicCnt, ok := userFavs[_tpTopic]; ok && topicCnt > 0 {
		tabDisplay = append(tabDisplay, _topicAct)
	}
	if spTopicCnt, sok := userFavs[_toOgvFilm]; sok && spTopicCnt > 0 {
		tabDisplay = append(tabDisplay, _ogvFilm)
	}
	for _, t := range secondTabArr {
		for _, dt := range tabDisplay {
			if t == dt {
				rly.Items = append(rly.Items, secondTabMap[t])
			}
		}
	}
	return
}

// Tab fav tab.
//
//nolint:gocognit
func (s *Service) Tab(c context.Context, param *favorite.TabParam) (tab []*favorite.TabItem, err error) {
	var (
		pn, ps     = 1, 5
		tabDisplay = []string{_playlist}
		lock       sync.Mutex
		userFavs   = make(map[int32]int64)
		hasTagList bool
	)
	plat := model.Plat(param.MobiApp, param.Device)
	if model.IsPad(plat) {
		tabDisplay = []string{_av}
	}
	g, ctx := errgroup.WithContext(c)
	if !model.IsOverseas(plat) {
		g.Go(func() (err error) {
			types := []int32{_tpCheese, _tpTopic, _toOgvFilm, _tpNewTopic}
			if userFavs, err = s.favDao.UserFavs(ctx, types, param.Mid); err != nil {
				log.Error("s.favDao.UserFavs err(%+v)", err)
				err = nil
			}
			return
		})
		g.Go(func() error {
			// 动态老话题没进收藏夹 是用tag的订阅关系
			args := &dyntopicapi.SubTopicsReq{Uid: param.Mid}
			tagList, err := s.topicDao.SubTopics(ctx, args)
			if err != nil {
				log.Error("s.topicDao.SubTopics args=%+v, error=%+v", args, err)
				return nil
			}
			if len(tagList.Topics) > 0 {
				hasTagList = true
			}
			return nil
		})
	}
	if param.TeenagersMode == 0 && param.LessonsMode == 0 {
		g.Go(func() (err error) {
			var bangumiFav, cinemaFav int
			if bangumiFav, cinemaFav, err = s.bangumiDao.FavDisplay(ctx, param.Mid); err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if bangumiFav == 1 {
				lock.Lock()
				tabDisplay = append(tabDisplay, _bangumi)
				lock.Unlock()
			}
			if cinemaFav == 1 {
				lock.Lock()
				tabDisplay = append(tabDisplay, _cinema)
				lock.Unlock()
			}
			return
		})
		if (plat == model.PlatAndroid && param.Build > s.c.FavBuildLimit.ComicAndroid) || (plat == model.PlatIPhone && param.Build > s.c.FavBuildLimit.ComicIOS) {
			g.Go(func() (err error) {
				comics, err := s.comicDao.FavComics(ctx, param.Mid, pn, ps)
				if err != nil {
					log.Error("%v", err)
					err = nil
					return
				}
				if len(comics) > 0 {
					lock.Lock()
					tabDisplay = append(tabDisplay, _comic)
					lock.Unlock()
				}
				return
			})
		}
	}
	if !model.IsPad(plat) || feature.GetBuildLimit(c, "service.hdFavoriteArticle", nil) {
		g.Go(func() error {
			article := s.Article(ctx, param.Mid, pn, ps)
			if article != nil && article.Count > 0 {
				lock.Lock()
				tabDisplay = append(tabDisplay, _article)
				lock.Unlock()
			}
			return nil
		})
		if !model.IsOverseas(plat) {
			g.Go(func() error {
				albums := s.Albums(ctx, param.Mid, param.AccessKey, param.ActionKey, param.Device, param.MobiApp, param.Platform, param.Build, pn, ps)
				if albums != nil && albums.PageInfo != nil && albums.Count > 0 {
					lock.Lock()
					tabDisplay = append(tabDisplay, _albums)
					lock.Unlock()
				}
				return nil
			})
			g.Go(func() (err error) {
				favShow, err := s.audioDao.Fav(ctx, param.Mid)
				if err != nil {
					log.Error("%+v", err)
					err = nil
					return
				}
				if favShow != nil {
					favShowMenu := false
					// 安卓粉>=6.57  ios粉>=6.56
					if (model.IsIPhone(plat) && param.Build >= 65600000) || (model.IsAndroidPick(plat) && param.Build >= 6570000) {
						favShowMenu = favShow.Menu || favShow.HasMenuCreated || favShow.HasCollection
					} else {
						favShowMenu = favShow.Menu
					}
					if favShowMenu {
						lock.Lock()
						tabDisplay = append(tabDisplay, _menu)
						lock.Unlock()
					}
				}
				return
			})
			if param.TeenagersMode == 0 && param.LessonsMode == 0 {
				g.Go(func() (err error) {
					var ticket int32
					if ticket, err = s.ticketDao.FavCount(ctx, param.Mid); err != nil {
						log.Error("%+v", err)
						err = nil
						return
					}
					if ticket > 0 {
						lock.Lock()
						tabDisplay = append(tabDisplay, _ticket)
						lock.Unlock()
					}
					return
				})
				g.Go(func() (err error) {
					var product int32
					if product, err = s.mallDao.FavCount(ctx, param.Mid); err != nil {
						log.Error("%+v", err)
						err = nil
						return
					}
					if product > 0 {
						lock.Lock()
						tabDisplay = append(tabDisplay, _product)
						lock.Unlock()
					}
					return
				})
				if showWorkshop(ctx) {
					g.Go(func() (err error) {
						var workshop int64
						if workshop, err = s.workshopDao.FavCount(ctx, param.Mid); err != nil {
							log.Error("%+v", err)
							err = nil
							return
						}
						if workshop > 0 {
							lock.Lock()
							tabDisplay = append(tabDisplay, _workshop)
							lock.Unlock()
						}
						return
					})
				}
				g.Go(func() (err error) {
					var checkin int32
					if checkin, err = s.checkinDao.CheckinCount(ctx, param.Mid); err != nil {
						log.Error("%+v", err)
						err = nil
						return
					}
					if checkin > 0 {
						lock.Lock()
						tabDisplay = append(tabDisplay, _checkin)
						lock.Unlock()
					}
					return
				})

			}
		}
	}
	if (plat == model.PlatAndroid && param.Build >= s.c.FavBuildLimit.NoteAndroid) || (plat == model.PlatIPhone && param.Build >= s.c.FavBuildLimit.NoteIOS) {
		g.Go(func() error {
			noteCnt, err := s.noteDao.NoteCount(ctx, param.Mid)
			if err != nil {
				log.Warn("noteWarn Tab err(%+v)", err)
				return nil
			}
			if noteCnt > 0 {
				lock.Lock()
				tabDisplay = append(tabDisplay, _note)
				lock.Unlock()
			}
			return nil
		})
	}
	if (plat == model.PlatIPhone && param.Build > s.c.FavBuildLimit.ChannelTabIOSBuild) || (plat == model.PlatAndroid && param.Build > s.c.FavBuildLimit.ChannelTabAndroidBuild) {
		g.Go(func() error {
			reply, err := s.channelDao.ChannelFav(ctx, param.Mid, 1, "")
			if err != nil {
				log.Error("s.channelDao.ChannelFav error(%+v) mid(%d)", err, param.Mid)
				return nil
			}
			if len(reply.List) > 0 {
				lock.Lock()
				tabDisplay = append(tabDisplay, _channel)
				lock.Unlock()
			}
			return nil
		})
	}
	//nolint:errcheck
	g.Wait()
	if param.Filtered != "1" && !model.IsPad(plat) && !model.IsOverseas(plat) {
		if (plat == model.PlatIPhone && param.Build > 8961) || (plat == model.PlatAndroid && param.Build >= 5510000) {
			if topicCnt, ok := userFavs[_tpTopic]; ok && topicCnt > 0 {
				tabDisplay = append(tabDisplay, _specialTopic)
			} else if spTopicCnt, sok := userFavs[_toOgvFilm]; sok && spTopicCnt > 0 {
				tabDisplay = append(tabDisplay, _specialTopic)
			}
		} else {
			if topicCnt, ok := userFavs[_tpTopic]; ok && topicCnt > 0 {
				tabDisplay = append(tabDisplay, _topic)
			}
		}
	}
	if isTopicListFavTabShow(userFavs, hasTagList, plat, param.Build) {
		tabDisplay = append(tabDisplay, _topicList)
	}
	if param.TeenagersMode == 0 && param.LessonsMode == 0 && s.cheeseDao.HasCheese(plat, param.Build, true) {
		cheeseName := _cheese
		if model.IsPad(plat) {
			cheeseName = _cheeseIPad
		}
		if cheeseCnt, ok := userFavs[_tpCheese]; ok && cheeseCnt > 0 {
			tabDisplay = append(tabDisplay, cheeseName)
		}
	}
	//国际版收藏页仅保留视频、追番、追剧、专栏
	for _, t := range tabArr {
		for _, dt := range tabDisplay {
			if t == dt {
				tab = append(tab, tabMap[t])
			}
		}
	}
	return
}

// ios拆分出去了，目前只用判断安卓版本
func showWorkshop(ctx context.Context) bool {
	return pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", int64(6720000))
	}).MustFinish()

}

func (s *Service) ChannelFav(ctx context.Context, mid, ps int64, offset, spmid string) (*favorite.ChannelFav, error) {
	reply, err := s.channelDao.ChannelFav(ctx, mid, ps, offset)
	if err != nil {
		log.Error("s.channelDao.ChannelFav error(%+v) mid(%d)", err, mid)
		return nil, err
	}
	channelFav := &favorite.ChannelFav{
		HasMore:      reply.HasMore,
		Offest:       reply.Offset,
		Total:        reply.Total,
		ViewMoreLink: s.c.Custom.ChannelLink,
	}
	//版本判断
	var isHightBuild bool
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsMobiAppIPhone().And().Build(">=", s.c.BuildLimit.OGVChanIOSBuild)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", s.c.BuildLimit.OGVChanAndroidBuild)
	}).FinishOr(false) {
		isHightBuild = true
	}
	for _, v := range reply.List {
		if v == nil {
			continue
		}
		// 电影频道
		url := channelModel.FillURI(channelModel.GotoChannelNew, strconv.FormatInt(v.Cid, 10), 0, 0, 0, nil)
		if v.BizType == channelApi.ChannelBizlType_MOVIE && isHightBuild {
			url = channelModel.FillURI(channelModel.GotoChannelMedia, fmt.Sprintf("%d", v.Cid), 0, 0, 0, model.ChannelHandler(fmt.Sprintf("biz_id=%d&biz_type=0&source=%s", v.Cid, spmid)))
		}
		temp := &favorite.SubChannel{
			Cid:           v.Cid,
			Cname:         v.Cname,
			FeaturedCnt:   v.FeaturedCnt,
			SubscribedCnt: v.SubscribedCnt,
			Icon:          v.Icon,
			Url:           url,
		}
		channelFav.List = append(channelFav.List, temp)
	}
	return channelFav, nil
}

func isTopicListFavTabShow(userFavs map[int32]int64, hasTagList bool, plat int8, build int) bool {
	if (model.IsIPhone(plat) && build < 64700000) || (model.IsAndroid(plat) && build < 6470000) || (model.IsIPadHD(plat) && build < 32900000) {
		return false
	}
	if newTopicCnt, ok := userFavs[_tpNewTopic]; ok && newTopicCnt > 0 {
		return true
	}
	return hasTagList
}
