package archive

import (
	"context"
	"fmt"
	ugcmdl "go-gateway/app/app-svr/ugc-season/service/api"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model/archive"
	"go-gateway/pkg/idsafe/bvid"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	mngActApi "git.bilibili.co/bapis/bapis-go/manager/service/active"
)

// Arc is
func (d *Dao) Arc(c context.Context, aid int64) (*api.Arc, error) {
	a, err := d.arcCache(c, aid)
	if err != nil {
		a, err = func() (*api.Arc, error) {
			setCache := false
			if err == redis.ErrNil {
				d.infoProm.Incr("ArcMissed")
				setCache = true
			} else {
				log.Error("d.arcCache err(%+v) aid(%d)", err, aid)
			}
			if a, err = d.getArcFromTaishan(c, aid); err == nil {
				if setCache {
					var ca = &api.Arc{}
					*ca = *a
					d.addCache(func() {
						_ = d.setArcRdsCache(context.Background(), ca)
					})
					d.infoProm.Incr("SetArcCacheByTaishan")
				}
				return a, nil
			}
			var ip string
			if a, ip, err = d.RawArc(c, aid); err != nil {
				log.Error("d.RawArc err(%+v) aid(%d)", err, aid)
				return nil, err
			}
			if a == nil {
				return nil, ecode.NothingFound
			}
			d.fillArc(c, a, ip, setCache)
			return a, nil
		}()
		if err != nil {
			log.Error("%+v", err)
			return nil, err
		}
	}
	// 历史缓存内存在账号的昵称与头像，不再使用，实时读取账号信息前将其置空
	a.Author.Name = ""
	a.Author.Face = ""
	func() { // set type name
		typeName, ok := d.tNamem[int16(a.TypeID)]
		if !ok {
			log.Error("日志报警 Arc接口 aid(%d) typeID(%d) not exist", a.Aid, a.TypeID)
			return
		}
		a.TypeName = typeName
	}()
	eg2 := errgroup.WithContext(c)
	var infoReply *accapi.InfoReply
	if a.Author.Mid > 0 {
		d.infoProm.Incr("ArcAccount")
		//获取账户信息
		eg2.Go(func(ctx context.Context) error {
			var err error
			if infoReply, err = d.acc.Info3(ctx, &accapi.MidReq{Mid: a.Author.Mid}); err != nil {
				// account error时，不影响稿件信息返回，仅不展示up主昵称和头像
				log.Error("日志报警 Arc接口 aid(%d) d.acc.Info3 error(%+v)", a.Aid, err)
			}
			return nil
		})
	}
	//获取inner attr
	var arcIn *api.ArcInternal
	eg2.Go(func(ctx context.Context) error {
		incRly, e := d.ArcsInner(ctx, []int64{aid})
		if e != nil {
			//  error时，不影响稿件信息返回,仅autoplay不准确
			log.Error("日志报警 Arc接口 aid(%d) d.ArcsInner error(%+v)", a.Aid, err)
			return nil
		}
		if _, ok := incRly[aid]; ok {
			arcIn = incRly[aid]
		}
		return nil
	})
	_ = eg2.Wait() //错误可降级
	// set account info
	if infoReply.GetInfo() != nil {
		a.Author.Name = infoReply.GetInfo().GetName()
		a.Author.Face = infoReply.GetInfo().GetFace()
	}
	//set autoplay
	a.Rights.Autoplay = api.CalcAutoplayV2(a, arcIn)
	// 生成短链（老字段可能有使用过暂不下线）
	a.ShortLink = d.handleShortLink(aid)
	a.ShortLinkV2 = d.handleShortLink(aid)
	return a, nil
}

// nolint:gocognit
func (d *Dao) Arcs(c context.Context, aids []int64, mid int64, mobiApp, device string) (map[int64]*api.Arc, map[int64]*api.ArcInternal, error) {
	am := make(map[int64]*api.Arc)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		var (
			missRdsAid, missTsAid []int64
			saveTs, saveDB        map[int64]*api.Arc
			setCache              = true
			ips                   map[int64]string
		)
		if am, missRdsAid, err = d.arcCaches(ctx, aids); err != nil {
			log.Error("d.arcCaches Redis aids(%+v) err(%+v)", aids, err)
			d.infoProm.Incr("ArcsCacheErr")
			am = map[int64]*api.Arc{}
			setCache = false
		}

		if len(missRdsAid) == 0 {
			return nil
		}
		if saveTs, missTsAid, err = d.batchGetArcFromTaishan(ctx, missRdsAid); err != nil {
			log.Error("d.batchGetArcFromTaishan missRdsAid(%+v) err(%+v)", missRdsAid, err)
			if ecode.EqualError(ecode.NothingFound, err) {
				d.infoProm.Incr("ArcsCacheByTaishanNothingFound")
			} else {
				d.infoProm.Incr("ArcsCacheByTaishanErr")
			}
		} else {
			for _, a := range saveTs {
				am[a.Aid] = a
				if setCache {
					var ca = &api.Arc{}
					*ca = *a
					d.addCache(func() {
						_ = d.setArcRdsCache(context.Background(), ca)
					})
					d.infoProm.Incr("SetArcCacheByTaishan")
				}
			}
		}

		if len(missTsAid) == 0 {
			return nil
		}
		if saveDB, ips, err = d.RawArcs(ctx, missTsAid); err != nil {
			log.Error("d.RawArcs aids(%v) missed(%v) err(%+v)", aids, missTsAid, err)
			return err
		}
		d.fillArcs(ctx, saveDB, ips, setCache)
		for _, a := range saveDB {
			am[a.Aid] = a
		}
		return nil
	})
	var stm map[int64]*api.Stat
	eg.Go(func(ctx context.Context) (err error) {
		var (
			missed []int64
			missm  map[int64]*api.Stat
		)
		if stm, missed, err = d.statRedisCaches(ctx, aids); err != nil {
			log.Error("d.statRedisCaches(%v) error(%+v)", aids, err)
		}
		if len(missed) == 0 {
			return nil
		}
		if missm, err = d.RawStats(ctx, missed); err != nil {
			log.Error("d.RawStats(%v) error(%+v)", missed, err)
			return nil
		}
		for aid, st := range missm {
			stm[aid] = st
			var cst = &api.Stat{}
			*cst = *st
			d.addCache(func() {
				_ = d.addStatRedisCache(context.TODO(), cst)
			})
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait(%v) error(%v)", aids, err)
		return nil, nil, err
	}
	var upMids, actSids, innerAids []int64
	upIDMap := make(map[int64]struct{})
	actSidsMap := make(map[int64]struct{})
	for aid, a := range am {
		// 历史缓存内存在账号的昵称与头像，不再使用，实时读取账号信息前将其置空
		a.Author.Name = ""
		a.Author.Face = ""
		if _, ok := upIDMap[a.Author.Mid]; !ok && a.Author.Mid > 0 {
			upMids = append(upMids, a.Author.Mid)
			upIDMap[a.Author.Mid] = struct{}{}
		}
		if st, ok := stm[aid]; ok {
			a.Stat = *st
			a.FillStat()
		}
		if _, ok := actSidsMap[a.SeasonID]; !ok && a.SeasonID > 0 && a.AttrValV2(api.AttrBitV2ActSeason) == api.AttrYes {
			actSids = append(actSids, a.SeasonID)
			actSidsMap[a.SeasonID] = struct{}{}
		}
		// 生成短链（老字段可能有使用过暂不下线）
		a.ShortLink = d.handleShortLink(aid)
		a.ShortLinkV2 = d.handleShortLink(aid)
		// 需要获取inner attr的ids
		innerAids = append(innerAids, aid)
	}
	eg2 := errgroup.WithContext(c)
	accInfos := make(map[int64]*accapi.Info)
	if len(upMids) > 0 {
		eg2.Go(func(ctx context.Context) (err error) {
			accInfosReply, err := d.acc.Infos3(ctx, &accapi.MidsReq{Mids: upMids})
			d.infoProm.Incr("ArcsAccount")
			if err != nil {
				// ignore account error
				log.Error("d.acc.Infos3(%v) error(%+v) or resp is empty", upMids, err)
				return nil
			}
			accInfos = accInfosReply.GetInfos()
			return nil
		})
	}
	// 大型活动配置信息
	actSeason := make(map[int64]*mngActApi.Color)
	if len(actSids) > 0 && mobiApp != "" {
		eg2.Go(func(ctx context.Context) (err error) {
			if actSeason, err = d.mngdao.ActSeasonColor(ctx, actSids, mid, mobiApp, device); err != nil {
				log.Error("s.mngdao.ActSeasonColor err(%+v) sids(%+v) mid(%d) mobiApp(%s)", err, actSids, mid, mobiApp)
			}
			return nil
		})
	}
	//获取inner attr
	innerArcs := make(map[int64]*api.ArcInternal)
	if len(innerAids) > 0 {
		eg2.Go(func(ctx context.Context) (e error) {
			if innerArcs, e = d.ArcsInner(ctx, aids); e != nil {
				log.Error("d.ArcsInner err(%+v) aids(%+v) mid(%d) mobiApp(%s)", e, innerAids, mid, mobiApp)
				e = nil
			}
			return
		})
	}
	if err := eg2.Wait(); err != nil {
		log.Error("Arcs eg.wait() err(%+v)", err)
	}
	for _, a := range am {
		if m, ok := accInfos[a.Author.Mid]; ok {
			a.Author.Name = m.Name
			a.Author.Face = m.Face
			func() { // set type name
				typeName, ok := d.tNamem[int16(a.TypeID)]
				if !ok {
					log.Error("日志报警 fillArc aid(%d) typeID(%d) not exist", a.Aid, a.TypeID)
					return
				}
				a.TypeName = typeName
			}()
		}
		if sc, ok := actSeason[a.SeasonID]; ok {
			if sc == nil {
				continue
			}
			a.SeasonTheme = &api.SeasonTheme{
				BgColor:         strings.TrimPrefix(sc.BgColor, "#"),
				SelectedBgColor: strings.TrimPrefix(sc.SelectedBgColor, "#"),
				TextColor:       strings.TrimPrefix(sc.TextColor, "#"),
			}
		}
		//right.autoplay重新赋值
		var inc *api.ArcInternal
		if _, ok := innerArcs[a.Aid]; ok {
			inc = innerArcs[a.Aid]
		}
		a.Rights.Autoplay = api.CalcAutoplayV2(a, inc)
	}
	return am, innerArcs, nil
}

// fillArc is
func (d *Dao) fillArc(c context.Context, a *api.Arc, ip string, setCache bool) {
	d.infoProm.Incr("缓存失败-fillArc")
	a.Fill() // set attribute and rights
	var (
		g       = errgroup.WithContext(c)
		staffs  []*api.StaffInfo
		expand  *archive.ArcExpand
		addit   *archive.Addit
		episode *archive.SeasonEpisode
		err     error
	)
	err = d.transIpv6ToLocation(c, a, ip)
	if err != nil {
		setCache = false
	}
	// set 联合投稿
	if a.AttrVal(archive.AttrBitIsCooperation) == archive.AttrYes {
		g.Go(func(c context.Context) (err error) {
			staffs, err = d.RawStaff(c, a.Aid)
			if err != nil {
				setCache = false
				log.Error("日志报警 fillArc aid(%d) d.RawStaff error(%+v)", a.Aid, err)
				return
			}
			return err
		})
	}
	if a.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes {
		g.Go(func(c context.Context) (err error) {
			missExpands, err := d.RawArchiveExpand(c, []int64{a.Aid})
			if err != nil {
				setCache = false
				log.Error("日志报警 fillArc aid(%d) d.RawArchiveExpand error(%+v)", a.Aid, err)
				return
			}
			expand = missExpands[a.Aid]
			return err
		})
	}
	if a.AttrValV2(api.AttrBitV2Pay) == api.AttrYes {
		g.Go(func(c context.Context) (err error) {
			missAddits, err := d.RawAddits(c, []int64{a.Aid})
			if err != nil {
				setCache = false
				log.Error("日志报警 fillArc aid(%d) d.RawAddits error(%+v)", a.Aid, err)
				return
			}
			addit = missAddits[a.Aid]
			return err
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("日志报警 fillArc eg.wait aid(%d) err(%+v)", a.Aid, err)
		return
	}
	if addit != nil && addit.AttrVal(api.PaySubTypeAttrBitSeason) == int64(api.AttrYes) {
		episode, err = d.RawSeasonEpisode(c, a.SeasonID, a.Aid)
		if err != nil {
			setCache = false
			log.Error("日志报警 fillArc aid(%d) d.RawSeasonEpisode error(%+v)", a.Aid, err)
			return
		}
	}
	a.StaffInfo = staffs
	if expand != nil {
		a.Premiere = &api.Premiere{
			StartTime: expand.PremiereTime.Time().Unix(),
			RoomId:    expand.RoomId,
		}
	}
	if addit != nil {
		a.Pay = &api.PayInfo{
			PayAttr: addit.Subtype,
		}
	}
	if episode != nil {
		a.Rights.ArcPayFreeWatch = episode.AttrVal(ugcmdl.EpisodeAttrSnFreeWatch)
	}
	log.Error("fillArc read from db success, setCache(%t) arc(%+v)", setCache, a)
	if !setCache {
		d.infoProm.Incr("ArcNotSetCache")
		return
	}
	var ca = &api.Arc{}
	*ca = *a
	d.addCache(func() {
		_ = d.setArcRdsCache(context.Background(), ca)
	})
}

// fillArcs is
// nolint:gocognit
func (d *Dao) fillArcs(c context.Context, am map[int64]*api.Arc, ips map[int64]string, setCache bool) {
	d.infoProm.Incr("缓存失败-fillArcs")
	if len(am) == 0 {
		return
	}
	var (
		staffAids    []int64
		premiereAids []int64
		payAids      []int64
		staffs       map[int64][]*api.StaffInfo
		missExpands  map[int64]*archive.ArcExpand
		missAddits   map[int64]*archive.Addit
		missEpisodes = make(map[int64]*archive.SeasonEpisode)
		g            = errgroup.WithContext(c)
		err          error
	)
	for _, a := range am {
		a.Fill()
		if a.AttrVal(archive.AttrBitIsCooperation) == archive.AttrYes {
			staffAids = append(staffAids, a.Aid)
		}
		if a.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes {
			premiereAids = append(premiereAids, a.Aid)
		}
		if a.AttrValV2(api.AttrBitV2Pay) == api.AttrYes {
			payAids = append(payAids, a.Aid)
		}
	}

	if len(premiereAids) > 0 {
		g.Go(func(c context.Context) (err error) {
			missExpands, err = d.RawArchiveExpand(c, premiereAids)
			if err != nil {
				setCache = false
				log.Error("日志报警 fillArcs aids(%+v) d.RawArchiveExpand error(%+v)", premiereAids, err)
				return
			}
			return err
		})
	}
	if len(payAids) > 0 {
		g.Go(func(c context.Context) (err error) {
			missAddits, err = d.RawAddits(c, payAids)
			if err != nil {
				setCache = false
				log.Error("日志报警 fillArcs aids(%+v) d.RawAddits error(%+v)", payAids, err)
				return
			}
			return err
		})
	}
	if len(staffAids) > 0 {
		g.Go(func(c context.Context) (err error) {
			staffs, err = d.RawStaffs(c, staffAids)
			if err != nil {
				setCache = false
				log.Error("日志报警 fillArcs aids(%+v) d.RawStaffs error(%+v)", staffAids, err)
				return
			}
			return err
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("fillArcs eg.wait err(%+v)", err)
		return
	}

	seasonPayAids := make([]int64, 0, len(payAids))
	for _, ad := range missAddits {
		if ad.AttrVal(api.PaySubTypeAttrBitSeason) == int64(api.AttrYes) {
			seasonPayAids = append(seasonPayAids, ad.Aid)
		}
	}
	for _, aid := range seasonPayAids {
		if am[aid] == nil {
			continue
		}
		episode, err := d.RawSeasonEpisode(c, am[aid].SeasonID, aid)
		if err != nil {
			setCache = false
			log.Error("日志报警 fillArcs aid(%d) seasonId(%d) d.RawSeasonEpisode error(%+v)", aid, am[aid].SeasonID, err)
		} else {
			missEpisodes[aid] = episode
		}
	}

	for _, a := range am {
		if ip, ok := ips[a.Aid]; ok {
			err := d.transIpv6ToLocation(c, a, ip)
			if err != nil {
				setCache = false
			}
		}
		if a.AttrVal(archive.AttrBitIsCooperation) == archive.AttrYes && staffs != nil {
			staff, ok := staffs[a.Aid]
			if !ok {
				log.Error("日志报警 fillArcs staff is nil aid(%d)", a.Aid)
			} else {
				a.StaffInfo = staff
			}
		}
		if a.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes && missExpands != nil {
			expand, ok := missExpands[a.Aid]
			if !ok {
				log.Error("日志报警 fillArcs expand is nil aid(%d)", a.Aid)
			} else {
				a.Premiere = &api.Premiere{
					StartTime: expand.PremiereTime.Time().Unix(),
					RoomId:    expand.RoomId,
				}
			}
		}
		if a.AttrValV2(api.AttrBitV2Pay) == api.AttrYes && missAddits != nil {
			addit, ok := missAddits[a.Aid]
			if !ok {
				log.Error("日志报警 fillArcs addit is nil aid(%d)", a.Aid)
			} else {
				a.Pay = &api.PayInfo{
					PayAttr: addit.Subtype,
				}
			}
			if missEpisodes[a.Aid] != nil {
				a.Rights.ArcPayFreeWatch = missEpisodes[a.Aid].AttrVal(ugcmdl.EpisodeAttrSnFreeWatch)
			}
		}
		log.Error("fillArcs read from db success, setCache(%t) arc(%+v)", setCache, a)

		if setCache {
			var ca = &api.Arc{}
			*ca = *a
			d.addCache(func() {
				_ = d.setArcRdsCache(context.Background(), ca)
			})
		}
	}
}

// SimpleArc is
func (d *Dao) SimpleArc(c context.Context, aid int64) (*api.SimpleArc, error) {
	sa, err := d.sArcCache(c, aid)
	if err == nil {
		return sa, nil
	}
	if err == redis.ErrNil {
		d.missProm.Incr("simpleArc")
		var (
			g       = errgroup.WithContext(c)
			arc     *api.Arc
			ps      []*api.Page
			expand  *archive.ArcExpand
			addit   *archive.Addit
			episode *archive.SeasonEpisode
		)
		g.Go(func(c context.Context) (err error) {
			arc, _, err = d.RawArc(c, aid)
			return err
		})
		g.Go(func(c context.Context) (err error) {
			ps, err = d.RawPages(c, aid)
			return err
		})
		if err = g.Wait(); err != nil {
			log.Error("SimpleArc eg.wait err(%+v)", err)
			return nil, err
		}
		if arc == nil || len(ps) == 0 {
			return nil, ecode.NothingFound
		}
		if arc.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes {
			missExpands, err := d.RawArchiveExpand(c, []int64{aid})
			if err != nil {
				return nil, err
			}
			expand = missExpands[aid]
		}
		if arc.AttrValV2(api.AttrBitV2Pay) == api.AttrYes {
			missAddits, err := d.RawAddits(c, []int64{aid})
			if err != nil {
				return nil, err
			}
			addit = missAddits[aid]
			if addit != nil && addit.AttrVal(api.PaySubTypeAttrBitSeason) == int64(api.AttrYes) {
				episode, err = d.RawSeasonEpisode(c, arc.SeasonID, aid)
				if err != nil {
					return nil, err
				}
			}
		}
		d.infoProm.Incr("SimpleArc-回源db")
		sa = d.fillSimpleArc(arc, ps, expand, addit, episode)
		return sa, nil
	}
	log.Error("d.sArcCache aid(%d) err(%+v)", aid, err)
	sa, err = d.getSimpleArcFromTaishan(c, aid)
	if err != nil {
		log.Error("d.getSimpleArcFromTaishan %+v", err)
		return nil, err
	}
	return sa, nil
}

// SimpleArcs is
func (d *Dao) SimpleArcs(c context.Context, aids []int64) (map[int64]*api.SimpleArc, error) {
	sas, err := d.sArcCaches(c, aids)
	if err != nil {
		d.infoProm.Incr("SimpleArcs-读缓存失败")
		log.Error("SimpleArcs d.sArcCaches err(%+v)", err)
		sas, err = d.batchGetSimpleArcFromTaishan(c, aids)
		if err != nil {
			d.infoProm.Incr("SimpleArcs-读taishan失败")
			log.Error("SimpleArcs d.batchGetSimpleArcFromTaishan err(%+v)", err)
			return nil, err
		}
		return sas, nil
	}
	var (
		missAids     []int64
		missArcs     map[int64]*api.Arc
		missPages    map[int64][]*api.Page
		missExpands  map[int64]*archive.ArcExpand
		missAddits   map[int64]*archive.Addit
		missEpisodes = make(map[int64]*archive.SeasonEpisode)
		eg           = errgroup.WithContext(c)
	)
	for _, aid := range aids {
		if _, ok := sas[aid]; !ok {
			missAids = append(missAids, aid)
		}
	}
	if len(missAids) == 0 {
		return sas, nil
	}
	d.infoProm.Incr("SimpleArcs-回源db")
	eg.Go(func(ctx context.Context) (err error) {
		missArcs, _, err = d.RawArcs(ctx, missAids)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		missPages, err = d.RawVideosByAids(ctx, missAids)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Error("SimpleArcs eg.Wait aids(%+v) error(%+v)", aids, err)
		return nil, err
	}

	premiereAids := make([]int64, 0, len(missAids))
	payAids := make([]int64, 0, len(missAids))
	for _, a := range missArcs {
		if a.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes {
			premiereAids = append(premiereAids, a.Aid)
		}
		if a.AttrValV2(api.AttrBitV2Pay) == api.AttrYes {
			payAids = append(payAids, a.Aid)
		}
	}
	if len(premiereAids) > 0 {
		missExpands, err = d.RawArchiveExpand(c, premiereAids)
		if err != nil {
			return nil, err
		}
	}
	if len(payAids) > 0 {
		missAddits, err = d.RawAddits(c, payAids)
		if err != nil {
			return nil, err
		}
		seasonPayAids := make([]int64, 0, len(payAids))
		for _, ad := range missAddits {
			if ad.AttrVal(api.PaySubTypeAttrBitSeason) == int64(api.AttrYes) {
				seasonPayAids = append(seasonPayAids, ad.Aid)
			}
		}
		for _, aid := range seasonPayAids {
			if missArcs[aid] == nil {
				continue
			}
			episode, err := d.RawSeasonEpisode(c, missArcs[aid].SeasonID, aid)
			if err != nil {
				return nil, err
			}
			missEpisodes[aid] = episode
		}
	}

	for _, aid := range missAids {
		arc, aok := missArcs[aid]
		pages, pok := missPages[aid]
		if !aok || !pok {
			continue
		}
		sas[aid] = d.fillSimpleArc(arc, pages, missExpands[aid], missAddits[aid], missEpisodes[aid])
	}
	return sas, nil
}

func (d *Dao) fillSimpleArc(arc *api.Arc, ps []*api.Page, expand *archive.ArcExpand, addit *archive.Addit, episode *archive.SeasonEpisode) *api.SimpleArc {
	if arc == nil || len(ps) == 0 {
		return nil
	}
	var cids []int64
	for _, p := range ps {
		cids = append(cids, p.Cid)
	}
	sa := &api.SimpleArc{
		Aid:         arc.Aid,
		Cids:        cids,
		TypeId:      arc.TypeID,
		Copyright:   arc.Copyright,
		State:       arc.State,
		Access:      arc.Access,
		Attribute:   arc.Attribute,
		Duration:    arc.Duration,
		RedirectUrl: arc.RedirectURL,
		Mid:         arc.Author.Mid,
		SeasonId:    arc.SeasonID,
		AttributeV2: arc.AttributeV2,
		Pubdate:     int64(arc.PubDate),
		Rights: &api.SimpleRights{
			ArcPay: arc.AttrValV2(api.AttrBitV2Pay),
		},
	}
	if expand != nil {
		sa.Premiere = &api.Premiere{
			StartTime: expand.PremiereTime.Time().Unix(),
			RoomId:    expand.RoomId,
		}
	}
	if addit != nil {
		sa.Pay = &api.PayInfo{
			PayAttr: addit.Subtype,
		}
	}
	if episode != nil {
		sa.Rights.ArcPayFreeWatch = episode.AttrVal(ugcmdl.EpisodeAttrSnFreeWatch)
	}
	log.Error("fillSimpleArc read from db success, simple arc(%+v)", sa)

	var ca = &api.SimpleArc{}
	*ca = *sa
	d.addCache(func() {
		if err := d.setSArcCache(context.Background(), ca); err != nil {
			log.Error("d.setSArcCache err(%+v) ca(%+v)", err, ca)
			return
		}
	})
	return sa
}

func (d *Dao) handleShortLink(aid int64) string {
	bv, err := bvid.AvToBv(aid)
	if err != nil {
		log.Error("handleShortLink AvToBv error(%+v) aid(%d)", err, aid)
		return d.shareHost + fmt.Sprintf("av%d", aid)
	}
	return d.shareHost + bv
}

func (d *Dao) transIpv6ToLocation(c context.Context, arc *api.Arc, ip string) (err error) {
	if len(ip) == 0 {
		return
	}
	res, err := d.locDao.Info2(c, ip)
	if err != nil {
		return err
	}
	arc.PubLocation = res.Show
	return
}
