package grpc

import (
	"context"

	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"

	v1 "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/conf"
	"go-gateway/app/app-svr/archive/service/service"
)

type server struct {
	srv *service.Service
	c   *conf.Config
}

// New grpc server
func New(cfg *warden.ServerConfig, srv *service.Service, c *conf.Config) (wsvr *warden.Server, err error) {
	wsvr = warden.NewServer(cfg)
	limiter := quota.New(c.QuotaConf)
	wsvr.Use(limiter.Limit())
	v1.RegisterArchiveServer(wsvr.Server(), &server{srv: srv, c: c})
	wsvr, err = wsvr.Start()
	return
}

// Types get all types
func (s *server) Types(c context.Context, noArg *v1.NoArgRequest) (resp *v1.TypesReply, err error) {
	types := s.srv.AllTypes(c)
	resp = new(v1.TypesReply)
	resp.Types = make(map[int32]*v1.Tp)
	for _, tp := range types {
		resp.Types[int32(tp.ID)] = &v1.Tp{
			ID:   int32(tp.ID),
			Pid:  int32(tp.Pid),
			Name: tp.Name,
		}
	}
	return
}

// Arc get archive
func (s *server) Arc(c context.Context, req *v1.ArcRequest) (resp *v1.ArcReply, err error) {
	resp = new(v1.ArcReply)
	var a *v1.Arc
	a, err = s.srv.ArcWithStat(c, req)
	if err != nil {
		return nil, err
	}
	resp.Arc = a
	return
}

// Arcs get archives
func (s *server) Arcs(c context.Context, req *v1.ArcsRequest) (resp *v1.ArcsReply, err error) {
	resp = new(v1.ArcsReply)
	resp.Arcs = make(map[int64]*v1.Arc)
	as, err := s.srv.Archives3(c, req.Aids, req.Mid, req.MobiApp, req.Device)
	if err != nil {
		return
	}
	if len(as) == 0 {
		return
	}
	for aid, a := range as {
		resp.Arcs[aid] = a
	}
	return
}

// ArcsWithPlayurl get arcs with playurl
func (s *server) ArcsWithPlayurl(c context.Context, req *v1.ArcsWithPlayurlRequest) (resp *v1.ArcsWithPlayurlReply, err error) {
	resp = new(v1.ArcsWithPlayurlReply)
	resp.ArcWithPlayurl, err = s.srv.ArcsWithPlayurl(c, req)
	return
}

func (s *server) ArcsPlayer(c context.Context, req *v1.ArcsPlayerRequest) (resp *v1.ArcsPlayerReply, err error) {
	resp = new(v1.ArcsPlayerReply)
	arcsPlayer, err := s.srv.ArcsPlayerSvr(c, req)
	if err != nil {
		return nil, err
	}
	resp.ArcsPlayer = arcsPlayer
	return resp, nil
}

// View get archive and page
func (s *server) View(c context.Context, req *v1.ViewRequest) (resp *v1.ViewReply, err error) {
	resp = new(v1.ViewReply)
	v, err := s.srv.View3(c, req.Aid, req.Mid)
	if err != nil {
		return
	}
	resp = v
	return
}

// SteinsGateView get archive and page for steins-gate
func (s *server) SteinsGateView(c context.Context, req *v1.SteinsGateViewRequest) (resp *v1.SteinsGateViewReply, err error) {
	resp = new(v1.SteinsGateViewReply)
	v, err := s.srv.SteinsGateView(c, req.Aid, req.Mid)
	if err != nil {
		return
	}
	resp = v
	return
}

// Views get archives and pages
func (s *server) Views(c context.Context, req *v1.ViewsRequest) (resp *v1.ViewsReply, err error) {
	resp = new(v1.ViewsReply)
	resp.Views = make(map[int64]*v1.ViewReply)
	vs, err := s.srv.Views3(c, req.Aids, req.Mid, req.MobiApp, req.Device)
	if err != nil {
		return
	}
	if len(vs) == 0 {
		return
	}
	resp.Views = vs
	return
}

// SteinsGateViews get archives and pages for steins-gate
func (s *server) SteinsGateViews(c context.Context, req *v1.SteinsGateViewsRequest) (resp *v1.SteinsGateViewsReply, err error) {
	resp = new(v1.SteinsGateViewsReply)
	resp.Views = make(map[int64]*v1.SteinsGateViewReply)
	vs, err := s.srv.SteinsGateViews(c, req.Aids, req.Mid)
	if err != nil {
		return
	}
	if len(vs) == 0 {
		return
	}
	resp.Views = vs
	return
}

func (s *server) Stat(c context.Context, req *v1.StatRequest) (resp *v1.StatReply, err error) {
	resp = new(v1.StatReply)
	stat, err := s.srv.Stat3(c, req.Aid)
	if err != nil {
		return
	}
	resp.Stat = stat
	return
}

func (s *server) Stats(c context.Context, req *v1.StatsRequest) (resp *v1.StatsReply, err error) {
	resp = new(v1.StatsReply)
	resp.Stats = make(map[int64]*v1.Stat)
	stats, err := s.srv.Stats3(c, req.Aids)
	if err != nil {
		return
	}
	resp.Stats = stats
	return
}

func (s *server) Page(c context.Context, req *v1.PageRequest) (resp *v1.PageReply, err error) {
	resp = new(v1.PageReply)
	resp.Pages, err = s.srv.Page3(c, req.Aid)
	return
}

func (s *server) Video(c context.Context, req *v1.VideoRequest) (resp *v1.VideoReply, err error) {
	resp = new(v1.VideoReply)
	resp.Page, err = s.srv.Video3(c, req.Aid, req.Cid)
	return
}

func (s *server) Description(c context.Context, req *v1.DescriptionRequest) (resp *v1.DescriptionReply, err error) {
	resp = new(v1.DescriptionReply)
	resp.Desc, resp.DescV2Parse, err = s.srv.Description(c, req.Aid)
	return
}

func (s *server) Descriptions(c context.Context, req *v1.DescriptionsRequest) (resp *v1.DescriptionsReply, err error) {
	resp = new(v1.DescriptionsReply)
	resp.Description, err = s.srv.Descriptions(c, req.Aids)
	return
}

// VideoShot is
func (s *server) VideoShot(c context.Context, req *v1.VideoShotRequest) (resp *v1.VideoShotReply, err error) {
	return s.srv.Videoshot(c, req.Aid, req.Cid, req.Common)
}

// UpCount is
func (s *server) UpCount(c context.Context, req *v1.UpCountRequest) (resp *v1.UpCountReply, err error) {
	resp = new(v1.UpCountReply)
	count, err := s.srv.UpperCount(c, req.Mid)
	if err != nil {
		return
	}
	resp.Count = int64(count)
	return
}

// UpPass is
func (s *server) UpsPassed(c context.Context, req *v1.UpsPassedRequest) (resp *v1.UpsPassedReply, err error) {
	resp = new(v1.UpsPassedReply)
	upp, err := s.srv.UppersAidPubTime(c, req.Mids, int(req.Pn), int(req.Ps))
	if err != nil || len(upp) == 0 {
		return
	}
	var upPassed = make(map[int64]*v1.UpPassedInfo)
	for k, v := range upp {
		var tmpUp = new(v1.UpPassedInfo)
		var tmpInfo []*v1.UpPassed
		for _, up := range v {
			tmpInfo = append(tmpInfo, &v1.UpPassed{
				Aid:       up.Aid,
				PubDate:   up.PubDate,
				Copyright: int32(up.Copyright),
			})
		}
		tmpUp.UpPassedInfo = tmpInfo
		upPassed[k] = tmpUp
	}
	resp.UpsPassed = upPassed
	return
}

// UpArcs is
func (s *server) UpArcs(c context.Context, req *v1.UpArcsRequest) (resp *v1.UpArcsReply, err error) {
	resp = new(v1.UpArcsReply)
	upArc, err := s.srv.UpperPassed3(c, req.Mid, int(req.Pn), int(req.Ps))
	if err != nil {
		return
	}
	resp.Arcs = upArc
	return
}

// UpArcs is
func (s *server) Creators(c context.Context, req *v1.CreatorsRequest) (resp *v1.CreatorsReply, err error) {
	resp = new(v1.CreatorsReply)
	creators, err := s.srv.Creators(c, req.Aids)
	if err != nil {
		return
	}
	resp.Info = creators
	return
}

// SimpleArc get archive
func (s *server) SimpleArc(c context.Context, req *v1.SimpleArcRequest) (*v1.SimpleArcReply, error) {
	resp := new(v1.SimpleArcReply)
	a, err := s.srv.SimpleArc(c, req)
	if err != nil {
		return nil, err
	}
	resp.Arc = a
	return resp, nil
}

// SimpleArcs get archives
func (s *server) SimpleArcs(c context.Context, req *v1.SimpleArcsRequest) (*v1.SimpleArcsReply, error) {
	resp := new(v1.SimpleArcsReply)
	as, err := s.srv.SimpleArcs(c, req.Aids, req.Mid)
	if err != nil {
		return nil, err
	}
	resp.Arcs = as
	return resp, nil
}

func (s *server) ArcsRedirectPolicy(c context.Context, req *v1.ArcsRedirectPolicyRequest) (*v1.ArcsRedirectPolicyReply, error) {
	resp := new(v1.ArcsRedirectPolicyReply)
	redirects, err := s.srv.ArcsRedirectPolicy(c, req.Aids)
	if err != nil {
		return nil, err
	}
	resp.RedirectPolicy = redirects
	return resp, nil
}

func (s *server) ArcRedirectPolicyAdd(c context.Context, req *v1.ArcRedirectPolicyAddRequest) (*v1.NoReply, error) {
	resp := new(v1.NoReply)
	err := s.srv.ArcRedirectPolicyAddSrv(c, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *server) UpPremiereArcs(c context.Context, req *v1.UpPremiereArcsRequest) (*v1.UpPremiereArcsReply, error) {
	resp := new(v1.UpPremiereArcsReply)
	upArcs, err := s.srv.UpPremiereArcsSrv(c, req)
	if err != nil {
		return nil, err
	}
	resp.UpArcs = upArcs
	return resp, nil
}

func (s *server) ArcsInner(c context.Context, req *v1.ArcsInnerRequest) (*v1.ArcsInnerReply, error) {
	resp := new(v1.ArcsInnerReply)
	as, err := s.srv.ArcsInner(c, req.Aids)
	if err != nil {
		return nil, err
	}
	resp.Items = as
	return resp, nil
}
