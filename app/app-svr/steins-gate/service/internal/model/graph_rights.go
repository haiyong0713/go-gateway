package model

import "go-gateway/app/app-svr/steins-gate/service/api"

const (
	_graphVersionIntervention = 2
)

func BuildRestrict(graph *api.GraphInfo) bool {
	return graph.Version == _graphVersionIntervention

}
