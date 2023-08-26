package web

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	"go-common/library/xstr"
	"go-gateway/pkg/idsafe/bvid"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web-goblin/interface/dao/web"
	webmdl "go-gateway/app/web-svr/web-goblin/interface/model/web"
)

var (
	_emptyMiArc  = make([]*webmdl.Mi, 0)
	_emptyOutArc = make([]*webmdl.OutArchive, 0)
)

const (
	_tagBlkSize = 50
	_tagArcType = 3
	_miType     = 1
)

func (s *Service) OutArc(c context.Context, pn, ps int64) ([]*webmdl.OutArchive, int64) {
	count := int64(len(s.outArcs))
	start := (pn - 1) * ps
	end := start + ps - 1
	if count == 0 || count < start {
		return _emptyOutArc, count
	}
	var outArcs []*webmdl.OutArc
	if count > end {
		outArcs = s.outArcs[start : end+1]
	} else {
		outArcs = s.outArcs[start:]
	}
	var aids []int64
	for _, v := range outArcs {
		if v.Aid > 0 {
			aids = append(aids, v.Aid)
		}
	}
	if len(aids) > 0 {
		arcsReply, err := s.arcGRPC.Arcs(c, &api.ArcsRequest{Aids: aids})
		if err != nil {
			log.Error("OutArc s.arcGRPC.Arcs(%v) error(%v)", aids, err)
			return _emptyOutArc, count
		}
		var res []*webmdl.OutArchive
		for _, v := range outArcs {
			if v != nil && v.Aid > 0 {
				if arc, ok := arcsReply.GetArcs()[v.Aid]; ok && arc != nil && arc.IsNormal() {
					bvidStr, _ := bvid.AvToBv(v.Aid)
					res = append(res, &webmdl.OutArchive{
						ID:       bvidStr,
						Cover:    arc.Pic,
						Title:    arc.Title,
						H5URL:    fmt.Sprintf(s.c.Rule.H5PlayURL, bvidStr),
						Score:    v.SnapView,
						AppURL:   fmt.Sprintf("bilibili://video/%d", arc.Aid),
						Duration: arc.Duration,
					})
				}

			}
		}
		return res, count
	}
	return _emptyOutArc, count
}

// FullShort  xiao mi  FullShort .
func (s *Service) FullShort(c context.Context, pn, ps int64, source string) (res []*webmdl.Mi, err error) {
	var (
		aids []int64
		ip   = metadata.String(c, metadata.RemoteIP)
		m    = make(map[int64]string)
	)
	if aids, err = s.aids(pn, ps); err != nil {
		return
	}
	if res, err = s.archiveWithTag(c, aids, ip, m, source); err != nil {
		log.Error("s.archiveWithTag  error(%v)", err)
	}
	return
}

func (s *Service) archiveWithTag(c context.Context, aids []int64, ip string, op map[int64]string, source string) (list []*webmdl.Mi, err error) {
	var (
		tagErr   error
		viewsRep *api.ViewsReply
		tags     map[int64][]*webmdl.Tag
		mutex    = sync.Mutex{}
		tempTags []string
	)
	if viewsRep, err = s.arcGRPC.Views(c, &api.ViewsRequest{Aids: aids}); err != nil {
		log.Error("s.arcGRPC.Views aids(%s) error(%v)", xstr.JoinInts(aids), err)
		return
	}
	if len(viewsRep.Views) == 0 {
		return
	}
	aidsLen := len(aids)
	tags = make(map[int64][]*webmdl.Tag, aidsLen)
	group := new(errgroup.Group)
	for i := 0; i < aidsLen; i += _tagBlkSize {
		var partAids []int64
		if i+_tagBlkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_tagBlkSize]
		}
		group.Go(func() (err error) {
			var tmpRes map[int64][]*webmdl.Tag
			arg := &webmdl.ArgResTags{Oids: partAids, Type: _tagArcType, RealIP: ip}
			if tmpRes, tagErr = s.tag.ResTags(context.Background(), arg); tagErr != nil {
				web.PromError("ResTags接口错误", "s.tag.ResTag(%+v) error(%v)", arg, tagErr)
				return
			}
			mutex.Lock()
			for aid, tmpTags := range tmpRes {
				tags[aid] = tmpTags
			}
			mutex.Unlock()
			return nil
		})
	}
	if err = group.Wait(); err != nil {
		return
	}
	for _, aid := range aids {
		if view, ok := viewsRep.Views[aid]; ok && view != nil && view.Arc != nil && view.IsNormal() {
			miArc := new(webmdl.Mi)
			tempTags = []string{}
			miArc.FromArchive(view.Arc, view.Pages, op[aid], source)
			if tag, ok := tags[aid]; ok {
				for _, v := range tag {
					tempTags = append(tempTags, v.Name)
				}
			}
			if len(tempTags) == 0 {
				miArc.Tags = ""
			} else {
				miArc.Tags = strings.Join(tempTags, ",")
			}
			list = append(list, miArc)
		}
	}
	if len(list) == 0 {
		list = _emptyMiArc
	}
	return
}

func (s *Service) aids(pn, ps int64) (res []int64, err error) {
	var start, end int64
	if pn > 1 {
		start = pn*ps + 1
	} else {
		start = 1
	}
	end = start + ps
	if end > s.c.Rule.MaxAid {
		log.Warn("aids(%d,%d) maxAid(%d)", pn, ps, s.c.Rule.MaxAid)
		err = ecode.RequestErr
		return
	}
	for i := start; i < end; i++ {
		res = append(res, i)
	}
	return
}

func (s *Service) loadOutArcs() {
	var (
		id  int64
		tmp []*webmdl.OutArc
	)
	for {
		arcs, err := s.dao.OutArcs(context.Background(), _miType, id)
		if err != nil {
			log.Error("loadOutArcs s.dao.OutArcs error(%v)", err)
			return
		}
		if len(arcs) == 0 {
			break
		}
		for _, v := range arcs {
			if v == nil {
				continue
			}
			id = v.ID
			tmp = append(tmp, v)
		}
	}
	s.outArcs = tmp
}

func (s *Service) loadBaiduArcContent() {
	data, err := s.dao.ReadURLContent(context.Background(), s.c.Rule.BaiduPushArcURL)
	if err != nil {
		log.Error("loadBaiduArcContent s.dao.ReadURLContent url:%s error(%v)", s.c.Rule.BaiduPushArcURL, err)
		return
	}
	if len(data) <= len(s.baiduPushContent) {
		return
	}
	s.baiduPushContent = data
}

func (s *Service) BaiduPushArcContent(_ context.Context) []byte {
	return s.baiduPushContent
}
