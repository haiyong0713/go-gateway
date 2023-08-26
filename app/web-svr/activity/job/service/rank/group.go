package rank

import (
	"context"

	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
	sourcemdl "go-gateway/app/web-svr/activity/job/model/source"
)

// Group 分组
func (s *Service) Group(c context.Context, rankType int, archive []*sourcemdl.Archive) map[int64]*sourcemdl.ArchiveGroup {
	return s.group(c, rankType, archive)
}

// group 分组
func (s *Service) group(c context.Context, rankType int, archive []*sourcemdl.Archive) map[int64]*sourcemdl.ArchiveGroup {
	mapArchiveGroup := make(map[int64]*sourcemdl.ArchiveGroup)
	for _, v := range archive {
		var oid int64
		if rankType == rankmdl.RankTypeArchive {
			oid = v.Aid
		}
		if rankType == rankmdl.RankTypeUp {
			oid = v.Mid
		}

		if arc, ok := mapArchiveGroup[oid]; ok {
			arc.Archive = append(arc.Archive, v)
			arc.NewScore = arc.NewScore + v.Score
			continue
		}
		arc := &sourcemdl.ArchiveGroup{}
		arc.Archive = append(arc.Archive, v)
		arc.NewScore = v.Score
		arc.OID = oid
		arc.MID = v.Mid
		mapArchiveGroup[oid] = arc
	}
	return mapArchiveGroup
}
