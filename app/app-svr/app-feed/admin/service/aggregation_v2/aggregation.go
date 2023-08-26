package aggregation_v2

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	aggmdl "go-gateway/app/app-svr/app-feed/admin/model/aggregation_v2"
	showmdl "go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/pkg/idsafe/bvid"
)

// 根据播放量降序
type Mtls []*aggmdl.Materiel

func (v Mtls) Len() int { return len(v) }
func (v Mtls) Less(i, j int) bool {
	var iv, jv int32
	if v[i] != nil {
		iv = v[i].ViewInt
	}
	if v[j] != nil {
		jv = v[j].ViewInt
	}
	return iv > jv
}
func (v Mtls) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// 根据播放量升序
type Mtls2 []*aggmdl.Materiel

func (v Mtls2) Len() int { return len(v) }
func (v Mtls2) Less(i, j int) bool {
	var iv, jv int32
	if v[i] != nil {
		iv = v[i].ViewInt
	}
	if v[j] != nil {
		jv = v[j].ViewInt
	}
	return iv < jv
}
func (v Mtls2) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// 根据增速降序
type Mtls3 []*aggmdl.Materiel

func (v Mtls3) Len() int { return len(v) }
func (v Mtls3) Less(i, j int) bool {
	var iv, jv int32
	if v[i] != nil {
		iv = v[i].ViewInt
	}
	if v[j] != nil {
		jv = v[j].ViewInt
	}
	return iv > jv
}
func (v Mtls3) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// 根据增速升序
type Mtls4 []*aggmdl.Materiel

func (v Mtls4) Len() int { return len(v) }
func (v Mtls4) Less(i, j int) bool {
	var iv, jv int32
	if v[i] != nil {
		iv = v[i].ViewInt
	}
	if v[j] != nil {
		jv = v[j].ViewInt
	}
	return iv < jv
}
func (v Mtls4) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

const (
	_subtitle       = "当前热门视频"
	_title          = "更多关于【%s】的热门视频"
	_aggregationImg = "https://i0.hdslb.com/bfs/archive/30768b1189c4fbf919e92edb5b3c7fb2ae403eb1.jpg"
)

func formTitle(title string) string {
	return fmt.Sprintf(_title, title)
}

//nolint:gocognit
func (s *Service) List(c context.Context, hotWord, filter string, state int) (res *aggmdl.List, err error) {
	var (
		aggMC   map[string][]*showmdl.AggregationItem
		aggDB   []*aggmdl.Aggregation
		filters []string
	)
	if filter != "" {
		filters = strings.Split(filter, ",")
	}
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if aggMC, err = s.aggDao2.ListCache(ctx, filters); err != nil {
			log.Error("%v", err)
		}
		return
	})
	g.Go(func() (err error) {
		if aggDB, err = s.aggDao2.Aggregations(ctx, hotWord, filter); err != nil {
			log.Error("%v", err)
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	res = &aggmdl.List{
		Sort: aggmdl.ListSort,
	}
	// 聚合全网热点(mc数据)
	var resTmp []*aggmdl.Aggregation
	//nolint:gosimple
	if aggMC != nil {
		for _, aggcaches := range aggMC {
			for _, aggcache := range aggcaches {
				if aggcache == nil {
					continue
				}
				a := &aggmdl.Aggregation{}
				a.FormAggregationMc(aggcache)
				// 数据补全
				for _, agg := range aggDB {
					if agg == nil {
						continue
					}
					if (a.PlatType == agg.PlatType) && (a.Hotword == agg.Hotword) {
						a.ID = agg.ID
						a.State = agg.State
						a.ActiveState = agg.ActiveState
						a.Title = agg.Title
						a.Subtitle = agg.Subtitle
						a.Cover = agg.Cover
					}
				}
				resTmp = append(resTmp, a)
			}
		}
	}
	// 聚合db数据
	for _, agg := range aggDB {
		if agg == nil {
			continue
		}
		// 聚合B站热门
		if agg.Plat == aggmdl.GotoBiliPopular {
			// 取一周内的B站热门
			if (time.Now().Unix() - agg.MTime) < 86400*7 {
				agg.Idx = "-"
				resTmp = append(resTmp, agg)
			}
		} else {
			// 聚合非B站热门
			var isExist bool
			//nolint:gosimple
			if aggMC != nil {
				for plat, aggcaches := range aggMC {
					if agg.Plat != plat {
						continue
					}
					for _, aggcache := range aggcaches {
						if aggcache == nil {
							continue
						}
						if aggcache.Title == agg.Hotword {
							isExist = true
						}
					}
				}
			}
			if !isExist {
				agg.Idx = "-"
				resTmp = append(resTmp, agg)
			}
		}
	}
	// 状态筛选
	for _, rt := range resTmp {
		// 过审物料变更文案
		if rt.State == aggmdl.StatePass {
			var ais *showmdl.AggAI
			if ais, err = s.aggDao2.AggAI(c, rt.ID); err != nil {
				log.Error("%v", err)
				err = nil
			}
			if ais != nil {
				rt.ArcCnt = len(ais.CardList)
				rt.NewArcCnt = ais.UpCnt
			}
		}
		switch state {
		case aggmdl.StatePass:
			if rt.State == aggmdl.StatePass {
				res.Items = append(res.Items, rt)
			}
		case aggmdl.StateDown:
			if rt.ActiveState == aggmdl.ActiveStateOffLine {
				res.Items = append(res.Items, rt)
			}
		default:
			res.Items = append(res.Items, rt)
		}
	}
	return
}

//nolint:gocognit
func (s *Service) Operate(c context.Context, plat, hotword string, state int) (err error) {
	var tx *sql.Tx
	if tx, err = s.aggDao2.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var hot *aggmdl.Hotword
	if hot, err = s.aggDao2.AggregationByPlatHotWord(c, plat, hotword); err != nil {
		log.Error("%v", err)
		return
	}
	// state = 0 待审
	// state = 1 && active_state = 0 过审 待上线
	// state = 1 && active_state = 1 过审 上线
	// state = 1 && active_state = 2 过审 下线
	// state = 2 拒绝
	if hot != nil {
		switch state {
		case aggmdl.StateWait:
			// 过审的热词保持过审,其他状态强制置为待审
			if hot.State != aggmdl.StatePass {
				if _, err = s.aggDao2.UpAggregationState(tx, plat, hotword, state); err != nil {
					log.Error("%v", err)
					return
				}
			}
			// 上线状态强制置为 未配置上线
			if _, err = s.aggDao2.UpAggregationActiveState(tx, plat, hotword, aggmdl.ActiveStateNew); err != nil {
				log.Error("%v", err)
			}
		case aggmdl.StatePass:
			// 置为 通过审核
			if _, err = s.aggDao2.UpAggregationState(tx, plat, hotword, state); err != nil {
				log.Error("%v", err)
			}
		case aggmdl.StateForbid:
			if hot.Plat == "artificial" && hot.State == aggmdl.StateWait {
				// 如果是人工添加的并且处于待审状态 清空所有数据
				if _, err = s.aggDao2.DelHotWord(tx, hot.ID); err != nil {
					log.Error("%v", err)
					return
				}
				if _, err = s.aggDao2.DelHotWordTag(tx, hot.ID); err != nil {
					log.Error("%v", err)
					return
				}
				if _, err = s.aggDao2.DelHotWordVideo(tx, hot.ID); err != nil {
					log.Error("%v", err)
				}
			} else {
				// 置为 拒绝
				if _, err = s.aggDao2.UpAggregationState(tx, plat, hotword, state); err != nil {
					log.Error("%v", err)
					return
				}
				// 上线状态强制置为 未配置上线
				if _, err = s.aggDao2.UpAggregationActiveState(tx, plat, hotword, aggmdl.ActiveStateNew); err != nil {
					log.Error("%v", err)
				}
			}
		case aggmdl.StateDown:
			if hot.State == aggmdl.StatePass && hot.ActiveState == aggmdl.ActiveStateOnline {
				if _, err = s.aggDao2.UpAggregationActiveState(tx, plat, hotword, aggmdl.ActiveStateOffLine); err != nil {
					log.Error("%v", err)
				}
			}
		}
	} else {
		// 非人工 非B站热门 通过审核落库
		if state == aggmdl.StatePass && plat != "artificial" && plat != "bili_popular" {
			if _, err = s.aggDao2.AddAggregation(tx, plat, hotword, formTitle(hotword), _subtitle, _aggregationImg, aggmdl.StatePass, aggmdl.ActiveStateNew); err != nil {
				log.Error("%v", err)
			}
		}
	}
	return
}

func (s *Service) Add(c context.Context, hotword string) (err error) {
	var tx *sql.Tx
	if tx, err = s.aggDao2.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var hot *aggmdl.Hotword
	if hot, err = s.aggDao2.AggregationByPlatHotWord(c, "artificial", hotword); err != nil {
		log.Error("%v", err)
		return
	}
	if hot != nil {
		err = ecode.Error(ecode.RequestErr, "热词已存在")
		return
	}
	if _, err = s.aggDao2.AddAggregation(tx, "artificial", hotword, formTitle(hotword), _subtitle, _aggregationImg, aggmdl.StateWait, aggmdl.ActiveStateNew); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) Save(c context.Context, id int64, title, subTitle, cover string) (err error) {
	var tx *sql.Tx
	if tx, err = s.aggDao2.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var hot *aggmdl.Hotword
	if hot, err = s.aggDao2.AggregationByID(c, id); err != nil {
		log.Error("%v", err)
		return
	}
	if hot != nil {
		if _, err = s.aggDao2.UpAggregationOlineConfig(tx, id, title, subTitle, cover); err != nil {
			log.Error("%v", err)
			return
		}
		if hot.State == aggmdl.StatePass && hot.ActiveState == aggmdl.ActiveStateNew {
			if _, err = s.aggDao2.UpAggregationActiveState(tx, hot.Plat, hot.HotTitle, aggmdl.ActiveStateOnline); err != nil {
				log.Error("%v", err)
			}
		}
	}
	return
}

//nolint:gocognit
func (s *Service) Materiels(c context.Context, id int64, sortParam, orderParam string) (res *aggmdl.MaterielsList, err error) {
	res = &aggmdl.MaterielsList{}
	// 获取db物料、缓存的ai物料、缓存的播放量
	var (
		aggs    []*aggmdl.Materiel
		ais     *showmdl.AggAI
		arcm    map[int64]*showmdl.ArcInfo
		aggTags []*aggmdl.AggregationTag
		hotword *aggmdl.Hotword
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if hotword, err = s.aggDao2.AggregationByID(c, id); err != nil {
			log.Error("%v", err)
			err = nil
		}
		return
	})
	g.Go(func() (err error) {
		if aggTags, err = s.aggDao2.Tags(ctx, id); err != nil {
			log.Error("%v", err)
			err = nil
		}
		return
	})
	g.Go(func() (err error) {
		if aggs, err = s.aggDao2.Materiels(ctx, id); err != nil {
			log.Error("%v", err)
			err = nil
		}
		return
	})
	g.Go(func() (err error) {
		if ais, err = s.aggDao2.AggAI(ctx, id); err != nil {
			log.Error("%v", err)
			err = nil
		}
		return
	})
	g.Go(func() (err error) {
		if arcm, err = s.aggDao2.AggArc(ctx, id); err != nil {
			log.Error("%v", err)
			err = nil
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	// 生成tags列表
	var (
		tids, aids []int64
		tags       map[int64]*taggrpc.Tag
		mtlTags    []*aggmdl.AggregationTag
	)
	for _, aggTag := range aggTags {
		if aggTag == nil {
			continue
		}
		mtlTags = append(mtlTags, &aggmdl.AggregationTag{
			ID:    aggTag.ID,
			HotID: aggTag.HotID,
			TagID: aggTag.TagID,
			State: aggTag.State,
		})
		tids = append(tids, aggTag.TagID)
	}
	if len(tids) > 0 {
		if tags, err = s.aggDao2.TagByTagID(c, tids); err != nil {
			log.Error("%v", err)
			err = nil
		}
		if tags != nil {
			for _, mtlTag := range mtlTags {
				mt := &aggmdl.AggregationTag{}
				*mt = *mtlTag
				if tag, ok := tags[mt.TagID]; ok && tag != nil {
					mt.Title = tag.Name
				}
				res.Tags = append(res.Tags, mt)
			}
		}
	}
	for _, agg := range aggs {
		if agg == nil || agg.ID == 0 {
			continue
		}
		aids = append(aids, agg.OID)
	}
	if arcm == nil && len(aids) > 0 {
		if arcm, err = s.Arcs(c, aids); err != nil {
			//nolint:govet
			log.Error("%v")
			err = nil
		}
	}
	// 物料拆分和补全
	var aggOK, aggAIs, aggFobid []*aggmdl.Materiel
	for _, agg := range aggs {
		if agg == nil {
			continue
		}
		agg2 := &aggmdl.Materiel{}
		*agg2 = *agg
		if a, ok := arcm[agg.OID]; ok && a != nil {
			agg2.ViewInt = a.View
			agg2.View = aggmdl.StatString(int(agg2.ViewInt), "")
			agg2.ViewSpeedInt = a.ViewSpeed
			agg2.ViewSpeed = aggmdl.StatString(int(agg2.ViewSpeedInt), "")
			agg2.Title = a.Title
			agg2.Author = a.Author
		}
		switch agg.State {
		case aggmdl.MaterialStateOK:
			aggOK = append(aggOK, agg2)
		case aggmdl.MaterialStateFobid:
			aggFobid = append(aggFobid, agg2)
		}
	}
	if ais != nil {
		res.ArcCnt = len(ais.CardList)
		res.NewArcCnt = ais.UpCnt
		for _, ai := range ais.CardList {
			if ai == nil {
				continue
			}
			agg := &aggmdl.Materiel{HotID: id, State: aggmdl.MaterialStateOK}
			agg.FormAI(ai, arcm)
			aggAIs = append(aggAIs, agg)
		}
	}
	if hotword != nil {
		res.Hotword = hotword.HotTitle
	}
	// 聚合物料列表：人工 > 河童 > 禁止 排序聚合
	res.Items = append(res.Items, s.sortMaterial(aggOK, sortParam, orderParam)...)
	res.Items = append(res.Items, s.sortMaterial(aggAIs, sortParam, orderParam)...)
	res.Items = append(res.Items, s.sortMaterial(aggFobid, sortParam, orderParam)...)
	return
}

func (s *Service) sortMaterial(ms []*aggmdl.Materiel, sortParam, orderParam string) (res []*aggmdl.Materiel) {
	if sortParam == "view" && orderParam == "desc" {
		sort.Sort(Mtls(ms))
	} else if sortParam == "view" && orderParam == "asc" {
		sort.Sort(Mtls2(ms))
	} else if sortParam == "view_speed" && orderParam == "desc" {
		sort.Sort(Mtls3(ms))
	} else if sortParam == "view_speed" && orderParam == "asc" {
		sort.Sort(Mtls4(ms))
	}
	res = ms
	return
}

func (s *Service) ViewAdd(c context.Context, hotID int64, oids []string) (err error) {
	var aids []int64
	for _, oid := range oids {
		var aid int64
		if aid, err = strconv.ParseInt(oid, 10, 64); err != nil {
			log.Error("%v", err)
			if aid, err = bvid.BvToAv(oid); err != nil {
				log.Error("%v", err)
				continue
			}
			aids = append(aids, aid)
			continue
		}
		aids = append(aids, aid)
	}
	var tx *sql.Tx
	if tx, err = s.aggDao2.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.aggDao2.AddViews(tx, "人工", hotID, aids, aggmdl.MaterialStateOK); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ViewOperate(c context.Context, hotID, oid int64, state int, source string) (err error) {
	var id int64
	if id, err = s.aggDao2.View(c, hotID, oid, source); err != nil {
		log.Error("%v", err)
		return
	}
	var tx *sql.Tx
	if tx, err = s.aggDao2.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if id > 0 {
		if _, err = s.aggDao2.UpView(tx, id, state); err != nil {
			log.Error("%v", err)
		}
	} else {
		if _, err = s.aggDao2.AddView(tx, source, hotID, oid, state); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

func (s *Service) TagAdd(c context.Context, hotID, tagID int64) (err error) {
	var id int64
	if id, err = s.aggDao2.TagID(c, hotID, tagID); err != nil {
		log.Error("%v", err)
		return
	}
	var tx *sql.Tx
	if tx, err = s.aggDao2.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if id == 0 {
		if _, err = s.aggDao2.AddTag(tx, hotID, tagID); err != nil {
			log.Error("%v", err)
		}
	} else {
		if _, err = s.aggDao2.UpTag(tx, id, aggmdl.TagStateOK); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

func (s *Service) TagDel(c context.Context, id int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.aggDao2.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.aggDao2.UpTag(tx, id, aggmdl.TagStateDel); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) Arcs(c context.Context, aids []int64) (res map[int64]*showmdl.ArcInfo, err error) {
	var arcs map[int64]*arcgrpc.Arc
	if arcs, err = s.arcDao.Arcs(c, aids); err != nil {
		log.Error("%v", err)
		return
	}
	res = make(map[int64]*showmdl.ArcInfo)
	//nolint:gosimple
	if arcs != nil {
		for aid, arc := range arcs {
			if aid == 0 || arc == nil {
				continue
			}
			res[aid] = &showmdl.ArcInfo{
				ID:     arc.Aid,
				Title:  arc.Title,
				Author: arc.Author.Name,
				View:   arc.Stat.View,
			}
		}
	}
	return
}
