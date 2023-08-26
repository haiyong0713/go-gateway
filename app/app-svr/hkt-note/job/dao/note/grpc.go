package note

import (
	archive "git.bilibili.co/bapis/bapis-go/archive/service"
	crm "git.bilibili.co/bapis/bapis-go/crm/service/profile-manager"
	upArc "git.bilibili.co/bapis/bapis-go/up-archive/service"
	note "go-gateway/app/app-svr/hkt-note/service/api"

	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
)

type grpc struct {
	crm     crm.ProfileManagerClient
	archive archive.ArchiveClient
	upArc   upArc.UpArchiveClient
	note    note.HktNoteClient
}

func NewGrpc() *grpc {
	var conf struct {
		Crm     *warden.ClientConfig
		UpArc   *warden.ClientConfig
		Archive *warden.ClientConfig
		Note    *warden.ClientConfig
	}
	if err := paladin.Get("grpc.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	crmClient, err := crm.NewClient(conf.Crm)
	if err != nil {
		panic(err)
	}
	arcClient, err := archive.NewClient(conf.Archive)
	if err != nil {
		panic(err)
	}
	upArcClient, err := upArc.NewClient(conf.UpArc)
	if err != nil {
		panic(err)
	}
	noteClient, err := note.NewClient(conf.Note)
	if err != nil {
		panic(err)
	}
	return &grpc{
		crm:     crmClient,
		archive: arcClient,
		upArc:   upArcClient,
		note:    noteClient,
	}
}
