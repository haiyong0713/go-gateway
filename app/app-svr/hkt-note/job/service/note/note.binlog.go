package note

import (
	"context"
	"encoding/json"
	"fmt"
	replyAPI "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	"go-common/library/database/taishan"
	xtime "go-common/library/time"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/app-svr/hkt-note/common"
	"go-gateway/app/app-svr/hkt-note/job/model/article"
	"go-gateway/app/app-svr/hkt-note/job/model/note"
	ntmdl "go-gateway/app/app-svr/hkt-note/job/model/note"
)

func (s *Service) consumeNoteBinlog() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(10 * time.Millisecond)
		msg, ok := <-s.noteBinlogSub.Messages()
		if !ok {
			log.Warn("noteWarn binlog consumeDetailBinlog quit")
			return
		}
		log.Warn("consumeNoteBinlog msg.value(%s)", string(msg.Value))
		if err := msg.Commit(); err != nil {
			log.Error("noteError consumeNoteBinlog commit error(%v)", err)
			continue
		}
		m := &note.Binlog{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("noteError consumeNoteBinlog msg(%+v) error(%v)", msg.Value, err)
			continue
		}
		switch m.ToTableName() {
		case "note_detail":
			data := &note.NoteDetailBlog{}
			if err := json.Unmarshal(msg.Value, data); err != nil {
				log.Error("noteError consumeNoteBinlog msg(%+v) error(%v)", msg.Value, err)
				continue
			}
			if data.New == nil {
				log.Error("noteError consumeNoteBinlog empty result")
				continue
			}
			s.treatNoteDetailBinlog(context.Background(), data.New)
		case "article_detail":
			data := &article.ArtDetailBlog{}
			if err := json.Unmarshal(msg.Value, data); err != nil {
				log.Error("noteError consumeNoteBinlog msg(%+v) error(%v)", msg.Value, err)
				continue
			}
			if data.New == nil {
				log.Error("noteError consumeNoteBinlog empty result")
				continue
			}
			log.Warn("consumeNoteBinlog article_detail msg (%s)", string(msg.Value))
			publicStat, commentOperation := data.ToPublicStatus()
			//todo 整合下这里的入参
			s.treatArtDetailBinlog(context.Background(), data.New, publicStat, commentOperation, data.New.Cvid)
		default:
			log.Error("noteError consumeNoteBinlog msg(%+v) tableName(%s) invalid", msg.Value, m.ToTableName())
			continue
		}
	}
}

func (s *Service) treatArtDetailBinlog(c context.Context, msg *article.ArtDetailDB, artStat int, commentOperation bool, curCvid int64) {
	var err error
	defer func() {
		if err != nil {
			log.Error("artError treatArtDetailBinlog msg(%+v) error(%v)", msg, err)
			s.dao.AddCacheRetry(c, note.KeyRetryArtDtlBinlog, fmt.Sprintf("%d-%d", curCvid, msg.NoteId), time.Now().Unix())
		}
	}()
	// 更新cvid&note_id维度的详情缓存
	if err = s.artDao.AddCacheArtDetail(c, msg.NoteId, msg.ToDtlCache(), article.TpArtDetailNoteId, s.artDao.ArtExpire); err != nil {
		return
	}
	if artStat != article.ArtNoChange {
		// 更新客态笔记的详情&正文缓存
		if err = s.artDao.AddCacheArtDetail(c, curCvid, msg.ToDtlCache(), article.TpArtDetailCvid, s.artDao.ArtExpire); err != nil {
			return
		}
		var cont *article.ArtContCache
		if cont, err = s.artDao.ArtContByVer(c, curCvid, msg.PubVersion); err != nil {
			return
		}
		if err = s.artDao.AddCacheArtContent(c, curCvid, cont); err != nil {
			return
		}
		// 更新稿件维度客态笔记数缓存
		var total int
		if total, err = s.artDao.ArtCountInArc(c, msg.Oid, msg.OidType); err != nil {
			return
		}
		if err = s.artDao.AddCacheArtCntInArc(c, msg.Oid, msg.OidType, total); err != nil {
			return
		}
	}
	// 客态笔记状态变更
	listVal := article.ToArtListVal(curCvid, msg.NoteId)
	switch artStat { // 笔记专栏缓存更新
	// 新版本客态笔记可看
	case article.ArtCanView:
		pubtime := msg.Pubtime
		if pubtime == "0000-00-00 00:00:00" { // 先发后审，无发布时间
			pubtime = msg.Ctime
		}
		score, _ := time.ParseInLocation("2006-01-02 15:04:05", pubtime, time.Local)
		if err = s.artDao.AddCacheArtList(c, s.artDao.ArcListKey(msg.Oid, msg.OidType), listVal, score.Unix()); err != nil {
			return
		}
		if err = s.artDao.AddCacheArtList(c, s.artDao.UserListKey(msg.Mid), listVal, score.Unix()); err != nil {
			return
		}

		// 处理笔记在评论区的展示
		s.processNoteReply(c, msg)
		if commentOperation {
			//公开笔记的状态满足评论运营位
			_ = s.setCommentOperationForUper(c, msg)
		}
	// 该cvid不可看，被删除了
	case article.ArtCantView:
		if err = s.artDao.RemCacheArtList(c, s.artDao.ArcListKey(msg.Oid, msg.OidType), listVal); err != nil {
			return
		}
		if err = s.artDao.RemCacheArtList(c, s.artDao.UserListKey(msg.Mid), listVal); err != nil {
			return
		}
		// 新版本客态笔记不可看时，继续用老版本
		// 若up删除自己公开笔记，下线笔记运营位
		_ = s.OfflineCommentOperation(c, msg)
	default:
	}
}

func (s *Service) setCommentOperationForUper(ctx context.Context, newData *article.ArtDetailDB) (err error) {
	//判断视频类型是否为ugc
	if newData.OidType != ntmdl.OidTypeUGC {
		return nil
	}
	//判断是否是uper主本人公开的笔记
	aids := []int64{newData.Oid}
	arcsReply, err := s.dao.Arcs(ctx, aids)
	if err != nil || arcsReply.Arcs == nil {
		return err
	}
	if arcReply, ok := arcsReply.Arcs[newData.Oid]; ok {
		if arcReply.Author.Mid != newData.Mid {
			//不是uper主发布的笔记
			return nil
		}
	}
	alreadySet, err := s.artDao.GetPubSuccessCvidsBeforeAssignedVersion(ctx, newData.Cvid, newData.PubVersion)
	if err != nil {
		log.Errorc(ctx, "setCommentOperationForUper GetPubSuccessCvidsBeforeAssignedVersion err %v cvid %v pubVersion %v", err, newData.Cvid, newData.PubVersion)
		return err
	}
	if alreadySet {
		//说明该cvid有过成功公开记录，已经设置过评论运营
		log.Warnc(ctx, "setCommentOperationForUper already set  cvid %v,pubVersion %v", newData.Cvid, newData.PubVersion)
		return nil
	}
	//通知评论设置该稿件的笔记运营位
	link := fmt.Sprintf(s.c.NoteCfg.ReplyCfg.ReplyOperationWebUrl, newData.Cvid)
	replyReq := &replyAPI.AddOperationReq{
		Type:          1,
		Oids:          []int64{newData.Oid},
		OperationType: replyAPI.OperationType_NOTE,
		Link:          link,
		Title:         s.c.NoteCfg.ReplyCfg.ReplyOperationTitle,
		Subtitle:      s.c.NoteCfg.ReplyCfg.ReplyOperationSubTitle,
		Icon:          s.c.NoteCfg.ReplyCfg.ReplyOperationIcon,
		StartTime:     xtime.Time(time.Now().Unix()),
		EndTime:       xtime.Time(s.c.NoteCfg.ReplyCfg.ReplyOperationEndTime),
	}
	resp, err := s.artDao.AddReplyOperation(ctx, replyReq)
	if err != nil {
		log.Errorc(ctx, "AddReplyOperation err  req %v err %v", replyReq, err)
		return err
	}
	if resp.Id == 0 {
		log.Errorc(ctx, "AddReplyOperation err req %v resp %v", replyReq, resp)
		return err
	}
	if err = s.recordCvidToOpid(ctx, newData.Cvid, resp.Id); err != nil {
		log.Errorc(ctx, "recordCvidToOpid err  cvid %v opid %v err %v", newData.Cvid, resp.Id, err)
		return err
	}
	return nil
}

func (s *Service) processNoteReply(ctx context.Context, newData *article.ArtDetailDB) {
	// 判断是否走评论区展示新样式
	isRichTextFormat := s.isNoteReplyRichTextFormat(ctx, newData.NoteId)
	if isRichTextFormat {
		// 新样式仅首次公开发布时进行reply/add
		alreadySet, err := s.artDao.GetPubSuccessCvidsBeforeAssignedVersion(ctx, newData.Cvid, newData.PubVersion)
		if err != nil {
			log.Errorc(ctx, "processNoteReply GetPubSuccessCvidsBeforeAssignedVersion err %v cvid %v pubVersion %v", err, newData.Cvid, newData.PubVersion)
			return
		}
		if alreadySet {
			//说明该cvid有过成功公开记录，已经设置过评论运营
			log.Warnc(ctx, "processNoteReply already set  cvid %v,pubVersion %v", newData.Cvid, newData.PubVersion)
			return
		}
		// 这里不做retry，后续改为railgun重试
		webUrl := fmt.Sprintf(s.c.NoteCfg.ReplyCfg.RichTextWebUrl, newData.Cvid)
		replyCont := fmt.Sprintf(s.c.NoteCfg.ReplyCfg.Template, newData.Summary, webUrl)
		replyAddRes, err := s.artDao.ReplyAddWithRes(ctx, newData.Mid, newData.Oid, replyCont)
		if err != nil || replyAddRes == nil || replyAddRes.Data == nil {
			return
		}
		log.Warnc(ctx, "processNoteReply record mapping cvid %v rpid %v", newData.Cvid, replyAddRes.Data.Rpid)
		_ = s.recordCvidRpidMapping(ctx, newData.Cvid, replyAddRes.Data.Rpid)
	} else {
		// 老样式
		// 自动发评论
		if newData.AutoComment == article.AutoComment || newData.PubFrom == article.PubFromReply {
			webUrl := fmt.Sprintf(s.c.NoteCfg.ReplyCfg.WebUrl, newData.Cvid)
			replyCont := fmt.Sprintf(s.c.NoteCfg.ReplyCfg.Template, newData.Summary, webUrl)
			if replyErr := retry.WithAttempts(ctx, "autoReply-retry", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
				return s.artDao.ReplyAdd(c, newData.Mid, newData.Oid, replyCont)
			}); replyErr != nil {
				log.Warn("artWarn treatArtDetailBinlog reply err(%+v)", replyErr)
			}
		}
	}
}

// 判断该公开笔记在评论区的样式是否是新的样式，明确获取新样式type为新，否则视为旧
// https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661002736749
func (s *Service) isNoteReplyRichTextFormat(ctx context.Context, noteId int64) (richTextFormat bool) {
	var (
		record              *taishan.Record
		noteReplyFormatInfo = &common.TaishanNoteReplyFormatInfo{}
	)
	taishanKey := fmt.Sprintf(common.Note_Reply_Format_Taishan_Key, noteId)
	record, err := s.artDao.GetTaishan(ctx, taishanKey, common.TaishanConfig.NoteReply)
	if err != nil {
		log.Errorc(ctx, "isNoteReplyRichTextFormat GetTaiShan err %v key %v", err, taishanKey)
		return false
	}
	if record != nil && len(record.Columns) > 0 && len(record.Columns[0].Value) > 0 {
		err = json.Unmarshal(record.Columns[0].Value, noteReplyFormatInfo)
		if err != nil {
			log.Errorc(ctx, "isNoteReplyRichTextFormat Unmarshal err %v and key %v", err, taishanKey)
			return false
		}
		if noteReplyFormatInfo != nil && noteReplyFormatInfo.FormatType == common.Note_Comment_Format_Type_New {
			return true
		}
	}
	return false
}

func (s *Service) treatNoteDetailBinlog(c context.Context, msg *note.NtDetailDB) {
	// if deleted, remove cache
	if msg.Deleted == note.Deleted {
		if err := s.dao.DelKey(c, s.dao.DetailKey(msg.NoteId)); err != nil {
			log.Error("noteError consumeNoteBinlog msg(%+v) err(%+v)", msg, err)
			s.dao.AddCacheRetry(c, note.KeyRetryDel, s.dao.DetailKey(msg.NoteId), time.Now().Unix())
		}
		if err := s.dao.DelKey(c, s.dao.ContentKey(msg.NoteId)); err != nil {
			log.Error("noteError consumeNoteBinlog msg(%+v) err(%+v)", msg, err)
			s.dao.AddCacheRetry(c, note.KeyRetryDel, s.dao.ContentKey(msg.NoteId), time.Now().Unix())
		}
	}

	// update note_detail cache
	if err := s.dao.AddCacheNoteDetail(c, msg.NoteId, msg.ToDtlCache()); err != nil {
		log.Error("noteError consumeNoteBinlog msg(%+v) err(%+v)", msg, err)
		s.dao.AddCacheRetry(c, note.KeyRetryDetail, fmt.Sprintf("%d-%d", msg.NoteId, msg.Mid), time.Now().Unix()+300)
	}

	// update note_user db&cache
	if upUserErr := s.updateUser(c, msg.Mid); upUserErr != nil {
		log.Error("noteError consumeNoteBinlog msg(%+v) err(%+v)", msg, upUserErr)
		s.dao.AddCacheRetry(c, note.KeyRetryUser, strconv.FormatInt(msg.Mid, 10), time.Now().Unix()+300)
	}

	// update note_list zset
	func() {
		if msg.Deleted == note.Deleted {
			if err := s.dao.RemCacheNoteList(c, msg.Mid, fmt.Sprintf("%d-%d", msg.NoteId, msg.Aid)); err != nil {
				s.dao.AddCacheRetry(c, note.KeyRetryListRem, fmt.Sprintf("%d-%d-%d", msg.NoteId, msg.Mid, msg.Aid), time.Now().Unix()+300)
			}
			return
		}
		if msg.Deleted == note.NotDeleted {
			mtime, _ := time.ParseInLocation("2006-01-02 15:04:05", msg.Mtime, time.Local)
			if err := s.dao.AddCacheNoteList(c, msg.Mid, fmt.Sprintf("%d-%d", msg.NoteId, msg.Aid), mtime.Unix()); err != nil {
				log.Error("noteError consumeNoteBinlog msg(%+v) err(%+v)", msg, err)
				s.dao.AddCacheRetry(c, note.KeyRetryList, fmt.Sprintf("%d-%d-%d", msg.NoteId, msg.Mid, msg.Aid), time.Now().Unix()+300)
			}
		}
	}()

	// update note_aid cache
	if msg.Deleted == note.Deleted {
		if err := s.dao.DelKey(c, s.dao.AidKey(msg.Mid, msg.Aid, msg.OidType)); err != nil {
			log.Error("noteError consumeNoteBinlog msg(%+v) err(%+v)", msg, err)
			s.dao.AddCacheRetry(c, note.KeyRetryDel, s.dao.AidKey(msg.Mid, msg.Aid, msg.OidType), time.Now().Unix())
		}
		return
	}
	if err := s.dao.AddCacheNoteAid(c, msg.Mid, msg.Aid, msg.NoteId, msg.OidType); err != nil {
		log.Error("noteError consumeNoteBinlog msg(%+v) err(%+v)", msg, err)
		s.dao.AddCacheRetry(c, note.KeyRetryAid, fmt.Sprintf("%d-%d-%d", msg.Mid, msg.Aid, msg.OidType), time.Now().Unix()+300)
	}
}

func (s *Service) updateUser(c context.Context, mid int64) error {
	// calculate user's new note_count & note_size
	newSize, newCnt, err := s.dao.NoteUserData(c, mid)
	if err != nil {
		return err
	}
	userCache := &note.UserCache{
		Mid:       mid,
		NoteSize:  newSize,
		NoteCount: newCnt,
	}
	// refresh cache
	if err := s.dao.DelKey(c, s.dao.UserKey(mid)); err != nil {
		return err
	}
	// update db
	if err := s.dao.UpNoteUser(c, userCache); err != nil {
		return err
	}
	return nil
}

func (s *Service) OfflineCommentOperation(ctx context.Context, newData *article.ArtDetailDB) (err error) {
	// 判断泰山中状态
	var (
		opidInfo = &common.TaishanCvidMappingOpidInfo{}
	)
	key := fmt.Sprintf(common.Cvid_Mapping_Opid_Taishan_Key, newData.Cvid)
	record, err := s.artDao.GetTaishan(ctx, key, common.TaishanConfig.NoteReply)
	if err != nil {
		log.Errorc(ctx, "OfflineCommentOperation query opidInfo from taiShan failed:(%v), key:%v", err, key)
		return
	}
	if record == nil || len(record.Columns) == 0 || len(record.Columns[0].Value) == 0 {
		log.Infoc(ctx, "OfflineCommentOperation query opidInfo from taiShan record isEmpty:(%v), key:%v", record, key)
		return
	}
	err = json.Unmarshal(record.Columns[0].Value, opidInfo)
	if err != nil {
		log.Errorc(ctx, "OfflineCommentOperation opidInfo from taiShan Unmarshal err %v and key %v", err, key)
		return
	}
	if opidInfo.Status != common.Cvid_Opid_Attached || opidInfo.Opid == 0 {
		log.Warnc(ctx, "OfflineCommentOperation opidInfo Status from taishan is deleted opidInfo %v", opidInfo)
		return
	}
	// 调用评论提供的下线运营位接口
	if _, err = s.artDao.OfflineReplyOperation(ctx, opidInfo.Opid); err != nil {
		log.Errorc(ctx, "OfflineReplyOperation err  req %v err %v", newData.Oid, err)
		return
	}
	if err = s.artDao.UnbindArticleCommentTaishan(ctx, newData.Cvid, opidInfo.Opid); err != nil {
		log.Errorc(ctx, "OfflineCommentOperation UnbindArticleCommentTaishan err%v and key %v", opidInfo, key)
		return
	}
	return nil
}

func (s *Service) recordCvidToOpid(ctx context.Context, cvid int64, opid int64) (err error) {
	cvidKey := fmt.Sprintf(common.Cvid_Mapping_Opid_Taishan_Key, cvid)
	cvidMappingOpidInfo := &common.TaishanCvidMappingOpidInfo{
		Opid:   opid,
		Status: common.Cvid_Opid_Attached,
	}
	cvidMappingOpidValue, err := json.Marshal(cvidMappingOpidInfo)
	if err != nil {
		log.Errorc(ctx, "recordCvidToOpid marshal err %v and cvidMappingOpidInfo %v", err, cvidMappingOpidInfo)
		return
	}
	if err = s.artDao.PutTaishan(ctx, cvidKey, cvidMappingOpidValue, common.TaishanConfig.NoteReply); err != nil {
		return err
	}
	return
}
