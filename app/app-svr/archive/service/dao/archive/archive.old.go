package archive

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model"
	"go-gateway/app/app-svr/archive/service/model/archive"

	"github.com/golang/protobuf/ptypes/empty"
)

// Videos3 get archive videos by aid.
func (d *Dao) Videos3(c context.Context, aid int64) (ps []*api.Page, err error) {
	ps, err = d.pageCache(c, aid)
	if err == nil {
		return ps, nil
	}
	setCache := false
	if err == redis.ErrNil {
		setCache = true
	}
	log.Error("d.pageCache err(%+v) aid(%d)", err, aid)
	ps, err = d.getPagesFromTaishan(c, aid)
	if err != nil {
		log.Error("d.getPagesFromTaishan err(%+v) aid(%d)", err, aid)
		ps, err = d.RawPages(c, aid)
		if err != nil {
			log.Error("d.RawPages err(%+v) aid(%d)", err, aid)
			return nil, err
		}
	}
	if len(ps) == 0 {
		log.Warn("no passed video aid(%d)", aid)
		ps = []*api.Page{}
	}
	if setCache {
		d.addCache(func() {
			_ = d.addPageCache(context.Background(), aid, ps)
		})
	}
	return ps, nil
}

// VideosByAidCids get videos by aidCids
func (d *Dao) VideosByAidCids(c context.Context, aidCids map[int64][]int64) (map[int64][]*api.Page, error) {
	vs, missMap, redisErr := d.videoAidCidsCache(c, aidCids)
	if redisErr == nil && len(missMap) == 0 {
		d.hitProm.Incr("VideosByAidCids_redis_hit")
		return vs, nil
	}

	if redisErr != nil {
		log.Error("d.videoAidCidsCache aidCids(%+v) redis error(%+v)", aidCids, redisErr)
		missMap = aidCids
	}

	resFromTaishan, missCids, taishanErr := d.batchGetVideoFromTaishan(c, missMap)
	if taishanErr != nil {
		log.Error("d.videoAidCidsCache missMap(%+v) taishan error(%+v)", missMap, taishanErr)
		for _, cids := range missMap {
			missCids = append(missCids, cids...)
		}
	}

	//合并结果
	for aid, values := range resFromTaishan {
		if vs == nil {
			vs = make(map[int64][]*api.Page)
		}
		if _, ok := vs[aid]; !ok {
			vs[aid] = values
		} else {
			vs[aid] = append(vs[aid], values...)
		}
	}

	if taishanErr == nil && len(resFromTaishan) > 0 {
		d.hitProm.Incr("VideosByAidCids_taishan_hit")
		d.rewriteToRedisFromTaishan(c, resFromTaishan)
		if len(missCids) == 0 {
			return vs, nil
		}
	}

	resFromDb, err := d.RawVideosByCids(c, missCids)
	if err != nil {
		d.missProm.Incr("VideosByAidCids_db_error_miss")
		log.Error("d.RawVideosByCids(%+v) error(%+v)", missCids, err)
		return nil, err
	}

	//合并结果
	for aid, values := range resFromDb {
		if vs == nil {
			vs = make(map[int64][]*api.Page)
		}
		if _, ok := vs[aid]; !ok {
			vs[aid] = values
		} else {
			vs[aid] = append(vs[aid], values...)
		}
	}

	d.hitProm.Incr("VideosByAidCids_db_hit")
	d.rewriteToRedisFromTaishan(c, resFromDb)
	d.batchPutVideoToTaishan(c, resFromDb)

	return vs, nil
}

// VideosByAids3 get videos by aids
func (d *Dao) VideosByAids3(c context.Context, aids []int64) (vs map[int64][]*api.Page, err error) {
	vs, missRdsAid, psErr := d.pagesCache(c, aids)
	setCache := true
	if psErr != nil {
		log.Error("d.pagesCache(%+v) error(%+v)", aids, psErr)
		setCache = false
	}
	if len(missRdsAid) == 0 {
		return vs, nil
	}
	if vs == nil {
		vs = make(map[int64][]*api.Page)
	}
	saveTs, missTsAid, tsErr := d.batchGetPagesFromTaishan(c, missRdsAid)
	if tsErr != nil {
		if ecode.EqualError(ecode.NothingFound, tsErr) {
			d.infoProm.Incr("PagesCacheByTaishanNothingFound")
			return vs, nil
		}
		log.Error("d.batchGetPagesFromTaishan missRdsAid(%+v) err(%+v)", missRdsAid, tsErr)
		d.infoProm.Incr("PagesCacheByTaishanErr")
	}
	for aid, v := range saveTs {
		vs[aid] = v
		if !setCache {
			continue
		}
		caid := aid
		cv := v
		d.addCache(func() {
			_ = d.addPageCache(context.Background(), caid, cv)
		})
	}
	if len(missTsAid) == 0 {
		return vs, nil
	}
	saveDB, err := d.RawVideosByAids(c, missTsAid)
	if err != nil {
		log.Error("d.RawVideosByAids(%v) error(%v)", missTsAid, err)
		return nil, err
	}
	for aid, v := range saveDB {
		vs[aid] = v
		if !setCache {
			continue
		}
		caid := aid
		cv := v
		d.addCache(func() {
			_ = d.addPageCache(context.Background(), caid, cv)
		})

	}
	return vs, nil
}

// Video3 get video by aid & cid.
func (d *Dao) Video3(c context.Context, aid, cid int64) (*api.Page, error) {
	p, err := d.videoCache(c, aid, cid)
	if err == nil {
		d.hitProm.Incr("Video3_redis_hit")
		return p, nil
	}

	// redis miss, get from taishan
	p, err = d.getVideoFromTaishan(c, aid, cid)
	if err == nil {
		d.hitProm.Incr("Video3_taishan_hit")
		// taishan hit, write to redis
		if err = d.addVideoCache(c, aid, cid, p); err != nil {
			log.Error("d.video3 aid(%d) cid(%d) write to redis error(%v)", aid, cid, err)
		}
		return p, nil
	}

	// taishan miss, get from db
	p, err = d.RawPage(c, aid, cid)
	if err != nil {
		d.missProm.Incr("Video3_db_error_miss")
		log.Error("d.video3 aid(%d) cid(%d) db error(%v)", aid, cid, err)
		return nil, err
	}
	if p == nil {
		d.missProm.Incr("Video3_db_miss")
		log.Warn("d.video3 aid(%d) cid(%d) no passed video", aid, cid)
		return nil, ecode.NothingFound
	}

	d.hitProm.Incr("Video3_db_hit")
	d.addCache(func() {
		if err = d.addVideoCache(context.Background(), aid, cid, p); err != nil {
			log.Error("d.video3 aid(%d) cid(%d) write to redis error(%v)", aid, cid, err)
		}
		if err = d.setVideoToTaishan(context.Background(), aid, cid, p); err != nil {
			log.Error("d.video3 aid(%d) cid(%d) write to redis error(%v)", aid, cid, err)
		}
	})

	return p, nil
}

// get Description from by aid
func (d *Dao) DescriptionV2(c context.Context, aid int64) (*archive.Addit, error) {
	//先从taishan获取description、desc_v2
	addit, err := d.getDescFromTaishanV2(c, aid)
	if err != nil && err != ecode.NothingFound {
		return nil, err
	}
	if addit != nil {
		return addit, nil
	}
	//addit表中获取
	if addit, err = d.RawAddit(c, aid); err != nil {
		log.Error("d.Addit(%d) error(%v)", aid, err)
		return nil, err
	}
	var desc string
	if addit == nil || addit.Description == "" {
		//archive中获取description
		var a *api.Arc
		if a, _, err = d.RawArc(c, aid); err != nil {
			log.Error("d.RawArc(%d) error(%v)", aid, err)
			return nil, err
		}
		if a != nil && a.Desc != "" {
			desc = a.Desc
		}
		if addit == nil {
			addit = &archive.Addit{Aid: aid, Description: desc}
		} else {
			addit.Description = desc
		}
	}
	//回源taishanV2
	d.rewriteTaishanDescV2(aid, addit.Description, addit.DescV2)
	return addit, nil
}

func (d *Dao) GetDescV2Params(additDescV2 string) []*api.DescV2 {
	if additDescV2 == "" {
		return nil
	}
	var descV2Arc []*archive.DescV2FromArchive
	res := []*api.DescV2{}
	if err := json.Unmarshal([]byte(additDescV2), &descV2Arc); err != nil {
		log.Error("descV2.Unmarshal(%s) error(%+v)", additDescV2, err)
		return nil
	}
	if len(descV2Arc) == 0 {
		return nil
	}
	for _, v := range descV2Arc {
		bizId, _ := strconv.ParseInt(v.BizId, 10, 64)
		res = append(res, &api.DescV2{
			RawText: v.RawText,
			Type:    api.DescType(v.Type),
			BizId:   bizId,
		})
	}
	return res
}

func (d *Dao) Descriptions(c context.Context, aids []int64) (map[int64]*api.DescriptionReply, error) {
	arcs := make(map[int64]*api.Arc)                   //多个稿件
	resp4ResultAddit := make(map[int64]*archive.Addit) //获取result_addit
	//从taishan获取description、desc_v2
	resp, noFoundCacheAid := d.batchGetDescV2FromTaishan(c, aids)
	if len(noFoundCacheAid) == 0 {
		return resp, nil
	}
	var g = errgroup.WithContext(c)
	g.Go(func(ctx context.Context) (err error) {
		//addit表中获取
		resp4ResultAddit, err = d.RawAddits(ctx, noFoundCacheAid)
		if err != nil {
			log.Error("d.RawAddits(%v) error(%v)", noFoundCacheAid, err)
		}
		return nil
	})
	g.Go(func(ctx context.Context) (err error) {
		//从archive表中获取
		arcs, _, err = d.RawArcs(ctx, noFoundCacheAid)
		if err != nil {
			log.Error("d.RawArcs(%v) error(%v)", noFoundCacheAid, err)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Error("Descriptions g.Wait() err:%+v", err)
		return nil, err
	}

	for _, aid := range noFoundCacheAid {
		var desc, descV2 string
		if addit, ok := resp4ResultAddit[aid]; ok && addit != nil {
			desc = addit.Description
			descV2 = addit.DescV2
		}
		if arc, ok := arcs[aid]; ok && desc == "" {
			desc = arc.GetDesc()
		}
		resp[aid] = &api.DescriptionReply{DescV2Parse: d.GetDescV2Params(descV2), Desc: desc}
		//回源taishan
		d.rewriteTaishanDescV2(aid, desc, descV2)
	}
	return resp, nil
}

func (d *Dao) rewriteTaishanDescV2(aid int64, desc, descV2 string) {
	addit := &archive.Addit{
		Aid:         aid,
		Description: desc,
		DescV2:      descV2,
	}
	val, err := json.Marshal(addit)
	if err != nil {
		log.Error("rewriteTaishanDescV2 Marshal error(%v)", err)
		return
	}
	//回源taishan
	d.addCache(func() {
		if err = d.setTaishan(context.Background(), []byte(model.DescKeyV2(aid)), val); err != nil {
			log.Error("rewriteTaishanDescV2 setTaishan error(%v)", err)
		}
	})
}

func (d *Dao) GetArcsRedirect(c context.Context, aids []int64) (map[int64]*api.RedirectPolicy, error) {
	//taishan获取数据
	resp, noFoundCacheAid := d.batchGetRedirectTaishan(c, aids)
	if len(noFoundCacheAid) == 0 {
		return resp, nil
	}
	//redirect表中获取
	redirects, err := d.RawRedirects(c, noFoundCacheAid)
	if err != nil {
		log.Error("d.RawRedirects(%v) error(%v)", noFoundCacheAid, err)
		return nil, err
	}
	//数据拼接
	for _, aid := range noFoundCacheAid {
		redirect, ok := redirects[aid]
		if ok {
			resp[redirect.Aid] = &api.RedirectPolicy{
				Aid:            redirect.Aid,
				RedirectType:   redirect.RedirectType,
				RedirectTarget: redirect.RedirectTarget,
				PolicyType:     redirect.PolicyType,
				PolicyId:       redirect.PolicyId,
			}
		} else {
			//未找到也写入taishan，保证下次请求走taishan
			redirect = &archive.ArcRedirect{
				Aid: aid,
			}
		}
		//回源taishan
		d.rewriteTaishanRedirectKey(redirect)
	}
	return resp, nil
}

func (d *Dao) AddRedirect(ctx context.Context, req *api.ArcRedirectPolicyAddRequest) error {
	redirect := &archive.ArcRedirect{
		Aid:            req.Aid,
		RedirectType:   req.RedirectType,
		RedirectTarget: req.RedirectTarget,
		PolicyType:     req.PolicyType,
		PolicyId:       req.PolicyId,
	}
	//数据库
	err := d.InsertRedirect(ctx, redirect)
	if err != nil {
		return err
	}
	//删除缓存
	d.delTaishanRedirectKey(redirect)
	return nil
}

func (d *Dao) rewriteTaishanRedirectKey(redirect *archive.ArcRedirect) {
	d.addCache(func() {
		val, err := json.Marshal(redirect)
		if err != nil {
			log.Error("rewriteTaishanRedirectKey Marshal error(%v)", err)
			return
		}
		if err := d.setTaishan(context.Background(), []byte(model.RedirectKey(redirect.Aid)), val); err != nil {
			log.Error("rewriteTaishanRedirectKey setTaishan error(%v)", err)
		}
	})
}

func (d *Dao) delTaishanRedirectKey(redirect *archive.ArcRedirect) {
	d.addCache(func() {
		val, err := json.Marshal(redirect)
		if err != nil {
			log.Error("delTaishanRedirectKey Marshal error(%v)", err)
			return
		}
		if err := d.delTaishan(context.Background(), []byte(model.RedirectKey(redirect.Aid)), val); err != nil {
			log.Error("delTaishanRedirectKey delTaishan error(%v)", err)
		}
	})
}

func (d *Dao) loadTypes() {
	var (
		types map[int16]*archive.ArcType
		nm    = make(map[int16]string)
		err   error
	)
	if types, err = d.RawTypes(context.TODO()); err != nil {
		log.Error("d.Types error(%v)", err)
		return
	}
	for _, t := range types {
		nm[t.ID] = t.Name
	}
	d.tNamem = nm
}

func (d *Dao) cacheproc() {
	for {
		f, ok := <-d.cacheCh
		if !ok {
			return
		}
		f()
	}
}

func (d *Dao) addCache(f func()) {
	select {
	case d.cacheCh <- f:
	default:
		log.Warn("d.cacheCh is full")
	}
}

func (d *Dao) loadShortLink() {
	resp, err := d.suClient.CurrentHost(context.Background(), &empty.Empty{})
	//CurrentHost格式为https://b23.tv/
	if err != nil || resp == nil || resp.CurrentHost == "" {
		log.Error("failed to fetch currentHost resp error(%+v)", err)
		d.shareHost = d.c.Custom.ShortLinkHost
		return
	}
	d.shareHost = resp.CurrentHost
}

func (d *Dao) rewriteToRedisFromTaishan(c context.Context, vs map[int64][]*api.Page) {
	// write to redis
	for aid, v := range vs {
		aid := aid
		v := v
		_ = d.cache.Do(c, func(c context.Context) {
			_ = d.addMultiVideoCache(context.Background(), aid, v)
		})
	}
}
