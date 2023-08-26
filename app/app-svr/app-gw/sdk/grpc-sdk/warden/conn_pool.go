package warden

import (
	"context"
	"sync"

	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"
)

var _connPool = newPool()

type pool struct {
	sync.RWMutex
	pool map[string]*grpc.ClientConn
}

func newPool() *pool {
	return &pool{
		pool: map[string]*grpc.ClientConn{},
	}
}

func (p *pool) conn(target string, config *warden.ClientConfig, opt ...grpc.DialOption) (*grpc.ClientConn, error) {
	p.RLock()
	cc, ok := p.pool[target]
	p.RUnlock()
	if ok {
		return cc, nil
	}
	client := warden.NewClient(config, opt...)
	newCC, err := client.Dial(context.Background(), target)
	if err != nil {
		return nil, err
	}
	p.Lock()
	defer p.Unlock()
	if cc, ok := p.pool[target]; ok {
		newCC.Close()
		return cc, nil
	}
	p.pool[target] = newCC
	return newCC, nil
}

// Dial external service connection on-demand or return the origin connection.
func pooledClientConn(appid string, config *warden.ClientConfig, opt ...grpc.DialOption) (*grpc.ClientConn, error) {
	return _connPool.conn(appid, config, opt...)
}
