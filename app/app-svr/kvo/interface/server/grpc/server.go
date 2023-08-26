package grpc

import (
	"context"

	pb "go-gateway/app/app-svr/kvo/interface/api"
	"go-gateway/app/app-svr/kvo/interface/model/module"
	"go-gateway/app/app-svr/kvo/interface/service"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
)

// New new a grpc server.
func New(cfg *warden.ServerConfig, svc *service.Service) *warden.Server {
	ws := warden.NewServer(cfg)
	pb.RegisterKvoServer(ws.Server(), &server{svc: svc})
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}

type server struct {
	svc *service.Service
}

func (s *server) AddDoc(c context.Context, req *pb.AddDocReq) (res *pb.AddDocReply, err error) {
	body := pb.ReqToConfigModify(req)
	return new(pb.AddDocReply), s.svc.AddUserDoc(c, req.GetMid(), body, req.GetPlatform(), req.GetBuvid(), req.GetModule())
}

func (s *server) GetDoc(c context.Context, req *pb.GetDocReq) (res *pb.GetDocReply, err error) {
	var data *module.Setting
	res = new(pb.GetDocReply)
	if req.GetMid() == 0 && req.GetBuvid() == "" {
		err = ecode.RequestErr
		return
	}
	if req.Mid > 0 {
		data, err = s.svc.DocumentMid(c, req.GetMid(), req.GetModule(), 0, 0, req.GetPlatform())
	} else {
		data, err = s.svc.DocumentBuvid(c, req.Buvid, req.GetModule(), req.GetPlatform())
	}
	if err != nil {
		if ecode.Cause(err) != ecode.NotModified {
			log.Error("grpc kvoSvr.Document(req:%+v) error(%v)", req, err)
		}
		return
	}
	res.Data = string(data.Data)
	return
}
