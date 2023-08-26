package view

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	viewm "go-gateway/app/app-svr/app-view/interface/model/view"
	noteApi "go-gateway/app/app-svr/hkt-note/service/api"
	tagapi "go-gateway/app/app-svr/hkt-note/service/api"

	dmapi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	replyapi "git.bilibili.co/bapis/bapis-go/community/interface/reply"

	"github.com/pkg/errors"
)

const (
	_reply          = "reply"
	_replySelection = "reply_selection"
	_dm             = "danmaku"
)

func (s *Service) Dots(c context.Context, aid, mid int64, plat int8) (*viewm.DotsReply, error) {
	reply, err := s.arcDao.SimpleArc(c, aid)
	if err != nil || reply == nil {
		return nil, errors.Wrap(err, "s.arcClient.SimpleArc reply error")
	}
	var (
		replyArcTag  *tagapi.ArcNotesCountReply
		dmUpRes      *dmapi.SubjectInfosByAidReply
		dmClosed     bool
		replyRes     *replyapi.SubjectInteractionStatusReply
		replyStatRes *replyapi.SubjectInteractionStatusReply // 评论区管控状态
		arcForbid    bool                                    // 稿件分区/mid管控
	)
	group := errgroup.WithContext(c)
	// up管控相关
	isUp := reply.Mid == mid
	if isUp {
		group.Go(func(ctx context.Context) error {
			var err error
			req := &replyapi.SubjectInteractionStatusReq{Oid: aid, Mid: mid, Type: _avTypeAv}
			replyRes, err = s.replyClient.SubjectInteractionStatus(ctx, req)
			if err != nil {
				log.Error("s.replyClient.SubjectInteractionStatus error(%+v)", err)
				return err
			}
			return nil
		})
		// up是否有权限开闭弹幕池
		group.Go(func(ctx context.Context) error {
			var err error
			req := &dmapi.SubjectInfosByAidReq{Aid: aid, Mid: mid, Type: _avTypeAv}
			dmUpRes, err = s.dmClient.SubjectInfosByAid(ctx, req)
			if err != nil {
				log.Error("s.dmClient.SubjectInfosByAid error(%+v)", err)
				return err
			}
			return nil
		})
	}
	// 笔记相关
	if len(reply.Cids) > 0 {
		group.Go(func(ctx context.Context) error {
			subRes, err := s.dmDao.SubjectInfos(ctx, _avTypeAv, plat, reply.Cids[0])
			if err != nil {
				log.Error("s.dmClient.SubjectInfos aid(%d) error(%+v)", aid, err)
				return err
			}
			if len(subRes) == 0 {
				log.Error("s.dmClient.SubjectInfos aid(%d) res nil", aid)
				return ecode.NothingFound
			}
			for _, v := range subRes {
				if v.Closed {
					dmClosed = true
				}
			}
			return nil
		})
	}
	if mid > 0 {
		group.Go(func(ctx context.Context) error {
			var err error
			req := &replyapi.SubjectInteractionStatusReq{Oid: aid, Mid: mid, Type: _avTypeAv}
			replyStatRes, err = s.replyClient.SubjectStatus(ctx, req)
			if err != nil {
				log.Error("s.replyClient.SubjectStatus error(%+v)", err)
				return err
			}
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		var err error
		replyArcTag, err = s.noteClient.ArcNotesCount(ctx, &tagapi.ArcNotesCountReq{Oid: aid})
		if err != nil || replyArcTag == nil {
			log.Error("s.hktNoteClient.ArcTag error(%+v)", err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		fbdRes, e := s.noteClient.ArcsForbid(ctx, &noteApi.ArcsForbidReq{Aids: []int64{aid}})
		if e != nil {
			log.Warn("noteWarn ArcsForbid aid(%d) err(%+v)", aid, e)
			return nil
		}
		if fbdRes != nil && fbdRes.Items != nil {
			arcForbid = fbdRes.Items[aid]
		}
		return nil
	})
	//弹幕与评论接口互不影响
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return handleDotsRes(dmUpRes, replyRes, replyStatRes, replyArcTag, isUp, dmClosed, arcForbid, mid)
}

func handleDotsRes(rawDmRes *dmapi.SubjectInfosByAidReply, rawReplyRes, replyStatRes *replyapi.SubjectInteractionStatusReply, replyArcTag *tagapi.ArcNotesCountReply, isUp, dmClosed, arcForbid bool, mid int64) (*viewm.DotsReply, error) {
	//danmaku与reply 互不影响
	// up主交互管理
	interM := func() *viewm.InteractionManagement {
		im := &viewm.InteractionManagement{}
		if !isUp {
			return im
		}
		if rawReplyRes == nil && rawDmRes == nil {
			return im
		}
		unifyStatus := make([]*viewm.InteractionStatus, 0)
		if rawReplyRes != nil {
			if rawReplyRes.UpClose != nil && rawReplyRes.UpClose.CanModify {
				unifyStatus = append(unifyStatus, &viewm.InteractionStatus{Status: rawReplyRes.UpClose.Status, Name: _reply})
			}
			if rawReplyRes.UpSelection != nil && rawReplyRes.UpSelection.CanModify {
				unifyStatus = append(unifyStatus, &viewm.InteractionStatus{Status: rawReplyRes.UpSelection.Status, Name: _replySelection})
			}
		}
		if rawDmRes != nil && rawDmRes.CanModify {
			unifyStatus = append(unifyStatus, &viewm.InteractionStatus{Status: int64(rawDmRes.Status), Name: _dm})
		}
		if len(unifyStatus) == 0 {
			return im
		}
		im.CanShow = true
		im.InteractionStatus = unifyStatus
		return im
	}()
	// 评论区、弹幕是否开放
	var (
		canDmShow    = !dmClosed
		canReplyShow = replyStatRes != nil && !replyStatRes.IsControlled
	)
	noteM := &viewm.NoteManagement{CanShow: canReplyShow && canDmShow && !arcForbid}
	if replyArcTag != nil {
		noteM.Count = replyArcTag.NotesCount
	}
	//未登录用户端上写死了一定展示按钮，不能根据上面的条件判断
	if mid == 0 {
		noteM.CanShow = true
	}
	return &viewm.DotsReply{InteractionManagement: interM, NoteManagement: noteM}, nil
}
