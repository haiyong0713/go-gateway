package middle

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/conf"
)

var (
	ctx = context.Background()
)

func WithService(f func(s *Middle)) {
	dir, _ := filepath.Abs("../cmd/space-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	log.Init(conf.Conf.Log)
	f(New(conf.Conf))
}

func TestMiddle_AccIsAllowd(t *testing.T) {
	WithService(func(m *Middle) {
		fmt.Println(m.AccIsAllowed(ctx, 111004350))
	})
}
