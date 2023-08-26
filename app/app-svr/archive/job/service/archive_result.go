package service

import (
	"context"
	"strconv"
	"time"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/archive/job/model/archive"
	"go-gateway/app/app-svr/archive/job/model/retry"
	"go-gateway/app/app-svr/archive/service/api"

	freyaComp "git.bilibili.co/bapis/bapis-go/pgc/service/freya/component"
	vuapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

func (s *Service) updateResultCache(nw *api.Arc, old *api.Arc, withState bool) {
	var (
		c      = context.TODO()
		oldMid int64
		now    = time.Now()
		err    error
	)
	defer func() { // multi error retry only once
		bm.MetricServerReqDur.Observe(int64(time.Since(now)/time.Millisecond), "arcUpdate", "job")
		bm.MetricServerReqCodeTotal.Inc("arcUpdate", "job", strconv.FormatInt(int64(ecode.Cause(err).Code()), 10))
		if err != nil {
			rt := &retry.Info{Action: retry.FailUpCache}
			rt.Data.Aid = nw.Aid
			rt.Data.State = nw.State
			rt.Data.WithState = withState
			s.PushFail(c, rt, retry.FailList)
			log.Error("updateResultCache aid(%d) error(%+v)", nw.Aid, err)
		}
	}()
	if old != nil && old.Author.Mid != nw.Author.Mid {
		oldMid = old.Author.Mid
	}
	if nw.State >= 0 {
		if err = s.addUpperPassed(c, nw.Aid); err != nil {
			return
		}
	} else {
		if err = s.delUpperPassedCache(c, nw.Aid, nw.Author.Mid); err != nil {
			return
		}
	}
	if oldMid != 0 {
		if err = s.delUpperPassedCache(c, nw.Aid, oldMid); err != nil {
			return
		}
	}

	arc, ip, err := s.resultDao.RawArc(c, nw.Aid)
	if err != nil || arc == nil {
		log.Error("RawArc err(%+v) or aid not exist(%d)", err, nw.Aid)
		return
	}
	//获取ip地址
	s.transIpv6ToLocation(c, arc, ip)
	if err = s.setArcCache(c, arc); err != nil {
		return
	}
	vs, err := s.resultDao.RawVideos(c, nw.Aid)
	if err != nil {
		log.Error("RawVideos err(%+v) aid(%d)", err, nw.Aid)
		return
	}
	if err = s.setVideosPageCache(c, nw.Aid, vs); err != nil {
		return
	}
	if err = s.setSimpleArcCache(c, arc, vs); err != nil {
		return
	}
	if err = s.initStatCache(c, nw.Aid); err != nil {
		return
	}
}

// nolint:gocognit
func (s *Service) tranResult(c context.Context, aid int64) (changed bool, upCids []int64, delCids []int64, err error) {
	var (
		tx           *sql.Tx
		rows         int64
		a            *archive.Archive
		aResult      *api.Arc
		ad           *archive.Addit
		vs           []*archive.Video
		biz          *archive.Biz
		videosCnt    int
		staff        []*archive.Staff
		ffs          map[int64]*archive.VideoFF
		cids         []int64
		arcDelay     *vuapi.ArchivesDelayReply
		pgcRoom      *freyaComp.CreateResp
		premiereTime int64
		roomId       int64
		ap           *archive.ArcExpand
		payBiz       *archive.Biz
	)
	if a, err = s.archiveDao.RawArchive(c, aid); err != nil || a == nil {
		log.Error("s.arc.Archive(%d) error(%+v)", aid, err)
		return
	}
	if ad, err = s.archiveDao.RawAddit(c, aid); err != nil {
		log.Error("s.arc.Addit(%d) error(%+v)", aid, err)
		return
	}
	if ad == nil {
		ad = &archive.Addit{Aid: aid}
	}
	if !a.IsSyncState() {
		log.Info("archive(%d) state(%d) cant change", aid, a.State)
		// FIXME: eeeee
		if s.isPGC(ad.UpFrom) && !s.hadPassed(c, aid) {
			log.Info("archive(%d) is PGC first change", aid)
		} else {
			return
		}
	}
	if ad.Desc == a.Content {
		log.Warn("addit aid(%d) archive and addit is same", aid)
		ad.Desc = ""
	}
	if vs, err = s.archiveDao.RawVideos(c, aid); err != nil {
		log.Error("s.arc.Videos2(%d) error(%+v)", aid, err)
		return
	}
	//获取desc_v2
	if biz, err = s.archiveDao.RawBiz(c, aid, archive.BizStateOk, archive.BizTypeArchive); err != nil {
		log.Error("s.arc.RawBiz(%d) error(%+v)", aid, err)
		return
	}

	//付费稿件检查
	if ad.AttrVal(archive.AttrPay) == archive.InnerAttrYes {
		s.Prom.Incr("付费稿件-消息")
		//获取付费类型
		if payBiz, err = s.archiveDao.RawBiz(c, aid, archive.BizStateOk, archive.BizTypeArchivePay); err != nil {
			log.Error("s.arc.RawBiz(%d) get payBiz error(%+v)", aid, err)
			return
		}
	}

	//首映稿件检查
	if (a.State >= 0 || a.State == api.StateForbidUserDelay) && ad.AttrVal(archive.AttrPremiere) == archive.InnerAttrYes {
		s.Prom.Incr("首映稿件-消息")
		//查询是否已经有首映记录
		ap, err = s.resultDao.RawArchiveExpand(c, aid)
		if err != nil {
			log.Error("首映稿件检查 s.resultDao.RawArchiveExpand error aid(%d) err(%+v)", aid, err)
			return
		}
		//还未记录
		if ap == nil {
			s.Prom.Incr("首映稿件-新增")
			//获取首映时间
			if arcDelay, err = s.archiveDao.ArchivesDelay(c, aid); err != nil {
				log.Error("日志告警 首映稿件检查 s.archiveDao.ArchivesDelay aid(%d) error(%+v)", aid, err)
				return
			}
			if arcDelay != nil && arcDelay.Archives[aid] != nil {
				premiereTime = arcDelay.Archives[aid].Dtime
			}
			// 创建聊天室
			if pgcRoom, err = s.pgcDao.Create4UGCPremiere(c, a, premiereTime); err != nil {
				log.Error("日志告警 首映稿件检查 s.pgcDao.Create4UGCPremiere aid(%d) error(%+v)", aid, err)
				return
			}
			if pgcRoom != nil {
				roomId = pgcRoom.RoomId
			}
		}
	}

	if biz == nil {
		biz = &archive.Biz{Aid: aid}
	}
	//check cid
	for _, v := range vs {
		if v.Cid == 0 && v.Status == archive.VideoStatusSubmit {
			// NOTE: 刚上传，没必要同步去
			log.Error("aid(%d) vid(%d) cid(%d) videoStatus(%d) return", v.Aid, v.ID, v.Cid, v.Status)
			return
		}
		if (v.Status == archive.VideoStatusAccess || v.Status == archive.VideoStatusOpen) && v.State == archive.VideoRelationBind {
			videosCnt++
			cids = append(cids, v.Cid)
		}
	}
	if len(cids) > 0 {
		if ffs, err = s.archiveDao.RawVideoFistFrames(c, cids); err != nil {
			log.Error("s.archiveDao.RawVideoFistFrames(%+v) error(%+v)", cids, err)
			return
		}
	}
	if aResult, _, err = s.resultDao.RawArc(c, aid); err != nil {
		log.Error("s.resultDao.Archive(%d) error(%+v)", aid, err)
		return
	}
	if tx, err = s.resultDao.BeginTran(c); err != nil {
		log.Error("s.result.BeginTran(%d) error(%+v)", aid, err)
		return
	}

	if err = s.resultDao.TxAddAddit(tx, aid, ad, biz, payBiz); err != nil {
		_ = tx.Rollback()
		log.Error("s.resultDao.TxAddAddit error(%+v)", err)
		return
	}

	//首映信息写入db
	if premiereTime != 0 && roomId != 0 {
		expand := &archive.ArcExpand{
			Aid:          aid,
			Mid:          a.Mid,
			ArcType:      1, //首映稿件
			PremiereTime: time.Unix(premiereTime, 0),
			RoomId:       roomId,
		}
		if err = s.resultDao.TxArchiveExpand(tx, expand); err != nil {
			_ = tx.Rollback()
			log.Error("首映稿件检查 s.resultDao.TxArchiveExpand expand(%+v) error(%+v)", expand, err)
			return
		}
		s.Prom.Incr("首映稿件-写入db")
	}
	var (
		duration               int
		firstCid               int64
		dimensions, firstFrame string
	)
	for _, v := range vs {
		if (v.Status == archive.VideoStatusAccess || v.Status == archive.VideoStatusOpen) && v.State == archive.VideoRelationBind {
			var tmpFF string
			if ff, ok := ffs[v.Cid]; ok {
				tmpFF = ff.FirstFrame
			}
			if _, err = s.resultDao.TxAddVideo(tx, v, tmpFF); err != nil {
				_ = tx.Rollback()
				log.Error("s.result.TxAddVideo(%d, %d) error(%+v)", aid, v.Cid, err)
				break
			}
			duration += int(v.Duration)
			upCids = append(upCids, v.Cid)
			if v.Index == 1 && v.SrcType == "vupload" {
				firstCid = v.Cid
				dimensions = v.Dimensions
				firstFrame = tmpFF
			}
		} else {
			if _, err = s.resultDao.TxDelVideoByCid(tx, aid, v.Cid); err != nil {
				_ = tx.Rollback()
				log.Error("s.result.TxDelVideoByCid(%d, %d) error(%+v)", aid, v.Cid, err)
				break
			}
			delCids = append(delCids, v.Cid)
		}
	}
	a.Duration = duration
	// 更新联合投稿人
	if a.AttrVal(archive.AttrBitIsCooperation) == archive.AttrYes {
		if staff, err = s.archiveDao.RawStaff(c, aid); err != nil {
			_ = tx.Rollback()
			log.Error("s.archiveDao.Staff aid(%d) error(%+v)", aid, err)
			return
		}
		if err = s.resultDao.TxDelStaff(tx, aid); err != nil {
			_ = tx.Rollback()
			log.Error("s.result.TxDelStaff aid(%d) error(%+v)", aid, err)
			return
		}
		if staff != nil {
			if err = s.resultDao.TxAddStaff(tx, staff); err != nil {
				_ = tx.Rollback()
				log.Error("s.result.TxAddStaff aid(%d) error(%+v)", aid, err)
				return
			}
		}
	} else { //从联合投稿改为非联合投稿的 删除staff数据
		if aResult != nil && aResult.AttrVal(archive.AttrBitIsCooperation) == archive.AttrYes {
			if err = s.resultDao.TxDelStaff(tx, aid); err != nil {
				_ = tx.Rollback()
				log.Error("s.result.TxDelStaff aid(%d) error(%+v)", aid, err)
				return
			}
		}
	}
	if rows, err = s.resultDao.TxAddArchive(tx, a, ad, videosCnt, firstCid, dimensions, firstFrame); err != nil {
		_ = tx.Rollback()
		log.Error("s.result.TxAddArchive(%d) error(%+v)", aid, err)
		return
	}
	if rows == 0 {
		if _, err = s.resultDao.TxUpArchive(tx, a, ad, videosCnt, firstCid, dimensions, firstFrame); err != nil {
			_ = tx.Rollback()
			log.Error("s.result.TxUpArchive(%d) error(%+v)", aid, err)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit(%d) error(%+v)", aid, err)
		return
	}
	arcState := "normal"
	if a.State < 0 {
		arcState = "abnormal"
	}
	log.Info("aid(%d) upCids(%d) delCids(%d) state(%s) db updated", aid, len(upCids), len(delCids), arcState)
	if a.State >= 0 && (firstCid == 0 || videosCnt == 0 || (len(upCids) == 0 && len(delCids) == 0)) {
		log.Error("日志告警 过审了非正常稿件 aid(%d) firstCid(%d) videosCnt(%d) ", aid, firstCid, videosCnt)
	}
	changed = true
	return
}

func (s *Service) transIpv6ToLocation(c context.Context, arc *api.Arc, ip string) {
	if len(ip) == 0 {
		return
	}
	res, err := s.locDao.Info2WithRetry(c, ip)
	if err == nil && res != nil {
		arc.PubLocation = res.Show
	}
}
