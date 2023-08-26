package poll

import (
	"context"
	"sync"
	"time"

	xecode "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	dao "go-gateway/app/web-svr/activity/interface/dao/poll"
	model "go-gateway/app/web-svr/activity/interface/model/poll"

	"go-common/library/database/sql"
	"go-common/library/log"
)

type pollStatTopStore struct {
	sync.RWMutex
	store map[int64][]*model.PollOptionStat
}

// Get is
func (ps *pollStatTopStore) Get(pollID int64) ([]*model.PollOptionStat, bool) {
	ps.RLock()
	defer ps.RUnlock()
	out, ok := ps.store[pollID]
	return out, ok
}

func (ps *pollStatTopStore) Set(pollID int64, in []*model.PollOptionStat) {
	ps.Lock()
	defer ps.Unlock()
	ps.store[pollID] = in
}

// Poll is
type Poll struct {
	manageMid        map[int64]struct{}
	dao              *dao.Dao
	pollStatTopStore pollStatTopStore
}

func (p *Poll) loadAllStatTop(ctx context.Context) error {
	log.Info("Load all poll stat top start at: %+v", time.Now())

	allPollMeta, err := p.dao.AllPollMeta(ctx)
	if err != nil {
		return err
	}
	for _, pm := range allPollMeta {
		// nil 过滤
		if pm == nil {
			continue
		}
		stats, err := p.dao.PollOptionStatTop(ctx, pm.Id, 10)
		if err != nil {
			log.Error("Failed to load poll option stat top: poll: %d: %+v", pm.Id, err)
			continue
		}
		p.pollStatTopStore.Set(pm.Id, stats)
	}

	log.Info("Load all poll stat top finished at: %+v", time.Now())
	return nil
}

func (p *Poll) loadstattopproc() {
	for {
		time.Sleep(time.Millisecond * 200)

		if err := p.loadAllStatTop(context.Background()); err != nil {
			log.Error("Failed to load all poll stat top: %+v", err)
			time.Sleep(time.Second * 5)
		}
	}
}

// New is
func New(conf *conf.Config) *Poll {
	poll := &Poll{
		manageMid: map[int64]struct{}{},
		dao:       dao.New(conf),
		pollStatTopStore: pollStatTopStore{
			store: map[int64][]*model.PollOptionStat{},
		},
	}
	for _, mid := range conf.Rule.PollManageMid {
		poll.manageMid[mid] = struct{}{}
	}
	if err := poll.loadAllStatTop(context.Background()); err != nil {
		panic(err)
	}
	go poll.loadstattopproc()
	return poll
}

func (p *Poll) pollMeta(ctx context.Context, id int64) (*model.PollMeta, error) {
	pm, ok := p.dao.PollMeta(ctx, id)
	if !ok {
		return nil, xecode.ActivityPollNotExist
	}
	return pm, nil
}

// PollMeta is
func (p *Poll) PollMeta(ctx context.Context, req *api.PollMetaReq) (*api.PollMetaReply, error) {
	pm, err := p.pollMeta(ctx, req.PollId)
	if err != nil {
		return nil, err
	}
	reply := &api.PollMetaReply{
		Id:          pm.Id,
		Title:       pm.Title,
		UniqueTable: pm.UniqueTable,
		Repeatable:  pm.Repeatable,
		DailyChance: pm.DailyChance,
		VoteMaximum: pm.VoteMaximum,
		EndAt:       pm.EndAt,
	}
	return reply, nil
}

func (p *Poll) pollOptions(ctx context.Context, pollID int64) ([]*model.PollOption, error) {
	options, ok := p.dao.ListPollOption(ctx, pollID)
	if !ok {
		return []*model.PollOption{}, nil
	}
	return options, nil
}

// PollOptions is
func (p *Poll) PollOptions(ctx context.Context, req *api.PollOptionsReq) (*api.PollOptionsReply, error) {
	options, err := p.pollOptions(ctx, req.PollId)
	if err != nil {
		return nil, err
	}
	reply := &api.PollOptionsReply{
		Options: make([]*api.PollOption, 0, len(options)),
	}
	for _, o := range options {
		ro := &api.PollOption{
			Id:     o.Id,
			PollId: o.PollId,
			Title:  o.Title,
			Image:  o.Image,
			Group:  o.Group,
		}
		reply.Options = append(reply.Options, ro)
	}
	return reply, nil
}

func asOptionMap(in []*model.PollOption) map[int64]*model.PollOption {
	out := make(map[int64]*model.PollOption, len(in))
	for _, o := range in {
		out[o.Id] = o
	}
	return out
}

// PollVote is
func (p *Poll) PollVote(ctx context.Context, req *api.PollVoteReq) error {
	// 检查投票话题是否存在
	pm, err := p.pollMeta(ctx, req.PollId)
	if err != nil {
		return err
	}
	now := time.Now()

	if pm.EndAt > 0 {
		if now.Unix() > pm.EndAt {
			return xecode.ActivityPollEnd
		}
	}

	// 检查投票数量
	if int64(len(req.Vote)) != pm.VoteMaximum {
		return xecode.ActivityLackOfPollVote
	}

	// 检查选项是否合法
	options, err := p.pollOptions(ctx, req.PollId)
	if err != nil {
		return err
	}
	optionMap := asOptionMap(options)
	for _, v := range req.Vote {
		_, ok := optionMap[v.PollOptionId]
		if !ok {
			return xecode.ActivityPollOptionNotExist
		}
	}

	// 检查用户提交次数
	if err := func() error {
		if !pm.Repeatable {
			stat, err := p.dao.LastPollVoteUserStat(ctx, req.Mid, req.PollId)
			if err != nil {
				return err
			}
			if stat == nil {
				if !p.dao.InitPollVoteUserStat(ctx, req.Mid, req.PollId, now) {
					return xecode.ActivityPollAlreadyVoted
				}
				return nil
			}
			if stat.VoteCount > 0 {
				return xecode.ActivityPollAlreadyVoted
			}
			return nil
		}

		stat, err := p.dao.PollVoteUserStatByDate(ctx, req.Mid, req.PollId, now)
		if err != nil {
			return err
		}
		if stat == nil {
			p.dao.InitPollVoteUserStat(ctx, req.Mid, req.PollId, now)
			return nil
		}
		if stat.VoteCount >= pm.DailyChance {
			return xecode.ActivityPollExceededDailyChance
		}
		return nil
	}(); err != nil {
		return err
	}

	defer func() {
		p.dao.DelCacheLastPollVoteUserStat(ctx, req.Mid, req.PollId)
		p.dao.DelCachePollVoteUserStatByDate(ctx, req.Mid, req.PollId, now)
	}()
	maxVoteCount := int64(1)
	if pm.Repeatable {
		maxVoteCount = pm.DailyChance
	}
	if err := p.dao.Transact(ctx, func(tx *sql.Tx) error {
		hasChance, err := p.dao.TxIncrDailyUserVoteStat(tx, req.Mid, req.PollId, now, maxVoteCount)
		if err != nil {
			return err
		}
		if !hasChance {
			return xecode.ActivityPollExceededDailyChance
		}

		for _, v := range req.Vote {
			if err := p.dao.TxAddVote(tx, &model.PollVote{
				PollId:       req.PollId,
				Mid:          req.Mid,
				PollOptionId: v.PollOptionId,
				TicketCount:  v.Count,
				VoteAt:       now.Unix(),
			}); err != nil {
				return err
			}
			if err := p.dao.TxIncrOptionStat(tx, req.PollId, v.PollOptionId, v.Count); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// PollOptionStatTop is
func (p *Poll) PollOptionStatTop(ctx context.Context, req *api.PollOptionStatTopReq) (*api.PollOptionStatTopReply, error) {
	pm, err := p.pollMeta(ctx, req.PollId)
	if err != nil {
		return nil, err
	}

	options, err := p.pollOptions(ctx, req.PollId)
	if err != nil {
		return nil, err
	}
	optionMap := asOptionMap(options)

	reply := &api.PollOptionStatTopReply{
		OptionStats: []*api.PollOptionStatReply{},
	}
	res, ok := p.pollStatTopStore.Get(pm.Id)
	if !ok {
		return reply, nil
	}
	for _, s := range res {
		opt, ok := optionMap[s.PollOptionId]
		if !ok {
			log.Warn("Skip this option: %q", s.PollOptionId)
			continue
		}
		as := &api.PollOptionStatReply{
			Id:           s.Id,
			PollId:       s.PollId,
			PollOptionId: s.PollOptionId,
			TicketSum:    s.TicketSum,
			VoteSum:      s.VoteSum,
			PollOption: &api.PollOption{
				Id:     opt.Id,
				PollId: opt.PollId,
				Title:  opt.Title,
				Image:  opt.Image,
				Group:  opt.Group,
			},
		}
		reply.OptionStats = append(reply.OptionStats, as)
	}
	return reply, nil
}

// PollS9Vote is
func (p *Poll) PollS9Vote(ctx context.Context, req *api.PollVoteReq) error {
	pm, err := p.pollMeta(ctx, req.PollId)
	if err != nil {
		return err
	}
	if pm.Title != "2019英雄联盟总决赛" {
		return xecode.ActivityPollVoteInvalid
	}

	countRequired := map[int64]struct{}{
		1:  {},
		2:  {},
		3:  {},
		4:  {},
		5:  {},
		6:  {},
		7:  {},
		8:  {},
		9:  {},
		10: {},
	}
	for _, v := range req.Vote {
		if v.Count <= 0 || v.Count > 10 {
			return xecode.ActivityPollVoteInvalid
		}
		if _, ok := countRequired[v.Count]; !ok {
			return xecode.ActivityPollVoteInvalid
		}
		delete(countRequired, v.Count)
	}
	if len(countRequired) != 0 {
		return xecode.ActivityPollVoteInvalid
	}
	return p.PollVote(ctx, req)
}

// PollVoted is
func (p *Poll) PollVoted(ctx context.Context, req *api.PollVotedReq) (*api.PollVotedReply, error) {
	reply := &api.PollVotedReply{
		Mid:            req.Mid,
		PollId:         req.PollId,
		Voted:          false,
		DailyVoteCount: 0,
	}

	func() {
		stat, err := p.dao.LastPollVoteUserStat(ctx, req.Mid, req.PollId)
		if err != nil {
			log.Error("Failed to get last poll vote user stat: %+v: %+v", req, err)
			return
		}
		if stat != nil {
			reply.Voted = true
			return
		}
	}()

	func() {
		now := time.Now()
		stat, err := p.dao.PollVoteUserStatByDate(ctx, req.Mid, req.PollId, now)
		if err != nil {
			log.Error("Failed to get daily poll vote user stat: %+v: %+v", req, err)
			return
		}
		if stat != nil {
			reply.DailyVoteCount = stat.VoteCount
			return
		}
	}()

	return reply, nil
}

// PollMOptions is
func (p *Poll) PollMOptions(ctx context.Context, pollID int64) ([]*model.PollOption, error) {
	return p.dao.PollOptions(ctx, pollID)
}

// PollMOptionDelete is
func (p *Poll) PollMOptionDelete(ctx context.Context, pollOptionID int64) error {
	return p.dao.PollOptionsDelete(ctx, pollOptionID)
}

// PollMOptionsAdd is
func (p *Poll) PollMOptionsAdd(ctx context.Context, pollID int64, title string, image string, group string) error {
	return p.dao.PollOptionsAdd(ctx, pollID, title, image, group)
}

// PollMOptionsUpdate is
func (p *Poll) PollMOptionsUpdate(ctx context.Context, pollOptionID int64, title string, image string, group string) error {
	return p.dao.PollOptionsUpdate(ctx, pollOptionID, title, image, group)
}

// PollM is
func (p *Poll) PollM(ctx context.Context, mid int64) bool {
	_, ok := p.manageMid[mid]
	return ok
}
