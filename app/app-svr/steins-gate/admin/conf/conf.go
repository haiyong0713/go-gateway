package conf

import (
	"flag"

	"go-common/library/conf"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

var (
	confPath string
	Conf     = &Config{}
	client   *conf.Client
)

type Config struct {
	MySQL       *MySQL
	XLog        *log.Config
	Server      *bm.ServerConfig
	HTTPClient  *bm.ClientConfig
	VideoClient *bm.ClientConfig
	Host        *Host
	Bvc         *Bvc
	Archive     *warden.ClientConfig
}

type Host struct {
	Videoup string
	Bvc     string
}

// MySQL def.
type MySQL struct {
	Steinsgate *sql.Config
}

type Bvc struct {
	Key string
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init init config.
func Init() (err error) {
	if confPath != "" {
		_, err = toml.DecodeFile(confPath, &Conf)
		return
	}
	err = remote()
	return
}

func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	client.Watch("steins-gate-admin.toml")
	//nolint:biligowordcheck
	go func() {
		for range client.Event() {
			log.Info("config reload")
			if load() != nil {
				log.Error("config reload error (%v)", err)
			}
		}
	}()
	return
}

func load() (err error) {
	var (
		s       string
		ok      bool
		tmpConf *Config
	)
	if s, ok = client.Toml2(); !ok {
		return errors.New("load config center error")
	}
	if _, err = toml.Decode(s, &tmpConf); err != nil {
		return errors.New("could not decode config")
	}
	*Conf = *tmpConf
	return

}
