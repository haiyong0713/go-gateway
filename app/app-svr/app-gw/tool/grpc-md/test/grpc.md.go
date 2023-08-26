package test

import (
	"regexp"
	"strings"
)

type App struct {
	Name     string
	Packages map[string]*Package
}

func (a *App) Copy() *App {
	ret := &App{
		Name:     a.Name,
		Packages: make(map[string]*Package),
	}
	for k, v := range a.Packages {
		ret.Packages[k] = v.Copy()
	}
	return ret
}

type Package struct {
	Name     string
	Services map[string]*Service
}

func (p *Package) Copy() *Package {
	ret := &Package{
		Name:     p.Name,
		Services: make(map[string]*Service),
	}
	for k, v := range p.Services {
		ret.Services[k] = v.Copy()
	}
	return ret
}

type Service struct {
	Name    string
	Methods []string
}

func (s *Service) Copy() *Service {
	ret := &Service{
		Name: s.Name,
	}
	//nolint:gosimple
	for _, v := range s.Methods {
		ret.Methods = append(ret.Methods, v)
	}
	return ret
}

func GetByAppID(appid string) (*App, bool) {
	app, ok := grpcMd[appid]
	if !ok {
		return nil, false
	}
	return app.Copy(), true
}

func GetByAppIDs(appids []string) (map[string]*App, bool) {
	ret := map[string]*App{}
	for _, appid := range appids {
		app, ok := grpcMd[appid]
		if !ok {
			continue
		}
		ret[appid] = app.Copy()
	}
	if len(ret) == 0 {
		return nil, false
	}
	return ret, true
}

func PrefixMatch(prefix string) (map[string]*App, bool) {
	ret := map[string]*App{}
	for key := range grpcMd {
		if strings.HasPrefix(key, prefix) {
			ret[key] = grpcMd[key].Copy()
		}
	}
	if len(ret) == 0 {
		return nil, false
	}
	return ret, true
}

func FuzzyMatch(candidate string) (map[string]*App, bool, error) {
	ret := map[string]*App{}
	r, err := regexp.Compile(candidate)
	if err != nil {
		return nil, false, err
	}
	for key := range grpcMd {
		if r.Match([]byte(key)) {
			ret[key] = grpcMd[key].Copy()
		}
	}
	if len(ret) == 0 {
		return nil, false, nil
	}
	return ret, true, nil
}

func GetAllAppID() []string {
	ret := []string{}
	for appid := range grpcMd {
		ret = append(ret, appid)
	}
	return ret
}

var (
	grpcMd = map[string]*App{
		//nolint:gofmt
		"main.dynamic.feed": {Name: "main.dynamic.feed",
			Packages: map[string]*Package{
				"dynamic.service.feed.v1": {Name: "dynamic.service.feed.v1",
					Services: map[string]*Service{
						"Feed": {Name: "Feed",
							Methods: []string{
								"/dynamic.service.feed.v1.Feed/UpdateNum",
							},
						},
					},
				},
			},
		},
	}
)
