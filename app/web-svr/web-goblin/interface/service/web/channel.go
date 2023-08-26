package web

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web-goblin/interface/model/web"
)

const (
	_chRqCnt     = 40
	_chDisplayID = 1
	_chTypeArc   = 3
	_chFrom      = 1
)

var _emptyArcs = make([]*api.Arc, 0)

// Channel .
func (s *Service) Channel(c context.Context, id, mid int64, buvid string) (channel *web.Channel, err error) {
	var (
		aids   []int64
		arcs   *api.ArcsReply
		tagErr error
	)
	ip := metadata.String(c, metadata.RemoteIP)
	channel = new(web.Channel)
	if cards, ok := s.channelCards[id]; ok {
		for _, card := range cards {
			aids = append(aids, card.Value)
		}
	}
	group, errCtx := errgroup.WithContext(c)
	group.Go(func() error {
		arg := &web.ArgChannelResource{
			Tid:        id,
			Mid:        mid,
			RequestCNT: int32(_chRqCnt),
			DisplayID:  _chDisplayID,
			Type:       _chTypeArc,
			Buvid:      buvid,
			From:       _chFrom,
			RealIP:     ip,
		}
		if channelResource, chErr := s.tag.ChannelResources(errCtx, arg); chErr != nil {
			log.Error("Channel s.tag.Resources error(%v)", chErr)
		} else if channelResource != nil {
			aids = append(aids, channelResource.Oids...)
		}
		return nil
	})
	group.Go(func() error {
		if channel.Tag, tagErr = s.tag.InfoByID(errCtx, &web.ArgID{ID: id, Mid: mid}); tagErr != nil {
			log.Error("Channel s.tag.InfoByID(%d, %d) error(%v)", id, mid, err)
			return tagErr
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		return
	}
	if len(aids) == 0 {
		channel.Archives = _emptyArcs
		return
	}
	if arcs, err = s.arcGRPC.Arcs(c, &api.ArcsRequest{Aids: aids}); err != nil {
		log.Error("Channel s.arc.Archives3(%v) error(%v)", aids, err)
		err = nil
		channel.Archives = _emptyArcs
		return
	}
	if arcs == nil || len(arcs.Arcs) == 0 {
		return
	}
	for _, aid := range aids {
		if arc, ok := arcs.Arcs[aid]; ok && arc.IsNormal() {
			channel.Archives = append(channel.Archives, arc)
		}
	}
	if len(channel.Archives) > _chRqCnt {
		channel.Archives = channel.Archives[:_chRqCnt]
	}
	return
}
