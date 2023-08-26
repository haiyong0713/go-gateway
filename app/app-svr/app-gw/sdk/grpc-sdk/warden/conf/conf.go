package conf

import (
	"path"

	"go-common/library/conf"
	wardensdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

var _defName = "grpc-client-sdk.toml"

func defaultPath(filePath string) string {
	dir := path.Dir(filePath)
	return path.Join(dir, _defName)
}

func New(filePath string, client *conf.Client) *wardensdk.InterceptorBuilder {
	if filePath == "" {
		if client == nil {
			panic(errors.New("client should not be nil if `filePath` is not present"))
		}
		builder, err := fromRemote("grpc-client-sdk.toml", client)
		if err != nil {
			panic(err)
		}
		return builder
	}
	builder, err := fromLocal(defaultPath(filePath))
	if err != nil {
		panic(err)
	}
	return builder
}

func OnEventReload(builder *wardensdk.InterceptorBuilder, client *conf.Client) error {
	content, ok := client.Toml2()
	if !ok {
		return errors.New("failed to load content from config center")
	}
	builderConfig, err := loadContent(content)
	if err != nil {
		return err
	}
	if err := builder.Reload(*builderConfig); err != nil {
		return err
	}
	return nil
}

func fromLocal(filePath string) (*wardensdk.InterceptorBuilder, error) {
	cfg := struct {
		SDKBuilderConfig *wardensdk.SDKBuilderConfig
	}{}
	if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
		return nil, err
	}
	if cfg.SDKBuilderConfig == nil {
		return nil, errors.Errorf("invalid sdk builder config from file: %q: maybe empty config", filePath)
	}
	return wardensdk.NewBuilder(*cfg.SDKBuilderConfig), nil
}

func fromContent(content string) (*wardensdk.InterceptorBuilder, error) {
	builderConfig, err := loadContent(content)
	if err != nil {
		return nil, err
	}
	return wardensdk.NewBuilder(*builderConfig), nil
}

func loadContent(content string) (*wardensdk.SDKBuilderConfig, error) {
	cfg := struct {
		SDKBuilderConfig *wardensdk.SDKBuilderConfig
	}{}
	if _, err := toml.Decode(content, &cfg); err != nil {
		return nil, errors.Wrapf(err, "invalid sdk builder config: %s", content)
	}
	if cfg.SDKBuilderConfig == nil {
		return nil, errors.Errorf("invalid sdk builder config: %s: maybe empty config", content)
	}
	return cfg.SDKBuilderConfig, nil
}

func fromRemote(filename string, client *conf.Client) (*wardensdk.InterceptorBuilder, error) {
	content, ok := client.Toml2()
	if !ok {
		return nil, errors.New("failed to load content from config center")
	}

	builder, err := fromContent(content)
	if err != nil {
		return nil, err
	}
	client.Watch(filename)
	return builder, nil
}
