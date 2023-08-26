package like

import (
	"context"
	"encoding/json"
	"strconv"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/ecode"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
)

func (s *Service) BdfSchoolList(c context.Context) (data []*lmdl.List, err error) {
	var (
		subject  *lmdl.SubjectItem
		likeList []*lmdl.List
	)
	if subject, err = s.dao.ActSubject(c, s.c.Bdf.Sid); err != nil {
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if likeList, err = s.orderByCtime(c, s.c.Bdf.Sid, 1, s.c.Bdf.SchoolCount, subject, 0); err != nil {
		log.Error("BdfSchoolList s.orderByCtime sid(%d) ps(%d) error(%v)", s.c.Bdf.Sid, s.c.Bdf.SchoolCount, err)
		return
	}
	// add count score
	if err = s.contentAccount(c, likeList, s.c.Bdf.Sid); err != nil {
		log.Error("BdfSchoolList contentAccount error(%v)", err)
		return
	}
	data = likeList
	return
}

func (s *Service) BdfSchoolArcs(c context.Context, lids []int64) (list map[int64][]*lmdl.SimpleArc, err error) {
	var aids []int64
	for _, lid := range lids {
		pieceAids, ok := s.bdfData[lid]
		if !ok || len(pieceAids) == 0 {
			continue
		}
		aids = append(aids, pieceAids...)
	}
	if len(aids) == 0 {
		err = xecode.NothingFound
		return
	}
	archives, err := s.archives(c, aids)
	if err != nil {
		log.Error("BdfSchoolArcs s.archives aids(%v) %v", aids, err)
		err = nil
		return
	}
	list = make(map[int64][]*lmdl.SimpleArc, len(lids))
	for _, lid := range lids {
		pieceAids, ok := s.bdfData[lid]
		if !ok || len(pieceAids) == 0 {
			list[lid] = make([]*lmdl.SimpleArc, 0)
			continue
		}
		var tmp []*lmdl.SimpleArc
		for _, aid := range pieceAids {
			if arc, ok := archives[aid]; ok && arc != nil && arc.IsNormal() {
				tmp = append(tmp, lmdl.CopyFromArc(arc))
			}
		}
		list[lid] = tmp
	}
	return
}

func (s *Service) loadBdfAids() {
	data, err := s.dao.SourceItem(context.Background(), s.c.Bdf.DataSid)
	if err != nil {
		log.Error("loadBdfAids s.dao.SourceItem sid(%d) error(%v)", s.c.Bdf.DataSid, err)
		return
	}
	tmp := new(struct {
		List []*struct {
			Name string        `json:"name"`
			Data *lmdl.BdfData `json:"data"`
		} `json:"list"`
	})
	if err = json.Unmarshal(data, tmp); err != nil {
		log.Error("loadBdfAids s.dao.SourceItem(%d) error(%v)", s.c.Scholarship.ArcVid, err)
		return
	}
	if len(tmp.List) == 0 {
		log.Error("loadBdfAids data len 0")
		return
	}
	tmpData := make(map[int64][]int64, len(tmp.List))
	for _, v := range tmp.List {
		if v == nil || v.Data == nil {
			continue
		}
		i, err := strconv.ParseInt(v.Name, 10, 64)
		if err != nil || i <= 0 {
			continue
		}
		aids, err := xstr.SplitInts(v.Data.Aids)
		if err != nil {
			continue
		}
		tmpData[i] = aids
	}
	s.bdfData = tmpData
}
