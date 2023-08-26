package gitlab

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	sagamdl "go-gateway/app/app-svr/fawkes/service/model/saga"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	GitMergeEvent   = "inner.git.merge"
	GitCommentEvent = "inner.git.comment"

	GitJobStatusChangeEvent = "git.job.status.change"
)

type MergeArgs struct {
	AppKey string
	HookMr *sagamdl.HookMR
}

type CommentArgs struct {
	AppKey      string
	HookComment *sagamdl.HookComment
}

// MergeAction merge事件
func (s *Service) MergeAction(args MergeArgs) (err error) {
	attr := args.HookMr.ObjectAttributes
	action := attr.Action
	state := attr.State
	mergeArgByte, err := json.Marshal(args)
	if err != nil {
		log.Warn("IID: %d, MergeArgs: %s", attr.IID, mergeArgByte)
	}
	if action == "open" && state == "opened" {
		// 开始一个mr
		var id int64
		id, err = s.fkDao.MergeInfoInsert(context.Background(), attr.IID, args.AppKey, attr.Source.PathWithNamespace, state, action, args.HookMr.User.UserName, attr.Title, time.Time{}, time.Time{})
		if err != nil || id <= 0 {
			err = errors.Wrap(err, fmt.Sprintf("MergeInfoInsert ERROR mergeID: %d, appKey: %s, state: %s, action: %s", attr.IID, args.AppKey, state, action))
		}
	} else if action == "merge" && state == "merged" {
		// merge成功 更新merge成功时间
		_, err = s.fkDao.MergedTimeUpdate(context.Background(), attr.IID, state, action, time.Now())
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("MergedTimeUpdate ERROR mergeID: %d, appKey: %s, state: %s, action: %s", attr.IID, args.AppKey, state, action))
		}
	} else if action == "close" && state == "closed" {
		// 关闭
		_, err = s.fkDao.MergedStateUpdate(context.Background(), attr.IID, state, action)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("MergedTimeUpdate ERROR mergeID: %d, appKey: %s, state: %s, action: %s", attr.IID, args.AppKey, state, action))
		}
	}
	return
}

// CommentAction merge comment事件
func (s *Service) CommentAction(args CommentArgs) (err error) {
	if args.HookComment.MergeRequest == nil || args.HookComment.ObjectAttributes == nil {
		marshal, _ := json.Marshal(args)
		log.Warn("mrCommentArgs: %s", marshal)
		return
	}
	CommentArgByte, err := json.Marshal(args)
	if err != nil {
		log.Warn("IID: %d, MergeArgs: %s", args.HookComment.MergeRequest.IID, CommentArgByte)
	}
	mergeId := args.HookComment.MergeRequest.IID
	command := args.HookComment.ObjectAttributes.Note
	if command == "+mr" || command == "+merge" {
		// 合并开始 更新开始合并时间
		_, err = s.fkDao.MergeStartTimeUpdate(context.Background(), mergeId, time.Now())
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("MergedTimeUpdate ERROR mergeID: %d, appKey: %s, command: %s", mergeId, args.AppKey, command))
		}
	}
	return
}

func (s *Service) SubscribeAsync(topic string, fn interface{}) (err error) {
	err = s.event.SubscribeAsync(topic, fn, false)
	return
}
