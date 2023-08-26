package model

import (
	"fmt"
	"strconv"

	"go-common/library/time"
	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	RecordInvalid       = -1
	GraphStatePass      = 1
	GraphStateSubmit    = -30
	GraphStateRepulse   = -20
	GraphStateReSubmit  = -10
	GraphIsPreview      = 1
	RootEdge            = 1
	RootCursor          = 0
	IllegalCursor       = -1
	MaxChoiceLen        = 200
	SecondBiggestChoice = 199
	EdgeGraph           = 1
	InterventionGraph   = 2
)

// GraphStateAudits .
var GraphStateAudits = map[int]struct{}{
	GraphStateSubmit:   {},
	GraphStateRepulse:  {},
	GraphStateReSubmit: {},
}

// GraphDB is graph struct in DB
type GraphDB struct {
	api.GraphInfo
	State  int
	Script string
	Ctime  time.Time
}

// IsPass def.
func (v *GraphDB) IsPass() bool {
	return v.State == GraphStatePass
}

func (v *GraphDB) IsEdgeGraph() bool {
	return v.Version == EdgeGraph
}

func IsEdgeGraph(in *api.GraphInfo) bool {
	return in.Version == EdgeGraph
}

func (v *GraphDB) IsInterventionGraph() bool {
	return v.Version == InterventionGraph
}

func IsInterventionGraph(in *api.GraphInfo) bool {
	return in.Version == InterventionGraph
}

// ToGraphInfo builds a graphInfo structure from GraphDB
func (v *GraphDB) ToGraphInfo(fiNid, fiCid int64) *api.GraphInfo {
	return &api.GraphInfo{
		Id:                         v.Id,
		Aid:                        v.Aid,
		GlobalVars:                 v.GlobalVars,
		RegionalVars:               v.RegionalVars,
		FirstNid:                   fiNid,
		FirstCid:                   fiCid,
		Version:                    v.Version,
		SkinId:                     v.SkinId,
		NoTutorial:                 v.NoTutorial,
		NoBacktracking:             v.NoBacktracking,
		NoEvaluation:               v.NoEvaluation,
		GuestOverwriteRegionalVars: v.GuestOverwriteRegionalVars,
	}
}

// SteinsCid gives the real first cid of the steins-gate video
type SteinsCid struct {
	Aid   int64  `json:"aid"`
	Cid   int64  `json:"cid"`
	Route string `json:"route"`
}

// Key def.
func (v *SteinsCid) Key() string {
	return fmt.Sprintf("%d_%d", v.Aid, v.Cid)
}

func GetAvID(input string) (aid int64, err error) {
	if aid, err = strconv.ParseInt(input, 10, 64); err != nil {
		err = nil
		if aid, err = bvid.BvToAv(input); err != nil {
			return 0, ecode.BvidIllegal
		}
	}
	return

}
