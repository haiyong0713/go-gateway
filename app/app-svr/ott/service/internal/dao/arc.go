package dao

import (
	"context"
	"net/url"
	"strconv"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"

	"github.com/pkg/errors"
)

const (
	_simpleArchivePath = "/videoup/simplearchive"
	_iosSmall          = 16
	_androidSmall      = 17
)

func (d *Dao) Arcs(c context.Context, aids []int64) (arcs map[int64]*arcmdl.Arc, err error) {
	var (
		req      = &arcmdl.ArcsRequest{Aids: aids}
		arcReply *arcmdl.ArcsReply
	)
	if arcReply, err = d.arcClient.Arcs(c, req); err != nil {
		err = errors.Wrapf(err, "d.Arc req(%+v)", req)
		return
	}
	if arcReply == nil || arcReply.Arcs == nil {
		err = errors.Wrapf(ecode.NothingFound, "d.Arc req(%+v)", req)
		return
	}
	arcs = arcReply.Arcs
	return
}

func (d *Dao) ArcType(c context.Context) (typeReply *arcmdl.TypesReply, err error) {
	req := &arcmdl.NoArgRequest{}
	if typeReply, err = d.arcClient.Types(c, req); err != nil {
		err = errors.Wrapf(err, "d.ArcView s.arcClient.View req(%+v)", req)
		return nil, err
	}
	return typeReply, nil
}

func (d *Dao) SimpleArchives(c context.Context, aids []int64) (res map[int64]bool, err error) {
	mutex := sync.Mutex{}
	res = make(map[int64]bool)
	eg := errgroup.WithContext(c)
	for _, aid := range aids {
		id := aid
		eg.Go(func(c context.Context) (err error) {
			isSmall, _ := d.simpleArchive(c, id)
			mutex.Lock()
			res[id] = isSmall
			mutex.Unlock()
			return
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait() error(%+v)", err)
	}
	return
}

func (d *Dao) simpleArchive(c context.Context, aid int64) (isSmall bool, err error) {
	params := url.Values{}
	params.Set("aid", strconv.FormatInt(aid, 10))
	var res struct {
		Code int          `json:"code"`
		Data *VideoUpView `json:"data"`
	}
	if err = d.client.Get(c, d.c.Cfg.ArchiveHost+_simpleArchivePath, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.c.Cfg.ArchiveHost+_simpleArchivePath+"?"+params.Encode())
		return
	}
	return res.Data.UpFrom == _iosSmall || res.Data.UpFrom == _androidSmall, nil
}

type VideoUpView struct {
	Aid    int64 `json:"Aid"`
	UpFrom int64 `json:"up_from"`
}
