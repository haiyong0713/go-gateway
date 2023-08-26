package view

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	egv2 "go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-view/interface/model/share"
	arcCode "go-gateway/app/app-svr/archive/ecode"
	arc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	acc "git.bilibili.co/bapis/bapis-go/account/service"
	rel "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

func (s *Service) ShareInfo(c context.Context, params *share.InfoParam) (*share.InfoReply, error) {
	aid, err := bvid.BvToAv(params.Bvid)
	if err != nil {
		return nil, ecode.RequestErr
	}

	var (
		accInfos *acc.InfosReply
		arcInfo  *arc.Arc
		stat     *rel.StatReply
	)
	arcInfo, err = s.arcDao.Archive(c, aid)
	if err != nil {
		log.Error("shareInfo s.arcDao.Archive params(%+v), error(%+v)", params, err)
		if ecode.EqualError(ecode.NothingFound, err) {
			err = arcCode.ArchiveNotExist
		}
		return nil, err
	}
	if !(arcInfo.IsNormal() || arcInfo.IsNormalPremiere()) {
		log.Error("shareInfo arc invalid status(%d) attr(%d)", arcInfo.State, arcInfo.AttributeV2)
		return nil, arcCode.ArchiveNotExist
	}

	g := egv2.WithContext(c)
	//拉取用户最新数据
	mids := []int64{arcInfo.Author.Mid}
	if params.Mid > 0 {
		mids = append(mids, params.Mid)
	}
	g.Go(func(ctx context.Context) (err error) {
		accInfos, err = s.accDao.GetInfos(ctx, mids)
		if err != nil {
			log.Error("shareInfo s.accDao.GetInfos error, params(%+v) mids(%+v),err(%+v)", params, mids, err)
		}
		return nil
	})
	g.Go(func(ctx context.Context) (err error) {
		stat, err = s.relDao.Stat(ctx, arcInfo.Author.Mid)
		if err != nil {
			log.Error("shareInfo s.relDao.Stat error, params(%+v) mid(%d),err(%+v)", params, arcInfo.Author.Mid, err)
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("ShareInfo g.Wait error %+v", err)
	}

	return buildInfoReply(params.Mid, arcInfo, accInfos, stat), nil
}

func buildInfoReply(mid int64, arc *arc.Arc, accountInfos *acc.InfosReply, stat *rel.StatReply) *share.InfoReply {
	reply := &share.InfoReply{}

	premiereStatus := 0
	if arc.Premiere != nil {
		premiereStatus = int(arc.Premiere.State)
	}
	reply.Arc = &share.ArcReply{
		Aid:            arc.Aid,
		Title:          arc.Title,
		Pic:            arc.Pic,
		Duration:       arc.Duration,
		PremiereStatus: premiereStatus,
	}
	if accountInfos != nil && accountInfos.Infos != nil && accountInfos.Infos[mid] != nil {
		reply.Requester = &share.RequesterReply{
			Name: accountInfos.Infos[mid].Name,
			Face: accountInfos.Infos[mid].Face,
		}
	}
	if accountInfos != nil && accountInfos.Infos != nil && accountInfos.Infos[arc.Author.Mid] != nil {
		var fans int64
		if stat != nil {
			fans = stat.Follower
		}
		reply.Author = &share.AuthorReply{
			Name: accountInfos.Infos[arc.Author.Mid].Name,
			Face: accountInfos.Infos[arc.Author.Mid].Face,
			Fans: fans,
		}
	}
	return reply
}
