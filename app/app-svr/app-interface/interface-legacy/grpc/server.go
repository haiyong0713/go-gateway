package grpc

import (
	"go-common/library/net/rpc/warden"

	"go-gateway/app/app-svr/app-interface/interface-legacy/http"
)

// New Interface warden rpc server
func New(c *warden.ServerConfig, svr *http.Server) *warden.Server {
	ws := warden.NewServer(c)
	if err := newSearch(ws, svr); err != nil {
		panic(err)
	}
	if err := newHistory(ws, svr); err != nil {
		panic(err)
	}
	if err := newSpace(ws, svr); err != nil {
		panic(err)
	}
	if err := newMedia(ws, svr); err != nil {
		panic(err)
	}
	if err := newTeenagers(ws, svr); err != nil {
		panic(err)
	}
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}
