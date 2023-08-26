package ranklist

import (
	"context"
	"math/rand"
	"sort"
	"strings"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	activityapi "git.bilibili.co/bapis/bapis-go/activity/service"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-feed/interface/model/sets"
	"go-gateway/app/app-svr/app-show/interface/conf"
	accdao "go-gateway/app/app-svr/app-show/interface/dao/account"
	actdao "go-gateway/app/app-svr/app-show/interface/dao/act"
	arcdao "go-gateway/app/app-svr/app-show/interface/dao/archive"
	dao "go-gateway/app/app-svr/app-show/interface/dao/rank-list"
	model "go-gateway/app/app-svr/app-show/interface/model/rank-list"
	archiveapi "go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

const (
	_maxFloatingLen     = 5
	_maxArchiveRankSize = 50
)

// Service is
type Service struct {
	c   *conf.Config
	arc *arcdao.Dao
	acc *accdao.Dao
	act *actdao.Dao
	dao *dao.Dao
}

type fanoutArgs struct {
	Aids   []int64
	Mids   []int64
	ActIDs []int64
}

type fanoutResult struct {
	Archives        map[int64]*archiveapi.Arc
	Accounts        map[int64]*accountapi.Info
	ActSubProtocols map[int64]*activityapi.ActSubProtocolReply
}

type floatingArgs struct {
	actAids        []int64
	fromViewAid    int64
	archiveRanking map[int64]int64
	promotText     []string
}

func New(c *conf.Config) *Service {
	return &Service{
		c:   c,
		arc: arcdao.New(c),
		acc: accdao.New(c),
		act: actdao.New(c),
		dao: dao.New(c),
	}
}

// RankListIndex is
func (s *Service) RankListIndex(ctx context.Context, req *model.IndexReq) (*model.IndexReply, error) {
	meta, err := s.dao.RankMeta(ctx, req.ID)
	if err != nil {
		if err == redis.ErrNil {
			return nil, errors.WithStack(ecode.NothingFound)
		}
		return nil, err
	}
	out := &model.IndexReply{
		ID:     meta.ID,
		Title:  meta.RankConfig.Title,
		Cover:  meta.RankConfig.Cover,
		State:  meta.RankState,
		Tids:   meta.RankConfig.Tids,
		ActIDs: meta.RankConfig.ActIDs,
	}
	out.Description = meta.RankConfig.Description

	args := &fanoutArgs{}
	var actAids []int64
	if req.Mid != 0 {
		actAids = s.dao.UpActivityArchive(ctx, req.Mid, meta.RankConfig.ActIDs)
		if len(actAids) > _maxFloatingLen {
			actAids = actAids[:_maxFloatingLen]
		}
		args.Aids = append(args.Aids, actAids...)
	}
	fromViewAid := req.ResolveFromViewArchive()
	if fromViewAid > 0 {
		args.Aids = append(args.Aids, fromViewAid)
	}
	collectFanoutArgs(meta, args)
	fanoutResult := s.doRankListFanout(ctx, args, req.Mid, req.MobiApp, req.Device)

	out.Tags = refineTags(fanoutResult.ActSubProtocols)

	var archiveRanking map[int64]int64
	switch meta.RankState {
	case model.StatePending, model.StateVoting, model.StateStopped:
		s.withRankArchive(fanoutResult, meta.RankVideos, &out.RankArchive)
		archiveRanking = makeArchiveRanking(out.RankArchive)
	case model.StateRankFinished:
		s.withRankPayload(fanoutResult, meta.FinalRank, &out.RankPayload)
		archiveRanking = makeArchiveRanking(firstArchiveRank(out.RankPayload))
	default:
		log.Error("Unrecognized rank list state: %d on id: %d: %+v", out.State, out.ID, meta)
	}

	floatingArgs := &floatingArgs{
		actAids:        actAids,
		fromViewAid:    fromViewAid,
		archiveRanking: archiveRanking,
		promotText:     meta.RankConfig.HelpTips,
	}
	s.withFloating(fanoutResult, floatingArgs, &out.Floating)

	return out, nil
}

func refineTags(in map[int64]*activityapi.ActSubProtocolReply) []string {
	tagSet := sets.NewString()
	for _, v := range in {
		if v.Protocol == nil {
			continue
		}
		for _, t := range strings.Split(v.Protocol.Tags, ",") {
			if t == "" {
				continue
			}
			tagSet.Insert(t)
		}
	}
	return tagSet.List()
}

func firstArchiveRank(in []*model.RankPayload) []*model.RankArchive {
	for _, i := range in {
		if i.Mode == model.PayloadModeArchive {
			return i.ArchiveList
		}
	}
	return nil
}

func makeArchiveRanking(in []*model.RankArchive) map[int64]int64 {
	out := make(map[int64]int64, len(in))
	for _, i := range in {
		out[i.Aid] = i.Ranking
	}
	return out
}

func (s *Service) doRankListFanout(ctx context.Context, args *fanoutArgs, mid int64, mobiApp, device string) *fanoutResult {
	result := &fanoutResult{}

	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		reply, err := s.arc.ArchivesPB(ctx, args.Aids, mid, mobiApp, device)
		if err != nil {
			log.Error("Failed to get archives: %+v: %+v", args.Aids, err)
			return nil
		}
		result.Archives = reply
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		reply, err := s.acc.Infos3GRPC(ctx, args.Mids)
		if err != nil {
			log.Error("Failed to get account infos: %+v: %+v", args.Mids, err)
			return nil
		}
		result.Accounts = reply
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		reply, err := s.act.ActSubsProtocol(ctx, args.ActIDs)
		if err != nil {
			log.Error("Failed to get act subs protocol: %+v: %+v", args.ActIDs, err)
			return nil
		}
		result.ActSubProtocols = reply
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Failed to execute fanout errgroup: %+v: %+v", args, err)
		return nil
	}

	return result
}

func collectFanoutArgs(meta *model.Meta, dst *fanoutArgs) {
	dst.ActIDs = meta.RankConfig.ActIDs
	switch meta.RankState {
	case model.StatePending, model.StateVoting, model.StateStopped:
		dst.Aids = append(dst.Aids, meta.RankVideos...)
	case model.StateRankFinished:
		for _, item := range meta.FinalRank {
			switch item.Mode {
			case model.PayloadModeArchive:
				dst.Aids = append(dst.Aids, item.List...)
			case model.PayloadModeAccount:
				dst.Mids = append(dst.Mids, item.List...)
			default:
				log.Warn("Unrecognized final rank mode: %d: %+v", item.Mode, meta)
				continue
			}
		}
	default:
		log.Warn("Unrecognized rank state: %d: %+v", meta.RankState, meta)
	}
}

func (s *Service) withFloating(result *fanoutResult, args *floatingArgs, dst *[]*model.Floating) {
	if args.fromViewAid > 0 {
		func() {
			fromViewArc, ok := result.Archives[args.fromViewAid]
			if !ok {
				log.Warn("Failed to get watching archive: %d", args.fromViewAid)
				return
			}
			fi := &model.Floating{
				Type: model.FloatTypeWatching,
				Archive: model.ArchiveSchema{
					Aid:       fromViewArc.Aid,
					Title:     fromViewArc.Title,
					Pic:       fromViewArc.Pic,
					MissionID: fromViewArc.MissionID,
					Author: model.AuthorSchema{
						Mid:  fromViewArc.Author.Mid,
						Name: fromViewArc.Author.Name,
						Face: fromViewArc.Author.Face,
					},
				},
			}
			fi.RankStatus.Ranking, fi.RankStatus.Ranked = args.archiveRanking[args.fromViewAid]
			*dst = append(*dst, fi)
		}()
	}

	actUpArcs := make([]*model.Floating, 0, len(args.actAids))
	for _, aid := range args.actAids {
		arc, ok := result.Archives[aid]
		if !ok {
			continue
		}
		fi := &model.Floating{
			Type: model.FloatTypeOwner,
			Archive: model.ArchiveSchema{
				Aid:       arc.Aid,
				Title:     arc.Title,
				Pic:       arc.Pic,
				MissionID: arc.MissionID,
				Author: model.AuthorSchema{
					Mid:  arc.Author.Mid,
					Name: arc.Author.Name,
					Face: arc.Author.Face,
				},
			},
			Text: randChoice(args.promotText),
		}
		fi.RankStatus.Ranking, fi.RankStatus.Ranked = args.archiveRanking[aid]
		actUpArcs = append(actUpArcs, fi)
	}
	sort.Slice(actUpArcs, func(i, j int) bool {
		return actUpArcs[i].RankStatus.Ranking < actUpArcs[j].RankStatus.Ranking
	})
	*dst = append(*dst, actUpArcs...)
	if len(*dst) > _maxFloatingLen {
		*dst = (*dst)[:_maxFloatingLen]
	}
}

func randChoice(in []string) string {
	return in[rand.Intn(len(in))]
}

func (s *Service) withRankPayload(result *fanoutResult, final []*model.FinalRankItem, dst *[]*model.RankPayload) {
	for _, item := range final {
		rp := &model.RankPayload{
			Mode:  item.Mode,
			Title: item.Title,
		}
		switch item.Mode {
		case model.PayloadModeAccount:
			rp.AccountList = constructRankAccount(item, result.Accounts)
		case model.PayloadModeArchive:
			rp.ArchiveList = constructRankArchive(item, result.Archives)
		default:
			log.Warn("Unrecognized final rank mode: %d: %+v", item.Mode, final)
			continue
		}
		*dst = append(*dst, rp)
	}
}

func constructRankArchive(item *model.FinalRankItem, archives map[int64]*archiveapi.Arc) []*model.RankArchive {
	out := make([]*model.RankArchive, 0, len(item.List))
	for _, aid := range item.List {
		arc, ok := archives[aid]
		if !ok {
			log.Warn("Unexpected archive with aid: %d", aid)
			continue
		}
		as := model.ArchiveSchema{
			Aid:       arc.Aid,
			Title:     arc.Title,
			Pic:       arc.Pic,
			MissionID: arc.MissionID,
			Author: model.AuthorSchema{
				Mid:  arc.Author.Mid,
				Name: arc.Author.Name,
				Face: arc.Author.Face,
			},
		}
		out = append(out, &model.RankArchive{
			ArchiveSchema: as,
		})
	}
	out = shrinkSize(out, _maxArchiveRankSize)
	rankArchiveSlice(out).markRanking()
	return out
}

func constructRankAccount(item *model.FinalRankItem, accounts map[int64]*accountapi.Info) []*model.RankAuthor {
	out := make([]*model.RankAuthor, 0, len(item.List))
	for _, mid := range item.List {
		accinfo, ok := accounts[mid]
		if !ok {
			log.Warn("Unexpected account info with mid: %d", mid)
			continue
		}
		as := model.AuthorSchema{
			Mid:  accinfo.Mid,
			Name: accinfo.Name,
			Face: accinfo.Face,
		}
		out = append(out, &model.RankAuthor{
			AuthorSchema: as,
		})
	}
	rankAuthorSlice(out).markRanking()
	return out
}

func (s *Service) withRankArchive(result *fanoutResult, aids []int64, dst *[]*model.RankArchive) {
	for _, aid := range aids {
		arc, ok := result.Archives[aid]
		if !ok {
			log.Warn("Unexcepted archive in rank archive: %d", aid)
			continue
		}
		as := model.ArchiveSchema{
			Aid:       arc.Aid,
			Title:     arc.Title,
			Pic:       arc.Pic,
			MissionID: arc.MissionID,
			Author: model.AuthorSchema{
				Mid:  arc.Author.Mid,
				Name: arc.Author.Name,
				Face: arc.Author.Face,
			},
		}
		*dst = append(*dst, &model.RankArchive{
			ArchiveSchema: as,
		})
	}
	*dst = shrinkSize(*dst, _maxArchiveRankSize)
	rankArchiveSlice(*dst).markRanking()
}

type rankArchiveSlice []*model.RankArchive

func shrinkSize(in []*model.RankArchive, size int) rankArchiveSlice {
	if len(in) > size {
		return in[:size]
	}
	return in
}

func (r rankArchiveSlice) markRanking() {
	for i, ra := range r {
		ra.Ranking = int64(i) + 1
	}
}

type rankAuthorSlice []*model.RankAuthor

func (r rankAuthorSlice) markRanking() {
	for i, ra := range r {
		ra.Ranking = int64(i) + 1
	}
}
