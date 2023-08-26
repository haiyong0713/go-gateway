package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/archive/service/api"

	"go-common/library/sync/errgroup.v2"
)

// Description update page by aid & cid
func (s *Service) Description(c context.Context, aid int64) (des string, descV2 []*api.DescV2, err error) {
	addit, err := s.arc.DescriptionV2(c, aid)
	if err != nil {
		return "", nil, err
	}
	if addit == nil {
		return "", nil, ecode.NothingFound
	}
	//数据拼装
	descV2 = s.arc.GetDescV2Params(addit.DescV2)
	des = addit.Description
	return
}

// Descriptions
func (s *Service) Descriptions(c context.Context, aids []int64) (map[int64]*api.DescriptionReply, error) {
	resp, err := s.arc.Descriptions(c, aids)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Page3 get page by aid
func (s *Service) Page3(c context.Context, aid int64) (ps []*api.Page, err error) {
	ps, err = s.arc.Videos3(c, aid)
	return
}

// View3 get archive with video.
func (s *Service) View3(c context.Context, aid int64, mid int64) (v *api.ViewReply, err error) {
	var (
		a  *api.Arc
		ps []*api.Page
	)
	var g = errgroup.WithContext(c)
	g.Go(func(c context.Context) (err error) {
		a, err = s.ArcWithStat(c, &api.ArcRequest{Aid: aid, Mid: mid})
		return
	})
	g.Go(func(c context.Context) (err error) {
		ps, err = s.arc.Videos3(c, aid)
		return
	})
	if err = g.Wait(); err != nil {
		return
	}
	v = &api.ViewReply{Arc: a}
	if a.IsSteinsGate() { // if steinsGate arc, pick the guide arc
		if len(s.steinsGuidePages) == 0 {
			err = ecode.NothingFound
			return
		}
		v.Pages = s.steinsGuidePages[:1]
		v.FirstCid = s.steinsGuidePages[0].Cid
	} else {
		v.Pages = ps
	}
	return
}

// SteinsGateView get steinsGate arc's view ( pages )
func (s *Service) SteinsGateView(c context.Context, aid int64, mid int64) (v *api.SteinsGateViewReply, err error) {
	caller := metadata.String(c, metadata.Caller)
	if _, ok := s.authorisedCallers[caller]; !ok {
		err = ecode.AccessDenied
		log.Warn("SteinsGateView Not Authorised Caller %s", caller)
		return
	}
	var (
		a  *api.Arc
		ps []*api.Page
	)
	var g = errgroup.WithContext(c)
	g.Go(func(c context.Context) (err error) {
		a, err = s.ArcWithStat(c, &api.ArcRequest{Aid: aid, Mid: mid})
		return err
	})
	g.Go(func(c context.Context) (err error) {
		ps, err = s.arc.Videos3(c, aid)
		return err
	})
	if err = g.Wait(); err != nil {
		return
	}
	v = &api.SteinsGateViewReply{Arc: a, Pages: ps}
	return
}

// Views3 get archives with videos.
func (s *Service) Views3(c context.Context, aids []int64, mid int64, mobiApp, device string) (vm map[int64]*api.ViewReply, err error) {
	if len(aids) == 0 {
		err = ecode.RequestErr
		return
	}
	var (
		am map[int64]*api.Arc
		vs map[int64][]*api.Page
	)
	var g = errgroup.WithContext(c)
	g.Go(func(c context.Context) (err error) {
		am, err = s.Archives3(c, aids, mid, mobiApp, device)
		return
	})
	g.Go(func(c context.Context) (err error) {
		vs, err = s.arc.VideosByAids3(c, aids) // always pick the steinsAid's video info to accelerate
		return
	})
	if err = g.Wait(); err != nil {
		return
	}
	vm = make(map[int64]*api.ViewReply, len(am))
	for _, a := range am {
		var pages = vs[a.Aid]
		if a.IsSteinsGate() { // give guide arc's page
			if len(s.steinsGuidePages) == 0 {
				log.Error("steins guide av wrong")
				continue
			}
			a.FirstCid = s.steinsGuidePages[0].Cid
			pages = s.steinsGuidePages[:1]
		}
		vm[a.Aid] = &api.ViewReply{
			Arc:   a,
			Pages: pages,
		}
	}
	return
}

// SteinsGateViews get archives with videos for steins-gate
func (s *Service) SteinsGateViews(c context.Context, aids []int64, mid int64) (vm map[int64]*api.SteinsGateViewReply, err error) {
	caller := metadata.String(c, metadata.Caller)
	if _, ok := s.authorisedCallers[caller]; !ok {
		err = ecode.AccessDenied
		log.Warn("SteinsGateViews Not Authorised Caller %s", caller)
		return
	}
	if len(aids) == 0 {
		err = ecode.RequestErr
		return
	}
	var (
		am map[int64]*api.Arc
		vs map[int64][]*api.Page
	)
	var g = errgroup.WithContext(c)
	g.Go(func(c context.Context) (err error) {
		am, err = s.Archives3(c, aids, mid, "", "")
		return
	})
	g.Go(func(c context.Context) (err error) {
		vs, err = s.arc.VideosByAids3(c, aids) // always pick the steinsAid's video info to accelerate
		return
	})
	if err = g.Wait(); err != nil {
		return
	}
	vm = make(map[int64]*api.SteinsGateViewReply, len(am))
	for _, a := range am {
		vm[a.Aid] = &api.SteinsGateViewReply{Arc: a, Pages: vs[a.Aid]}
	}
	return
}

// Video3 get video by aid & cid
func (s *Service) Video3(c context.Context, aid, cid int64) (video *api.Page, err error) {
	video, err = s.arc.Video3(c, aid, cid)
	return
}
